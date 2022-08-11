package mediary_test

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/dir01/mediary"
	"github.com/dir01/mediary/mocks"
	"github.com/gojuno/minimock/v3"
)

func TestGetMetadata(t *testing.T) {
	type testCase struct {
		Name string

		StorageGetResponse *mediary.Metadata
		StorageGetError    error

		DownloaderNotFound bool
		DownloaderResponse *mediary.Metadata
		DownloaderError    error

		StorageSaveResponse error

		ExpectedResponse *mediary.Metadata
		ExpectedError    error
	}

	someContext := context.WithValue(context.Background(), "some-key", "some-value")
	url := "magnet:?xt=urn:btih:deadbeef"

	for _, tc := range []testCase{
		{
			Name:               "metadata is found in storage",
			StorageGetResponse: &mediary.Metadata{Name: "some-name"},
			ExpectedResponse:   &mediary.Metadata{Name: "some-name"},
		},
		{
			Name:               "storage get returns nil, downloader responds",
			StorageGetResponse: nil,
			DownloaderResponse: &mediary.Metadata{Name: "some-name"},
			ExpectedResponse:   &mediary.Metadata{Name: "some-name"},
		},
		{
			Name:               "storage get errors, downloader responds",
			StorageGetError:    fmt.Errorf("some-error"),
			DownloaderResponse: &mediary.Metadata{Name: "some-name"},
			ExpectedResponse:   &mediary.Metadata{Name: "some-name"},
		},
		{
			Name:                "storage get errors, downloader responds, storage set errors",
			StorageGetError:     fmt.Errorf("some-error"),
			DownloaderResponse:  &mediary.Metadata{Name: "some-name"},
			StorageSaveResponse: fmt.Errorf("storage-save-error"),
			ExpectedResponse:    &mediary.Metadata{Name: "some-name"},
		},
		{
			Name:               "storage get returns nil, downloader does not match",
			StorageGetResponse: nil,
			DownloaderNotFound: true,
			ExpectedResponse:   nil,
			ExpectedError:      fmt.Errorf("no downloader found for url: %s", url),
		},
		{
			Name:            "downloader errors",
			DownloaderError: fmt.Errorf("some-error"),
			ExpectedError:   fmt.Errorf("error getting metadata from downloader: some-error"),
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			mc := minimock.NewController(t)
			downloader := mocks.NewDownloaderMock(mc)
			storage := mocks.NewStorageMock(mc)
			service := mediary.NewService([]mediary.Downloader{downloader}, storage)

			storage.GetMetadataMock.
				Expect(someContext, url).
				Return(tc.StorageGetResponse, tc.StorageGetError)

			if tc.DownloaderNotFound {
				downloader.MatchesMock.Return(false)
			} else {
				downloader.MatchesMock.Return(true)
			}

			downloader.GetMetadataMock.
				Expect(someContext, url).
				Return(tc.DownloaderResponse, tc.DownloaderError)

			storage.SaveMetadataMock.Return(tc.StorageSaveResponse)

			result, err := service.GetMetadata(someContext, url)

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
