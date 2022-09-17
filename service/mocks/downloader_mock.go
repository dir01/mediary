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

// DownloaderMock implements service.Downloader
type DownloaderMock struct {
	t minimock.Tester

	funcAcceptsURL          func(url string) (b1 bool)
	inspectFuncAcceptsURL   func(url string)
	afterAcceptsURLCounter  uint64
	beforeAcceptsURLCounter uint64
	AcceptsURLMock          mDownloaderMockAcceptsURL

	funcDownload          func(ctx context.Context, url string, filepaths []string) (filepathsMap map[string]string, err error)
	inspectFuncDownload   func(ctx context.Context, url string, filepaths []string)
	afterDownloadCounter  uint64
	beforeDownloadCounter uint64
	DownloadMock          mDownloaderMockDownload

	funcGetMetadata          func(ctx context.Context, url string) (mp1 *mm_service.Metadata, err error)
	inspectFuncGetMetadata   func(ctx context.Context, url string)
	afterGetMetadataCounter  uint64
	beforeGetMetadataCounter uint64
	GetMetadataMock          mDownloaderMockGetMetadata
}

// NewDownloaderMock returns a mock for service.Downloader
func NewDownloaderMock(t minimock.Tester) *DownloaderMock {
	m := &DownloaderMock{t: t}
	if controller, ok := t.(minimock.MockController); ok {
		controller.RegisterMocker(m)
	}

	m.AcceptsURLMock = mDownloaderMockAcceptsURL{mock: m}
	m.AcceptsURLMock.callArgs = []*DownloaderMockAcceptsURLParams{}

	m.DownloadMock = mDownloaderMockDownload{mock: m}
	m.DownloadMock.callArgs = []*DownloaderMockDownloadParams{}

	m.GetMetadataMock = mDownloaderMockGetMetadata{mock: m}
	m.GetMetadataMock.callArgs = []*DownloaderMockGetMetadataParams{}

	return m
}

type mDownloaderMockAcceptsURL struct {
	mock               *DownloaderMock
	defaultExpectation *DownloaderMockAcceptsURLExpectation
	expectations       []*DownloaderMockAcceptsURLExpectation

	callArgs []*DownloaderMockAcceptsURLParams
	mutex    sync.RWMutex
}

// DownloaderMockAcceptsURLExpectation specifies expectation struct of the Downloader.AcceptsURL
type DownloaderMockAcceptsURLExpectation struct {
	mock    *DownloaderMock
	params  *DownloaderMockAcceptsURLParams
	results *DownloaderMockAcceptsURLResults
	Counter uint64
}

// DownloaderMockAcceptsURLParams contains parameters of the Downloader.AcceptsURL
type DownloaderMockAcceptsURLParams struct {
	url string
}

// DownloaderMockAcceptsURLResults contains results of the Downloader.AcceptsURL
type DownloaderMockAcceptsURLResults struct {
	b1 bool
}

// Expect sets up expected params for Downloader.AcceptsURL
func (mmAcceptsURL *mDownloaderMockAcceptsURL) Expect(url string) *mDownloaderMockAcceptsURL {
	if mmAcceptsURL.mock.funcAcceptsURL != nil {
		mmAcceptsURL.mock.t.Fatalf("DownloaderMock.AcceptsURL mock is already set by Set")
	}

	if mmAcceptsURL.defaultExpectation == nil {
		mmAcceptsURL.defaultExpectation = &DownloaderMockAcceptsURLExpectation{}
	}

	mmAcceptsURL.defaultExpectation.params = &DownloaderMockAcceptsURLParams{url}
	for _, e := range mmAcceptsURL.expectations {
		if minimock.Equal(e.params, mmAcceptsURL.defaultExpectation.params) {
			mmAcceptsURL.mock.t.Fatalf("Expectation set by When has same params: %#v", *mmAcceptsURL.defaultExpectation.params)
		}
	}

	return mmAcceptsURL
}

