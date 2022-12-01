//go:build gen_docs
// +build gen_docs

package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/dir01/mediary/downloader/torrent"
	http2 "github.com/dir01/mediary/http"
	"github.com/dir01/mediary/media_processor"
	"github.com/dir01/mediary/service"
	"github.com/dir01/mediary/service/jobs_queue"
	"github.com/dir01/mediary/storage"
	"github.com/dir01/mediary/uploader"
	"github.com/go-redis/redis"
	"go.uber.org/zap"
)

var logger, _ = zap.NewDevelopment()

const (
	magnetURL      = "magnet:?xt=urn:btih:FB0B49D5E3E18E29868C680D2F7BC00D67987D56&tr=http%3A%2F%2Fbt.t-ru.org"
	testBucketName = "some-bucket"
)

func TestApplication(t *testing.T) {
	s3Client, teardownS3, err := getS3Client(context.Background(), testBucketName)
	defer teardownS3()
	if err != nil {
		t.Fatalf("error creating s3 client: %v", err)
	}

	downloader, err := torrent.NewTorrentDownloader(os.TempDir(), logger)
	if err != nil {
		t.Fatalf("error creating torrent downloader: %v", err)
	}

	redisURL, teardownRedis, err := GetFakeRedisURL(context.Background())
	defer teardownRedis()
	if err != nil {
		t.Fatalf("error getting redis url: %v", err)
	}
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		t.Fatalf("error parsing redis url: %v", err)
	}
	redisClient := redis.NewClient(opt)
	defer func() { _ = redisClient.Close() }()

	queue, err := jobs_queue.NewRedisJobsQueue(redisClient, logger, 10, "mediary:")
	if err != nil {
		t.Fatalf("error initializing redis jobs queue: %v", err)
	}

	store := storage.NewRedisStorage(redisClient, "mediary:")
	mediaProcessor, err := media_processor.NewFFMpegMediaProcessor(logger)
	if err != nil {
		t.Fatalf("error creating media processor: %v", err)
	}

	upl, err := uploader.New()
	if err != nil {
		t.Fatalf("error creating uploader: %v", err)
	}

	svc := service.NewService(downloader, store, queue, mediaProcessor, upl, logger)
	mux := http2.PrepareHTTPServerMux(svc)

	docs := newDocsHelper(
		t, mux, "../README.md",
		"<!-- start autogenerated samples -->",
		"<!-- stop autogenerated samples -->",
	)
	defer docs.Finish()

	t.Run("metadata with timeout", func(t *testing.T) {
		expectedResponse := `{"status": "accepted"}`
		docs.InsertText(`### '''/metadata''' - Timeouts

By default, the endpoint will timeout pretty quickly, 
probably sooner than it takes to fetch metadata of a torrent, for example.

In such cases, the endpoint will return a '''202 Accepted''' status code and a message '''%s'''

Feel free to repeat your request later: metadata is still being fetched in background.
`, expectedResponse)

		docs.PerformRequestForDocs("GET",
			`/metadata?url=`+magnetURL,
			nil,
			http.StatusAccepted,
			func(rr *httptest.ResponseRecorder) {
				if rr.Body.String() != expectedResponse {
					fmt.Println(rr.Body.String())
					t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expectedResponse)
				}
			},
		)
	})

	t.Run("metadata with long-polling", func(t *testing.T) {
		docs.InsertText(`### '''/metadata/long-polling'''

In case you'd rather wait for the metadata to be fetched, you can use the long-polling endpoint.

It will not return a response until the metadata is fetched.

There is still a timeout on the request, but it's pretty long (5 minutes).`)

		docs.PerformRequestForDocs("GET",
			`/metadata/long-polling?url=`+magnetURL,
			nil,
			http.StatusOK,
			func(rr *httptest.ResponseRecorder) {
				assertMatchesGoldenFile(t, rr.Body.Bytes(), "metadata_long_polling.json")
			},
		)
	})

	t.Run("cached metadata", func(t *testing.T) {
		docs.InsertText(`### '''/metadata''' - Cached

It goes without saying, that once the metadata is fetched, it is cached.

So all consecutive requests for the same URL will return the same metadata, and immediately.`)

		docs.PerformRequestForDocs(
			"GET",
			`/metadata?url=`+magnetURL,
			nil,
			http.StatusOK,
			func(rr *httptest.ResponseRecorder) {
				assertMatchesGoldenFile(t, rr.Body.Bytes(), "metadata_cached.json")
			},
		)
	})

	t.Run("job creation and status", func(t *testing.T) {
		docs.InsertText(`### '''/jobs''' 

POST to '''/jobs''' will schedule for background execution a process of downloading, converting/processing and uploading the media.
Only required parameters are '''url''' and '''type'''. '''type''' signifies the type of operation to be performed. 
Each operation can require some additional parameters, passed as '''params'''. For example, '''concatenate''' job
requires a list of files to be concatenated and, optionally, an '''audioCodec''' to be used for the output file.`)

		presignClient := s3.NewPresignClient(s3Client)
		presignResult, err := presignClient.PresignPutObject(context.TODO(), &s3.PutObjectInput{
			Bucket: aws.String(testBucketName),
			Key:    aws.String("some-path/some-file.some-ext"),
		})
		if err != nil {
			t.Fatal(fmt.Errorf("failed to presign: %w", err))
		}
		urlStr := presignResult.URL

		payload := strings.NewReader(fmt.Sprintf(`{
			"url": "%s",
			"type": "concatenate",
			"params": {
				"filepaths": [
					"01-001.mp3",
					"01-002.mp3"
				],
				"audioCodec": "mp3",
				"uploadUrl": "%s"
			}
		}`, "magnet:?xt=urn:btih:58C665647C1A34019A0DC99C9046BD459F006B73&tr=http%3A%2F%2Fbt3.t-ru.org", urlStr,
		))

		var jobID string
		docs.PerformRequestForDocs(
			"POST",
			"/jobs",
			payload,
			http.StatusAccepted,
			func(rr *httptest.ResponseRecorder) {
				var job struct {
					ID string `json:"id"`
				}
				err := json.Unmarshal(rr.Body.Bytes(), &job)
				if err != nil {
					t.Errorf("failed to unmarshal job ID: %s", err)
				}
				jobID = job.ID
			},
		)

		docs.InsertText(`### '''/jobs/:id'''

Since jobs can run for a long time, job creation api responds immediately with a job ID.
To check the status of the job, you can use the '''/jobs/:id''' endpoint.`)

		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Minute)
		defer cancel()

		var jobStatus string
		startTime := time.Now()
	loop:
		for {
			select {
			case <-ctx.Done():
				t.Errorf("job %s did not finish in time", jobID)
				break loop
			case <-time.After(100 * time.Millisecond):

				docs.PerformRequest("GET", "/jobs/"+jobID, nil, http.StatusOK, func(rr *httptest.ResponseRecorder) {
					var job struct {
						Status string `json:"status"`
					}
					err := json.Unmarshal(rr.Body.Bytes(), &job)
					if err != nil {
						t.Errorf("failed to unmarshal job ID: %s", err)
					}
					if job.Status != jobStatus {
						if jobStatus == "" {
							docs.InsertText("%s after starting the job:", time.Since(startTime).Round(time.Second))
						} else {
							docs.InsertText("%s later:", time.Since(startTime).Round(time.Second))
						}
						startTime = time.Now()
						jobStatus = job.Status
						docs.PerformRequestForDocs("GET", "/jobs/"+jobID, nil, http.StatusOK, nil)
					}
				})
				if jobStatus == "complete" {
					break loop
				}
			}
		}
	})
}
