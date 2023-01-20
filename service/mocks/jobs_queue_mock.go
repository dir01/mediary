package mocks

// Code generated by http://github.com/gojuno/minimock (dev). DO NOT EDIT.

import (
	"context"
	"sync"
	mm_atomic "sync/atomic"
	mm_time "time"

	"github.com/gojuno/minimock/v3"
)

// JobsQueueMock implements service.JobsQueue
type JobsQueueMock struct {
	t minimock.Tester

	funcPublish          func(ctx context.Context, jobType string, payload any) (err error)
	inspectFuncPublish   func(ctx context.Context, jobType string, payload any)
	afterPublishCounter  uint64
	beforePublishCounter uint64
	PublishMock          mJobsQueueMockPublish

	funcRun          func()
	inspectFuncRun   func()
	afterRunCounter  uint64
	beforeRunCounter uint64
	RunMock          mJobsQueueMockRun

	funcShutdown          func()
	inspectFuncShutdown   func()
	afterShutdownCounter  uint64
	beforeShutdownCounter uint64
	ShutdownMock          mJobsQueueMockShutdown

	funcSubscribe          func(ctx context.Context, jobType string, f func(payloadBytes []byte) error)
	inspectFuncSubscribe   func(ctx context.Context, jobType string, f func(payloadBytes []byte) error)
	afterSubscribeCounter  uint64
	beforeSubscribeCounter uint64
	SubscribeMock          mJobsQueueMockSubscribe
}

// NewJobsQueueMock returns a mock for service.JobsQueue
func NewJobsQueueMock(t minimock.Tester) *JobsQueueMock {
	m := &JobsQueueMock{t: t}
	if controller, ok := t.(minimock.MockController); ok {
		controller.RegisterMocker(m)
	}

	m.PublishMock = mJobsQueueMockPublish{mock: m}
	m.PublishMock.callArgs = []*JobsQueueMockPublishParams{}

	m.RunMock = mJobsQueueMockRun{mock: m}

	m.ShutdownMock = mJobsQueueMockShutdown{mock: m}

	m.SubscribeMock = mJobsQueueMockSubscribe{mock: m}
	m.SubscribeMock.callArgs = []*JobsQueueMockSubscribeParams{}

	return m
}

type mJobsQueueMockPublish struct {
	mock               *JobsQueueMock
	defaultExpectation *JobsQueueMockPublishExpectation
	expectations       []*JobsQueueMockPublishExpectation

	callArgs []*JobsQueueMockPublishParams
	mutex    sync.RWMutex
}

// JobsQueueMockPublishExpectation specifies expectation struct of the JobsQueue.Publish
type JobsQueueMockPublishExpectation struct {
	mock    *JobsQueueMock
	params  *JobsQueueMockPublishParams
	results *JobsQueueMockPublishResults
	Counter uint64
}

// JobsQueueMockPublishParams contains parameters of the JobsQueue.Publish
type JobsQueueMockPublishParams struct {
	ctx     context.Context
	jobType string
	payload any
}

// JobsQueueMockPublishResults contains results of the JobsQueue.Publish
type JobsQueueMockPublishResults struct {
	err error
}

// Expect sets up expected params for JobsQueue.Publish
func (mmPublish *mJobsQueueMockPublish) Expect(ctx context.Context, jobType string, payload any) *mJobsQueueMockPublish {
	if mmPublish.mock.funcPublish != nil {
		mmPublish.mock.t.Fatalf("JobsQueueMock.Publish mock is already set by Set")
	}

	if mmPublish.defaultExpectation == nil {
		mmPublish.defaultExpectation = &JobsQueueMockPublishExpectation{}
	}

	mmPublish.defaultExpectation.params = &JobsQueueMockPublishParams{ctx, jobType, payload}
	for _, e := range mmPublish.expectations {
		if minimock.Equal(e.params, mmPublish.defaultExpectation.params) {
			mmPublish.mock.t.Fatalf("Expectation set by When has same params: %#v", *mmPublish.defaultExpectation.params)
		}
	}

	return mmPublish
}

