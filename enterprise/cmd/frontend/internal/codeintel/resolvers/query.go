package resolvers

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/inconshreveable/log15"
	"github.com/opentracing/opentracing-go/log"
	"github.com/pkg/errors"

	store "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/dbstore"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/lsifstore"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

var ErrMissingDump = errors.New("missing dump")

type ResolvedCodeIntelligenceRange struct {
	Range       lsifstore.Range
	Definitions []ResolvedLocation
	References  []ResolvedLocation
	HoverText   string
}

type ResolvedLocation struct {
	Dump  store.Dump
	Path  string
	Range lsifstore.Range
}

type ResolvedDiagnostic struct {
	Dump       store.Dump
	Diagnostic lsifstore.Diagnostic
}

// AdjustedLocation is similar to a ResolvedLocation, but with fields denoting
// the commit and range adjusted for the target commit (when the requested commit is not indexed).
type AdjustedLocation struct {
	Dump           store.Dump
	Path           string
	AdjustedCommit string
	AdjustedRange  lsifstore.Range
}

// AdjustedDiagnostic is similar to a ResolvedDiagnostic, but with fields denoting
// the commit and range adjusted for the target commit (when the requested commit is not indexed).
type AdjustedDiagnostic struct {
	lsifstore.Diagnostic
	Dump           store.Dump
	AdjustedCommit string
	AdjustedRange  lsifstore.Range
}

// AdjustedCodeIntelligenceRange is similar to a CodeIntelligenceRange,
// but with adjusted definition and reference locations.
type AdjustedCodeIntelligenceRange struct {
	Range       lsifstore.Range
	Definitions []AdjustedLocation
	References  []AdjustedLocation
	HoverText   string
}

// QueryResolver is the main interface to bundle-related operations exposed to the GraphQL API. This
// resolver consolidates the logic for bundle operations and is not itself concerned with GraphQL/API
// specifics (auth, validation, marshaling, etc.). This resolver is wrapped by a symmetrics resolver
// in this package's graphql subpackage, which is exposed directly by the API.
type QueryResolver interface {
	Ranges(ctx context.Context, startLine, endLine int) ([]AdjustedCodeIntelligenceRange, error)
	Definitions(ctx context.Context, line, character int) ([]AdjustedLocation, error)
	References(ctx context.Context, line, character, limit int, rawCursor string) ([]AdjustedLocation, string, error)
	Hover(ctx context.Context, line, character int) (string, lsifstore.Range, bool, error)
	Diagnostics(ctx context.Context, limit int) ([]AdjustedDiagnostic, int, error)
}

type queryResolver struct {
	dbStore          DBStore
	lsifStore        LSIFStore
	positionAdjuster PositionAdjuster
	repositoryID     int
	commit           string
	path             string
	uploads          []store.Dump
	operations       *operations
}

