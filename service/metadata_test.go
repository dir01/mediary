package service_test

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/dir01/mediary/downloader"
	"github.com/dir01/mediary/service"
	"github.com/dir01/mediary/service/mocks"
	"github.com/gojuno/minimock/v3"
	"go.uber.org/zap"
)

var logger, _ = zap.NewDevelopment()

func TestGetMetadata(t *testing.T) {
	type testCase struct {
		Name string

		StorageGetResponse *service.Metadata
		StorageGetError    error

		DownloaderNotFound bool
		DownloaderResponse *service.Metadata
		DownloaderError    error

		StorageSaveResponse error

		ExpectedResponse *service.Metadata
		ExpectedError    error
	}

	someContext := context.WithValue(context.Background(), "some-key", "some-value")
	url := "magnet:?xt=urn:btih:deadbeef"

	for _, tc := range []testCase{
		{
			Name:               "metadata is found in storage",
			StorageGetResponse: &service.Metadata{Name: "some-name"},
			ExpectedResponse:   &service.Metadata{Name: "some-name"},
		},
		{
			Name:               "storage get returns nil, downloader responds",
			StorageGetResponse: nil,
			DownloaderResponse: &service.Metadata{Name: "some-name"},
			ExpectedResponse:   &service.Metadata{Name: "some-name"},
		},
		{
			Name:               "storage get errors, downloader responds",
			StorageGetError:    fmt.Errorf("some-error"),
			DownloaderResponse: &service.Metadata{Name: "some-name"},
			ExpectedResponse:   &service.Metadata{Name: "some-name"},
		},
		{
			Name:                "storage get errors, downloader responds, storage set errors",
			StorageGetError:     fmt.Errorf("some-error"),
			DownloaderResponse:  &service.Metadata{Name: "some-name"},
			StorageSaveResponse: fmt.Errorf("storage-save-error"),
			ExpectedResponse:    &service.Metadata{Name: "some-name"},
		},
		{
			Name:               "storage get returns nil, downloader does not match",
			StorageGetResponse: nil,
			DownloaderNotFound: true,
			ExpectedResponse:   nil,
			ExpectedError:      downloader.ErrUrlNotSupported,
		},
		{
			Name:            "downloader errors",
			DownloaderError: fmt.Errorf("some-error"),
			ExpectedError:   fmt.Errorf("error getting metadata from downloader: some-error"),
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			mc := minimock.NewController(t)
			dwn := mocks.NewDownloaderMock(mc)
			storage := mocks.NewStorageMock(mc)
			queue := mocks.NewJobsQueueMock(mc)
			queue.
				SubscribeMock.Set(func(f1 func(jobId string) error) {}).
				PublishMock.Set(func(ctx context.Context, jobId string) (err error) { return nil })
			svc := service.NewService(dwn, storage, queue, nil, nil, logger)

			storage.GetMetadataMock.
				Expect(someContext, url).
				Return(tc.StorageGetResponse, tc.StorageGetError)

			if tc.DownloaderNotFound {
				dwn.AcceptsURLMock.Return(false)
			} else {
				dwn.AcceptsURLMock.Return(true)
			}

			dwn.GetMetadataMock.
				Expect(someContext, url).
				Return(tc.DownloaderResponse, tc.DownloaderError)

			storage.SaveMetadataMock.Return(tc.StorageSaveResponse)

			result, err := svc.GetMetadata(someContext, url)

			if tc.ExpectedError != nil {
				if err == nil || err.Error() != tc.ExpectedError.Error() {
					t.Errorf("expected error: %v, got: %v", tc.ExpectedError, err)
				}
			} else if tc.ExpectedError == nil && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if !reflect.DeepEqual(result, tc.ExpectedResponse) {
				t.Errorf("expected result %v, got %v", tc.ExpectedResponse, result)
			}
		})
	}
}