// Inspect accepts an inspector function that has same arguments as the JobsQueue.Publish
func (mmPublish *mJobsQueueMockPublish) Inspect(f func(ctx context.Context, jobType string, payload any)) *mJobsQueueMockPublish {
	if mmPublish.mock.inspectFuncPublish != nil {
		mmPublish.mock.t.Fatalf("Inspect function is already set for JobsQueueMock.Publish")
	}

	mmPublish.mock.inspectFuncPublish = f

	return mmPublish
}

// Return sets up results that will be returned by JobsQueue.Publish
func (mmPublish *mJobsQueueMockPublish) Return(err error) *JobsQueueMock {
	if mmPublish.mock.funcPublish != nil {
		mmPublish.mock.t.Fatalf("JobsQueueMock.Publish mock is already set by Set")
	}

	if mmPublish.defaultExpectation == nil {
		mmPublish.defaultExpectation = &JobsQueueMockPublishExpectation{mock: mmPublish.mock}
	}
	mmPublish.defaultExpectation.results = &JobsQueueMockPublishResults{err}
	return mmPublish.mock
}

//Set uses given function f to mock the JobsQueue.Publish method
func (mmPublish *mJobsQueueMockPublish) Set(f func(ctx context.Context, jobType string, payload any) (err error)) *JobsQueueMock {
	if mmPublish.defaultExpectation != nil {
		mmPublish.mock.t.Fatalf("Default expectation is already set for the JobsQueue.Publish method")
	}

	if len(mmPublish.expectations) > 0 {
		mmPublish.mock.t.Fatalf("Some expectations are already set for the JobsQueue.Publish method")
	}

	mmPublish.mock.funcPublish = f
	return mmPublish.mock
}

// When sets expectation for the JobsQueue.Publish which will trigger the result defined by the following
// Then helper
func (mmPublish *mJobsQueueMockPublish) When(ctx context.Context, jobType string, payload any) *JobsQueueMockPublishExpectation {
	if mmPublish.mock.funcPublish != nil {
		mmPublish.mock.t.Fatalf("JobsQueueMock.Publish mock is already set by Set")
	}

	expectation := &JobsQueueMockPublishExpectation{
		mock:   mmPublish.mock,
		params: &JobsQueueMockPublishParams{ctx, jobType, payload},
	}
	mmPublish.expectations = append(mmPublish.expectations, expectation)
	return expectation
}

// Then sets up JobsQueue.Publish return parameters for the expectation previously defined by the When method
func (e *JobsQueueMockPublishExpectation) Then(err error) *JobsQueueMock {
	e.results = &JobsQueueMockPublishResults{err}
	return e.mock
}

// Publish implements service.JobsQueue
func (mmPublish *JobsQueueMock) Publish(ctx context.Context, jobType string, payload any) (err error) {
	mm_atomic.AddUint64(&mmPublish.beforePublishCounter, 1)
	defer mm_atomic.AddUint64(&mmPublish.afterPublishCounter, 1)

	if mmPublish.inspectFuncPublish != nil {
		mmPublish.inspectFuncPublish(ctx, jobType, payload)
	}

	mm_params := &JobsQueueMockPublishParams{ctx, jobType, payload}

	// Record call args
	mmPublish.PublishMock.mutex.Lock()
	mmPublish.PublishMock.callArgs = append(mmPublish.PublishMock.callArgs, mm_params)
	mmPublish.PublishMock.mutex.Unlock()

	for _, e := range mmPublish.PublishMock.expectations {
		if minimock.Equal(e.params, mm_params) {
			mm_atomic.AddUint64(&e.Counter, 1)
			return e.results.err
		}
	}

	if mmPublish.PublishMock.defaultExpectation != nil {
		mm_atomic.AddUint64(&mmPublish.PublishMock.defaultExpectation.Counter, 1)
		mm_want := mmPublish.PublishMock.defaultExpectation.params
		mm_got := JobsQueueMockPublishParams{ctx, jobType, payload}
		if mm_want != nil && !minimock.Equal(*mm_want, mm_got) {
			mmPublish.t.Errorf("JobsQueueMock.Publish got unexpected parameters, want: %#v, got: %#v%s\n", *mm_want, mm_got, minimock.Diff(*mm_want, mm_got))
		}

		mm_results := mmPublish.PublishMock.defaultExpectation.results
		if mm_results == nil {
			mmPublish.t.Fatal("No results are set for the JobsQueueMock.Publish")
		}
		return (*mm_results).err
	}
	if mmPublish.funcPublish != nil {
		return mmPublish.funcPublish(ctx, jobType, payload)
	}
	mmPublish.t.Fatalf("Unexpected call to JobsQueueMock.Publish. %v %v %v", ctx, jobType, payload)
	return
}