// Inspect accepts an inspector function that has same arguments as the Downloader.AcceptsURL
func (mmAcceptsURL *mDownloaderMockAcceptsURL) Inspect(f func(url string)) *mDownloaderMockAcceptsURL {
	if mmAcceptsURL.mock.inspectFuncAcceptsURL != nil {
		mmAcceptsURL.mock.t.Fatalf("Inspect function is already set for DownloaderMock.AcceptsURL")
	}

	mmAcceptsURL.mock.inspectFuncAcceptsURL = f

	return mmAcceptsURL
}

// Return sets up results that will be returned by Downloader.AcceptsURL
func (mmAcceptsURL *mDownloaderMockAcceptsURL) Return(b1 bool) *DownloaderMock {
	if mmAcceptsURL.mock.funcAcceptsURL != nil {
		mmAcceptsURL.mock.t.Fatalf("DownloaderMock.AcceptsURL mock is already set by Set")
	}

	if mmAcceptsURL.defaultExpectation == nil {
		mmAcceptsURL.defaultExpectation = &DownloaderMockAcceptsURLExpectation{mock: mmAcceptsURL.mock}
	}
	mmAcceptsURL.defaultExpectation.results = &DownloaderMockAcceptsURLResults{b1}
	return mmAcceptsURL.mock
}

//Set uses given function f to mock the Downloader.AcceptsURL method
func (mmAcceptsURL *mDownloaderMockAcceptsURL) Set(f func(url string) (b1 bool)) *DownloaderMock {
	if mmAcceptsURL.defaultExpectation != nil {
		mmAcceptsURL.mock.t.Fatalf("Default expectation is already set for the Downloader.AcceptsURL method")
	}

	if len(mmAcceptsURL.expectations) > 0 {
		mmAcceptsURL.mock.t.Fatalf("Some expectations are already set for the Downloader.AcceptsURL method")
	}

	mmAcceptsURL.mock.funcAcceptsURL = f
	return mmAcceptsURL.mock
}

// When sets expectation for the Downloader.AcceptsURL which will trigger the result defined by the following
// Then helper
func (mmAcceptsURL *mDownloaderMockAcceptsURL) When(url string) *DownloaderMockAcceptsURLExpectation {
	if mmAcceptsURL.mock.funcAcceptsURL != nil {
		mmAcceptsURL.mock.t.Fatalf("DownloaderMock.AcceptsURL mock is already set by Set")
	}

	expectation := &DownloaderMockAcceptsURLExpectation{
		mock:   mmAcceptsURL.mock,
		params: &DownloaderMockAcceptsURLParams{url},
	}
	mmAcceptsURL.expectations = append(mmAcceptsURL.expectations, expectation)
	return expectation
}

// Then sets up Downloader.AcceptsURL return parameters for the expectation previously defined by the When method
func (e *DownloaderMockAcceptsURLExpectation) Then(b1 bool) *DownloaderMock {
	e.results = &DownloaderMockAcceptsURLResults{b1}
	return e.mock
}

