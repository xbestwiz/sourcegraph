package resolvers

import (
	"context"
	"strings"
	"time"

	"github.com/opentracing/opentracing-go/log"

	store "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/dbstore"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/lsifstore"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

const slowHoverRequestThreshold = time.Second

// Hover returns the hover text and range for the symbol at the given position. If there are
// multiple bundles associated with this resolver, the hover text and range from the first
// bundle with any results will be returned.
func (r *queryResolver) Hover(ctx context.Context, line, character int) (_ string, _ lsifstore.Range, _ bool, err error) {
	ctx, endObservation := observeResolver(ctx, &err, "Hover", r.operations.hover, slowHoverRequestThreshold, observation.Args{
		LogFields: []log.Field{
			log.Int("repositoryID", r.repositoryID),
			log.String("commit", r.commit),
			log.String("path", r.path),
			log.String("uploadIDs", strings.Join(r.uploadIDs(), ", ")),
			log.Int("line", line),
			log.Int("character", character),
		},
	})
	defer endObservation()

	position := lsifstore.Position{
		Line:      line,
		Character: character,
	}

	type TEMPORARY struct {
		Upload           store.Dump
		AdjustedPath     string
		AdjustedPosition lsifstore.Position
		Text             string
		Range            lsifstore.Range
	}
	var worklist []TEMPORARY

	for _, upload := range r.uploads {
		adjustedPath, adjustedPosition, ok, err := r.positionAdjuster.AdjustPosition(ctx, upload.Commit, r.path, position, false)
		if err != nil {
			return "", lsifstore.Range{}, false, err
		}
		if !ok {
			continue
		}

		worklist = append(worklist, TEMPORARY{
			Upload:           upload,
			AdjustedPath:     adjustedPath,
			AdjustedPosition: adjustedPosition,
		})
	}

	for i, w := range worklist {
		// TODO - batch these requests
		text, r, exists, err := r.lsifStore.Hover(ctx, w.Upload.ID, strings.TrimPrefix(w.AdjustedPath, w.Upload.Root), w.AdjustedPosition.Line, w.AdjustedPosition.Character)
		if err != nil {
			return "", lsifstore.Range{}, false, err
		}
		if !exists || text == "" {
			continue
		}

		worklist[i].Text = text
		worklist[i].Range = r
	}

	for _, w := range worklist {
		if w.Text != "" {
			_, adjustedRange, ok, err := r.positionAdjuster.AdjustRange(ctx, w.Upload.Commit, r.path, w.Range, true)
			if err != nil || !ok {
				return "", lsifstore.Range{}, false, err
			}

			return w.Text, adjustedRange, true, nil
		}
	}

	return "", lsifstore.Range{}, false, nil
}