// PublishAfterCounter returns a count of finished JobsQueueMock.Publish invocations
func (mmPublish *JobsQueueMock) PublishAfterCounter() uint64 {
	return mm_atomic.LoadUint64(&mmPublish.afterPublishCounter)
}

// PublishBeforeCounter returns a count of JobsQueueMock.Publish invocations
func (mmPublish *JobsQueueMock) PublishBeforeCounter() uint64 {
	return mm_atomic.LoadUint64(&mmPublish.beforePublishCounter)
}

// Calls returns a list of arguments used in each call to JobsQueueMock.Publish.
// The list is in the same order as the calls were made (i.e. recent calls have a higher index)
func (mmPublish *mJobsQueueMockPublish) Calls() []*JobsQueueMockPublishParams {
	mmPublish.mutex.RLock()

	argCopy := make([]*JobsQueueMockPublishParams, len(mmPublish.callArgs))
	copy(argCopy, mmPublish.callArgs)

	mmPublish.mutex.RUnlock()

	return argCopy
}

// MinimockPublishDone returns true if the count of the Publish invocations corresponds
// the number of defined expectations
func (m *JobsQueueMock) MinimockPublishDone() bool {
	for _, e := range m.PublishMock.expectations {
		if mm_atomic.LoadUint64(&e.Counter) < 1 {
			return false
		}
	}

	// if default expectation was set then invocations count should be greater than zero
	if m.PublishMock.defaultExpectation != nil && mm_atomic.LoadUint64(&m.afterPublishCounter) < 1 {
		return false
	}
	// if func was set then invocations count should be greater than zero
	if m.funcPublish != nil && mm_atomic.LoadUint64(&m.afterPublishCounter) < 1 {
		return false
	}
	return true
}

// MinimockPublishInspect logs each unmet expectation
func (m *JobsQueueMock) MinimockPublishInspect() {
	for _, e := range m.PublishMock.expectations {
		if mm_atomic.LoadUint64(&e.Counter) < 1 {
			m.t.Errorf("Expected call to JobsQueueMock.Publish with params: %#v", *e.params)
		}
	}

	// if default expectation was set then invocations count should be greater than zero
	if m.PublishMock.defaultExpectation != nil && mm_atomic.LoadUint64(&m.afterPublishCounter) < 1 {
		if m.PublishMock.defaultExpectation.params == nil {
			m.t.Error("Expected call to JobsQueueMock.Publish")
		} else {
			m.t.Errorf("Expected call to JobsQueueMock.Publish with params: %#v", *m.PublishMock.defaultExpectation.params)
		}
	}
	// if func was set then invocations count should be greater than zero
	if m.funcPublish != nil && mm_atomic.LoadUint64(&m.afterPublishCounter) < 1 {
		m.t.Error("Expected call to JobsQueueMock.Publish")
	}
}

type mJobsQueueMockRun struct {
	mock               *JobsQueueMock
	defaultExpectation *JobsQueueMockRunExpectation
	expectations       []*JobsQueueMockRunExpectation
}