// AcceptsURL implements service.Downloader
func (mmAcceptsURL *DownloaderMock) AcceptsURL(url string) (b1 bool) {
	mm_atomic.AddUint64(&mmAcceptsURL.beforeAcceptsURLCounter, 1)
	defer mm_atomic.AddUint64(&mmAcceptsURL.afterAcceptsURLCounter, 1)

	if mmAcceptsURL.inspectFuncAcceptsURL != nil {
		mmAcceptsURL.inspectFuncAcceptsURL(url)
	}

	mm_params := &DownloaderMockAcceptsURLParams{url}

	// Record call args
	mmAcceptsURL.AcceptsURLMock.mutex.Lock()
	mmAcceptsURL.AcceptsURLMock.callArgs = append(mmAcceptsURL.AcceptsURLMock.callArgs, mm_params)
	mmAcceptsURL.AcceptsURLMock.mutex.Unlock()

	for _, e := range mmAcceptsURL.AcceptsURLMock.expectations {
		if minimock.Equal(e.params, mm_params) {
			mm_atomic.AddUint64(&e.Counter, 1)
			return e.results.b1
		}
	}

	if mmAcceptsURL.AcceptsURLMock.defaultExpectation != nil {
		mm_atomic.AddUint64(&mmAcceptsURL.AcceptsURLMock.defaultExpectation.Counter, 1)
		mm_want := mmAcceptsURL.AcceptsURLMock.defaultExpectation.params
		mm_got := DownloaderMockAcceptsURLParams{url}
		if mm_want != nil && !minimock.Equal(*mm_want, mm_got) {
			mmAcceptsURL.t.Errorf("DownloaderMock.AcceptsURL got unexpected parameters, want: %#v, got: %#v%s\n", *mm_want, mm_got, minimock.Diff(*mm_want, mm_got))
		}

		mm_results := mmAcceptsURL.AcceptsURLMock.defaultExpectation.results
		if mm_results == nil {
			mmAcceptsURL.t.Fatal("No results are set for the DownloaderMock.AcceptsURL")
		}
		return (*mm_results).b1
	}
	if mmAcceptsURL.funcAcceptsURL != nil {
		return mmAcceptsURL.funcAcceptsURL(url)
	}
	mmAcceptsURL.t.Fatalf("Unexpected call to DownloaderMock.AcceptsURL. %v", url)
	return
}

// AcceptsURLAfterCounter returns a count of finished DownloaderMock.AcceptsURL invocations
func (mmAcceptsURL *DownloaderMock) AcceptsURLAfterCounter() uint64 {
	return mm_atomic.LoadUint64(&mmAcceptsURL.afterAcceptsURLCounter)
}

// AcceptsURLBeforeCounter returns a count of DownloaderMock.AcceptsURL invocations
func (mmAcceptsURL *DownloaderMock) AcceptsURLBeforeCounter() uint64 {
	return mm_atomic.LoadUint64(&mmAcceptsURL.beforeAcceptsURLCounter)
}

// Calls returns a list of arguments used in each call to DownloaderMock.AcceptsURL.
// The list is in the same order as the calls were made (i.e. recent calls have a higher index)
func (mmAcceptsURL *mDownloaderMockAcceptsURL) Calls() []*DownloaderMockAcceptsURLParams {
	mmAcceptsURL.mutex.RLock()

	argCopy := make([]*DownloaderMockAcceptsURLParams, len(mmAcceptsURL.callArgs))
	copy(argCopy, mmAcceptsURL.callArgs)

	mmAcceptsURL.mutex.RUnlock()

	return argCopy
}

// MinimockAcceptsURLDone returns true if the count of the AcceptsURL invocations corresponds
// the number of defined expectations
func (m *DownloaderMock) MinimockAcceptsURLDone() bool {
	for _, e := range m.AcceptsURLMock.expectations {
		if mm_atomic.LoadUint64(&e.Counter) < 1 {
			return false
		}
	}

	// if default expectation was set then invocations count should be greater than zero
	if m.AcceptsURLMock.defaultExpectation != nil && mm_atomic.LoadUint64(&m.afterAcceptsURLCounter) < 1 {
		return false
	}
	// if func was set then invocations count should be greater than zero
	if m.funcAcceptsURL != nil && mm_atomic.LoadUint64(&m.afterAcceptsURLCounter) < 1 {
		return false
	}
	return true
}

// MinimockAcceptsURLInspect logs each unmet expectation
func (m *DownloaderMock) MinimockAcceptsURLInspect() {
	for _, e := range m.AcceptsURLMock.expectations {
		if mm_atomic.LoadUint64(&e.Counter) < 1 {
			m.t.Errorf("Expected call to DownloaderMock.AcceptsURL with params: %#v", *e.params)
		}
	}

	// if default expectation was set then invocations count should be greater than zero
	if m.AcceptsURLMock.defaultExpectation != nil && mm_atomic.LoadUint64(&m.afterAcceptsURLCounter) < 1 {
		if m.AcceptsURLMock.defaultExpectation.params == nil {
			m.t.Error("Expected call to DownloaderMock.AcceptsURL")
		} else {
			m.t.Errorf("Expected call to DownloaderMock.AcceptsURL with params: %#v", *m.AcceptsURLMock.defaultExpectation.params)
		}
	}
	// if func was set then invocations count should be greater than zero
	if m.funcAcceptsURL != nil && mm_atomic.LoadUint64(&m.afterAcceptsURLCounter) < 1 {
		m.t.Error("Expected call to DownloaderMock.AcceptsURL")
	}
}

