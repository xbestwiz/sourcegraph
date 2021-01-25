package resolvers

// // Cursor holds the complete state necessary to page through a reference result set.
// type Cursor struct {
// 	Phase                  string                  // common
// 	DumpID                 int                     // common
// 	Path                   string                  // same-dump/same-dump-monikers/definition-monikers
// 	Line                   int                     // same-dump/same-dump-monikers
// 	Character              int                     // same-dump/same-dump-monikers
// 	Monikers               []lsifstore.MonikerData // same-dump/same-dump-monikers/definition-monikers
// 	SkipResults            int                     // same-dump/same-dump-monikers/definition-monikers
// 	Identifier             string                  // same-repo/remote-repo
// 	Scheme                 string                  // same-repo/remote-repo
// 	Name                   string                  // same-repo/remote-repo
// 	Version                string                  // same-repo/remote-repo
// 	DumpIDs                []int                   // same-repo/remote-repo
// 	TotalDumpsWhenBatching int                     // same-repo/remote-repo
// 	SkipDumpsWhenBatching  int                     // same-repo/remote-repo
// 	SkipDumpsInBatch       int                     // same-repo/remote-repo
// 	SkipResultsInDump      int                     // same-repo/remote-repo
// }

// // EncodeCursor returns an encoding of the given cursor suitable for a URL.
// func EncodeCursor(cursor Cursor) string {
// 	rawEncoded, _ := json.Marshal(cursor)
// 	return base64.RawURLEncoding.EncodeToString(rawEncoded)
// }

// // decodeCursor is the inverse of EncodeCursor.
// func decodeCursor(rawEncoded string) (Cursor, error) {
// 	raw, err := base64.RawURLEncoding.DecodeString(rawEncoded)
// 	if err != nil {
// 		return Cursor{}, err
// 	}

// 	var cursor Cursor
// 	err = json.Unmarshal([]byte(raw), &cursor)
// 	return cursor, err
// }

// // DecodeOrCreateCursor decodes and returns the raw cursor, or creates a new initial page cursor
// // if a raw cursor is not supplied.
// func DecodeOrCreateCursor(ctx context.Context, path string, line, character, uploadID int, rawCursor string, dbStore DBStore, lsifStore LSIFStore) (Cursor, error) {
// 	if rawCursor != "" {
// 		cursor, err := decodeCursor(rawCursor)
// 		if err != nil {
// 			return Cursor{}, err
// 		}

// 		return cursor, nil
// 	}

// 	dump, exists, err := dbStore.GetDumpByID(ctx, uploadID)
// 	if err != nil {
// 		return Cursor{}, errors.Wrap(err, "store.GetDumpByID")
// 	}
// 	if !exists {
// 		return Cursor{}, ErrMissingDump
// 	}

// 	pathInBundle := strings.TrimPrefix(path, dump.Root)
// 	rangeMonikers, err := lsifStore.MonikersByPosition(ctx, dump.ID, pathInBundle, line, character)
// 	if err != nil {
// 		return Cursor{}, errors.Wrap(err, "bundleClient.MonikersByPosition")
// 	}

// 	var flattened []lsifstore.MonikerData
// 	for _, monikers := range rangeMonikers {
// 		flattened = append(flattened, monikers...)
// 	}

// 	return Cursor{
// 		Phase:       "same-dump",
// 		DumpID:      dump.ID,
// 		Path:        pathInBundle,
// 		Line:        line,
// 		Character:   character,
// 		Monikers:    flattened,
// 		SkipResults: 0,
// 	}, nil
// }

// //
// //
// //

// type ReferencePageResolver struct {
// 	dbStore         DBStore
// 	lsifStore       LSIFStore
// 	gitserverClient GitserverClient
// 	repositoryID    int
// 	commit          string
// 	remoteDumpLimit int
// 	limit           int
// }

// func (s *ReferencePageResolver) ResolvePage(ctx context.Context, cursor Cursor) ([]ResolvedLocation, Cursor, bool, error) {
// 	var allLocations []ResolvedLocation