// JobsQueueMockRunExpectation specifies expectation struct of the JobsQueue.Run
type JobsQueueMockRunExpectation struct {
	mock *JobsQueueMock

	Counter uint64
}

// Expect sets up expected params for JobsQueue.Run
func (mmRun *mJobsQueueMockRun) Expect() *mJobsQueueMockRun {
	if mmRun.mock.funcRun != nil {
		mmRun.mock.t.Fatalf("JobsQueueMock.Run mock is already set by Set")
	}

	if mmRun.defaultExpectation == nil {
		mmRun.defaultExpectation = &JobsQueueMockRunExpectation{}
	}

	return mmRun
}

// Inspect accepts an inspector function that has same arguments as the JobsQueue.Run
func (mmRun *mJobsQueueMockRun) Inspect(f func()) *mJobsQueueMockRun {
	if mmRun.mock.inspectFuncRun != nil {
		mmRun.mock.t.Fatalf("Inspect function is already set for JobsQueueMock.Run")
	}

	mmRun.mock.inspectFuncRun = f

	return mmRun
}

// Return sets up results that will be returned by JobsQueue.Run
func (mmRun *mJobsQueueMockRun) Return() *JobsQueueMock {
	if mmRun.mock.funcRun != nil {
		mmRun.mock.t.Fatalf("JobsQueueMock.Run mock is already set by Set")
	}

	if mmRun.defaultExpectation == nil {
		mmRun.defaultExpectation = &JobsQueueMockRunExpectation{mock: mmRun.mock}
	}

	return mmRun.mock
}

//Set uses given function f to mock the JobsQueue.Run method
func (mmRun *mJobsQueueMockRun) Set(f func()) *JobsQueueMock {
	if mmRun.defaultExpectation != nil {
		mmRun.mock.t.Fatalf("Default expectation is already set for the JobsQueue.Run method")
	}

	if len(mmRun.expectations) > 0 {
		mmRun.mock.t.Fatalf("Some expectations are already set for the JobsQueue.Run method")
	}

	mmRun.mock.funcRun = f
	return mmRun.mock
}

// Run implements service.JobsQueue
func (mmRun *JobsQueueMock) Run() {
	mm_atomic.AddUint64(&mmRun.beforeRunCounter, 1)
	defer mm_atomic.AddUint64(&mmRun.afterRunCounter, 1)

	if mmRun.inspectFuncRun != nil {
		mmRun.inspectFuncRun()
	}

	if mmRun.RunMock.defaultExpectation != nil {
		mm_atomic.AddUint64(&mmRun.RunMock.defaultExpectation.Counter, 1)

		return

	}
	if mmRun.funcRun != nil {
		mmRun.funcRun()
		return
	}
	mmRun.t.Fatalf("Unexpected call to JobsQueueMock.Run.")

}

// RunAfterCounter returns a count of finished JobsQueueMock.Run invocations
func (mmRun *JobsQueueMock) RunAfterCounter() uint64 {
	return mm_atomic.LoadUint64(&mmRun.afterRunCounter)
}

// RunBeforeCounter returns a count of JobsQueueMock.Run invocations
func (mmRun *JobsQueueMock) RunBeforeCounter() uint64 {
	return mm_atomic.LoadUint64(&mmRun.beforeRunCounter)
}

// MinimockRunDone returns true if the count of the Run invocations corresponds
// the number of defined expectations
func (m *JobsQueueMock) MinimockRunDone() bool {
	for _, e := range m.RunMock.expectations {
		if mm_atomic.LoadUint64(&e.Counter) < 1 {
			return false
		}
	}

	// if default expectation was set then invocations count should be greater than zero
	if m.RunMock.defaultExpectation != nil && mm_atomic.LoadUint64(&m.afterRunCounter) < 1 {
		return false
	}
	// if func was set then invocations count should be greater than zero
	if m.funcRun != nil && mm_atomic.LoadUint64(&m.afterRunCounter) < 1 {
		return false
	}
	return true
}