type mDownloaderMockDownload struct {
	mock               *DownloaderMock
	defaultExpectation *DownloaderMockDownloadExpectation
	expectations       []*DownloaderMockDownloadExpectation

	callArgs []*DownloaderMockDownloadParams
	mutex    sync.RWMutex
}

// DownloaderMockDownloadExpectation specifies expectation struct of the Downloader.Download
type DownloaderMockDownloadExpectation struct {
	mock    *DownloaderMock
	params  *DownloaderMockDownloadParams
	results *DownloaderMockDownloadResults
	Counter uint64
}

// DownloaderMockDownloadParams contains parameters of the Downloader.Download
type DownloaderMockDownloadParams struct {
	ctx       context.Context
	url       string
	filepaths []string
}

// DownloaderMockDownloadResults contains results of the Downloader.Download
type DownloaderMockDownloadResults struct {
	filepathsMap map[string]string
	err          error
}

// Expect sets up expected params for Downloader.Download
func (mmDownload *mDownloaderMockDownload) Expect(ctx context.Context, url string, filepaths []string) *mDownloaderMockDownload {
	if mmDownload.mock.funcDownload != nil {
		mmDownload.mock.t.Fatalf("DownloaderMock.Download mock is already set by Set")
	}

	if mmDownload.defaultExpectation == nil {
		mmDownload.defaultExpectation = &DownloaderMockDownloadExpectation{}
	}

	mmDownload.defaultExpectation.params = &DownloaderMockDownloadParams{ctx, url, filepaths}
	for _, e := range mmDownload.expectations {
		if minimock.Equal(e.params, mmDownload.defaultExpectation.params) {
			mmDownload.mock.t.Fatalf("Expectation set by When has same params: %#v", *mmDownload.defaultExpectation.params)
		}
	}

	return mmDownload
}

// Inspect accepts an inspector function that has same arguments as the Downloader.Download
func (mmDownload *mDownloaderMockDownload) Inspect(f func(ctx context.Context, url string, filepaths []string)) *mDownloaderMockDownload {
	if mmDownload.mock.inspectFuncDownload != nil {
		mmDownload.mock.t.Fatalf("Inspect function is already set for DownloaderMock.Download")
	}

	mmDownload.mock.inspectFuncDownload = f

	return mmDownload
}

// Return sets up results that will be returned by Downloader.Download
func (mmDownload *mDownloaderMockDownload) Return(filepathsMap map[string]string, err error) *DownloaderMock {
	if mmDownload.mock.funcDownload != nil {
		mmDownload.mock.t.Fatalf("DownloaderMock.Download mock is already set by Set")
	}

	if mmDownload.defaultExpectation == nil {
		mmDownload.defaultExpectation = &DownloaderMockDownloadExpectation{mock: mmDownload.mock}
	}
	mmDownload.defaultExpectation.results = &DownloaderMockDownloadResults{filepathsMap, err}
	return mmDownload.mock
}

//Set uses given function f to mock the Downloader.Download method
func (mmDownload *mDownloaderMockDownload) Set(f func(ctx context.Context, url string, filepaths []string) (filepathsMap map[string]string, err error)) *DownloaderMock {
	if mmDownload.defaultExpectation != nil {
		mmDownload.mock.t.Fatalf("Default expectation is already set for the Downloader.Download method")
	}

	if len(mmDownload.expectations) > 0 {
		mmDownload.mock.t.Fatalf("Some expectations are already set for the Downloader.Download method")
	}

	mmDownload.mock.funcDownload = f
	return mmDownload.mock
}

