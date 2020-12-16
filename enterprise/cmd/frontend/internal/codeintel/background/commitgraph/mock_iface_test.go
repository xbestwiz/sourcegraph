// Code generated by github.com/efritz/go-mockgen 0.1.0; DO NOT EDIT.

package commitgraph

import (
	"context"
	gitserver "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/gitserver"
	dbstore "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/dbstore"
	"sync"
	"time"
)

// MockDBStore is a mock implementation of the DBStore interface (from the
// package
// github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/codeintel/background/commitgraph)
// used for unit testing.
type MockDBStore struct {
	// CalculateVisibleUploadsFunc is an instance of a mock function object
	// controlling the behavior of the method CalculateVisibleUploads.
	CalculateVisibleUploadsFunc *DBStoreCalculateVisibleUploadsFunc
	// DirtyRepositoriesFunc is an instance of a mock function object
	// controlling the behavior of the method DirtyRepositories.
	DirtyRepositoriesFunc *DBStoreDirtyRepositoriesFunc
	// GetUploadsFunc is an instance of a mock function object controlling
	// the behavior of the method GetUploads.
	GetUploadsFunc *DBStoreGetUploadsFunc
	// LockFunc is an instance of a mock function object controlling the
	// behavior of the method Lock.
	LockFunc *DBStoreLockFunc
	// OldestDumpForRepositoryFunc is an instance of a mock function object
	// controlling the behavior of the method OldestDumpForRepository.
	OldestDumpForRepositoryFunc *DBStoreOldestDumpForRepositoryFunc
}

// NewMockDBStore creates a new mock of the DBStore interface. All methods
// return zero values for all results, unless overwritten.
func NewMockDBStore() *MockDBStore {
	return &MockDBStore{
		CalculateVisibleUploadsFunc: &DBStoreCalculateVisibleUploadsFunc{
			defaultHook: func(context.Context, int, *gitserver.CommitGraph, string, int) error {
				return nil
			},
		},
		DirtyRepositoriesFunc: &DBStoreDirtyRepositoriesFunc{
			defaultHook: func(context.Context) (map[int]int, error) {
				return nil, nil
			},
		},
		GetUploadsFunc: &DBStoreGetUploadsFunc{
			defaultHook: func(context.Context, dbstore.GetUploadsOptions) ([]dbstore.Upload, int, error) {
				return nil, 0, nil
			},
		},
		LockFunc: &DBStoreLockFunc{
			defaultHook: func(context.Context, int, bool) (bool, dbstore.UnlockFunc, error) {
				return false, nil, nil
			},
		},
		OldestDumpForRepositoryFunc: &DBStoreOldestDumpForRepositoryFunc{
			defaultHook: func(context.Context, int) (dbstore.Dump, bool, error) {
				return dbstore.Dump{}, false, nil
			},
		},
	}
}

// NewMockDBStoreFrom creates a new mock of the MockDBStore interface. All
// methods delegate to the given implementation, unless overwritten.
func NewMockDBStoreFrom(i DBStore) *MockDBStore {
	return &MockDBStore{
		CalculateVisibleUploadsFunc: &DBStoreCalculateVisibleUploadsFunc{
			defaultHook: i.CalculateVisibleUploads,
		},
		DirtyRepositoriesFunc: &DBStoreDirtyRepositoriesFunc{
			defaultHook: i.DirtyRepositories,
		},
		GetUploadsFunc: &DBStoreGetUploadsFunc{
			defaultHook: i.GetUploads,
		},
		LockFunc: &DBStoreLockFunc{
			defaultHook: i.Lock,
		},
		OldestDumpForRepositoryFunc: &DBStoreOldestDumpForRepositoryFunc{
			defaultHook: i.OldestDumpForRepository,
		},
	}
}

// DBStoreCalculateVisibleUploadsFunc describes the behavior when the
// CalculateVisibleUploads method of the parent MockDBStore instance is
// invoked.
type DBStoreCalculateVisibleUploadsFunc struct {
	defaultHook func(context.Context, int, *gitserver.CommitGraph, string, int) error
	hooks       []func(context.Context, int, *gitserver.CommitGraph, string, int) error
	history     []DBStoreCalculateVisibleUploadsFuncCall
	mutex       sync.Mutex
}

