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

const slowDefinitionsRequestThreshold = time.Second

// Definitions returns the list of source locations that define the symbol at the given position.
// This may include remote definitions if the remote repository is also indexed. If there are multiple
// bundles associated with this resolver, the definitions from the first bundle with any results will
// be returned.
func (r *queryResolver) Definitions(ctx context.Context, line, character int) (_ []AdjustedLocation, err error) {
	ctx, endObservation := observeResolver(ctx, &err, "Definitions", r.operations.definitions, slowDefinitionsRequestThreshold, observation.Args{
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

	type TEMPORARY2 struct {
		Upload    store.Dump
		Locations []lsifstore.Location
	}
	type TEMPORARY struct {
		Upload           store.Dump
		AdjustedPath     string
		AdjustedPosition lsifstore.Position
		OrderedMonikers  []lsifstore.MonikerData
		Locations        []TEMPORARY2
	}
	var worklist []TEMPORARY

	for _, upload := range r.uploads {
		adjustedPath, adjustedPosition, ok, err := r.positionAdjuster.AdjustPosition(ctx, upload.Commit, r.path, position, false)
		if err != nil {
			return nil, err
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
		// TODO - batch these requests together
		locations, err := r.lsifStore.Definitions(ctx, w.Upload.ID, strings.TrimPrefix(w.AdjustedPath, w.Upload.Root), w.AdjustedPosition.Line, w.AdjustedPosition.Character)
		if err != nil {
			return nil, err
		}

		if len(locations) > 0 {
			worklist[i].Locations = append(worklist[i].Locations, TEMPORARY2{
				Upload:    w.Upload,
				Locations: locations,
			})
		}
	}

	// TODO - can early exit here

	for i, w := range worklist {
		if len(w.Locations) > 0 {
			continue
		}

		// TODO - batch these requests together
		rangeMonikers, err := r.lsifStore.MonikersByPosition(
			ctx,
			w.Upload.ID,
			strings.TrimPrefix(w.AdjustedPath, w.Upload.Root),
			w.AdjustedPosition.Line,
			w.AdjustedPosition.Character,
		)
		if err != nil {
			return nil, err
		}

		var orderedMonikers []lsifstore.MonikerData
		for _, monikers := range rangeMonikers {
			for _, moniker := range monikers {
				if moniker.Kind == "import" && moniker.PackageInformationID != "" {
					orderedMonikers = append(orderedMonikers, moniker)
				}
			}
		}

		// TODO - ensure uniqueness
		worklist[i].OrderedMonikers = orderedMonikers
	}

	for i, w := range worklist {
		for _, moniker := range w.OrderedMonikers {
			// TODO - batch these requests together
			pid, _, err := r.lsifStore.PackageInformation(ctx, w.Upload.ID, strings.TrimPrefix(w.AdjustedPath, w.Upload.Root), string(moniker.PackageInformationID))
			if err != nil {
				return nil, err
			}

			definitionUpload, exists, err := r.dbStore.GetPackage(ctx, moniker.Scheme, pid.Name, pid.Version)
			if err != nil {
				return nil, err
			}
			if !exists {
				continue
			}

			const defintionMonikersLimit = 100
			locations, _, err := r.lsifStore.MonikerResults(ctx, definitionUpload.ID, "definitions", moniker.Scheme, moniker.Identifier, 0, defintionMonikersLimit)
			if err != nil {
				return nil, err
			}

			if len(locations) > 0 {
				worklist[i].Locations = append(worklist[i].Locations, TEMPORARY2{
					Upload:    definitionUpload,
					Locations: locations,
				})
				break
			}
		}
	}

	for _, w := range worklist {
		if len(w.Locations) == 0 {
			continue
		}

		var adjustedLocations []AdjustedLocation
		for _, pair := range w.Locations {
			locations, err := r.adjustLocations(ctx, pair.Upload, pair.Locations)
			if err != nil {
				return nil, err
			}

			adjustedLocations = append(adjustedLocations, locations...)
		}

		return adjustedLocations, nil
	}

	return nil, nil
}
