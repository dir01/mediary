package mocks

// Code generated by http://github.com/gojuno/minimock (dev). DO NOT EDIT.

import (
	"context"
	"sync"
	mm_atomic "sync/atomic"
	mm_time "time"

	mm_service "github.com/dir01/mediary/service"
	"github.com/gojuno/minimock/v3"
)

// StorageMock implements service.Storage
type StorageMock struct {
	t minimock.Tester

	funcGetJob          func(ctx context.Context, id string) (jp1 *mm_service.Job, err error)
	inspectFuncGetJob   func(ctx context.Context, id string)
	afterGetJobCounter  uint64
	beforeGetJobCounter uint64
	GetJobMock          mStorageMockGetJob

	funcGetMetadata          func(ctx context.Context, url string) (mp1 *mm_service.Metadata, err error)
	inspectFuncGetMetadata   func(ctx context.Context, url string)
	afterGetMetadataCounter  uint64
	beforeGetMetadataCounter uint64
	GetMetadataMock          mStorageMockGetMetadata

	funcSaveJob          func(ctx context.Context, job *mm_service.Job) (err error)
	inspectFuncSaveJob   func(ctx context.Context, job *mm_service.Job)
	afterSaveJobCounter  uint64
	beforeSaveJobCounter uint64
	SaveJobMock          mStorageMockSaveJob

	funcSaveMetadata          func(ctx context.Context, metadata *mm_service.Metadata) (err error)
	inspectFuncSaveMetadata   func(ctx context.Context, metadata *mm_service.Metadata)
	afterSaveMetadataCounter  uint64
	beforeSaveMetadataCounter uint64
	SaveMetadataMock          mStorageMockSaveMetadata
}

// NewStorageMock returns a mock for service.Storage
func NewStorageMock(t minimock.Tester) *StorageMock {
	m := &StorageMock{t: t}
	if controller, ok := t.(minimock.MockController); ok {
		controller.RegisterMocker(m)
	}

	m.GetJobMock = mStorageMockGetJob{mock: m}
	m.GetJobMock.callArgs = []*StorageMockGetJobParams{}

	m.GetMetadataMock = mStorageMockGetMetadata{mock: m}
	m.GetMetadataMock.callArgs = []*StorageMockGetMetadataParams{}

	m.SaveJobMock = mStorageMockSaveJob{mock: m}
	m.SaveJobMock.callArgs = []*StorageMockSaveJobParams{}

	m.SaveMetadataMock = mStorageMockSaveMetadata{mock: m}
	m.SaveMetadataMock.callArgs = []*StorageMockSaveMetadataParams{}

	return m
}

type mStorageMockGetJob struct {
	mock               *StorageMock
	defaultExpectation *StorageMockGetJobExpectation
	expectations       []*StorageMockGetJobExpectation

	callArgs []*StorageMockGetJobParams
	mutex    sync.RWMutex
}

// StorageMockGetJobExpectation specifies expectation struct of the Storage.GetJob
type StorageMockGetJobExpectation struct {
	mock    *StorageMock
	params  *StorageMockGetJobParams
	results *StorageMockGetJobResults
	Counter uint64
}

// StorageMockGetJobParams contains parameters of the Storage.GetJob
type StorageMockGetJobParams struct {
	ctx context.Context
	id  string
}

// StorageMockGetJobResults contains results of the Storage.GetJob
type StorageMockGetJobResults struct {
	jp1 *mm_service.Job
	err error
}

// Expect sets up expected params for Storage.GetJob
func (mmGetJob *mStorageMockGetJob) Expect(ctx context.Context, id string) *mStorageMockGetJob {
	if mmGetJob.mock.funcGetJob != nil {
		mmGetJob.mock.t.Fatalf("StorageMock.GetJob mock is already set by Set")
	}

	if mmGetJob.defaultExpectation == nil {
		mmGetJob.defaultExpectation = &StorageMockGetJobExpectation{}
	}

	mmGetJob.defaultExpectation.params = &StorageMockGetJobParams{ctx, id}
	for _, e := range mmGetJob.expectations {
		if minimock.Equal(e.params, mmGetJob.defaultExpectation.params) {
			mmGetJob.mock.t.Fatalf("Expectation set by When has same params: %#v", *mmGetJob.defaultExpectation.params)
		}
	}

	return mmGetJob
}

// Inspect accepts an inspector function that has same arguments as the Storage.GetJob
func (mmGetJob *mStorageMockGetJob) Inspect(f func(ctx context.Context, id string)) *mStorageMockGetJob {
	if mmGetJob.mock.inspectFuncGetJob != nil {
		mmGetJob.mock.t.Fatalf("Inspect function is already set for StorageMock.GetJob")
	}

	mmGetJob.mock.inspectFuncGetJob = f

	return mmGetJob
}

// Return sets up results that will be returned by Storage.GetJob
func (mmGetJob *mStorageMockGetJob) Return(jp1 *mm_service.Job, err error) *StorageMock {
	if mmGetJob.mock.funcGetJob != nil {
		mmGetJob.mock.t.Fatalf("StorageMock.GetJob mock is already set by Set")
	}

	if mmGetJob.defaultExpectation == nil {
		mmGetJob.defaultExpectation = &StorageMockGetJobExpectation{mock: mmGetJob.mock}
	}
	mmGetJob.defaultExpectation.results = &StorageMockGetJobResults{jp1, err}
	return mmGetJob.mock
}