// CalculateVisibleUploads delegates to the next hook function in the queue
// and stores the parameter and result values of this invocation.
func (m *MockDBStore) CalculateVisibleUploads(v0 context.Context, v1 int, v2 *gitserver.CommitGraph, v3 string, v4 int) error {
	r0 := m.CalculateVisibleUploadsFunc.nextHook()(v0, v1, v2, v3, v4)
	m.CalculateVisibleUploadsFunc.appendCall(DBStoreCalculateVisibleUploadsFuncCall{v0, v1, v2, v3, v4, r0})
	return r0
}

// SetDefaultHook sets function that is called when the
// CalculateVisibleUploads method of the parent MockDBStore instance is
// invoked and the hook queue is empty.
func (f *DBStoreCalculateVisibleUploadsFunc) SetDefaultHook(hook func(context.Context, int, *gitserver.CommitGraph, string, int) error) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// CalculateVisibleUploads method of the parent MockDBStore instance inovkes
// the hook at the front of the queue and discards it. After the queue is
// empty, the default hook function is invoked for any future action.
func (f *DBStoreCalculateVisibleUploadsFunc) PushHook(hook func(context.Context, int, *gitserver.CommitGraph, string, int) error) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultDefaultHook with a function that returns
// the given values.
func (f *DBStoreCalculateVisibleUploadsFunc) SetDefaultReturn(r0 error) {
	f.SetDefaultHook(func(context.Context, int, *gitserver.CommitGraph, string, int) error {
		return r0
	})
}

// PushReturn calls PushDefaultHook with a function that returns the given
// values.
func (f *DBStoreCalculateVisibleUploadsFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, int, *gitserver.CommitGraph, string, int) error {
		return r0
	})
}