// 	for {
// 		locations, newCursor, hasNewCursor, err := s.dispatchCursorHandler(ctx, cursor)
// 		if err != nil {
// 			return nil, Cursor{}, false, err
// 		}

// 		s.limit -= len(locations)
// 		allLocations = append(allLocations, locations...)

// 		if !hasNewCursor || s.limit <= 0 {
// 			return allLocations, newCursor, hasNewCursor, err
// 		}

// 		cursor = newCursor
// 	}
// }

// func (s *ReferencePageResolver) dispatchCursorHandler(ctx context.Context, cursor Cursor) ([]ResolvedLocation, Cursor, bool, error) {
// 	fns := map[string]func(context.Context, Cursor) ([]ResolvedLocation, Cursor, bool, error){
// 		"same-dump":           s.handleSameDumpCursor,
// 		"definition-monikers": s.handleDefinitionMonikersCursor,
// 		"same-repo":           s.handleSameRepoCursor,
// 		"remote-repo":         s.handleRemoteRepoCursor,
// 	}

// 	fn, exists := fns[cursor.Phase]
// 	if !exists {
// 		return nil, Cursor{}, false, fmt.Errorf("unknown cursor phase %s", cursor.Phase)
// 	}

// 	locations, newCursor, hasNewCursor, err := fn(ctx, cursor)
// 	if err != nil {
// 		return nil, Cursor{}, false, pkgerrors.Wrap(err, cursor.Phase)
// 	}
// 	return locations, newCursor, hasNewCursor, nil
// }

// func (s *ReferencePageResolver) handleSameDumpCursor(ctx context.Context, cursor Cursor) ([]ResolvedLocation, Cursor, bool, error) {
// 	dump, exists, err := s.dbStore.GetDumpByID(ctx, cursor.DumpID)
// 	if err != nil {
// 		return nil, Cursor{}, false, pkgerrors.Wrap(err, "store.GetDumpByID")
// 	}
// 	if !exists {
// 		return nil, Cursor{}, false, ErrMissingDump
// 	}

// 	locations, err := s.lsifStore.References(ctx, dump.ID, cursor.Path, cursor.Line, cursor.Character)
// 	if err != nil {
// 		if err == lsifstore.ErrNotFound {
// 			log15.Warn("Bundle does not exist")
// 			return nil, Cursor{}, false, nil
// 		}
// 		return nil, Cursor{}, false, pkgerrors.Wrap(err, "bundleClient.References")
// 	}

// 	resolvedLocations := resolveLocationsWithDump(dump, sliceLocations(locations, cursor.SkipResults, cursor.SkipResults+s.limit))

// 	if newOffset := cursor.SkipResults + s.limit; newOffset <= len(locations) {
// 		newCursor := Cursor{
// 			Phase:       cursor.Phase,
// 			DumpID:      cursor.DumpID,
// 			Path:        cursor.Path,
// 			Line:        cursor.Line,
// 			Character:   cursor.Character,
// 			Monikers:    cursor.Monikers,
// 			SkipResults: newOffset,
// 		}
// 		return resolvedLocations, newCursor, true, nil
// 	}

// 	newCursor := Cursor{
// 		Phase:       "definition-monikers",
// 		DumpID:      cursor.DumpID,
// 		Path:        cursor.Path,
// 		Line:        cursor.Line,
// 		Character:   cursor.Character,
// 		Monikers:    cursor.Monikers,
// 		SkipResults: 0,
// 	}
// 	return resolvedLocations, newCursor, true, nil
// }

// func (s *ReferencePageResolver) handleDefinitionMonikersCursor(ctx context.Context, cursor Cursor) ([]ResolvedLocation, Cursor, bool, error) {
// 	var hasNextPhaseCursor = false
// 	var nextPhaseCursor Cursor
// 	for _, moniker := range cursor.Monikers {
// 		if moniker.PackageInformationID == "" {
// 			continue
// 		}