// MinimockRunInspect logs each unmet expectation
func (m *JobsQueueMock) MinimockRunInspect() {
	for _, e := range m.RunMock.expectations {
		if mm_atomic.LoadUint64(&e.Counter) < 1 {
			m.t.Error("Expected call to JobsQueueMock.Run")
		}
	}

	// if default expectation was set then invocations count should be greater than zero
	if m.RunMock.defaultExpectation != nil && mm_atomic.LoadUint64(&m.afterRunCounter) < 1 {
		m.t.Error("Expected call to JobsQueueMock.Run")
	}
	// if func was set then invocations count should be greater than zero
	if m.funcRun != nil && mm_atomic.LoadUint64(&m.afterRunCounter) < 1 {
		m.t.Error("Expected call to JobsQueueMock.Run")
	}
}

type mJobsQueueMockShutdown struct {
	mock               *JobsQueueMock
	defaultExpectation *JobsQueueMockShutdownExpectation
	expectations       []*JobsQueueMockShutdownExpectation
}

// JobsQueueMockShutdownExpectation specifies expectation struct of the JobsQueue.Shutdown
type JobsQueueMockShutdownExpectation struct {
	mock *JobsQueueMock

	Counter uint64
}

// Expect sets up expected params for JobsQueue.Shutdown
func (mmShutdown *mJobsQueueMockShutdown) Expect() *mJobsQueueMockShutdown {
	if mmShutdown.mock.funcShutdown != nil {
		mmShutdown.mock.t.Fatalf("JobsQueueMock.Shutdown mock is already set by Set")
	}

	if mmShutdown.defaultExpectation == nil {
		mmShutdown.defaultExpectation = &JobsQueueMockShutdownExpectation{}
	}

	return mmShutdown
}

// Inspect accepts an inspector function that has same arguments as the JobsQueue.Shutdown
func (mmShutdown *mJobsQueueMockShutdown) Inspect(f func()) *mJobsQueueMockShutdown {
	if mmShutdown.mock.inspectFuncShutdown != nil {
		mmShutdown.mock.t.Fatalf("Inspect function is already set for JobsQueueMock.Shutdown")
	}

	mmShutdown.mock.inspectFuncShutdown = f

	return mmShutdown
}

// Return sets up results that will be returned by JobsQueue.Shutdown
func (mmShutdown *mJobsQueueMockShutdown) Return() *JobsQueueMock {
	if mmShutdown.mock.funcShutdown != nil {
		mmShutdown.mock.t.Fatalf("JobsQueueMock.Shutdown mock is already set by Set")
	}

	if mmShutdown.defaultExpectation == nil {
		mmShutdown.defaultExpectation = &JobsQueueMockShutdownExpectation{mock: mmShutdown.mock}
	}

	return mmShutdown.mock
}

//Set uses given function f to mock the JobsQueue.Shutdown method
func (mmShutdown *mJobsQueueMockShutdown) Set(f func()) *JobsQueueMock {
	if mmShutdown.defaultExpectation != nil {
		mmShutdown.mock.t.Fatalf("Default expectation is already set for the JobsQueue.Shutdown method")
	}

	if len(mmShutdown.expectations) > 0 {
		mmShutdown.mock.t.Fatalf("Some expectations are already set for the JobsQueue.Shutdown method")
	}

	mmShutdown.mock.funcShutdown = f
	return mmShutdown.mock
}

// Shutdown implements service.JobsQueue
func (mmShutdown *JobsQueueMock) Shutdown() {
	mm_atomic.AddUint64(&mmShutdown.beforeShutdownCounter, 1)
	defer mm_atomic.AddUint64(&mmShutdown.afterShutdownCounter, 1)

	if mmShutdown.inspectFuncShutdown != nil {
		mmShutdown.inspectFuncShutdown()
	}

	if mmShutdown.ShutdownMock.defaultExpectation != nil {
		mm_atomic.AddUint64(&mmShutdown.ShutdownMock.defaultExpectation.Counter, 1)

		return

	}
	if mmShutdown.funcShutdown != nil {
		mmShutdown.funcShutdown()
		return
	}
	mmShutdown.t.Fatalf("Unexpected call to JobsQueueMock.Shutdown.")

}

