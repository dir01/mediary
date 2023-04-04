package service_test

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/dir01/mediary/service"
	"github.com/dir01/mediary/service/mocks"
	"github.com/gojuno/minimock/v3"
	"go.uber.org/zap"
)

var logger = zap.NewNop()

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
			ExpectedError:      errors.New("failed to get metadata: url not supported"),
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
				SubscribeMock.Set(func(ctx context.Context, jobType string, f1 func(payloadBytes []byte) error) {}).
				PublishMock.Set(func(ctx context.Context, jobType string, payload any) (err error) { return nil }).
				RunMock.Set(func() {}).
				ShutdownMock.Set(func() {})
			svc := service.NewService(dwn, storage, queue, nil, nil, logger)
			svc.Start()
			defer svc.Stop()

			storage.GetMetadataMock.Set(func(ctx context.Context, u string) (r *service.Metadata, err error) {
				if u != url {
					t.Fatalf("expected url %s, got %s", url, u)
				}
				return tc.StorageGetResponse, tc.StorageGetError
			})

			if tc.DownloaderNotFound {
				dwn.AcceptsURLMock.Return(false)
			} else {
				dwn.AcceptsURLMock.Return(true)
			}

			dwn.GetMetadataMock.Set(func(ctx context.Context, u string) (r *service.Metadata, err error) {
				if u != url {
					t.Fatalf("expected url %s, got %s", url, u)
				}
				return tc.DownloaderResponse, tc.DownloaderError
			})

			storage.SaveMetadataMock.Return(tc.StorageSaveResponse)

			result, err := svc.GetMetadata(context.TODO(), url)

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