// NewQueryResolver create a new query resolver with the given services. The methods of this
// struct return queries for the given repository, commit, and path, and will query only the
// bundles associated with the given dump objects.
func NewQueryResolver(
	dbStore DBStore,
	lsifStore LSIFStore,
	positionAdjuster PositionAdjuster,
	repositoryID int,
	commit string,
	path string,
	uploads []store.Dump,
	operations *operations,
) QueryResolver {
	return &queryResolver{
		dbStore:          dbStore,
		lsifStore:        lsifStore,
		positionAdjuster: positionAdjuster,
		operations:       operations,
		repositoryID:     repositoryID,
		commit:           commit,
		path:             path,
		uploads:          uploads,
	}
}

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

	Ranges := func(ctx context.Context, file string, startLine, endLine, uploadID int) (_ []ResolvedCodeIntelligenceRange, err error) {
		dump, exists, err := r.dbStore.GetDumpByID(ctx, uploadID)
		if err != nil {
			return nil, errors.Wrap(err, "store.GetDumpByID")
		}
		if !exists {
			return nil, ErrMissingDump
		}

		pathInBundle := strings.TrimPrefix(file, dump.Root)
		ranges, err := r.lsifStore.Ranges(ctx, dump.ID, pathInBundle, startLine, endLine)
		if err != nil {
			if err == lsifstore.ErrNotFound {
				log15.Warn("Bundle does not exist")
				return nil, nil
			}
			return nil, errors.Wrap(err, "bundleClient.Ranges")
		}

		var codeintelRanges []ResolvedCodeIntelligenceRange
		for _, r := range ranges {
			codeintelRanges = append(codeintelRanges, ResolvedCodeIntelligenceRange{
				Range:       r.Range,
				Definitions: resolveLocationsWithDump(dump, r.Definitions),
				References:  resolveLocationsWithDump(dump, r.References),
				HoverText:   r.HoverText,
			})
		}

		return codeintelRanges, nil
	}

	var adjustedRanges []AdjustedCodeIntelligenceRange
	for i := range r.uploads {
		adjustedPath, ok, err := r.positionAdjuster.AdjustPath(ctx, r.uploads[i].Commit, r.path, false)
		if err != nil {
			return nil, err
		}
		if !ok {
			continue
		}

		// TODO(efritz) - determine how to do best-effort line adjustments for this case
		ranges, err := Ranges(ctx, adjustedPath, startLine, endLine, r.uploads[i].ID)
		if err != nil {
			return nil, err
		}

		for _, rn := range ranges {
			adjustedDefinitions, err := r.adjustLocations(ctx, rn.Definitions)
			if err != nil {
				return nil, err
			}

			adjustedReferences, err := r.adjustLocations(ctx, rn.References)
			if err != nil {
				return nil, err
			}

			_, adjustedRange, err := r.adjustRange(ctx, r.uploads[i].RepositoryID, r.uploads[i].Commit, adjustedPath, rn.Range)
			if err != nil {
				return nil, err
			}

			adjustedRanges = append(adjustedRanges, AdjustedCodeIntelligenceRange{
				Range:       adjustedRange,
				Definitions: adjustedDefinitions,
				References:  adjustedReferences,
				HoverText:   rn.HoverText,
			})
		}
	}

	return adjustedRanges, nil
}

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

	const defintionMonikersLimit = 100

	Definitions := func(ctx context.Context, file string, line, character, uploadID int) (_ []ResolvedLocation, err error) {
		dump, exists, err := r.dbStore.GetDumpByID(ctx, uploadID)
		if err != nil {
			return nil, errors.Wrap(err, "store.GetDumpByID")
		}
		if !exists {
			return nil, ErrMissingDump
		}

		pathInBundle := strings.TrimPrefix(file, dump.Root)
		locations, err := r.lsifStore.Definitions(ctx, dump.ID, pathInBundle, line, character)
		if err != nil {
			if err == lsifstore.ErrNotFound {
				log15.Warn("Bundle does not exist")
				return nil, nil
			}
			return nil, errors.Wrap(err, "bundleClient.Definitions")
		}
		if len(locations) > 0 {
			return resolveLocationsWithDump(dump, locations), nil
		}

		rangeMonikers, err := r.lsifStore.MonikersByPosition(context.Background(), dump.ID, pathInBundle, line, character)
		if err != nil {
			if err == lsifstore.ErrNotFound {
				log15.Warn("Bundle does not exist")
				return nil, nil
			}
			return nil, errors.Wrap(err, "bundleClient.MonikersByPosition")
		}

		for _, monikers := range rangeMonikers {
			for _, moniker := range monikers {
				if moniker.Kind == "import" {
					locations, _, err := lookupMoniker(r.dbStore, r.lsifStore, dump.ID, pathInBundle, "definitions", moniker, 0, defintionMonikersLimit)
					if err != nil {
						return nil, err
					}
					if len(locations) > 0 {
						return locations, nil
					}
				} else {
					// This symbol was not imported from another bundle. We search the definitions
					// of our own bundle in case there was a definition that wasn't properly attached
					// to a result set but did have the correct monikers attached.

					locations, _, err := r.lsifStore.MonikerResults(context.Background(), dump.ID, "definitions", moniker.Scheme, moniker.Identifier, 0, defintionMonikersLimit)
					if err != nil {
						if err == lsifstore.ErrNotFound {
							log15.Warn("Bundle does not exist")
							return nil, nil
						}
						return nil, errors.Wrap(err, "bundleClient.MonikerResults")
					}
					if len(locations) > 0 {
						return resolveLocationsWithDump(dump, locations), nil
					}
				}
			}
		}

		return nil, nil
	}

	position := lsifstore.Position{Line: line, Character: character}

	for i := range r.uploads {
		adjustedPath, adjustedPosition, ok, err := r.positionAdjuster.AdjustPosition(ctx, r.uploads[i].Commit, r.path, position, false)
		if err != nil {
			return nil, err
		}
		if !ok {
			continue
		}

		locations, err := Definitions(ctx, adjustedPath, adjustedPosition.Line, adjustedPosition.Character, r.uploads[i].ID)
		if err != nil {
			return nil, err
		}
		if len(locations) == 0 {
			continue
		}

		return r.adjustLocations(ctx, locations)
	}

	return nil, nil
}

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

	// ErrIllegalLimit occurs when a zero-length page of references is requested
	var ErrIllegalLimit = errors.New("limit must be positive")

	// RemoteDumpLimit is the limit for fetching batches of remote dumps.
	const RemoteDumpLimit = 20

	References := func(ctx context.Context, repositoryID int, commit string, limit int, cursor Cursor) (_ []ResolvedLocation, _ Cursor, _ bool, err error) {
		if limit <= 0 {
			return nil, Cursor{}, false, ErrIllegalLimit
		}

		return NewReferencePageResolver(
			r.dbStore,
			r.lsifStore,
			repositoryID,
			commit,
			RemoteDumpLimit,
			limit).ResolvePage(ctx, cursor)
	}

	position := lsifstore.Position{Line: line, Character: character}

	// Decode a map of upload ids to the next url that serves
	// the new page of results. This may not include an entry
	// for every upload if their result sets have already been
	// exhausted.
	cursors, err := readCursor(rawCursor)
	if err != nil {
		return nil, "", err
	}

	// We need to maintain a symmetric map for the next page
	// of results that we can encode into the endCursor of
	// this request.
	newCursors := map[int]string{}

	var allLocations []ResolvedLocation
	for i := range r.uploads {
		rawCursor := ""
		if cursor, ok := cursors[r.uploads[i].ID]; ok {
			rawCursor = cursor
		} else if len(cursors) != 0 {
			// Result set is exhausted or newer than the first page
			// of results. Skip anything from this upload as it will
			// have duplicate results, or it will be out of order.
			continue
		}

		adjustedPath, adjustedPosition, ok, err := r.positionAdjuster.AdjustPosition(ctx, r.uploads[i].Commit, r.path, position, false)
		if err != nil {
			return nil, "", err
		}
		if !ok {
			continue
		}

		cursor, err := DecodeOrCreateCursor(adjustedPath, adjustedPosition.Line, adjustedPosition.Character, r.uploads[i].ID, rawCursor, r.dbStore, r.lsifStore)
		if err != nil {
			return nil, "", err
		}

		locations, newCursor, hasNewCursor, err := References(ctx, r.repositoryID, r.commit, limit, cursor)
		if err != nil {
			return nil, "", err
		}

		allLocations = append(allLocations, locations...)
		if hasNewCursor {
			newCursors[r.uploads[i].ID] = EncodeCursor(newCursor)
		}
	}

	endCursor, err := makeCursor(newCursors)
	if err != nil {
		return nil, "", err
	}

	adjustedLocations, err := r.adjustLocations(ctx, allLocations)
	if err != nil {
		return nil, "", err
	}

	return adjustedLocations, endCursor, nil
}

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

	const defintionMonikersLimit = 100

	definitionsRaw := func(ctx context.Context, dump store.Dump, pathInBundle string, line, character int) ([]ResolvedLocation, error) {
		locations, err := r.lsifStore.Definitions(ctx, dump.ID, pathInBundle, line, character)
		if err != nil {
			if err == lsifstore.ErrNotFound {
				log15.Warn("Bundle does not exist")
				return nil, nil
			}
			return nil, errors.Wrap(err, "bundleClient.Definitions")
		}
		if len(locations) > 0 {
			return resolveLocationsWithDump(dump, locations), nil
		}

		rangeMonikers, err := r.lsifStore.MonikersByPosition(context.Background(), dump.ID, pathInBundle, line, character)
		if err != nil {
			if err == lsifstore.ErrNotFound {
				log15.Warn("Bundle does not exist")
				return nil, nil
			}
			return nil, errors.Wrap(err, "bundleClient.MonikersByPosition")
		}

		for _, monikers := range rangeMonikers {
			for _, moniker := range monikers {
				if moniker.Kind == "import" {
					locations, _, err := lookupMoniker(r.dbStore, r.lsifStore, dump.ID, pathInBundle, "definitions", moniker, 0, defintionMonikersLimit)
					if err != nil {
						return nil, err
					}
					if len(locations) > 0 {
						return locations, nil
					}
				} else {
					// This symbol was not imported from another bundle. We search the definitions
					// of our own bundle in case there was a definition that wasn't properly attached
					// to a result set but did have the correct monikers attached.

					locations, _, err := r.lsifStore.MonikerResults(context.Background(), dump.ID, "definitions", moniker.Scheme, moniker.Identifier, 0, defintionMonikersLimit)
					if err != nil {
						if err == lsifstore.ErrNotFound {
							log15.Warn("Bundle does not exist")
							return nil, nil
						}
						return nil, errors.Wrap(err, "bundleClient.MonikerResults")
					}
					if len(locations) > 0 {
						return resolveLocationsWithDump(dump, locations), nil
					}
				}
			}
		}

		return nil, nil
	}

	definitionRaw := func(ctx context.Context, dump store.Dump, pathInBundle string, line, character int) (ResolvedLocation, bool, error) {
		resolved, err := definitionsRaw(ctx, dump, pathInBundle, line, character)
		if err != nil || len(resolved) == 0 {
			return ResolvedLocation{}, false, errors.Wrap(err, "api.definitionsRaw")
		}

		return resolved[0], true, nil
	}

	Hover := func(ctx context.Context, file string, line, character, uploadID int) (_ string, _ lsifstore.Range, _ bool, err error) {
		dump, exists, err := r.dbStore.GetDumpByID(ctx, uploadID)
		if err != nil {
			return "", lsifstore.Range{}, false, errors.Wrap(err, "store.GetDumpByID")
		}
		if !exists {
			return "", lsifstore.Range{}, false, ErrMissingDump
		}

		pathInBundle := strings.TrimPrefix(file, dump.Root)
		text, rn, exists, err := r.lsifStore.Hover(ctx, dump.ID, pathInBundle, line, character)
		if err != nil {
			if err == lsifstore.ErrNotFound {
				log15.Warn("Bundle does not exist")
				return "", lsifstore.Range{}, false, nil
			}
			return "", lsifstore.Range{}, false, errors.Wrap(err, "bundleClient.Hover")
		}
		if exists {
			return text, rn, true, nil
		}

		definition, exists, err := definitionRaw(ctx, dump, pathInBundle, line, character)
		if err != nil || !exists {
			return "", lsifstore.Range{}, false, errors.Wrap(err, "api.definitionRaw")
		}

		pathInDefinitionBundle := strings.TrimPrefix(definition.Path, definition.Dump.Root)

		text, rn, exists, err = r.lsifStore.Hover(ctx, definition.Dump.ID, pathInDefinitionBundle, definition.Range.Start.Line, definition.Range.Start.Character)
		if err != nil {
			if err == lsifstore.ErrNotFound {
				log15.Warn("Bundle does not exist")
				return "", lsifstore.Range{}, false, nil
			}
			return "", lsifstore.Range{}, false, errors.Wrap(err, "definitionBundleClient.Hover")
		}

		return text, rn, exists, nil
	}

	position := lsifstore.Position{Line: line, Character: character}

	for i := range r.uploads {
		adjustedPath, adjustedPosition, ok, err := r.positionAdjuster.AdjustPosition(ctx, r.uploads[i].Commit, r.path, position, false)
		if err != nil {
			return "", lsifstore.Range{}, false, err
		}
		if !ok {
			continue
		}

		text, rn, exists, err := Hover(ctx, adjustedPath, adjustedPosition.Line, adjustedPosition.Character, r.uploads[i].ID)
		if err != nil {
			return "", lsifstore.Range{}, false, err
		}
		if !exists || text == "" {
			continue
		}

		if _, adjustedRange, ok, err := r.positionAdjuster.AdjustRange(ctx, r.uploads[i].Commit, r.path, rn, true); err != nil {
			return "", lsifstore.Range{}, false, err
		} else if ok {
			return text, adjustedRange, true, nil
		}

		// Failed to adjust range. This _might_ happen in cases where the LSIF range
		// spans multiple lines which intersect a diff; the hover position on an earlier
		// line may not be edited, but the ending line of the expression may have been
		// edited or removed. This is rare and unfortunate, and we'll skip the result
		// in this case because we have low confidence that it will be rendered correctly.
		continue
	}

	return "", lsifstore.Range{}, false, nil
}

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

	Diagnostics := func(ctx context.Context, prefix string, uploadID, limit, offset int) (_ []ResolvedDiagnostic, _ int, err error) {
		dump, exists, err := r.dbStore.GetDumpByID(ctx, uploadID)
		if err != nil {
			return nil, 0, errors.Wrap(err, "store.GetDumpByID")
		}
		if !exists {
			return nil, 0, ErrMissingDump
		}

		pathInBundle := strings.TrimPrefix(prefix, dump.Root)
		diagnostics, totalCount, err := r.lsifStore.Diagnostics(ctx, dump.ID, pathInBundle, offset, limit)
		if err != nil {
			if err == lsifstore.ErrNotFound {
				log15.Warn("Bundle does not exist")
				return nil, 0, nil
			}
			return nil, 0, errors.Wrap(err, "bundleClient.Diagnostics")
		}

		return resolveDiagnosticsWithDump(dump, diagnostics), totalCount, nil
	}

	totalCount := 0
	var allDiagnostics []ResolvedDiagnostic
	for i := range r.uploads {
		adjustedPath, ok, err := r.positionAdjuster.AdjustPath(ctx, r.uploads[i].Commit, r.path, false)
		if err != nil {
			return nil, 0, err
		}
		if !ok {
			continue
		}

		l := limit - len(allDiagnostics)
		if l < 0 {
			l = 0
		}

		diagnostics, count, err := Diagnostics(ctx, adjustedPath, r.uploads[i].ID, l, 0)
		if err != nil {
			return nil, 0, err
		}

		totalCount += count
		allDiagnostics = append(allDiagnostics, diagnostics...)
	}

	adjustedDiagnostics := make([]AdjustedDiagnostic, 0, len(allDiagnostics))
	for i := range allDiagnostics {
		clientRange := lsifstore.Range{
			Start: lsifstore.Position{Line: allDiagnostics[i].Diagnostic.StartLine, Character: allDiagnostics[i].Diagnostic.StartCharacter},
			End:   lsifstore.Position{Line: allDiagnostics[i].Diagnostic.EndLine, Character: allDiagnostics[i].Diagnostic.EndCharacter},
		}

		adjustedCommit, adjustedRange, err := r.adjustRange(ctx, allDiagnostics[i].Dump.RepositoryID, allDiagnostics[i].Dump.Commit, allDiagnostics[i].Diagnostic.Path, clientRange)
		if err != nil {
			return nil, 0, err
		}

		adjustedDiagnostics = append(adjustedDiagnostics, AdjustedDiagnostic{
			Diagnostic:     allDiagnostics[i].Diagnostic,
			Dump:           allDiagnostics[i].Dump,
			AdjustedCommit: adjustedCommit,
			AdjustedRange:  adjustedRange,
		})
	}

	return adjustedDiagnostics, totalCount, nil
}