//Set uses given function f to mock the Storage.GetJob method
func (mmGetJob *mStorageMockGetJob) Set(f func(ctx context.Context, id string) (jp1 *mm_service.Job, err error)) *StorageMock {
	if mmGetJob.defaultExpectation != nil {
		mmGetJob.mock.t.Fatalf("Default expectation is already set for the Storage.GetJob method")
	}

	if len(mmGetJob.expectations) > 0 {
		mmGetJob.mock.t.Fatalf("Some expectations are already set for the Storage.GetJob method")
	}

	mmGetJob.mock.funcGetJob = f
	return mmGetJob.mock
}

// When sets expectation for the Storage.GetJob which will trigger the result defined by the following
// Then helper
func (mmGetJob *mStorageMockGetJob) When(ctx context.Context, id string) *StorageMockGetJobExpectation {
	if mmGetJob.mock.funcGetJob != nil {
		mmGetJob.mock.t.Fatalf("StorageMock.GetJob mock is already set by Set")
	}

	expectation := &StorageMockGetJobExpectation{
		mock:   mmGetJob.mock,
		params: &StorageMockGetJobParams{ctx, id},
	}
	mmGetJob.expectations = append(mmGetJob.expectations, expectation)
	return expectation
}

// Then sets up Storage.GetJob return parameters for the expectation previously defined by the When method
func (e *StorageMockGetJobExpectation) Then(jp1 *mm_service.Job, err error) *StorageMock {
	e.results = &StorageMockGetJobResults{jp1, err}
	return e.mock
}

// GetJob implements service.Storage
func (mmGetJob *StorageMock) GetJob(ctx context.Context, id string) (jp1 *mm_service.Job, err error) {
	mm_atomic.AddUint64(&mmGetJob.beforeGetJobCounter, 1)
	defer mm_atomic.AddUint64(&mmGetJob.afterGetJobCounter, 1)

	if mmGetJob.inspectFuncGetJob != nil {
		mmGetJob.inspectFuncGetJob(ctx, id)
	}

	mm_params := &StorageMockGetJobParams{ctx, id}

	// Record call args
	mmGetJob.GetJobMock.mutex.Lock()
	mmGetJob.GetJobMock.callArgs = append(mmGetJob.GetJobMock.callArgs, mm_params)
	mmGetJob.GetJobMock.mutex.Unlock()

	for _, e := range mmGetJob.GetJobMock.expectations {
		if minimock.Equal(e.params, mm_params) {
			mm_atomic.AddUint64(&e.Counter, 1)
			return e.results.jp1, e.results.err
		}
	}

	if mmGetJob.GetJobMock.defaultExpectation != nil {
		mm_atomic.AddUint64(&mmGetJob.GetJobMock.defaultExpectation.Counter, 1)
		mm_want := mmGetJob.GetJobMock.defaultExpectation.params
		mm_got := StorageMockGetJobParams{ctx, id}
		if mm_want != nil && !minimock.Equal(*mm_want, mm_got) {
			mmGetJob.t.Errorf("StorageMock.GetJob got unexpected parameters, want: %#v, got: %#v%s\n", *mm_want, mm_got, minimock.Diff(*mm_want, mm_got))
		}

		mm_results := mmGetJob.GetJobMock.defaultExpectation.results
		if mm_results == nil {
			mmGetJob.t.Fatal("No results are set for the StorageMock.GetJob")
		}
		return (*mm_results).jp1, (*mm_results).err
	}
	if mmGetJob.funcGetJob != nil {
		return mmGetJob.funcGetJob(ctx, id)
	}
	mmGetJob.t.Fatalf("Unexpected call to StorageMock.GetJob. %v %v", ctx, id)
	return
}

// GetJobAfterCounter returns a count of finished StorageMock.GetJob invocations
func (mmGetJob *StorageMock) GetJobAfterCounter() uint64 {
	return mm_atomic.LoadUint64(&mmGetJob.afterGetJobCounter)
}

// GetJobBeforeCounter returns a count of StorageMock.GetJob invocations
func (mmGetJob *StorageMock) GetJobBeforeCounter() uint64 {
	return mm_atomic.LoadUint64(&mmGetJob.beforeGetJobCounter)
}

// Calls returns a list of arguments used in each call to StorageMock.GetJob.
// The list is in the same order as the calls were made (i.e. recent calls have a higher index)
func (mmGetJob *mStorageMockGetJob) Calls() []*StorageMockGetJobParams {
	mmGetJob.mutex.RLock()

	argCopy := make([]*StorageMockGetJobParams, len(mmGetJob.callArgs))
	copy(argCopy, mmGetJob.callArgs)

	mmGetJob.mutex.RUnlock()

	return argCopy
}

// MinimockGetJobDone returns true if the count of the GetJob invocations corresponds
// the number of defined expectations
func (m *StorageMock) MinimockGetJobDone() bool {
	for _, e := range m.GetJobMock.expectations {
		if mm_atomic.LoadUint64(&e.Counter) < 1 {
			return false
		}
	}

	// if default expectation was set then invocations count should be greater than zero
	if m.GetJobMock.defaultExpectation != nil && mm_atomic.LoadUint64(&m.afterGetJobCounter) < 1 {
		return false
	}
	// if func was set then invocations count should be greater than zero
	if m.funcGetJob != nil && mm_atomic.LoadUint64(&m.afterGetJobCounter) < 1 {
		return false
	}
	return true
}