// ShutdownAfterCounter returns a count of finished JobsQueueMock.Shutdown invocations
func (mmShutdown *JobsQueueMock) ShutdownAfterCounter() uint64 {
	return mm_atomic.LoadUint64(&mmShutdown.afterShutdownCounter)
}

// ShutdownBeforeCounter returns a count of JobsQueueMock.Shutdown invocations
func (mmShutdown *JobsQueueMock) ShutdownBeforeCounter() uint64 {
	return mm_atomic.LoadUint64(&mmShutdown.beforeShutdownCounter)
}

// MinimockShutdownDone returns true if the count of the Shutdown invocations corresponds
// the number of defined expectations
func (m *JobsQueueMock) MinimockShutdownDone() bool {
	for _, e := range m.ShutdownMock.expectations {
		if mm_atomic.LoadUint64(&e.Counter) < 1 {
			return false
		}
	}

	// if default expectation was set then invocations count should be greater than zero
	if m.ShutdownMock.defaultExpectation != nil && mm_atomic.LoadUint64(&m.afterShutdownCounter) < 1 {
		return false
	}
	// if func was set then invocations count should be greater than zero
	if m.funcShutdown != nil && mm_atomic.LoadUint64(&m.afterShutdownCounter) < 1 {
		return false
	}
	return true
}

// MinimockShutdownInspect logs each unmet expectation
func (m *JobsQueueMock) MinimockShutdownInspect() {
	for _, e := range m.ShutdownMock.expectations {
		if mm_atomic.LoadUint64(&e.Counter) < 1 {
			m.t.Error("Expected call to JobsQueueMock.Shutdown")
		}
	}

	// if default expectation was set then invocations count should be greater than zero
	if m.ShutdownMock.defaultExpectation != nil && mm_atomic.LoadUint64(&m.afterShutdownCounter) < 1 {
		m.t.Error("Expected call to JobsQueueMock.Shutdown")
	}
	// if func was set then invocations count should be greater than zero
	if m.funcShutdown != nil && mm_atomic.LoadUint64(&m.afterShutdownCounter) < 1 {
		m.t.Error("Expected call to JobsQueueMock.Shutdown")
	}
}

type mJobsQueueMockSubscribe struct {
	mock               *JobsQueueMock
	defaultExpectation *JobsQueueMockSubscribeExpectation
	expectations       []*JobsQueueMockSubscribeExpectation

	callArgs []*JobsQueueMockSubscribeParams
	mutex    sync.RWMutex
}

// JobsQueueMockSubscribeExpectation specifies expectation struct of the JobsQueue.Subscribe
type JobsQueueMockSubscribeExpectation struct {
	mock   *JobsQueueMock
	params *JobsQueueMockSubscribeParams

	Counter uint64
}

// JobsQueueMockSubscribeParams contains parameters of the JobsQueue.Subscribe
type JobsQueueMockSubscribeParams struct {
	ctx     context.Context
	jobType string
	f       func(payloadBytes []byte) error
}

// Expect sets up expected params for JobsQueue.Subscribe
func (mmSubscribe *mJobsQueueMockSubscribe) Expect(ctx context.Context, jobType string, f func(payloadBytes []byte) error) *mJobsQueueMockSubscribe {
	if mmSubscribe.mock.funcSubscribe != nil {
		mmSubscribe.mock.t.Fatalf("JobsQueueMock.Subscribe mock is already set by Set")
	}

	if mmSubscribe.defaultExpectation == nil {
		mmSubscribe.defaultExpectation = &JobsQueueMockSubscribeExpectation{}
	}

	mmSubscribe.defaultExpectation.params = &JobsQueueMockSubscribeParams{ctx, jobType, f}
	for _, e := range mmSubscribe.expectations {
		if minimock.Equal(e.params, mmSubscribe.defaultExpectation.params) {
			mmSubscribe.mock.t.Fatalf("Expectation set by When has same params: %#v", *mmSubscribe.defaultExpectation.params)
		}
	}

	return mmSubscribe
}

