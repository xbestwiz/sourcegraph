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

const slowReferencesRequestThreshold = time.Second

// References returns the list of source locations that reference the symbol at the given position.
// This may include references from other dumps and repositories. If there are multiple bundles
// associated with this resolver, results from all bundles will be concatenated and returned.
func (r *queryResolver) References(ctx context.Context, line, character, limit int, rawCursor string) (_ []AdjustedLocation, _ string, err error) {
	ctx, endObservation := observeResolver(ctx, &err, "References", r.operations.references, slowReferencesRequestThreshold, observation.Args{
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
	type QualifiedMoniker struct {
		lsifstore.MonikerData
		lsifstore.PackageInformationData
	}
	type TEMPORARY struct {
		Upload           store.Dump
		AdjustedPath     string
		AdjustedPosition lsifstore.Position
		OrderedMonikers  []QualifiedMoniker
		Locations        []TEMPORARY2
	}
	var worklist []TEMPORARY

	for _, upload := range r.uploads {
		adjustedPath, adjustedPosition, ok, err := r.positionAdjuster.AdjustPosition(ctx, upload.Commit, r.path, position, false)
		if err != nil {
			return nil, "", err
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

	//
	// PHASE 1 (same dump)
	//

	for i, w := range worklist {
		// TODO - batch these requests together
		locations, err := r.lsifStore.References(ctx, w.Upload.ID, strings.TrimPrefix(w.AdjustedPath, w.Upload.Root), w.AdjustedPosition.Line, w.AdjustedPosition.Character)
		if err != nil {
			return nil, "", err
		}

		if len(locations) > 0 {
			worklist[i].Locations = append(worklist[i].Locations, TEMPORARY2{
				Upload:    w.Upload,
				Locations: locations,
			})
		}
	}

	//
	// PHASE 2 (gather monikers)
	// TODO - take this from definitions
	//

	for i, w := range worklist {
		// TODO - batch these requests together
		rangeMonikers, err := r.lsifStore.MonikersByPosition(
			ctx,
			w.Upload.ID,
			strings.TrimPrefix(w.AdjustedPath, w.Upload.Root),
			w.AdjustedPosition.Line,
			w.AdjustedPosition.Character,
		)
		if err != nil {
			return nil, "", err
		}

		var orderedMonikers []QualifiedMoniker
		for _, monikers := range rangeMonikers {
			for _, moniker := range monikers {
				if moniker.PackageInformationID != "" {
					// TODO - batch these requests together
					packageInformationData, _, err := r.lsifStore.PackageInformation(ctx, w.Upload.ID, strings.TrimPrefix(w.AdjustedPath, w.Upload.Root), string(moniker.PackageInformationID))
					if err != nil {
						return nil, "", err
					}

					orderedMonikers = append(orderedMonikers, QualifiedMoniker{
						MonikerData:            moniker,
						PackageInformationData: packageInformationData,
					})
				}
			}
		}

		// TODO - ensure uniqueness
		worklist[i].OrderedMonikers = orderedMonikers
	}

	//
	// PHASE 3 (definition dump)
	//

	for i, w := range worklist {
		for _, moniker := range w.OrderedMonikers {
			// TODO - wtf
			// if moniker.Kind != "export" {
			// 	continue
			// }

			// TODO - batch these requests together
			definitionUpload, exists, err := r.dbStore.GetPackage(ctx, moniker.Scheme, moniker.Name, moniker.Version)
			if err != nil {
				return nil, "", err
			}
			if !exists {
				continue
			}
			if definitionUpload.ID == w.Upload.ID {
				continue
			}

			locations, _, err := r.lsifStore.MonikerResults(ctx, definitionUpload.ID, "references", moniker.Scheme, moniker.Identifier, 0, 10000000)
			if err != nil {
				return nil, "", err
			}

			// TODO - ensure deduplicated
			worklist[i].Locations = append(worklist[i].Locations, TEMPORARY2{
				Upload:    definitionUpload,
				Locations: locations,
			})
		}
	}

	//
	// PHASE 4 (same repo)
	//

	for i, w := range worklist {
		for _, moniker := range w.OrderedMonikers {
			// TODO - wtf
			// if moniker.Kind != "import" {
			// 	continue
			// }

			// TODO - batch these requests together
			_, pager, err := r.dbStore.SameRepoPager(ctx, r.repositoryID, r.commit, moniker.Scheme, moniker.Name, moniker.Version, 10000000)
			if err != nil {
				return nil, "", err
			}
			defer func() {
				err = pager.Done(err) // TODO
			}()

			// TODO - loop
			// TODO - check bloom filter
			references, err := pager.PageFromOffset(ctx, 0)
			if err != nil {
				return nil, "", err
			}

			// TODO - remove duplicate uploads
			for _, reference := range references {
				upload, exists, err := r.dbStore.GetDumpByID(ctx, reference.DumpID)
				if err != nil {
					return nil, "", err
				}
				if !exists {
					continue
				}

				// TODO - check for commit existence
				locations, _, err := r.lsifStore.MonikerResults(ctx, reference.DumpID, "references", reference.Scheme, moniker.Identifier, 0, 10000000)
				if err != nil {
					return nil, "", err
				}

				// TODO - ensure deduplicated
				worklist[i].Locations = append(worklist[i].Locations, TEMPORARY2{
					Upload:    upload,
					Locations: locations,
				})
			}
		}
	}

	//
	// PHASE 5 (remote repo)
	//

	for i, w := range worklist {
		for _, moniker := range w.OrderedMonikers {
			// TODO - wtf
			// if moniker.Kind != "import" {
			// 	continue
			// }

			// TODO - batch these requests together
			_, pager, err := r.dbStore.PackageReferencePager(ctx, moniker.Scheme, moniker.Name, moniker.Version, r.repositoryID, 10000000)
			if err != nil {
				return nil, "", err
			}
			defer func() {
				err = pager.Done(err) // TODO
			}()

			// TODO - loop
			// TODO - check bloom filter
			references, err := pager.PageFromOffset(ctx, 0)
			if err != nil {
				return nil, "", err
			}

			// TODO - remove duplicate uploads
			for _, reference := range references {
				upload, exists, err := r.dbStore.GetDumpByID(ctx, reference.DumpID)
				if err != nil {
					return nil, "", err
				}
				if !exists {
					continue
				}

				// TODO - get upload
				// TODO - check for commit existence
				locations, _, err := r.lsifStore.MonikerResults(ctx, reference.DumpID, "references", reference.Scheme, moniker.Identifier, 0, 10000000)
				if err != nil {
					return nil, "", err
				}

				// TODO - ensure deduplicated
				worklist[i].Locations = append(worklist[i].Locations, TEMPORARY2{
					Upload:    upload,
					Locations: locations,
				})
			}
		}
	}

	//
	//
	//

	var allAdjustedLocations []AdjustedLocation
	for _, w := range worklist {
		for _, pair := range w.Locations {
			adjustedLocations, err := r.adjustLocations(ctx, resolveLocationsWithDump(pair.Upload, pair.Locations))
			if err != nil {
				return nil, "", err
			}

			allAdjustedLocations = append(allAdjustedLocations, adjustedLocations...)
		}
	}

	// TODO - cursor
	return allAdjustedLocations, "", nil
}