// 		packageInformation, _, err := s.lsifStore.PackageInformation(ctx, cursor.DumpID, cursor.Path, string(moniker.PackageInformationID))
// 		if err != nil {
// 			if err == lsifstore.ErrNotFound {
// 				log15.Warn("Bundle does not exist")
// 				return nil, Cursor{}, false, nil
// 			}
// 			return nil, Cursor{}, false, pkgerrors.Wrap(err, "lsifStore.PackageInformation")
// 		}

// 		hasNextPhaseCursor = true
// 		nextPhaseCursor = Cursor{
// 			DumpID:                 cursor.DumpID,
// 			Phase:                  "same-repo",
// 			Scheme:                 moniker.Scheme,
// 			Identifier:             moniker.Identifier,
// 			Name:                   packageInformation.Name,
// 			Version:                packageInformation.Version,
// 			DumpIDs:                nil,
// 			TotalDumpsWhenBatching: 0,
// 			SkipDumpsWhenBatching:  0,
// 			SkipDumpsInBatch:       0,
// 			SkipResultsInDump:      0,
// 		}
// 		break
// 	}

// 	for _, moniker := range cursor.Monikers {
// 		if moniker.Kind != "import" {
// 			continue
// 		}

// 		locations, count, err := lookupMoniker(ctx, s.dbStore, s.lsifStore, cursor.DumpID, cursor.Path, "references", moniker, cursor.SkipResults, s.limit)
// 		if err != nil {
// 			return nil, Cursor{}, false, err
// 		}
// 		if len(locations) == 0 {
// 			continue
// 		}

// 		if newOffset := cursor.SkipResults + len(locations); newOffset < count {
// 			newCursor := Cursor{
// 				Phase:       cursor.Phase,
// 				DumpID:      cursor.DumpID,
// 				Path:        cursor.Path,
// 				Monikers:    cursor.Monikers,
// 				SkipResults: newOffset,
// 			}
// 			return locations, newCursor, true, nil
// 		}

// 		return locations, nextPhaseCursor, hasNextPhaseCursor, nil
// 	}

// 	return nil, nextPhaseCursor, hasNextPhaseCursor, nil

// }

// func (s *ReferencePageResolver) handleSameRepoCursor(ctx context.Context, cursor Cursor) ([]ResolvedLocation, Cursor, bool, error) {
// 	locations, newCursor, hasNewCursor, err := s.resolveLocationsViaReferencePager(ctx, cursor, func(ctx context.Context) (int, store.ReferencePager, error) {
// 		totalCount, pager, err := s.dbStore.SameRepoPager(ctx, s.repositoryID, s.commit, cursor.Scheme, cursor.Name, cursor.Version, s.remoteDumpLimit)
// 		if err != nil {
// 			return 0, nil, pkgerrors.Wrap(err, "store.SameRepoPager")
// 		}
// 		return totalCount, pager, nil
// 	})
// 	if err != nil || hasNewCursor {
// 		return locations, newCursor, hasNewCursor, err
// 	}

// 	newCursor = Cursor{
// 		DumpID:                 cursor.DumpID,
// 		Phase:                  "remote-repo",
// 		Scheme:                 cursor.Scheme,
// 		Identifier:             cursor.Identifier,
// 		Name:                   cursor.Name,
// 		Version:                cursor.Version,
// 		DumpIDs:                nil,
// 		TotalDumpsWhenBatching: 0,
// 		SkipDumpsWhenBatching:  0,
// 		SkipDumpsInBatch:       0,
// 		SkipResultsInDump:      0,
// 	}
// 	return locations, newCursor, true, nil
// }

// func (s *ReferencePageResolver) handleRemoteRepoCursor(ctx context.Context, cursor Cursor) ([]ResolvedLocation, Cursor, bool, error) {
// 	return s.resolveLocationsViaReferencePager(ctx, cursor, func(ctx context.Context) (int, store.ReferencePager, error) {
// 		totalCount, pager, err := s.dbStore.PackageReferencePager(ctx, cursor.Scheme, cursor.Name, cursor.Version, s.repositoryID, s.remoteDumpLimit)
// 		if err != nil {
// 			return 0, nil, pkgerrors.Wrap(err, "store.PackageReferencePager")
// 		}
// 		return totalCount, pager, nil
// 	})
// }