// When sets expectation for the Downloader.Download which will trigger the result defined by the following
// Then helper
func (mmDownload *mDownloaderMockDownload) When(ctx context.Context, url string, filepaths []string) *DownloaderMockDownloadExpectation {
	if mmDownload.mock.funcDownload != nil {
		mmDownload.mock.t.Fatalf("DownloaderMock.Download mock is already set by Set")
	}

	expectation := &DownloaderMockDownloadExpectation{
		mock:   mmDownload.mock,
		params: &DownloaderMockDownloadParams{ctx, url, filepaths},
	}
	mmDownload.expectations = append(mmDownload.expectations, expectation)
	return expectation
}

// Then sets up Downloader.Download return parameters for the expectation previously defined by the When method
func (e *DownloaderMockDownloadExpectation) Then(filepathsMap map[string]string, err error) *DownloaderMock {
	e.results = &DownloaderMockDownloadResults{filepathsMap, err}
	return e.mock
}

// Download implements service.Downloader
func (mmDownload *DownloaderMock) Download(ctx context.Context, url string, filepaths []string) (filepathsMap map[string]string, err error) {
	mm_atomic.AddUint64(&mmDownload.beforeDownloadCounter, 1)
	defer mm_atomic.AddUint64(&mmDownload.afterDownloadCounter, 1)

	if mmDownload.inspectFuncDownload != nil {
		mmDownload.inspectFuncDownload(ctx, url, filepaths)
	}

	mm_params := &DownloaderMockDownloadParams{ctx, url, filepaths}

	// Record call args
	mmDownload.DownloadMock.mutex.Lock()
	mmDownload.DownloadMock.callArgs = append(mmDownload.DownloadMock.callArgs, mm_params)
	mmDownload.DownloadMock.mutex.Unlock()

	for _, e := range mmDownload.DownloadMock.expectations {
		if minimock.Equal(e.params, mm_params) {
			mm_atomic.AddUint64(&e.Counter, 1)
			return e.results.filepathsMap, e.results.err
		}
	}

	if mmDownload.DownloadMock.defaultExpectation != nil {
		mm_atomic.AddUint64(&mmDownload.DownloadMock.defaultExpectation.Counter, 1)
		mm_want := mmDownload.DownloadMock.defaultExpectation.params
		mm_got := DownloaderMockDownloadParams{ctx, url, filepaths}
		if mm_want != nil && !minimock.Equal(*mm_want, mm_got) {
			mmDownload.t.Errorf("DownloaderMock.Download got unexpected parameters, want: %#v, got: %#v%s\n", *mm_want, mm_got, minimock.Diff(*mm_want, mm_got))
		}

		mm_results := mmDownload.DownloadMock.defaultExpectation.results
		if mm_results == nil {
			mmDownload.t.Fatal("No results are set for the DownloaderMock.Download")
		}
		return (*mm_results).filepathsMap, (*mm_results).err
	}
	if mmDownload.funcDownload != nil {
		return mmDownload.funcDownload(ctx, url, filepaths)
	}
	mmDownload.t.Fatalf("Unexpected call to DownloaderMock.Download. %v %v %v", ctx, url, filepaths)
	return
}

// DownloadAfterCounter returns a count of finished DownloaderMock.Download invocations
func (mmDownload *DownloaderMock) DownloadAfterCounter() uint64 {
	return mm_atomic.LoadUint64(&mmDownload.afterDownloadCounter)
}

// DownloadBeforeCounter returns a count of DownloaderMock.Download invocations
func (mmDownload *DownloaderMock) DownloadBeforeCounter() uint64 {
	return mm_atomic.LoadUint64(&mmDownload.beforeDownloadCounter)
}