// MinimockGetJobInspect logs each unmet expectation
func (m *StorageMock) MinimockGetJobInspect() {
	for _, e := range m.GetJobMock.expectations {
		if mm_atomic.LoadUint64(&e.Counter) < 1 {
			m.t.Errorf("Expected call to StorageMock.GetJob with params: %#v", *e.params)
		}
	}

	// if default expectation was set then invocations count should be greater than zero
	if m.GetJobMock.defaultExpectation != nil && mm_atomic.LoadUint64(&m.afterGetJobCounter) < 1 {
		if m.GetJobMock.defaultExpectation.params == nil {
			m.t.Error("Expected call to StorageMock.GetJob")
		} else {
			m.t.Errorf("Expected call to StorageMock.GetJob with params: %#v", *m.GetJobMock.defaultExpectation.params)
		}
	}
	// if func was set then invocations count should be greater than zero
	if m.funcGetJob != nil && mm_atomic.LoadUint64(&m.afterGetJobCounter) < 1 {
		m.t.Error("Expected call to StorageMock.GetJob")
	}
}

type mStorageMockGetMetadata struct {
	mock               *StorageMock
	defaultExpectation *StorageMockGetMetadataExpectation
	expectations       []*StorageMockGetMetadataExpectation

	callArgs []*StorageMockGetMetadataParams
	mutex    sync.RWMutex
}

// StorageMockGetMetadataExpectation specifies expectation struct of the Storage.GetMetadata
type StorageMockGetMetadataExpectation struct {
	mock    *StorageMock
	params  *StorageMockGetMetadataParams
	results *StorageMockGetMetadataResults
	Counter uint64
}

// StorageMockGetMetadataParams contains parameters of the Storage.GetMetadata
type StorageMockGetMetadataParams struct {
	ctx context.Context
	url string
}

// StorageMockGetMetadataResults contains results of the Storage.GetMetadata
type StorageMockGetMetadataResults struct {
	mp1 *mm_service.Metadata
	err error
}

// Expect sets up expected params for Storage.GetMetadata
func (mmGetMetadata *mStorageMockGetMetadata) Expect(ctx context.Context, url string) *mStorageMockGetMetadata {
	if mmGetMetadata.mock.funcGetMetadata != nil {
		mmGetMetadata.mock.t.Fatalf("StorageMock.GetMetadata mock is already set by Set")
	}

	if mmGetMetadata.defaultExpectation == nil {
		mmGetMetadata.defaultExpectation = &StorageMockGetMetadataExpectation{}
	}

	mmGetMetadata.defaultExpectation.params = &StorageMockGetMetadataParams{ctx, url}
	for _, e := range mmGetMetadata.expectations {
		if minimock.Equal(e.params, mmGetMetadata.defaultExpectation.params) {
			mmGetMetadata.mock.t.Fatalf("Expectation set by When has same params: %#v", *mmGetMetadata.defaultExpectation.params)
		}
	}

	return mmGetMetadata
}

// Inspect accepts an inspector function that has same arguments as the Storage.GetMetadata
func (mmGetMetadata *mStorageMockGetMetadata) Inspect(f func(ctx context.Context, url string)) *mStorageMockGetMetadata {
	if mmGetMetadata.mock.inspectFuncGetMetadata != nil {
		mmGetMetadata.mock.t.Fatalf("Inspect function is already set for StorageMock.GetMetadata")
	}

	mmGetMetadata.mock.inspectFuncGetMetadata = f

	return mmGetMetadata
}

// Return sets up results that will be returned by Storage.GetMetadata
func (mmGetMetadata *mStorageMockGetMetadata) Return(mp1 *mm_service.Metadata, err error) *StorageMock {
	if mmGetMetadata.mock.funcGetMetadata != nil {
		mmGetMetadata.mock.t.Fatalf("StorageMock.GetMetadata mock is already set by Set")
	}

	if mmGetMetadata.defaultExpectation == nil {
		mmGetMetadata.defaultExpectation = &StorageMockGetMetadataExpectation{mock: mmGetMetadata.mock}
	}
	mmGetMetadata.defaultExpectation.results = &StorageMockGetMetadataResults{mp1, err}
	return mmGetMetadata.mock
}

//Set uses given function f to mock the Storage.GetMetadata method
func (mmGetMetadata *mStorageMockGetMetadata) Set(f func(ctx context.Context, url string) (mp1 *mm_service.Metadata, err error)) *StorageMock {
	if mmGetMetadata.defaultExpectation != nil {
		mmGetMetadata.mock.t.Fatalf("Default expectation is already set for the Storage.GetMetadata method")
	}

	if len(mmGetMetadata.expectations) > 0 {
		mmGetMetadata.mock.t.Fatalf("Some expectations are already set for the Storage.GetMetadata method")
	}

	mmGetMetadata.mock.funcGetMetadata = f
	return mmGetMetadata.mock
}

// When sets expectation for the Storage.GetMetadata which will trigger the result defined by the following
// Then helper
func (mmGetMetadata *mStorageMockGetMetadata) When(ctx context.Context, url string) *StorageMockGetMetadataExpectation {
	if mmGetMetadata.mock.funcGetMetadata != nil {
		mmGetMetadata.mock.t.Fatalf("StorageMock.GetMetadata mock is already set by Set")
	}

	expectation := &StorageMockGetMetadataExpectation{
		mock:   mmGetMetadata.mock,
		params: &StorageMockGetMetadataParams{ctx, url},
	}
	mmGetMetadata.expectations = append(mmGetMetadata.expectations, expectation)
	return expectation
}