// func (s *ReferencePageResolver) resolveLocationsViaReferencePager(ctx context.Context, cursor Cursor, createPager func(context.Context) (int, store.ReferencePager, error)) ([]ResolvedLocation, Cursor, bool, error) {
// 	dumpID := cursor.DumpID
// 	scheme := cursor.Scheme
// 	identifier := cursor.Identifier
// 	limit := s.limit
// 	commitExistenceCache := map[int]map[string]bool{}

// 	if len(cursor.DumpIDs) == 0 {
// 		totalCount, pager, err := createPager(ctx)
// 		if err != nil {
// 			return nil, Cursor{}, false, err
// 		}

// 		identifier := cursor.Identifier
// 		offset := cursor.SkipDumpsWhenBatching
// 		limit := s.remoteDumpLimit
// 		newOffset := offset

// 		var packageReferences []lsifstore.PackageReference
// 		for len(packageReferences) < limit && newOffset < totalCount {
// 			page, err := pager.PageFromOffset(ctx, newOffset)
// 			if err != nil {
// 				return nil, Cursor{}, false, pager.Done(err)
// 			}

// 			if len(page) == 0 {
// 				// Shouldn't happen, but just in case of a bug we
// 				// don't want this to throw up into an infinite loop.
// 				break
// 			}

// 			filtered, scanned := applyBloomFilter(page, identifier, limit-len(packageReferences))
// 			packageReferences = append(packageReferences, filtered...)
// 			newOffset += scanned
// 		}

// 		var dumpIDs []int
// 		for _, ref := range packageReferences {
// 			dumpIDs = append(dumpIDs, ref.DumpID)
// 		}

// 		cursor.DumpIDs = dumpIDs
// 		cursor.SkipDumpsWhenBatching = newOffset
// 		cursor.TotalDumpsWhenBatching = totalCount

// 		if err := pager.Done(nil); err != nil {
// 			return nil, Cursor{}, false, err
// 		}
// 	}

// 	for i, batchDumpID := range cursor.DumpIDs {
// 		// Skip the remote reference that show up for ourselves - we've already gathered
// 		// these in the previous step of the references query.
// 		if i < cursor.SkipDumpsInBatch || batchDumpID == dumpID {
// 			continue
// 		}

// 		dump, exists, err := s.dbStore.GetDumpByID(ctx, batchDumpID)
// 		if err != nil {
// 			return nil, Cursor{}, false, pkgerrors.Wrap(err, "store.GetDumpByID")
// 		}
// 		if !exists {
// 			continue
// 		}

// 		if _, ok := commitExistenceCache[dump.RepositoryID]; !ok {
// 			commitExistenceCache[dump.RepositoryID] = map[string]bool{}
// 		}

// 		// We've already determined the target commit doesn't exist
// 		if exists, ok := commitExistenceCache[dump.RepositoryID][dump.Commit]; ok && !exists {
// 			continue
// 		}

// 		commitExists, err := s.gitserverClient.CommitExists(ctx, dump.RepositoryID, dump.Commit)
// 		if err != nil {
// 			return nil, Cursor{}, false, pkgerrors.Wrap(err, "gitserverClient.CommitExists")
// 		}

// 		// Cache result as we're likely to have multiple
// 		// dumps per commit if there are overlapping roots.
// 		commitExistenceCache[dump.RepositoryID][dump.Commit] = exists

// 		if !commitExists {
// 			continue
// 		}