// Calls returns a list of arguments used in each call to DownloaderMock.Download.
// The list is in the same order as the calls were made (i.e. recent calls have a higher index)
func (mmDownload *mDownloaderMockDownload) Calls() []*DownloaderMockDownloadParams {
	mmDownload.mutex.RLock()

	argCopy := make([]*DownloaderMockDownloadParams, len(mmDownload.callArgs))
	copy(argCopy, mmDownload.callArgs)

	mmDownload.mutex.RUnlock()

	return argCopy
}

// MinimockDownloadDone returns true if the count of the Download invocations corresponds
// the number of defined expectations
func (m *DownloaderMock) MinimockDownloadDone() bool {
	for _, e := range m.DownloadMock.expectations {
		if mm_atomic.LoadUint64(&e.Counter) < 1 {
			return false
		}
	}

	// if default expectation was set then invocations count should be greater than zero
	if m.DownloadMock.defaultExpectation != nil && mm_atomic.LoadUint64(&m.afterDownloadCounter) < 1 {
		return false
	}
	// if func was set then invocations count should be greater than zero
	if m.funcDownload != nil && mm_atomic.LoadUint64(&m.afterDownloadCounter) < 1 {
		return false
	}
	return true
}

// MinimockDownloadInspect logs each unmet expectation
func (m *DownloaderMock) MinimockDownloadInspect() {
	for _, e := range m.DownloadMock.expectations {
		if mm_atomic.LoadUint64(&e.Counter) < 1 {
			m.t.Errorf("Expected call to DownloaderMock.Download with params: %#v", *e.params)
		}
	}

	// if default expectation was set then invocations count should be greater than zero
	if m.DownloadMock.defaultExpectation != nil && mm_atomic.LoadUint64(&m.afterDownloadCounter) < 1 {
		if m.DownloadMock.defaultExpectation.params == nil {
			m.t.Error("Expected call to DownloaderMock.Download")
		} else {
			m.t.Errorf("Expected call to DownloaderMock.Download with params: %#v", *m.DownloadMock.defaultExpectation.params)
		}
	}
	// if func was set then invocations count should be greater than zero
	if m.funcDownload != nil && mm_atomic.LoadUint64(&m.afterDownloadCounter) < 1 {
		m.t.Error("Expected call to DownloaderMock.Download")
	}
}

type mDownloaderMockGetMetadata struct {
	mock               *DownloaderMock
	defaultExpectation *DownloaderMockGetMetadataExpectation
	expectations       []*DownloaderMockGetMetadataExpectation

	callArgs []*DownloaderMockGetMetadataParams
	mutex    sync.RWMutex
}

// DownloaderMockGetMetadataExpectation specifies expectation struct of the Downloader.GetMetadata
type DownloaderMockGetMetadataExpectation struct {
	mock    *DownloaderMock
	params  *DownloaderMockGetMetadataParams
	results *DownloaderMockGetMetadataResults
	Counter uint64
}

// DownloaderMockGetMetadataParams contains parameters of the Downloader.GetMetadata
type DownloaderMockGetMetadataParams struct {
	ctx context.Context
	url string
}

// DownloaderMockGetMetadataResults contains results of the Downloader.GetMetadata
type DownloaderMockGetMetadataResults struct {
	mp1 *mm_service.Metadata
	err error
}

// Expect sets up expected params for Downloader.GetMetadata
func (mmGetMetadata *mDownloaderMockGetMetadata) Expect(ctx context.Context, url string) *mDownloaderMockGetMetadata {
	if mmGetMetadata.mock.funcGetMetadata != nil {
		mmGetMetadata.mock.t.Fatalf("DownloaderMock.GetMetadata mock is already set by Set")
	}

	if mmGetMetadata.defaultExpectation == nil {
		mmGetMetadata.defaultExpectation = &DownloaderMockGetMetadataExpectation{}
	}

	mmGetMetadata.defaultExpectation.params = &DownloaderMockGetMetadataParams{ctx, url}
	for _, e := range mmGetMetadata.expectations {
		if minimock.Equal(e.params, mmGetMetadata.defaultExpectation.params) {
			mmGetMetadata.mock.t.Fatalf("Expectation set by When has same params: %#v", *mmGetMetadata.defaultExpectation.params)
		}
	}

	return mmGetMetadata
}