// Then sets up Storage.GetMetadata return parameters for the expectation previously defined by the When method
func (e *StorageMockGetMetadataExpectation) Then(mp1 *mm_service.Metadata, err error) *StorageMock {
	e.results = &StorageMockGetMetadataResults{mp1, err}
	return e.mock
}

// GetMetadata implements service.Storage
func (mmGetMetadata *StorageMock) GetMetadata(ctx context.Context, url string) (mp1 *mm_service.Metadata, err error) {
	mm_atomic.AddUint64(&mmGetMetadata.beforeGetMetadataCounter, 1)
	defer mm_atomic.AddUint64(&mmGetMetadata.afterGetMetadataCounter, 1)

	if mmGetMetadata.inspectFuncGetMetadata != nil {
		mmGetMetadata.inspectFuncGetMetadata(ctx, url)
	}

	mm_params := &StorageMockGetMetadataParams{ctx, url}

	// Record call args
	mmGetMetadata.GetMetadataMock.mutex.Lock()
	mmGetMetadata.GetMetadataMock.callArgs = append(mmGetMetadata.GetMetadataMock.callArgs, mm_params)
	mmGetMetadata.GetMetadataMock.mutex.Unlock()

	for _, e := range mmGetMetadata.GetMetadataMock.expectations {
		if minimock.Equal(e.params, mm_params) {
			mm_atomic.AddUint64(&e.Counter, 1)
			return e.results.mp1, e.results.err
		}
	}

	if mmGetMetadata.GetMetadataMock.defaultExpectation != nil {
		mm_atomic.AddUint64(&mmGetMetadata.GetMetadataMock.defaultExpectation.Counter, 1)
		mm_want := mmGetMetadata.GetMetadataMock.defaultExpectation.params
		mm_got := StorageMockGetMetadataParams{ctx, url}
		if mm_want != nil && !minimock.Equal(*mm_want, mm_got) {
			mmGetMetadata.t.Errorf("StorageMock.GetMetadata got unexpected parameters, want: %#v, got: %#v%s\n", *mm_want, mm_got, minimock.Diff(*mm_want, mm_got))
		}

		mm_results := mmGetMetadata.GetMetadataMock.defaultExpectation.results
		if mm_results == nil {
			mmGetMetadata.t.Fatal("No results are set for the StorageMock.GetMetadata")
		}
		return (*mm_results).mp1, (*mm_results).err
	}
	if mmGetMetadata.funcGetMetadata != nil {
		return mmGetMetadata.funcGetMetadata(ctx, url)
	}
	mmGetMetadata.t.Fatalf("Unexpected call to StorageMock.GetMetadata. %v %v", ctx, url)
	return
}

// GetMetadataAfterCounter returns a count of finished StorageMock.GetMetadata invocations
func (mmGetMetadata *StorageMock) GetMetadataAfterCounter() uint64 {
	return mm_atomic.LoadUint64(&mmGetMetadata.afterGetMetadataCounter)
}

// GetMetadataBeforeCounter returns a count of StorageMock.GetMetadata invocations
func (mmGetMetadata *StorageMock) GetMetadataBeforeCounter() uint64 {
	return mm_atomic.LoadUint64(&mmGetMetadata.beforeGetMetadataCounter)
}

// Calls returns a list of arguments used in each call to StorageMock.GetMetadata.
// The list is in the same order as the calls were made (i.e. recent calls have a higher index)
func (mmGetMetadata *mStorageMockGetMetadata) Calls() []*StorageMockGetMetadataParams {
	mmGetMetadata.mutex.RLock()

	argCopy := make([]*StorageMockGetMetadataParams, len(mmGetMetadata.callArgs))
	copy(argCopy, mmGetMetadata.callArgs)

	mmGetMetadata.mutex.RUnlock()

	return argCopy
}

// MinimockGetMetadataDone returns true if the count of the GetMetadata invocations corresponds
// the number of defined expectations
func (m *StorageMock) MinimockGetMetadataDone() bool {
	for _, e := range m.GetMetadataMock.expectations {
		if mm_atomic.LoadUint64(&e.Counter) < 1 {
			return false
		}
	}

	// if default expectation was set then invocations count should be greater than zero
	if m.GetMetadataMock.defaultExpectation != nil && mm_atomic.LoadUint64(&m.afterGetMetadataCounter) < 1 {
		return false
	}
	// if func was set then invocations count should be greater than zero
	if m.funcGetMetadata != nil && mm_atomic.LoadUint64(&m.afterGetMetadataCounter) < 1 {
		return false
	}
	return true
}

// MinimockGetMetadataInspect logs each unmet expectation
func (m *StorageMock) MinimockGetMetadataInspect() {
	for _, e := range m.GetMetadataMock.expectations {
		if mm_atomic.LoadUint64(&e.Counter) < 1 {
			m.t.Errorf("Expected call to StorageMock.GetMetadata with params: %#v", *e.params)
		}
	}

	// if default expectation was set then invocations count should be greater than zero
	if m.GetMetadataMock.defaultExpectation != nil && mm_atomic.LoadUint64(&m.afterGetMetadataCounter) < 1 {
		if m.GetMetadataMock.defaultExpectation.params == nil {
			m.t.Error("Expected call to StorageMock.GetMetadata")
		} else {
			m.t.Errorf("Expected call to StorageMock.GetMetadata with params: %#v", *m.GetMetadataMock.defaultExpectation.params)
		}
	}
	// if func was set then invocations count should be greater than zero
	if m.funcGetMetadata != nil && mm_atomic.LoadUint64(&m.afterGetMetadataCounter) < 1 {
		m.t.Error("Expected call to StorageMock.GetMetadata")
	}
}