// 		results, count, err := s.lsifStore.MonikerResults(ctx, batchDumpID, "references", scheme, identifier, cursor.SkipResultsInDump, limit)
// 		if err != nil {
// 			if err == lsifstore.ErrNotFound {
// 				log15.Warn("Bundle does not exist")
// 				return nil, Cursor{}, false, nil
// 			}
// 			return nil, Cursor{}, false, pkgerrors.Wrap(err, "bundleClient.MonikerResults")
// 		}
// 		if len(results) == 0 {
// 			continue
// 		}
// 		resolvedLocations := resolveLocationsWithDump(dump, results)

// 		if newResultOffset := cursor.SkipResultsInDump + len(results); newResultOffset < count {
// 			newCursor := cursor
// 			newCursor.SkipResultsInDump = newResultOffset
// 			return resolvedLocations, newCursor, true, nil
// 		}

// 		if i+1 < len(cursor.DumpIDs) {
// 			newCursor := cursor
// 			newCursor.SkipDumpsInBatch = i + 1
// 			newCursor.SkipResultsInDump = 0
// 			return resolvedLocations, newCursor, true, nil
// 		}

// 		if cursor.SkipDumpsWhenBatching < cursor.TotalDumpsWhenBatching {
// 			newCursor := cursor
// 			newCursor.DumpIDs = []int{}
// 			newCursor.SkipDumpsInBatch = 0
// 			newCursor.SkipResultsInDump = 0
// 			return resolvedLocations, newCursor, true, nil
// 		}

// 		return resolvedLocations, Cursor{}, false, nil
// 	}

// 	return nil, Cursor{}, false, nil
// }

// func hashLocation(location lsifstore.Location) string {
// 	return fmt.Sprintf(
// 		"%s:%d:%d:%d:%d",
// 		location.Path,
// 		location.Range.Start.Line,
// 		location.Range.Start.Character,
// 		location.Range.End.Line,
// 		location.Range.End.Character,
// 	)
// }

// func applyBloomFilter(packageReferences []lsifstore.PackageReference, identifier string, limit int) ([]lsifstore.PackageReference, int) {
// 	var filteredReferences []lsifstore.PackageReference
// 	for i, ref := range packageReferences {
// 		test, err := bloomfilter.DecodeAndTestFilter([]byte(ref.Filter), identifier)
// 		if err != nil || !test {
// 			continue
// 		}

// 		filteredReferences = append(filteredReferences, ref)

// 		if len(filteredReferences) >= limit {
// 			return filteredReferences, i + 1
// 		}
// 	}

// 	return filteredReferences, len(packageReferences)
// }

// func sliceLocations(locations []lsifstore.Location, lo, hi int) []lsifstore.Location {
// 	if lo >= len(locations) {
// 		return nil
// 	}
// 	if hi >= len(locations) {
// 		hi = len(locations)
// 	}
// 	return locations[lo:hi]
// }

// func lookupMoniker(
// 	ctx context.Context,
// 	dbStore DBStore,
// 	lsifStore LSIFStore,
// 	dumpID int,
// 	path string,
// 	modelType string,
// 	moniker lsifstore.MonikerData,
// 	skip int,
// 	take int,
// ) ([]ResolvedLocation, int, error) {
// 	if moniker.PackageInformationID == "" {
// 		return nil, 0, nil
// 	}

// 	pid, _, err := lsifStore.PackageInformation(ctx, dumpID, path, string(moniker.PackageInformationID))
// 	if err != nil {
// 		return nil, 0, err
// 	}

// 	dump, exists, err := dbStore.GetPackage(ctx, moniker.Scheme, pid.Name, pid.Version)
// 	if err != nil || !exists {
// 		return nil, 0, err
// 	}

// 	locations, count, err := lsifStore.MonikerResults(ctx, dump.ID, modelType, moniker.Scheme, moniker.Identifier, skip, take)
// 	if err != nil {
// 		return nil, 0, err
// 	}

// 	return resolveLocationsWithDump(dump, locations), count, nil
// }

// // ErrIllegalLimit occurs when a zero-length page of references is requested
// var ErrIllegalLimit = errors.New("limit must be positive")

// // remoteDumpLimit is the limit for fetching batches of remote dumps.
// const remoteDumpLimit = 20