// uploadIDs returns a slice of this query's matched upload identifiers.
func (r *queryResolver) uploadIDs() []string {
	uploadIDs := make([]string, 0, len(r.uploads))
	for i := range r.uploads {
		uploadIDs = append(uploadIDs, strconv.Itoa(r.uploads[i].ID))
	}

	return uploadIDs
}

// adjustLocations translates a list of resolved locations (relative to the indexed commit) into a list of
// equivalent locations in the requested commit.
func (r *queryResolver) adjustLocations(ctx context.Context, locations []ResolvedLocation) ([]AdjustedLocation, error) {
	adjustedLocations := make([]AdjustedLocation, 0, len(locations))
	for i := range locations {
		adjustedCommit, adjustedRange, err := r.adjustRange(ctx, locations[i].Dump.RepositoryID, locations[i].Dump.Commit, locations[i].Path, locations[i].Range)
		if err != nil {
			return nil, err
		}

		adjustedLocations = append(adjustedLocations, AdjustedLocation{
			Dump:           locations[i].Dump,
			Path:           locations[i].Path,
			AdjustedCommit: adjustedCommit,
			AdjustedRange:  adjustedRange,
		})
	}

	return adjustedLocations, nil
}

// adjustRange translates a range (relative to the indexed commit) into an equivalent range in the requested commit.
func (r *queryResolver) adjustRange(ctx context.Context, repositoryID int, commit, path string, rx lsifstore.Range) (string, lsifstore.Range, error) {
	if repositoryID != r.repositoryID {
		// No diffs exist for translation between repos
		return commit, rx, nil
	}

	if _, adjustedRange, ok, err := r.positionAdjuster.AdjustRange(ctx, commit, path, rx, true); err != nil {
		return "", lsifstore.Range{}, err
	} else if ok {
		return r.commit, adjustedRange, nil
	}

	return commit, rx, nil
}