// Inspect accepts an inspector function that has same arguments as the JobsQueue.Subscribe
func (mmSubscribe *mJobsQueueMockSubscribe) Inspect(f func(ctx context.Context, jobType string, f func(payloadBytes []byte) error)) *mJobsQueueMockSubscribe {
	if mmSubscribe.mock.inspectFuncSubscribe != nil {
		mmSubscribe.mock.t.Fatalf("Inspect function is already set for JobsQueueMock.Subscribe")
	}

	mmSubscribe.mock.inspectFuncSubscribe = f

	return mmSubscribe
}

// Return sets up results that will be returned by JobsQueue.Subscribe
func (mmSubscribe *mJobsQueueMockSubscribe) Return() *JobsQueueMock {
	if mmSubscribe.mock.funcSubscribe != nil {
		mmSubscribe.mock.t.Fatalf("JobsQueueMock.Subscribe mock is already set by Set")
	}

	if mmSubscribe.defaultExpectation == nil {
		mmSubscribe.defaultExpectation = &JobsQueueMockSubscribeExpectation{mock: mmSubscribe.mock}
	}

	return mmSubscribe.mock
}

//Set uses given function f to mock the JobsQueue.Subscribe method
func (mmSubscribe *mJobsQueueMockSubscribe) Set(f func(ctx context.Context, jobType string, f func(payloadBytes []byte) error)) *JobsQueueMock {
	if mmSubscribe.defaultExpectation != nil {
		mmSubscribe.mock.t.Fatalf("Default expectation is already set for the JobsQueue.Subscribe method")
	}

	if len(mmSubscribe.expectations) > 0 {
		mmSubscribe.mock.t.Fatalf("Some expectations are already set for the JobsQueue.Subscribe method")
	}

	mmSubscribe.mock.funcSubscribe = f
	return mmSubscribe.mock
}

// Subscribe implements service.JobsQueue
func (mmSubscribe *JobsQueueMock) Subscribe(ctx context.Context, jobType string, f func(payloadBytes []byte) error) {
	mm_atomic.AddUint64(&mmSubscribe.beforeSubscribeCounter, 1)
	defer mm_atomic.AddUint64(&mmSubscribe.afterSubscribeCounter, 1)

	if mmSubscribe.inspectFuncSubscribe != nil {
		mmSubscribe.inspectFuncSubscribe(ctx, jobType, f)
	}

	mm_params := &JobsQueueMockSubscribeParams{ctx, jobType, f}

	// Record call args
	mmSubscribe.SubscribeMock.mutex.Lock()
	mmSubscribe.SubscribeMock.callArgs = append(mmSubscribe.SubscribeMock.callArgs, mm_params)
	mmSubscribe.SubscribeMock.mutex.Unlock()

	for _, e := range mmSubscribe.SubscribeMock.expectations {
		if minimock.Equal(e.params, mm_params) {
			mm_atomic.AddUint64(&e.Counter, 1)
			return
		}
	}

	if mmSubscribe.SubscribeMock.defaultExpectation != nil {
		mm_atomic.AddUint64(&mmSubscribe.SubscribeMock.defaultExpectation.Counter, 1)
		mm_want := mmSubscribe.SubscribeMock.defaultExpectation.params
		mm_got := JobsQueueMockSubscribeParams{ctx, jobType, f}
		if mm_want != nil && !minimock.Equal(*mm_want, mm_got) {
			mmSubscribe.t.Errorf("JobsQueueMock.Subscribe got unexpected parameters, want: %#v, got: %#v%s\n", *mm_want, mm_got, minimock.Diff(*mm_want, mm_got))
		}

		return

	}
	if mmSubscribe.funcSubscribe != nil {
		mmSubscribe.funcSubscribe(ctx, jobType, f)
		return
	}
	mmSubscribe.t.Fatalf("Unexpected call to JobsQueueMock.Subscribe. %v %v %v", ctx, jobType, f)

}

