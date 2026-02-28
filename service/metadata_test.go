package service_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"reflect"
	"testing"

	"github.com/dir01/mediary/service"
	"github.com/dir01/mediary/service/mocks"
	"github.com/gojuno/minimock/v3"
)

var logger = slog.New(slog.NewTextHandler(io.Discard, nil))

func TestGetMetadata(t *testing.T) {
	type testCase struct {
		Name string

		StorageGetResponse *service.Metadata
		StorageGetError    error

		DownloaderAcceptsURL bool
		DownloaderResponse   *service.Metadata
		DownloaderError      error

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
			Name:                 "storage get returns nil, downloader responds",
			StorageGetResponse:   nil,
			DownloaderAcceptsURL: true,
			DownloaderResponse:   &service.Metadata{Name: "some-name"},
			ExpectedResponse:     &service.Metadata{Name: "some-name"},
		},
		{
			Name:                 "storage get errors, downloader responds",
			StorageGetError:      fmt.Errorf("some-error"),
			DownloaderAcceptsURL: true,
			DownloaderResponse:   &service.Metadata{Name: "some-name"},
			ExpectedResponse:     &service.Metadata{Name: "some-name"},
		},
		{
			Name:                 "storage get errors, downloader responds, storage set errors",
			StorageGetError:      fmt.Errorf("some-error"),
			DownloaderAcceptsURL: true,
			DownloaderResponse:   &service.Metadata{Name: "some-name"},
			StorageSaveResponse:  fmt.Errorf("storage-save-error"),
			ExpectedResponse:     &service.Metadata{Name: "some-name"},
		},
		{
			Name:                 "storage get returns nil, downloader does not match",
			StorageGetResponse:   nil,
			DownloaderAcceptsURL: false,
			ExpectedResponse:     nil,
			ExpectedError:        errors.New("failed to get metadata: url not supported"),
		},
		{
			Name:                 "downloader errors",
			DownloaderAcceptsURL: true,
			DownloaderError:      fmt.Errorf("some-error"),
			ExpectedError:        fmt.Errorf("error getting metadata from downloader: some-error"),
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			mc := minimock.NewController(t)
			dwn := mocks.NewDownloaderMock(mc)
			storage := mocks.NewStorageMock(mc)
			queue := mocks.NewJobsQueueMock(mc)
			queue.
				SubscribeMock.Optional().Set(func(ctx context.Context, jobType string, f1 func(context.Context, []byte) error) {}).
				RunMock.Optional().Set(func() {}).
				PublishMock.Optional().Set(func(ctx context.Context, jobType string, payload any) (err error) { return nil }).
				ShutdownMock.Optional().Set(func() {})

			svc := service.NewService(dwn, storage, queue, nil, nil, logger)
			svc.Start()
			defer svc.Stop()

			storage.
				GetMetadataMock.Optional().Set(func(ctx context.Context, u string) (r *service.Metadata, err error) {
				if u != url {
					t.Fatalf("expected url %s, got %s", url, u)
				}
				return tc.StorageGetResponse, tc.StorageGetError
			})

			if tc.DownloaderAcceptsURL {
				dwn.AcceptsURLMock.Optional().Return(true)
			} else {
				dwn.AcceptsURLMock.Optional().Return(false)
			}

			if tc.DownloaderResponse != nil || tc.DownloaderError != nil {
				dwn.GetMetadataMock.Set(func(ctx context.Context, u string) (r *service.Metadata, err error) {
					if u != url {
						t.Fatalf("expected url %s, got %s", url, u)
					}
					return tc.DownloaderResponse, tc.DownloaderError
				})
			}

			storage.SaveMetadataMock.Optional().Return(tc.StorageSaveResponse)

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
