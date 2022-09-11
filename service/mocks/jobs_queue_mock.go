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

	funcPublish          func(ctx context.Context, jobId string) (err error)
	inspectFuncPublish   func(ctx context.Context, jobId string)
	afterPublishCounter  uint64
	beforePublishCounter uint64
	PublishMock          mJobsQueueMockPublish

	funcSubscribe          func(f1 func(jobId string) error)
	inspectFuncSubscribe   func(f1 func(jobId string) error)
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
	ctx   context.Context
	jobId string
}

// JobsQueueMockPublishResults contains results of the JobsQueue.Publish
type JobsQueueMockPublishResults struct {
	err error
}

// Expect sets up expected params for JobsQueue.Publish
func (mmPublish *mJobsQueueMockPublish) Expect(ctx context.Context, jobId string) *mJobsQueueMockPublish {
	if mmPublish.mock.funcPublish != nil {
		mmPublish.mock.t.Fatalf("JobsQueueMock.Publish mock is already set by Set")
	}

	if mmPublish.defaultExpectation == nil {
		mmPublish.defaultExpectation = &JobsQueueMockPublishExpectation{}
	}

	mmPublish.defaultExpectation.params = &JobsQueueMockPublishParams{ctx, jobId}
	for _, e := range mmPublish.expectations {
		if minimock.Equal(e.params, mmPublish.defaultExpectation.params) {
			mmPublish.mock.t.Fatalf("Expectation set by When has same params: %#v", *mmPublish.defaultExpectation.params)
		}
	}

	return mmPublish
}

// Inspect accepts an inspector function that has same arguments as the JobsQueue.Publish
func (mmPublish *mJobsQueueMockPublish) Inspect(f func(ctx context.Context, jobId string)) *mJobsQueueMockPublish {
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
func (mmPublish *mJobsQueueMockPublish) Set(f func(ctx context.Context, jobId string) (err error)) *JobsQueueMock {
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
func (mmPublish *mJobsQueueMockPublish) When(ctx context.Context, jobId string) *JobsQueueMockPublishExpectation {
	if mmPublish.mock.funcPublish != nil {
		mmPublish.mock.t.Fatalf("JobsQueueMock.Publish mock is already set by Set")
	}

	expectation := &JobsQueueMockPublishExpectation{
		mock:   mmPublish.mock,
		params: &JobsQueueMockPublishParams{ctx, jobId},
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
func (mmPublish *JobsQueueMock) Publish(ctx context.Context, jobId string) (err error) {
	mm_atomic.AddUint64(&mmPublish.beforePublishCounter, 1)
	defer mm_atomic.AddUint64(&mmPublish.afterPublishCounter, 1)

	if mmPublish.inspectFuncPublish != nil {
		mmPublish.inspectFuncPublish(ctx, jobId)
	}

	mm_params := &JobsQueueMockPublishParams{ctx, jobId}

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
		mm_got := JobsQueueMockPublishParams{ctx, jobId}
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
		return mmPublish.funcPublish(ctx, jobId)
	}
	mmPublish.t.Fatalf("Unexpected call to JobsQueueMock.Publish. %v %v", ctx, jobId)
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
	f1 func(jobId string) error
}

// Expect sets up expected params for JobsQueue.Subscribe
func (mmSubscribe *mJobsQueueMockSubscribe) Expect(f1 func(jobId string) error) *mJobsQueueMockSubscribe {
	if mmSubscribe.mock.funcSubscribe != nil {
		mmSubscribe.mock.t.Fatalf("JobsQueueMock.Subscribe mock is already set by Set")
	}

	if mmSubscribe.defaultExpectation == nil {
		mmSubscribe.defaultExpectation = &JobsQueueMockSubscribeExpectation{}
	}

	mmSubscribe.defaultExpectation.params = &JobsQueueMockSubscribeParams{f1}
	for _, e := range mmSubscribe.expectations {
		if minimock.Equal(e.params, mmSubscribe.defaultExpectation.params) {
			mmSubscribe.mock.t.Fatalf("Expectation set by When has same params: %#v", *mmSubscribe.defaultExpectation.params)
		}
	}

	return mmSubscribe
}

// Inspect accepts an inspector function that has same arguments as the JobsQueue.Subscribe
func (mmSubscribe *mJobsQueueMockSubscribe) Inspect(f func(f1 func(jobId string) error)) *mJobsQueueMockSubscribe {
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
func (mmSubscribe *mJobsQueueMockSubscribe) Set(f func(f1 func(jobId string) error)) *JobsQueueMock {
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
func (mmSubscribe *JobsQueueMock) Subscribe(f1 func(jobId string) error) {
	mm_atomic.AddUint64(&mmSubscribe.beforeSubscribeCounter, 1)
	defer mm_atomic.AddUint64(&mmSubscribe.afterSubscribeCounter, 1)

	if mmSubscribe.inspectFuncSubscribe != nil {
		mmSubscribe.inspectFuncSubscribe(f1)
	}

	mm_params := &JobsQueueMockSubscribeParams{f1}

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
		mm_got := JobsQueueMockSubscribeParams{f1}
		if mm_want != nil && !minimock.Equal(*mm_want, mm_got) {
			mmSubscribe.t.Errorf("JobsQueueMock.Subscribe got unexpected parameters, want: %#v, got: %#v%s\n", *mm_want, mm_got, minimock.Diff(*mm_want, mm_got))
		}

		return

	}
	if mmSubscribe.funcSubscribe != nil {
		mmSubscribe.funcSubscribe(f1)
		return
	}
	mmSubscribe.t.Fatalf("Unexpected call to JobsQueueMock.Subscribe. %v", f1)

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
		m.MinimockSubscribeDone()
}