// SubscribeAfterCounter returns a count of finished JobsQueueMock.Subscribe invocations
func (mmSubscribe *JobsQueueMock) SubscribeAfterCounter() uint64 {
	return mm_atomic.LoadUint64(&mmSubscribe.afterSubscribeCounter)
}

// SubscribeBeforeCounter returns a count of JobsQueueMock.Subscribe invocations
func (mmSubscribe *JobsQueueMock) SubscribeBeforeCounter() uint64 {
	return mm_atomic.LoadUint64(&mmSubscribe.beforeSubscribeCounter)
}

// Calls returns a list of arguments used in each call to JobsQueueMock.Subscribe.
// The list is in the same order as the calls were made (i.e. recent calls have a higher index)
func (mmSubscribe *mJobsQueueMockSubscribe) Calls() []*JobsQueueMockSubscribeParams {
	mmSubscribe.mutex.RLock()

	argCopy := make([]*JobsQueueMockSubscribeParams, len(mmSubscribe.callArgs))
	copy(argCopy, mmSubscribe.callArgs)

	mmSubscribe.mutex.RUnlock()

	return argCopy
}

// MinimockSubscribeDone returns true if the count of the Subscribe invocations corresponds
// the number of defined expectations
func (m *JobsQueueMock) MinimockSubscribeDone() bool {
	for _, e := range m.SubscribeMock.expectations {
		if mm_atomic.LoadUint64(&e.Counter) < 1 {
			return false
		}
	}

	// if default expectation was set then invocations count should be greater than zero
	if m.SubscribeMock.defaultExpectation != nil && mm_atomic.LoadUint64(&m.afterSubscribeCounter) < 1 {
		return false
	}
	// if func was set then invocations count should be greater than zero
	if m.funcSubscribe != nil && mm_atomic.LoadUint64(&m.afterSubscribeCounter) < 1 {
		return false
	}
	return true
}

// MinimockSubscribeInspect logs each unmet expectation
func (m *JobsQueueMock) MinimockSubscribeInspect() {
	for _, e := range m.SubscribeMock.expectations {
		if mm_atomic.LoadUint64(&e.Counter) < 1 {
			m.t.Errorf("Expected call to JobsQueueMock.Subscribe with params: %#v", *e.params)
		}
	}

	// if default expectation was set then invocations count should be greater than zero
	if m.SubscribeMock.defaultExpectation != nil && mm_atomic.LoadUint64(&m.afterSubscribeCounter) < 1 {
		if m.SubscribeMock.defaultExpectation.params == nil {
			m.t.Error("Expected call to JobsQueueMock.Subscribe")
		} else {
			m.t.Errorf("Expected call to JobsQueueMock.Subscribe with params: %#v", *m.SubscribeMock.defaultExpectation.params)
		}
	}
	// if func was set then invocations count should be greater than zero
	if m.funcSubscribe != nil && mm_atomic.LoadUint64(&m.afterSubscribeCounter) < 1 {
		m.t.Error("Expected call to JobsQueueMock.Subscribe")
	}
}

// MinimockFinish checks that all mocked methods have been called the expected number of times
func (m *JobsQueueMock) MinimockFinish() {
	if !m.minimockDone() {
		m.MinimockPublishInspect()

		m.MinimockRunInspect()

		m.MinimockShutdownInspect()

		m.MinimockSubscribeInspect()
		m.t.FailNow()
	}
}

// MinimockWait waits for all mocked methods to be called the expected number of times
func (m *JobsQueueMock) MinimockWait(timeout mm_time.Duration) {
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

func (m *JobsQueueMock) minimockDone() bool {
	done := true
	return done &&
		m.MinimockPublishDone() &&
		m.MinimockRunDone() &&
		m.MinimockShutdownDone() &&
		m.MinimockSubscribeDone()
}
