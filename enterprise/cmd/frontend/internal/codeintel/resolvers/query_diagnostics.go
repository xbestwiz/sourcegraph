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

const slowDiagnosticsRequestThreshold = time.Second

// Diagnostics returns the diagnostics for documents with the given path prefix. If there are
// multiple bundles associated with this resolver, results from all bundles will be concatenated
// and returned.
func (r *queryResolver) Diagnostics(ctx context.Context, limit int) (_ []AdjustedDiagnostic, _ int, err error) {
	ctx, endObservation := observeResolver(ctx, &err, "Diagnostics", r.operations.diagnostics, slowDiagnosticsRequestThreshold, observation.Args{
		LogFields: []log.Field{
			log.Int("repositoryID", r.repositoryID),
			log.String("commit", r.commit),
			log.String("path", r.path),
			log.String("uploadIDs", strings.Join(r.uploadIDs(), ", ")),
			log.Int("limit", limit),
		},
	})
	defer endObservation()

	type TEMPORARY struct {
		Upload       store.Dump
		AdjustedPath string
		Diagnostics  []lsifstore.Diagnostic
		Count        int
	}
	var worklist []TEMPORARY

	for _, upload := range r.uploads {
		adjustedPath, ok, err := r.positionAdjuster.AdjustPath(ctx, upload.Commit, r.path, false)
		if err != nil {
			return nil, 0, err
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
		// TODO - batch these requests
		diagnostics, count, err := r.lsifStore.Diagnostics(ctx, w.Upload.ID, strings.TrimPrefix(w.AdjustedPath, w.Upload.Root), 0, limit)
		if err != nil {
			return nil, 0, err
		}

		worklist[i].Diagnostics = diagnostics
		worklist[i].Count = count
	}

	totalCount := 0
	var adjustedDiagnostics []AdjustedDiagnostic
	for _, w := range worklist {
		for _, diagnostic := range w.Diagnostics {
			diagnostic = lsifstore.Diagnostic{
				DumpID:         diagnostic.DumpID,
				Path:           w.Upload.Root + diagnostic.Path,
				DiagnosticData: diagnostic.DiagnosticData,
			}

			clientRange := lsifstore.Range{
				Start: lsifstore.Position{Line: diagnostic.StartLine, Character: diagnostic.StartCharacter},
				End:   lsifstore.Position{Line: diagnostic.EndLine, Character: diagnostic.EndCharacter},
			}

			adjustedCommit, adjustedRange, err := r.adjustRange(ctx, w.Upload.RepositoryID, w.Upload.Commit, diagnostic.Path, clientRange)
			if err != nil {
				return nil, 0, err
			}

			adjustedDiagnostics = append(adjustedDiagnostics, AdjustedDiagnostic{
				Diagnostic:     diagnostic,
				Dump:           w.Upload,
				AdjustedCommit: adjustedCommit,
				AdjustedRange:  adjustedRange,
			})
		}

		totalCount += w.Count
	}

	if len(adjustedDiagnostics) > limit {
		adjustedDiagnostics = adjustedDiagnostics[:limit]
	}

	return adjustedDiagnostics, totalCount, nil
}