// Inspect accepts an inspector function that has same arguments as the Downloader.GetMetadata
func (mmGetMetadata *mDownloaderMockGetMetadata) Inspect(f func(ctx context.Context, url string)) *mDownloaderMockGetMetadata {
	if mmGetMetadata.mock.inspectFuncGetMetadata != nil {
		mmGetMetadata.mock.t.Fatalf("Inspect function is already set for DownloaderMock.GetMetadata")
	}

	mmGetMetadata.mock.inspectFuncGetMetadata = f

	return mmGetMetadata
}

// Return sets up results that will be returned by Downloader.GetMetadata
func (mmGetMetadata *mDownloaderMockGetMetadata) Return(mp1 *mm_service.Metadata, err error) *DownloaderMock {
	if mmGetMetadata.mock.funcGetMetadata != nil {
		mmGetMetadata.mock.t.Fatalf("DownloaderMock.GetMetadata mock is already set by Set")
	}

	if mmGetMetadata.defaultExpectation == nil {
		mmGetMetadata.defaultExpectation = &DownloaderMockGetMetadataExpectation{mock: mmGetMetadata.mock}
	}
	mmGetMetadata.defaultExpectation.results = &DownloaderMockGetMetadataResults{mp1, err}
	return mmGetMetadata.mock
}

//Set uses given function f to mock the Downloader.GetMetadata method
func (mmGetMetadata *mDownloaderMockGetMetadata) Set(f func(ctx context.Context, url string) (mp1 *mm_service.Metadata, err error)) *DownloaderMock {
	if mmGetMetadata.defaultExpectation != nil {
		mmGetMetadata.mock.t.Fatalf("Default expectation is already set for the Downloader.GetMetadata method")
	}

	if len(mmGetMetadata.expectations) > 0 {
		mmGetMetadata.mock.t.Fatalf("Some expectations are already set for the Downloader.GetMetadata method")
	}

	mmGetMetadata.mock.funcGetMetadata = f
	return mmGetMetadata.mock
}

// When sets expectation for the Downloader.GetMetadata which will trigger the result defined by the following
// Then helper
func (mmGetMetadata *mDownloaderMockGetMetadata) When(ctx context.Context, url string) *DownloaderMockGetMetadataExpectation {
	if mmGetMetadata.mock.funcGetMetadata != nil {
		mmGetMetadata.mock.t.Fatalf("DownloaderMock.GetMetadata mock is already set by Set")
	}

	expectation := &DownloaderMockGetMetadataExpectation{
		mock:   mmGetMetadata.mock,
		params: &DownloaderMockGetMetadataParams{ctx, url},
	}
	mmGetMetadata.expectations = append(mmGetMetadata.expectations, expectation)
	return expectation
}

// Then sets up Downloader.GetMetadata return parameters for the expectation previously defined by the When method
func (e *DownloaderMockGetMetadataExpectation) Then(mp1 *mm_service.Metadata, err error) *DownloaderMock {
	e.results = &DownloaderMockGetMetadataResults{mp1, err}
	return e.mock
}