type mStorageMockSaveJob struct {
	mock               *StorageMock
	defaultExpectation *StorageMockSaveJobExpectation
	expectations       []*StorageMockSaveJobExpectation

	callArgs []*StorageMockSaveJobParams
	mutex    sync.RWMutex
}

// StorageMockSaveJobExpectation specifies expectation struct of the Storage.SaveJob
type StorageMockSaveJobExpectation struct {
	mock    *StorageMock
	params  *StorageMockSaveJobParams
	results *StorageMockSaveJobResults
	Counter uint64
}

// StorageMockSaveJobParams contains parameters of the Storage.SaveJob
type StorageMockSaveJobParams struct {
	ctx context.Context
	job *mm_service.Job
}

// StorageMockSaveJobResults contains results of the Storage.SaveJob
type StorageMockSaveJobResults struct {
	err error
}

// Expect sets up expected params for Storage.SaveJob
func (mmSaveJob *mStorageMockSaveJob) Expect(ctx context.Context, job *mm_service.Job) *mStorageMockSaveJob {
	if mmSaveJob.mock.funcSaveJob != nil {
		mmSaveJob.mock.t.Fatalf("StorageMock.SaveJob mock is already set by Set")
	}

	if mmSaveJob.defaultExpectation == nil {
		mmSaveJob.defaultExpectation = &StorageMockSaveJobExpectation{}
	}

	mmSaveJob.defaultExpectation.params = &StorageMockSaveJobParams{ctx, job}
	for _, e := range mmSaveJob.expectations {
		if minimock.Equal(e.params, mmSaveJob.defaultExpectation.params) {
			mmSaveJob.mock.t.Fatalf("Expectation set by When has same params: %#v", *mmSaveJob.defaultExpectation.params)
		}
	}

	return mmSaveJob
}

// Inspect accepts an inspector function that has same arguments as the Storage.SaveJob
func (mmSaveJob *mStorageMockSaveJob) Inspect(f func(ctx context.Context, job *mm_service.Job)) *mStorageMockSaveJob {
	if mmSaveJob.mock.inspectFuncSaveJob != nil {
		mmSaveJob.mock.t.Fatalf("Inspect function is already set for StorageMock.SaveJob")
	}

	mmSaveJob.mock.inspectFuncSaveJob = f

	return mmSaveJob
}

// Return sets up results that will be returned by Storage.SaveJob
func (mmSaveJob *mStorageMockSaveJob) Return(err error) *StorageMock {
	if mmSaveJob.mock.funcSaveJob != nil {
		mmSaveJob.mock.t.Fatalf("StorageMock.SaveJob mock is already set by Set")
	}

	if mmSaveJob.defaultExpectation == nil {
		mmSaveJob.defaultExpectation = &StorageMockSaveJobExpectation{mock: mmSaveJob.mock}
	}
	mmSaveJob.defaultExpectation.results = &StorageMockSaveJobResults{err}
	return mmSaveJob.mock
}

//Set uses given function f to mock the Storage.SaveJob method
func (mmSaveJob *mStorageMockSaveJob) Set(f func(ctx context.Context, job *mm_service.Job) (err error)) *StorageMock {
	if mmSaveJob.defaultExpectation != nil {
		mmSaveJob.mock.t.Fatalf("Default expectation is already set for the Storage.SaveJob method")
	}

	if len(mmSaveJob.expectations) > 0 {
		mmSaveJob.mock.t.Fatalf("Some expectations are already set for the Storage.SaveJob method")
	}

	mmSaveJob.mock.funcSaveJob = f
	return mmSaveJob.mock
}

// When sets expectation for the Storage.SaveJob which will trigger the result defined by the following
// Then helper
func (mmSaveJob *mStorageMockSaveJob) When(ctx context.Context, job *mm_service.Job) *StorageMockSaveJobExpectation {
	if mmSaveJob.mock.funcSaveJob != nil {
		mmSaveJob.mock.t.Fatalf("StorageMock.SaveJob mock is already set by Set")
	}

	expectation := &StorageMockSaveJobExpectation{
		mock:   mmSaveJob.mock,
		params: &StorageMockSaveJobParams{ctx, job},
	}
	mmSaveJob.expectations = append(mmSaveJob.expectations, expectation)
	return expectation
}

// Then sets up Storage.SaveJob return parameters for the expectation previously defined by the When method
func (e *StorageMockSaveJobExpectation) Then(err error) *StorageMock {
	e.results = &StorageMockSaveJobResults{err}
	return e.mock
}