// readCursor decodes a cursor into a map from upload ids to URLs that serves the next page of results.
func readCursor(after string) (map[int]string, error) {
	if after == "" {
		return nil, nil
	}

	var cursors map[int]string
	if err := json.Unmarshal([]byte(after), &cursors); err != nil {
		return nil, err
	}
	return cursors, nil
}

// makeCursor encodes a map from upload ids to URLs that serves the next page of results into a single string
// that can be sent back for use in cursor pagination.
func makeCursor(cursors map[int]string) (string, error) {
	if len(cursors) == 0 {
		return "", nil
	}

	encoded, err := json.Marshal(cursors)
	if err != nil {
		return "", err
	}
	return string(encoded), nil
}

func resolveLocationsWithDump(dump store.Dump, locations []lsifstore.Location) []ResolvedLocation {
	var resolvedLocations []ResolvedLocation
	for _, location := range locations {
		resolvedLocations = append(resolvedLocations, ResolvedLocation{
			Dump:  dump,
			Path:  dump.Root + location.Path,
			Range: location.Range,
		})
	}

	return resolvedLocations
}

func lookupMoniker(
	dbStore DBStore,
	lsifStore LSIFStore,
	dumpID int,
	path string,
	modelType string,
	moniker lsifstore.MonikerData,
	skip int,
	take int,
) ([]ResolvedLocation, int, error) {
	if moniker.PackageInformationID == "" {
		return nil, 0, nil
	}

	pid, _, err := lsifStore.PackageInformation(context.Background(), dumpID, path, string(moniker.PackageInformationID))
	if err != nil {
		if err == lsifstore.ErrNotFound {
			log15.Warn("Bundle does not exist")
			return nil, 0, nil
		}
		return nil, 0, errors.Wrap(err, "lsifStore.BundleClient")
	}

	dump, exists, err := dbStore.GetPackage(context.Background(), moniker.Scheme, pid.Name, pid.Version)
	if err != nil || !exists {
		return nil, 0, errors.Wrap(err, "store.GetPackage")
	}

	locations, count, err := lsifStore.MonikerResults(context.Background(), dump.ID, modelType, moniker.Scheme, moniker.Identifier, skip, take)
	if err != nil {
		if err == lsifstore.ErrNotFound {
			log15.Warn("Bundle does not exist")
			return nil, 0, nil
		}
		return nil, 0, errors.Wrap(err, "lsifStore.BundleClient")
	}

	return resolveLocationsWithDump(dump, locations), count, nil
}

func resolveDiagnosticsWithDump(dump store.Dump, diagnostics []lsifstore.Diagnostic) []ResolvedDiagnostic {
	var resolvedDiagnostics []ResolvedDiagnostic
	for _, diagnostic := range diagnostics {
		diagnostic.Path = dump.Root + diagnostic.Path
		resolvedDiagnostics = append(resolvedDiagnostics, ResolvedDiagnostic{
			Dump:       dump,
			Diagnostic: diagnostic,
		})
	}

	return resolvedDiagnostics
}