// GetMetadata implements service.Downloader
func (mmGetMetadata *DownloaderMock) GetMetadata(ctx context.Context, url string) (mp1 *mm_service.Metadata, err error) {
	mm_atomic.AddUint64(&mmGetMetadata.beforeGetMetadataCounter, 1)
	defer mm_atomic.AddUint64(&mmGetMetadata.afterGetMetadataCounter, 1)

	if mmGetMetadata.inspectFuncGetMetadata != nil {
		mmGetMetadata.inspectFuncGetMetadata(ctx, url)
	}

	mm_params := &DownloaderMockGetMetadataParams{ctx, url}

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
		mm_got := DownloaderMockGetMetadataParams{ctx, url}
		if mm_want != nil && !minimock.Equal(*mm_want, mm_got) {
			mmGetMetadata.t.Errorf("DownloaderMock.GetMetadata got unexpected parameters, want: %#v, got: %#v%s\n", *mm_want, mm_got, minimock.Diff(*mm_want, mm_got))
		}

		mm_results := mmGetMetadata.GetMetadataMock.defaultExpectation.results
		if mm_results == nil {
			mmGetMetadata.t.Fatal("No results are set for the DownloaderMock.GetMetadata")
		}
		return (*mm_results).mp1, (*mm_results).err
	}
	if mmGetMetadata.funcGetMetadata != nil {
		return mmGetMetadata.funcGetMetadata(ctx, url)
	}
	mmGetMetadata.t.Fatalf("Unexpected call to DownloaderMock.GetMetadata. %v %v", ctx, url)
	return
}

// GetMetadataAfterCounter returns a count of finished DownloaderMock.GetMetadata invocations
func (mmGetMetadata *DownloaderMock) GetMetadataAfterCounter() uint64 {
	return mm_atomic.LoadUint64(&mmGetMetadata.afterGetMetadataCounter)
}

// GetMetadataBeforeCounter returns a count of DownloaderMock.GetMetadata invocations
func (mmGetMetadata *DownloaderMock) GetMetadataBeforeCounter() uint64 {
	return mm_atomic.LoadUint64(&mmGetMetadata.beforeGetMetadataCounter)
}

// Calls returns a list of arguments used in each call to DownloaderMock.GetMetadata.
// The list is in the same order as the calls were made (i.e. recent calls have a higher index)
func (mmGetMetadata *mDownloaderMockGetMetadata) Calls() []*DownloaderMockGetMetadataParams {
	mmGetMetadata.mutex.RLock()

	argCopy := make([]*DownloaderMockGetMetadataParams, len(mmGetMetadata.callArgs))
	copy(argCopy, mmGetMetadata.callArgs)

	mmGetMetadata.mutex.RUnlock()

	return argCopy
}

// MinimockGetMetadataDone returns true if the count of the GetMetadata invocations corresponds
// the number of defined expectations
func (m *DownloaderMock) MinimockGetMetadataDone() bool {
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
func (m *DownloaderMock) MinimockGetMetadataInspect() {
	for _, e := range m.GetMetadataMock.expectations {
		if mm_atomic.LoadUint64(&e.Counter) < 1 {
			m.t.Errorf("Expected call to DownloaderMock.GetMetadata with params: %#v", *e.params)
		}
	}

	// if default expectation was set then invocations count should be greater than zero
	if m.GetMetadataMock.defaultExpectation != nil && mm_atomic.LoadUint64(&m.afterGetMetadataCounter) < 1 {
		if m.GetMetadataMock.defaultExpectation.params == nil {
			m.t.Error("Expected call to DownloaderMock.GetMetadata")
		} else {
			m.t.Errorf("Expected call to DownloaderMock.GetMetadata with params: %#v", *m.GetMetadataMock.defaultExpectation.params)
		}
	}
	// if func was set then invocations count should be greater than zero
	if m.funcGetMetadata != nil && mm_atomic.LoadUint64(&m.afterGetMetadataCounter) < 1 {
		m.t.Error("Expected call to DownloaderMock.GetMetadata")
	}
}

// MinimockFinish checks that all mocked methods have been called the expected number of times
func (m *DownloaderMock) MinimockFinish() {
	if !m.minimockDone() {
		m.MinimockAcceptsURLInspect()

		m.MinimockDownloadInspect()

		m.MinimockGetMetadataInspect()
		m.t.FailNow()
	}
}

// MinimockWait waits for all mocked methods to be called the expected number of times
func (m *DownloaderMock) MinimockWait(timeout mm_time.Duration) {
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

func (m *DownloaderMock) minimockDone() bool {
	done := true
	return done &&
		m.MinimockAcceptsURLDone() &&
		m.MinimockDownloadDone() &&
		m.MinimockGetMetadataDone()
}