// SaveJob implements service.Storage
func (mmSaveJob *StorageMock) SaveJob(ctx context.Context, job *mm_service.Job) (err error) {
	mm_atomic.AddUint64(&mmSaveJob.beforeSaveJobCounter, 1)
	defer mm_atomic.AddUint64(&mmSaveJob.afterSaveJobCounter, 1)

	if mmSaveJob.inspectFuncSaveJob != nil {
		mmSaveJob.inspectFuncSaveJob(ctx, job)
	}

	mm_params := &StorageMockSaveJobParams{ctx, job}

	// Record call args
	mmSaveJob.SaveJobMock.mutex.Lock()
	mmSaveJob.SaveJobMock.callArgs = append(mmSaveJob.SaveJobMock.callArgs, mm_params)
	mmSaveJob.SaveJobMock.mutex.Unlock()

	for _, e := range mmSaveJob.SaveJobMock.expectations {
		if minimock.Equal(e.params, mm_params) {
			mm_atomic.AddUint64(&e.Counter, 1)
			return e.results.err
		}
	}

	if mmSaveJob.SaveJobMock.defaultExpectation != nil {
		mm_atomic.AddUint64(&mmSaveJob.SaveJobMock.defaultExpectation.Counter, 1)
		mm_want := mmSaveJob.SaveJobMock.defaultExpectation.params
		mm_got := StorageMockSaveJobParams{ctx, job}
		if mm_want != nil && !minimock.Equal(*mm_want, mm_got) {
			mmSaveJob.t.Errorf("StorageMock.SaveJob got unexpected parameters, want: %#v, got: %#v%s\n", *mm_want, mm_got, minimock.Diff(*mm_want, mm_got))
		}

		mm_results := mmSaveJob.SaveJobMock.defaultExpectation.results
		if mm_results == nil {
			mmSaveJob.t.Fatal("No results are set for the StorageMock.SaveJob")
		}
		return (*mm_results).err
	}
	if mmSaveJob.funcSaveJob != nil {
		return mmSaveJob.funcSaveJob(ctx, job)
	}
	mmSaveJob.t.Fatalf("Unexpected call to StorageMock.SaveJob. %v %v", ctx, job)
	return
}

// SaveJobAfterCounter returns a count of finished StorageMock.SaveJob invocations
func (mmSaveJob *StorageMock) SaveJobAfterCounter() uint64 {
	return mm_atomic.LoadUint64(&mmSaveJob.afterSaveJobCounter)
}

// SaveJobBeforeCounter returns a count of StorageMock.SaveJob invocations
func (mmSaveJob *StorageMock) SaveJobBeforeCounter() uint64 {
	return mm_atomic.LoadUint64(&mmSaveJob.beforeSaveJobCounter)
}

// Calls returns a list of arguments used in each call to StorageMock.SaveJob.
// The list is in the same order as the calls were made (i.e. recent calls have a higher index)
func (mmSaveJob *mStorageMockSaveJob) Calls() []*StorageMockSaveJobParams {
	mmSaveJob.mutex.RLock()

	argCopy := make([]*StorageMockSaveJobParams, len(mmSaveJob.callArgs))
	copy(argCopy, mmSaveJob.callArgs)

	mmSaveJob.mutex.RUnlock()

	return argCopy
}

// MinimockSaveJobDone returns true if the count of the SaveJob invocations corresponds
// the number of defined expectations
func (m *StorageMock) MinimockSaveJobDone() bool {
	for _, e := range m.SaveJobMock.expectations {
		if mm_atomic.LoadUint64(&e.Counter) < 1 {
			return false
		}
	}

	// if default expectation was set then invocations count should be greater than zero
	if m.SaveJobMock.defaultExpectation != nil && mm_atomic.LoadUint64(&m.afterSaveJobCounter) < 1 {
		return false
	}
	// if func was set then invocations count should be greater than zero
	if m.funcSaveJob != nil && mm_atomic.LoadUint64(&m.afterSaveJobCounter) < 1 {
		return false
	}
	return true
}

// MinimockSaveJobInspect logs each unmet expectation
func (m *StorageMock) MinimockSaveJobInspect() {
	for _, e := range m.SaveJobMock.expectations {
		if mm_atomic.LoadUint64(&e.Counter) < 1 {
			m.t.Errorf("Expected call to StorageMock.SaveJob with params: %#v", *e.params)
		}
	}

	// if default expectation was set then invocations count should be greater than zero
	if m.SaveJobMock.defaultExpectation != nil && mm_atomic.LoadUint64(&m.afterSaveJobCounter) < 1 {
		if m.SaveJobMock.defaultExpectation.params == nil {
			m.t.Error("Expected call to StorageMock.SaveJob")
		} else {
			m.t.Errorf("Expected call to StorageMock.SaveJob with params: %#v", *m.SaveJobMock.defaultExpectation.params)
		}
	}
	// if func was set then invocations count should be greater than zero
	if m.funcSaveJob != nil && mm_atomic.LoadUint64(&m.afterSaveJobCounter) < 1 {
		m.t.Error("Expected call to StorageMock.SaveJob")
	}
}

type mStorageMockSaveMetadata struct {
	mock               *StorageMock
	defaultExpectation *StorageMockSaveMetadataExpectation
	expectations       []*StorageMockSaveMetadataExpectation

	callArgs []*StorageMockSaveMetadataParams
	mutex    sync.RWMutex
}

// StorageMockSaveMetadataExpectation specifies expectation struct of the Storage.SaveMetadata
type StorageMockSaveMetadataExpectation struct {
	mock    *StorageMock
	params  *StorageMockSaveMetadataParams
	results *StorageMockSaveMetadataResults
	Counter uint64
}

// StorageMockSaveMetadataParams contains parameters of the Storage.SaveMetadata
type StorageMockSaveMetadataParams struct {
	ctx      context.Context
	metadata *mm_service.Metadata
}

