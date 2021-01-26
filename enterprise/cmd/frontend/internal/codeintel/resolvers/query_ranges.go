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

const slowRangesRequestThreshold = time.Second

// Ranges returns code intelligence for the ranges that fall within the given range of lines. These
// results are partial and do not include references outside the current file, or any location that
// requires cross-linking of bundles (cross-repo or cross-root).
func (r *queryResolver) Ranges(ctx context.Context, startLine, endLine int) (_ []AdjustedCodeIntelligenceRange, err error) {
	ctx, endObservation := observeResolver(ctx, &err, "Ranges", r.operations.ranges, slowRangesRequestThreshold, observation.Args{
		LogFields: []log.Field{
			log.Int("repositoryID", r.repositoryID),
			log.String("commit", r.commit),
			log.String("path", r.path),
			log.String("uploadIDs", strings.Join(r.uploadIDs(), ", ")),
			log.Int("startLine", startLine),
			log.Int("endLine", endLine),
		},
	})
	defer endObservation()

	type TEMPORARY struct {
		Upload       store.Dump
		AdjustedPath string
		Ranges       []lsifstore.CodeIntelligenceRange
	}
	worklist := make([]TEMPORARY, 0, len(r.uploads))

	for _, upload := range r.uploads {
		// TODO - adjust line offsets as well as path
		adjustedPath, ok, err := r.positionAdjuster.AdjustPath(ctx, upload.Commit, r.path, false)
		if err != nil {
			return nil, err
		}
		if !ok {
			continue
		}

		worklist = append(worklist, TEMPORARY{
			Upload:       upload,
			AdjustedPath: adjustedPath,
		})
	}

	for i, w := range worklist {
		// TODO - batch these requests together
		ranges, err := r.lsifStore.Ranges(ctx, w.Upload.ID, strings.TrimPrefix(w.AdjustedPath, w.Upload.Root), startLine, endLine)
		if err != nil {
			return nil, err
		}

		worklist[i].Ranges = ranges
	}

	var ranges []AdjustedCodeIntelligenceRange
	for _, w := range worklist {
		for _, rn := range w.Ranges {
			_, adjustedRange, err := r.adjustRange(ctx, w.Upload.RepositoryID, w.Upload.Commit, w.AdjustedPath, rn.Range)
			if err != nil {
				return nil, err
			}

			adjustedDefinitions, err := r.adjustLocations(ctx, w.Upload, rn.Definitions)
			if err != nil {
				return nil, err
			}

			adjustedReferences, err := r.adjustLocations(ctx, w.Upload, rn.References)
			if err != nil {
				return nil, err
			}

			ranges = append(ranges, AdjustedCodeIntelligenceRange{
				Range:       adjustedRange,
				Definitions: adjustedDefinitions,
				References:  adjustedReferences,
				HoverText:   rn.HoverText,
			})
		}
	}

	return ranges, nil
}