func (f *DBStoreCalculateVisibleUploadsFunc) nextHook() func(context.Context, int, *gitserver.CommitGraph, string, int) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *DBStoreCalculateVisibleUploadsFunc) appendCall(r0 DBStoreCalculateVisibleUploadsFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of DBStoreCalculateVisibleUploadsFuncCall
// objects describing the invocations of this function.
func (f *DBStoreCalculateVisibleUploadsFunc) History() []DBStoreCalculateVisibleUploadsFuncCall {
	f.mutex.Lock()
	history := make([]DBStoreCalculateVisibleUploadsFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// DBStoreCalculateVisibleUploadsFuncCall is an object that describes an
// invocation of method CalculateVisibleUploads on an instance of
// MockDBStore.
type DBStoreCalculateVisibleUploadsFuncCall struct {
	// Arg0 is the value of the 1st argument passed to this method
	// invocation.
	Arg0 context.Context
	// Arg1 is the value of the 2nd argument passed to this method
	// invocation.
	Arg1 int
	// Arg2 is the value of the 3rd argument passed to this method
	// invocation.
	Arg2 *gitserver.CommitGraph
	// Arg3 is the value of the 4th argument passed to this method
	// invocation.
	Arg3 string
	// Arg4 is the value of the 5th argument passed to this method
	// invocation.
	Arg4 int
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 error
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c DBStoreCalculateVisibleUploadsFuncCall) Args() []interface{} {
	return []interface{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3, c.Arg4}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c DBStoreCalculateVisibleUploadsFuncCall) Results() []interface{} {
	return []interface{}{c.Result0}
}

// DBStoreDirtyRepositoriesFunc describes the behavior when the
// DirtyRepositories method of the parent MockDBStore instance is invoked.
type DBStoreDirtyRepositoriesFunc struct {
	defaultHook func(context.Context) (map[int]int, error)
	hooks       []func(context.Context) (map[int]int, error)
	history     []DBStoreDirtyRepositoriesFuncCall
	mutex       sync.Mutex
}

// DirtyRepositories delegates to the next hook function in the queue and
// stores the parameter and result values of this invocation.
func (m *MockDBStore) DirtyRepositories(v0 context.Context) (map[int]int, error) {
	r0, r1 := m.DirtyRepositoriesFunc.nextHook()(v0)
	m.DirtyRepositoriesFunc.appendCall(DBStoreDirtyRepositoriesFuncCall{v0, r0, r1})
	return r0, r1
}

// SetDefaultHook sets function that is called when the DirtyRepositories
// method of the parent MockDBStore instance is invoked and the hook queue
// is empty.
func (f *DBStoreDirtyRepositoriesFunc) SetDefaultHook(hook func(context.Context) (map[int]int, error)) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// DirtyRepositories method of the parent MockDBStore instance inovkes the
// hook at the front of the queue and discards it. After the queue is empty,
// the default hook function is invoked for any future action.
func (f *DBStoreDirtyRepositoriesFunc) PushHook(hook func(context.Context) (map[int]int, error)) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultDefaultHook with a function that returns
// the given values.
func (f *DBStoreDirtyRepositoriesFunc) SetDefaultReturn(r0 map[int]int, r1 error) {
	f.SetDefaultHook(func(context.Context) (map[int]int, error) {
		return r0, r1
	})
}

// PushReturn calls PushDefaultHook with a function that returns the given
// values.
func (f *DBStoreDirtyRepositoriesFunc) PushReturn(r0 map[int]int, r1 error) {
	f.PushHook(func(context.Context) (map[int]int, error) {
		return r0, r1
	})
}

func (f *DBStoreDirtyRepositoriesFunc) nextHook() func(context.Context) (map[int]int, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *DBStoreDirtyRepositoriesFunc) appendCall(r0 DBStoreDirtyRepositoriesFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of DBStoreDirtyRepositoriesFuncCall objects
// describing the invocations of this function.
func (f *DBStoreDirtyRepositoriesFunc) History() []DBStoreDirtyRepositoriesFuncCall {
	f.mutex.Lock()
	history := make([]DBStoreDirtyRepositoriesFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// DBStoreDirtyRepositoriesFuncCall is an object that describes an
// invocation of method DirtyRepositories on an instance of MockDBStore.
type DBStoreDirtyRepositoriesFuncCall struct {
	// Arg0 is the value of the 1st argument passed to this method
	// invocation.
	Arg0 context.Context
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 map[int]int
	// Result1 is the value of the 2nd result returned from this method
	// invocation.
	Result1 error
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c DBStoreDirtyRepositoriesFuncCall) Args() []interface{} {
	return []interface{}{c.Arg0}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c DBStoreDirtyRepositoriesFuncCall) Results() []interface{} {
	return []interface{}{c.Result0, c.Result1}
}

// DBStoreGetUploadsFunc describes the behavior when the GetUploads method
// of the parent MockDBStore instance is invoked.
type DBStoreGetUploadsFunc struct {
	defaultHook func(context.Context, dbstore.GetUploadsOptions) ([]dbstore.Upload, int, error)
	hooks       []func(context.Context, dbstore.GetUploadsOptions) ([]dbstore.Upload, int, error)
	history     []DBStoreGetUploadsFuncCall
	mutex       sync.Mutex
}

// GetUploads delegates to the next hook function in the queue and stores
// the parameter and result values of this invocation.
func (m *MockDBStore) GetUploads(v0 context.Context, v1 dbstore.GetUploadsOptions) ([]dbstore.Upload, int, error) {
	r0, r1, r2 := m.GetUploadsFunc.nextHook()(v0, v1)
	m.GetUploadsFunc.appendCall(DBStoreGetUploadsFuncCall{v0, v1, r0, r1, r2})
	return r0, r1, r2
}

// SetDefaultHook sets function that is called when the GetUploads method of
// the parent MockDBStore instance is invoked and the hook queue is empty.
func (f *DBStoreGetUploadsFunc) SetDefaultHook(hook func(context.Context, dbstore.GetUploadsOptions) ([]dbstore.Upload, int, error)) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// GetUploads method of the parent MockDBStore instance inovkes the hook at
// the front of the queue and discards it. After the queue is empty, the
// default hook function is invoked for any future action.
func (f *DBStoreGetUploadsFunc) PushHook(hook func(context.Context, dbstore.GetUploadsOptions) ([]dbstore.Upload, int, error)) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultDefaultHook with a function that returns
// the given values.
func (f *DBStoreGetUploadsFunc) SetDefaultReturn(r0 []dbstore.Upload, r1 int, r2 error) {
	f.SetDefaultHook(func(context.Context, dbstore.GetUploadsOptions) ([]dbstore.Upload, int, error) {
		return r0, r1, r2
	})
}

// PushReturn calls PushDefaultHook with a function that returns the given
// values.
func (f *DBStoreGetUploadsFunc) PushReturn(r0 []dbstore.Upload, r1 int, r2 error) {
	f.PushHook(func(context.Context, dbstore.GetUploadsOptions) ([]dbstore.Upload, int, error) {
		return r0, r1, r2
	})
}

func (f *DBStoreGetUploadsFunc) nextHook() func(context.Context, dbstore.GetUploadsOptions) ([]dbstore.Upload, int, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *DBStoreGetUploadsFunc) appendCall(r0 DBStoreGetUploadsFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of DBStoreGetUploadsFuncCall objects
// describing the invocations of this function.
func (f *DBStoreGetUploadsFunc) History() []DBStoreGetUploadsFuncCall {
	f.mutex.Lock()
	history := make([]DBStoreGetUploadsFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// DBStoreGetUploadsFuncCall is an object that describes an invocation of
// method GetUploads on an instance of MockDBStore.
type DBStoreGetUploadsFuncCall struct {
	// Arg0 is the value of the 1st argument passed to this method
	// invocation.
	Arg0 context.Context
	// Arg1 is the value of the 2nd argument passed to this method
	// invocation.
	Arg1 dbstore.GetUploadsOptions
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 []dbstore.Upload
	// Result1 is the value of the 2nd result returned from this method
	// invocation.
	Result1 int
	// Result2 is the value of the 3rd result returned from this method
	// invocation.
	Result2 error
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c DBStoreGetUploadsFuncCall) Args() []interface{} {
	return []interface{}{c.Arg0, c.Arg1}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c DBStoreGetUploadsFuncCall) Results() []interface{} {
	return []interface{}{c.Result0, c.Result1, c.Result2}
}

// DBStoreLockFunc describes the behavior when the Lock method of the parent
// MockDBStore instance is invoked.
type DBStoreLockFunc struct {
	defaultHook func(context.Context, int, bool) (bool, dbstore.UnlockFunc, error)
	hooks       []func(context.Context, int, bool) (bool, dbstore.UnlockFunc, error)
	history     []DBStoreLockFuncCall
	mutex       sync.Mutex
}

// Lock delegates to the next hook function in the queue and stores the
// parameter and result values of this invocation.
func (m *MockDBStore) Lock(v0 context.Context, v1 int, v2 bool) (bool, dbstore.UnlockFunc, error) {
	r0, r1, r2 := m.LockFunc.nextHook()(v0, v1, v2)
	m.LockFunc.appendCall(DBStoreLockFuncCall{v0, v1, v2, r0, r1, r2})
	return r0, r1, r2
}

// SetDefaultHook sets function that is called when the Lock method of the
// parent MockDBStore instance is invoked and the hook queue is empty.
func (f *DBStoreLockFunc) SetDefaultHook(hook func(context.Context, int, bool) (bool, dbstore.UnlockFunc, error)) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// Lock method of the parent MockDBStore instance inovkes the hook at the
// front of the queue and discards it. After the queue is empty, the default
// hook function is invoked for any future action.
func (f *DBStoreLockFunc) PushHook(hook func(context.Context, int, bool) (bool, dbstore.UnlockFunc, error)) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultDefaultHook with a function that returns
// the given values.
func (f *DBStoreLockFunc) SetDefaultReturn(r0 bool, r1 dbstore.UnlockFunc, r2 error) {
	f.SetDefaultHook(func(context.Context, int, bool) (bool, dbstore.UnlockFunc, error) {
		return r0, r1, r2
	})
}

// PushReturn calls PushDefaultHook with a function that returns the given
// values.
func (f *DBStoreLockFunc) PushReturn(r0 bool, r1 dbstore.UnlockFunc, r2 error) {
	f.PushHook(func(context.Context, int, bool) (bool, dbstore.UnlockFunc, error) {
		return r0, r1, r2
	})
}

func (f *DBStoreLockFunc) nextHook() func(context.Context, int, bool) (bool, dbstore.UnlockFunc, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *DBStoreLockFunc) appendCall(r0 DBStoreLockFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of DBStoreLockFuncCall objects describing the
// invocations of this function.
func (f *DBStoreLockFunc) History() []DBStoreLockFuncCall {
	f.mutex.Lock()
	history := make([]DBStoreLockFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// DBStoreLockFuncCall is an object that describes an invocation of method
// Lock on an instance of MockDBStore.
type DBStoreLockFuncCall struct {
	// Arg0 is the value of the 1st argument passed to this method
	// invocation.
	Arg0 context.Context
	// Arg1 is the value of the 2nd argument passed to this method
	// invocation.
	Arg1 int
	// Arg2 is the value of the 3rd argument passed to this method
	// invocation.
	Arg2 bool
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 bool
	// Result1 is the value of the 2nd result returned from this method
	// invocation.
	Result1 dbstore.UnlockFunc
	// Result2 is the value of the 3rd result returned from this method
	// invocation.
	Result2 error
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c DBStoreLockFuncCall) Args() []interface{} {
	return []interface{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c DBStoreLockFuncCall) Results() []interface{} {
	return []interface{}{c.Result0, c.Result1, c.Result2}
}

// DBStoreOldestDumpForRepositoryFunc describes the behavior when the
// OldestDumpForRepository method of the parent MockDBStore instance is
// invoked.
type DBStoreOldestDumpForRepositoryFunc struct {
	defaultHook func(context.Context, int) (dbstore.Dump, bool, error)
	hooks       []func(context.Context, int) (dbstore.Dump, bool, error)
	history     []DBStoreOldestDumpForRepositoryFuncCall
	mutex       sync.Mutex
}

// OldestDumpForRepository delegates to the next hook function in the queue
// and stores the parameter and result values of this invocation.
func (m *MockDBStore) OldestDumpForRepository(v0 context.Context, v1 int) (dbstore.Dump, bool, error) {
	r0, r1, r2 := m.OldestDumpForRepositoryFunc.nextHook()(v0, v1)
	m.OldestDumpForRepositoryFunc.appendCall(DBStoreOldestDumpForRepositoryFuncCall{v0, v1, r0, r1, r2})
	return r0, r1, r2
}

// SetDefaultHook sets function that is called when the
// OldestDumpForRepository method of the parent MockDBStore instance is
// invoked and the hook queue is empty.
func (f *DBStoreOldestDumpForRepositoryFunc) SetDefaultHook(hook func(context.Context, int) (dbstore.Dump, bool, error)) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// OldestDumpForRepository method of the parent MockDBStore instance inovkes
// the hook at the front of the queue and discards it. After the queue is
// empty, the default hook function is invoked for any future action.
func (f *DBStoreOldestDumpForRepositoryFunc) PushHook(hook func(context.Context, int) (dbstore.Dump, bool, error)) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultDefaultHook with a function that returns
// the given values.
func (f *DBStoreOldestDumpForRepositoryFunc) SetDefaultReturn(r0 dbstore.Dump, r1 bool, r2 error) {
	f.SetDefaultHook(func(context.Context, int) (dbstore.Dump, bool, error) {
		return r0, r1, r2
	})
}

// PushReturn calls PushDefaultHook with a function that returns the given
// values.
func (f *DBStoreOldestDumpForRepositoryFunc) PushReturn(r0 dbstore.Dump, r1 bool, r2 error) {
	f.PushHook(func(context.Context, int) (dbstore.Dump, bool, error) {
		return r0, r1, r2
	})
}

func (f *DBStoreOldestDumpForRepositoryFunc) nextHook() func(context.Context, int) (dbstore.Dump, bool, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *DBStoreOldestDumpForRepositoryFunc) appendCall(r0 DBStoreOldestDumpForRepositoryFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of DBStoreOldestDumpForRepositoryFuncCall
// objects describing the invocations of this function.
func (f *DBStoreOldestDumpForRepositoryFunc) History() []DBStoreOldestDumpForRepositoryFuncCall {
	f.mutex.Lock()
	history := make([]DBStoreOldestDumpForRepositoryFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// DBStoreOldestDumpForRepositoryFuncCall is an object that describes an
// invocation of method OldestDumpForRepository on an instance of
// MockDBStore.
type DBStoreOldestDumpForRepositoryFuncCall struct {
	// Arg0 is the value of the 1st argument passed to this method
	// invocation.
	Arg0 context.Context
	// Arg1 is the value of the 2nd argument passed to this method
	// invocation.
	Arg1 int
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 dbstore.Dump
	// Result1 is the value of the 2nd result returned from this method
	// invocation.
	Result1 bool
	// Result2 is the value of the 3rd result returned from this method
	// invocation.
	Result2 error
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c DBStoreOldestDumpForRepositoryFuncCall) Args() []interface{} {
	return []interface{}{c.Arg0, c.Arg1}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c DBStoreOldestDumpForRepositoryFuncCall) Results() []interface{} {
	return []interface{}{c.Result0, c.Result1, c.Result2}
}

// MockGitserverClient is a mock implementation of the GitserverClient
// interface (from the package
// github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/codeintel/background/commitgraph)
// used for unit testing.
type MockGitserverClient struct {
	// CommitDateFunc is an instance of a mock function object controlling
	// the behavior of the method CommitDate.
	CommitDateFunc *GitserverClientCommitDateFunc
	// CommitGraphFunc is an instance of a mock function object controlling
	// the behavior of the method CommitGraph.
	CommitGraphFunc *GitserverClientCommitGraphFunc
	// HeadFunc is an instance of a mock function object controlling the
	// behavior of the method Head.
	HeadFunc *GitserverClientHeadFunc
}

// NewMockGitserverClient creates a new mock of the GitserverClient
// interface. All methods return zero values for all results, unless
// overwritten.
func NewMockGitserverClient() *MockGitserverClient {
	return &MockGitserverClient{
		CommitDateFunc: &GitserverClientCommitDateFunc{
			defaultHook: func(context.Context, int, string) (time.Time, error) {
				return time.Time{}, nil
			},
		},
		CommitGraphFunc: &GitserverClientCommitGraphFunc{
			defaultHook: func(context.Context, int, gitserver.CommitGraphOptions) (*gitserver.CommitGraph, error) {
				return nil, nil
			},
		},
		HeadFunc: &GitserverClientHeadFunc{
			defaultHook: func(context.Context, int) (string, error) {
				return "", nil
			},
		},
	}
}

// NewMockGitserverClientFrom creates a new mock of the MockGitserverClient
// interface. All methods delegate to the given implementation, unless
// overwritten.
func NewMockGitserverClientFrom(i GitserverClient) *MockGitserverClient {
	return &MockGitserverClient{
		CommitDateFunc: &GitserverClientCommitDateFunc{
			defaultHook: i.CommitDate,
		},
		CommitGraphFunc: &GitserverClientCommitGraphFunc{
			defaultHook: i.CommitGraph,
		},
		HeadFunc: &GitserverClientHeadFunc{
			defaultHook: i.Head,
		},
	}
}

// GitserverClientCommitDateFunc describes the behavior when the CommitDate
// method of the parent MockGitserverClient instance is invoked.
type GitserverClientCommitDateFunc struct {
	defaultHook func(context.Context, int, string) (time.Time, error)
	hooks       []func(context.Context, int, string) (time.Time, error)
	history     []GitserverClientCommitDateFuncCall
	mutex       sync.Mutex
}

// CommitDate delegates to the next hook function in the queue and stores
// the parameter and result values of this invocation.
func (m *MockGitserverClient) CommitDate(v0 context.Context, v1 int, v2 string) (time.Time, error) {
	r0, r1 := m.CommitDateFunc.nextHook()(v0, v1, v2)
	m.CommitDateFunc.appendCall(GitserverClientCommitDateFuncCall{v0, v1, v2, r0, r1})
	return r0, r1
}

// SetDefaultHook sets function that is called when the CommitDate method of
// the parent MockGitserverClient instance is invoked and the hook queue is
// empty.
func (f *GitserverClientCommitDateFunc) SetDefaultHook(hook func(context.Context, int, string) (time.Time, error)) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// CommitDate method of the parent MockGitserverClient instance inovkes the
// hook at the front of the queue and discards it. After the queue is empty,
// the default hook function is invoked for any future action.
func (f *GitserverClientCommitDateFunc) PushHook(hook func(context.Context, int, string) (time.Time, error)) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultDefaultHook with a function that returns
// the given values.
func (f *GitserverClientCommitDateFunc) SetDefaultReturn(r0 time.Time, r1 error) {
	f.SetDefaultHook(func(context.Context, int, string) (time.Time, error) {
		return r0, r1
	})
}

// PushReturn calls PushDefaultHook with a function that returns the given
// values.
func (f *GitserverClientCommitDateFunc) PushReturn(r0 time.Time, r1 error) {
	f.PushHook(func(context.Context, int, string) (time.Time, error) {
		return r0, r1
	})
}

func (f *GitserverClientCommitDateFunc) nextHook() func(context.Context, int, string) (time.Time, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *GitserverClientCommitDateFunc) appendCall(r0 GitserverClientCommitDateFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of GitserverClientCommitDateFuncCall objects
// describing the invocations of this function.
func (f *GitserverClientCommitDateFunc) History() []GitserverClientCommitDateFuncCall {
	f.mutex.Lock()
	history := make([]GitserverClientCommitDateFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// GitserverClientCommitDateFuncCall is an object that describes an
// invocation of method CommitDate on an instance of MockGitserverClient.
type GitserverClientCommitDateFuncCall struct {
	// Arg0 is the value of the 1st argument passed to this method
	// invocation.
	Arg0 context.Context
	// Arg1 is the value of the 2nd argument passed to this method
	// invocation.
	Arg1 int
	// Arg2 is the value of the 3rd argument passed to this method
	// invocation.
	Arg2 string
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 time.Time
	// Result1 is the value of the 2nd result returned from this method
	// invocation.
	Result1 error
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c GitserverClientCommitDateFuncCall) Args() []interface{} {
	return []interface{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c GitserverClientCommitDateFuncCall) Results() []interface{} {
	return []interface{}{c.Result0, c.Result1}
}

// GitserverClientCommitGraphFunc describes the behavior when the
// CommitGraph method of the parent MockGitserverClient instance is invoked.
type GitserverClientCommitGraphFunc struct {
	defaultHook func(context.Context, int, gitserver.CommitGraphOptions) (*gitserver.CommitGraph, error)
	hooks       []func(context.Context, int, gitserver.CommitGraphOptions) (*gitserver.CommitGraph, error)
	history     []GitserverClientCommitGraphFuncCall
	mutex       sync.Mutex
}

// CommitGraph delegates to the next hook function in the queue and stores
// the parameter and result values of this invocation.
func (m *MockGitserverClient) CommitGraph(v0 context.Context, v1 int, v2 gitserver.CommitGraphOptions) (*gitserver.CommitGraph, error) {
	r0, r1 := m.CommitGraphFunc.nextHook()(v0, v1, v2)
	m.CommitGraphFunc.appendCall(GitserverClientCommitGraphFuncCall{v0, v1, v2, r0, r1})
	return r0, r1
}

// SetDefaultHook sets function that is called when the CommitGraph method
// of the parent MockGitserverClient instance is invoked and the hook queue
// is empty.
func (f *GitserverClientCommitGraphFunc) SetDefaultHook(hook func(context.Context, int, gitserver.CommitGraphOptions) (*gitserver.CommitGraph, error)) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// CommitGraph method of the parent MockGitserverClient instance inovkes the
// hook at the front of the queue and discards it. After the queue is empty,
// the default hook function is invoked for any future action.
func (f *GitserverClientCommitGraphFunc) PushHook(hook func(context.Context, int, gitserver.CommitGraphOptions) (*gitserver.CommitGraph, error)) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultDefaultHook with a function that returns
// the given values.
func (f *GitserverClientCommitGraphFunc) SetDefaultReturn(r0 *gitserver.CommitGraph, r1 error) {
	f.SetDefaultHook(func(context.Context, int, gitserver.CommitGraphOptions) (*gitserver.CommitGraph, error) {
		return r0, r1
	})
}

// PushReturn calls PushDefaultHook with a function that returns the given
// values.
func (f *GitserverClientCommitGraphFunc) PushReturn(r0 *gitserver.CommitGraph, r1 error) {
	f.PushHook(func(context.Context, int, gitserver.CommitGraphOptions) (*gitserver.CommitGraph, error) {
		return r0, r1
	})
}

func (f *GitserverClientCommitGraphFunc) nextHook() func(context.Context, int, gitserver.CommitGraphOptions) (*gitserver.CommitGraph, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *GitserverClientCommitGraphFunc) appendCall(r0 GitserverClientCommitGraphFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of GitserverClientCommitGraphFuncCall objects
// describing the invocations of this function.
func (f *GitserverClientCommitGraphFunc) History() []GitserverClientCommitGraphFuncCall {
	f.mutex.Lock()
	history := make([]GitserverClientCommitGraphFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// GitserverClientCommitGraphFuncCall is an object that describes an
// invocation of method CommitGraph on an instance of MockGitserverClient.
type GitserverClientCommitGraphFuncCall struct {
	// Arg0 is the value of the 1st argument passed to this method
	// invocation.
	Arg0 context.Context
	// Arg1 is the value of the 2nd argument passed to this method
	// invocation.
	Arg1 int
	// Arg2 is the value of the 3rd argument passed to this method
	// invocation.
	Arg2 gitserver.CommitGraphOptions
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 *gitserver.CommitGraph
	// Result1 is the value of the 2nd result returned from this method
	// invocation.
	Result1 error
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c GitserverClientCommitGraphFuncCall) Args() []interface{} {
	return []interface{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c GitserverClientCommitGraphFuncCall) Results() []interface{} {
	return []interface{}{c.Result0, c.Result1}
}

// GitserverClientHeadFunc describes the behavior when the Head method of
// the parent MockGitserverClient instance is invoked.
type GitserverClientHeadFunc struct {
	defaultHook func(context.Context, int) (string, error)
	hooks       []func(context.Context, int) (string, error)
	history     []GitserverClientHeadFuncCall
	mutex       sync.Mutex
}

// Head delegates to the next hook function in the queue and stores the
// parameter and result values of this invocation.
func (m *MockGitserverClient) Head(v0 context.Context, v1 int) (string, error) {
	r0, r1 := m.HeadFunc.nextHook()(v0, v1)
	m.HeadFunc.appendCall(GitserverClientHeadFuncCall{v0, v1, r0, r1})
	return r0, r1
}

// SetDefaultHook sets function that is called when the Head method of the
// parent MockGitserverClient instance is invoked and the hook queue is
// empty.
func (f *GitserverClientHeadFunc) SetDefaultHook(hook func(context.Context, int) (string, error)) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// Head method of the parent MockGitserverClient instance inovkes the hook
// at the front of the queue and discards it. After the queue is empty, the
// default hook function is invoked for any future action.
func (f *GitserverClientHeadFunc) PushHook(hook func(context.Context, int) (string, error)) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultDefaultHook with a function that returns
// the given values.
func (f *GitserverClientHeadFunc) SetDefaultReturn(r0 string, r1 error) {
	f.SetDefaultHook(func(context.Context, int) (string, error) {
		return r0, r1
	})
}

// PushReturn calls PushDefaultHook with a function that returns the given
// values.
func (f *GitserverClientHeadFunc) PushReturn(r0 string, r1 error) {
	f.PushHook(func(context.Context, int) (string, error) {
		return r0, r1
	})
}

func (f *GitserverClientHeadFunc) nextHook() func(context.Context, int) (string, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *GitserverClientHeadFunc) appendCall(r0 GitserverClientHeadFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of GitserverClientHeadFuncCall objects
// describing the invocations of this function.
func (f *GitserverClientHeadFunc) History() []GitserverClientHeadFuncCall {
	f.mutex.Lock()
	history := make([]GitserverClientHeadFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// GitserverClientHeadFuncCall is an object that describes an invocation of
// method Head on an instance of MockGitserverClient.
type GitserverClientHeadFuncCall struct {
	// Arg0 is the value of the 1st argument passed to this method
	// invocation.
	Arg0 context.Context
	// Arg1 is the value of the 2nd argument passed to this method
	// invocation.
	Arg1 int
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 string
	// Result1 is the value of the 2nd result returned from this method
	// invocation.
	Result1 error
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c GitserverClientHeadFuncCall) Args() []interface{} {
	return []interface{}{c.Arg0, c.Arg1}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c GitserverClientHeadFuncCall) Results() []interface{} {
	return []interface{}{c.Result0, c.Result1}
}