// StorageMockSaveMetadataResults contains results of the Storage.SaveMetadata
type StorageMockSaveMetadataResults struct {
	err error
}

// Expect sets up expected params for Storage.SaveMetadata
func (mmSaveMetadata *mStorageMockSaveMetadata) Expect(ctx context.Context, metadata *mm_service.Metadata) *mStorageMockSaveMetadata {
	if mmSaveMetadata.mock.funcSaveMetadata != nil {
		mmSaveMetadata.mock.t.Fatalf("StorageMock.SaveMetadata mock is already set by Set")
	}

	if mmSaveMetadata.defaultExpectation == nil {
		mmSaveMetadata.defaultExpectation = &StorageMockSaveMetadataExpectation{}
	}

	mmSaveMetadata.defaultExpectation.params = &StorageMockSaveMetadataParams{ctx, metadata}
	for _, e := range mmSaveMetadata.expectations {
		if minimock.Equal(e.params, mmSaveMetadata.defaultExpectation.params) {
			mmSaveMetadata.mock.t.Fatalf("Expectation set by When has same params: %#v", *mmSaveMetadata.defaultExpectation.params)
		}
	}

	return mmSaveMetadata
}

// Inspect accepts an inspector function that has same arguments as the Storage.SaveMetadata
func (mmSaveMetadata *mStorageMockSaveMetadata) Inspect(f func(ctx context.Context, metadata *mm_service.Metadata)) *mStorageMockSaveMetadata {
	if mmSaveMetadata.mock.inspectFuncSaveMetadata != nil {
		mmSaveMetadata.mock.t.Fatalf("Inspect function is already set for StorageMock.SaveMetadata")
	}

	mmSaveMetadata.mock.inspectFuncSaveMetadata = f

	return mmSaveMetadata
}

// Return sets up results that will be returned by Storage.SaveMetadata
func (mmSaveMetadata *mStorageMockSaveMetadata) Return(err error) *StorageMock {
	if mmSaveMetadata.mock.funcSaveMetadata != nil {
		mmSaveMetadata.mock.t.Fatalf("StorageMock.SaveMetadata mock is already set by Set")
	}

	if mmSaveMetadata.defaultExpectation == nil {
		mmSaveMetadata.defaultExpectation = &StorageMockSaveMetadataExpectation{mock: mmSaveMetadata.mock}
	}
	mmSaveMetadata.defaultExpectation.results = &StorageMockSaveMetadataResults{err}
	return mmSaveMetadata.mock
}

//Set uses given function f to mock the Storage.SaveMetadata method
func (mmSaveMetadata *mStorageMockSaveMetadata) Set(f func(ctx context.Context, metadata *mm_service.Metadata) (err error)) *StorageMock {
	if mmSaveMetadata.defaultExpectation != nil {
		mmSaveMetadata.mock.t.Fatalf("Default expectation is already set for the Storage.SaveMetadata method")
	}

	if len(mmSaveMetadata.expectations) > 0 {
		mmSaveMetadata.mock.t.Fatalf("Some expectations are already set for the Storage.SaveMetadata method")
	}

	mmSaveMetadata.mock.funcSaveMetadata = f
	return mmSaveMetadata.mock
}

// When sets expectation for the Storage.SaveMetadata which will trigger the result defined by the following
// Then helper
func (mmSaveMetadata *mStorageMockSaveMetadata) When(ctx context.Context, metadata *mm_service.Metadata) *StorageMockSaveMetadataExpectation {
	if mmSaveMetadata.mock.funcSaveMetadata != nil {
		mmSaveMetadata.mock.t.Fatalf("StorageMock.SaveMetadata mock is already set by Set")
	}

	expectation := &StorageMockSaveMetadataExpectation{
		mock:   mmSaveMetadata.mock,
		params: &StorageMockSaveMetadataParams{ctx, metadata},
	}
	mmSaveMetadata.expectations = append(mmSaveMetadata.expectations, expectation)
	return expectation
}

// Then sets up Storage.SaveMetadata return parameters for the expectation previously defined by the When method
func (e *StorageMockSaveMetadataExpectation) Then(err error) *StorageMock {
	e.results = &StorageMockSaveMetadataResults{err}
	return e.mock
}

// SaveMetadata implements service.Storage
func (mmSaveMetadata *StorageMock) SaveMetadata(ctx context.Context, metadata *mm_service.Metadata) (err error) {
	mm_atomic.AddUint64(&mmSaveMetadata.beforeSaveMetadataCounter, 1)
	defer mm_atomic.AddUint64(&mmSaveMetadata.afterSaveMetadataCounter, 1)

	if mmSaveMetadata.inspectFuncSaveMetadata != nil {
		mmSaveMetadata.inspectFuncSaveMetadata(ctx, metadata)
	}

	mm_params := &StorageMockSaveMetadataParams{ctx, metadata}

	// Record call args
	mmSaveMetadata.SaveMetadataMock.mutex.Lock()
	mmSaveMetadata.SaveMetadataMock.callArgs = append(mmSaveMetadata.SaveMetadataMock.callArgs, mm_params)
	mmSaveMetadata.SaveMetadataMock.mutex.Unlock()

	for _, e := range mmSaveMetadata.SaveMetadataMock.expectations {
		if minimock.Equal(e.params, mm_params) {
			mm_atomic.AddUint64(&e.Counter, 1)
			return e.results.err
		}
	}

	if mmSaveMetadata.SaveMetadataMock.defaultExpectation != nil {
		mm_atomic.AddUint64(&mmSaveMetadata.SaveMetadataMock.defaultExpectation.Counter, 1)
		mm_want := mmSaveMetadata.SaveMetadataMock.defaultExpectation.params
		mm_got := StorageMockSaveMetadataParams{ctx, metadata}
		if mm_want != nil && !minimock.Equal(*mm_want, mm_got) {
			mmSaveMetadata.t.Errorf("StorageMock.SaveMetadata got unexpected parameters, want: %#v, got: %#v%s\n", *mm_want, mm_got, minimock.Diff(*mm_want, mm_got))
		}

		mm_results := mmSaveMetadata.SaveMetadataMock.defaultExpectation.results
		if mm_results == nil {
			mmSaveMetadata.t.Fatal("No results are set for the StorageMock.SaveMetadata")
		}
		return (*mm_results).err
	}
	if mmSaveMetadata.funcSaveMetadata != nil {
		return mmSaveMetadata.funcSaveMetadata(ctx, metadata)
	}
	mmSaveMetadata.t.Fatalf("Unexpected call to StorageMock.SaveMetadata. %v %v", ctx, metadata)
	return
}

// SaveMetadataAfterCounter returns a count of finished StorageMock.SaveMetadata invocations
func (mmSaveMetadata *StorageMock) SaveMetadataAfterCounter() uint64 {
	return mm_atomic.LoadUint64(&mmSaveMetadata.afterSaveMetadataCounter)
}

// SaveMetadataBeforeCounter returns a count of StorageMock.SaveMetadata invocations
func (mmSaveMetadata *StorageMock) SaveMetadataBeforeCounter() uint64 {
	return mm_atomic.LoadUint64(&mmSaveMetadata.beforeSaveMetadataCounter)
}

// Calls returns a list of arguments used in each call to StorageMock.SaveMetadata.
// The list is in the same order as the calls were made (i.e. recent calls have a higher index)
func (mmSaveMetadata *mStorageMockSaveMetadata) Calls() []*StorageMockSaveMetadataParams {
	mmSaveMetadata.mutex.RLock()

	argCopy := make([]*StorageMockSaveMetadataParams, len(mmSaveMetadata.callArgs))
	copy(argCopy, mmSaveMetadata.callArgs)

	mmSaveMetadata.mutex.RUnlock()

	return argCopy
}

// MinimockSaveMetadataDone returns true if the count of the SaveMetadata invocations corresponds
// the number of defined expectations
func (m *StorageMock) MinimockSaveMetadataDone() bool {
	for _, e := range m.SaveMetadataMock.expectations {
		if mm_atomic.LoadUint64(&e.Counter) < 1 {
			return false
		}
	}

	// if default expectation was set then invocations count should be greater than zero
	if m.SaveMetadataMock.defaultExpectation != nil && mm_atomic.LoadUint64(&m.afterSaveMetadataCounter) < 1 {
		return false
	}
	// if func was set then invocations count should be greater than zero
	if m.funcSaveMetadata != nil && mm_atomic.LoadUint64(&m.afterSaveMetadataCounter) < 1 {
		return false
	}
	return true
}

// MinimockSaveMetadataInspect logs each unmet expectation
func (m *StorageMock) MinimockSaveMetadataInspect() {
	for _, e := range m.SaveMetadataMock.expectations {
		if mm_atomic.LoadUint64(&e.Counter) < 1 {
			m.t.Errorf("Expected call to StorageMock.SaveMetadata with params: %#v", *e.params)
		}
	}

	// if default expectation was set then invocations count should be greater than zero
	if m.SaveMetadataMock.defaultExpectation != nil && mm_atomic.LoadUint64(&m.afterSaveMetadataCounter) < 1 {
		if m.SaveMetadataMock.defaultExpectation.params == nil {
			m.t.Error("Expected call to StorageMock.SaveMetadata")
		} else {
			m.t.Errorf("Expected call to StorageMock.SaveMetadata with params: %#v", *m.SaveMetadataMock.defaultExpectation.params)
		}
	}
	// if func was set then invocations count should be greater than zero
	if m.funcSaveMetadata != nil && mm_atomic.LoadUint64(&m.afterSaveMetadataCounter) < 1 {
		m.t.Error("Expected call to StorageMock.SaveMetadata")
	}
}

// MinimockFinish checks that all mocked methods have been called the expected number of times
func (m *StorageMock) MinimockFinish() {
	if !m.minimockDone() {
		m.MinimockGetJobInspect()

		m.MinimockGetMetadataInspect()

		m.MinimockSaveJobInspect()

		m.MinimockSaveMetadataInspect()
		m.t.FailNow()
	}
}

// MinimockWait waits for all mocked methods to be called the expected number of times
func (m *StorageMock) MinimockWait(timeout mm_time.Duration) {
	timeoutCh := mm_time.After(timeout)
	for {
		if m.minimockDone() {
			return
		}
		select {
		case <-timeoutCh:
			m.MinimockFinish()
			return
		case <-mm_time.After(10 * mm_time.Millisecond):
		}
	}
}

func (m *StorageMock) minimockDone() bool {
	done := true
	return done &&
		m.MinimockGetJobDone() &&
		m.MinimockGetMetadataDone() &&
		m.MinimockSaveJobDone() &&
		m.MinimockSaveMetadataDone()
}
