# Mediary


## What is Mediary
Mediary is a service that can download, convert and upload media

Example use cases:

- Given a magnet link, download files A, B and C, glue them together and upload to this pre-signed S3 URL
- Given a YouTube video link, download audio, convert it to .mp3, and upload to this pre-signed S3 URL
- (to be done) Given a link to a single file, just take it and upload it to this pre-signed S3 URL


## API
- `GET /metadata` - returns metadata for a given media link. You are going to need it if you want 
to pick and choose which files should be processed.
- `POST /job` - creates a task to upload media. Describes the source URL, files at source URL
    to be processed, what transformation to apply and where to upload the result.
- `GET /job/{id}` - returns the status of a job.


## Examples
<!-- start autogenerated samples -->
### `/metadata` - Timeouts

By default, the endpoint will time out pretty quickly, 
probably sooner than it takes to fetch metadata of a torrent, for example.

In such cases, the endpoint will return a `202 Accepted` status code and a message `{"status": "accepted"}`

Feel free to repeat your request later: metadata is still being fetched in background.


```
$ curl -X GET '/metadata?url=magnet:?xt=urn:btih:FB0B49D5E3E18E29868C680D2F7BC00D67987D56&tr=http%3A%2F%2Fbt.t-ru.org'
{"status": "accepted"}
```


### `/metadata/long-polling`

In case you'd rather wait for the metadata to be fetched, you can use the long-polling endpoint.

It will not return a response until the metadata is fetched.

There is still a timeout on the request, but it's pretty long (5 minutes).

```
$ curl -X GET '/metadata/long-polling?url=magnet:?xt=urn:btih:FB0B49D5E3E18E29868C680D2F7BC00D67987D56&tr=http%3A%2F%2Fbt.t-ru.org'
{
  "url": "magnet:?xt=urn:btih:FB0B49D5E3E18E29868C680D2F7BC00D67987D56",
  "name": "Джо Диспенза - Медитации к Силе подсознания [Александр Шаронов]",
  "variants": [
    {
      "id": "вступление.mp3",
      "length_bytes": 1181881
    },
    {
      "id": "глава 1.mp3",
      "length_bytes": 40623850
    },
    {
      "id": "глава 2.mp3",
      "length_bytes": 42107250
    }
  ],
  "allow_multiple_variants": true,
  "downloader_name": "torrent"
}
```


### `/metadata` - Cached

It goes without saying, that once the metadata is fetched, it is cached.

So all consecutive requests for the same URL will return the same metadata, and immediately.

```
$ curl -X GET '/metadata?url=magnet:?xt=urn:btih:FB0B49D5E3E18E29868C680D2F7BC00D67987D56&tr=http%3A%2F%2Fbt.t-ru.org'
{
  "url": "magnet:?xt=urn:btih:FB0B49D5E3E18E29868C680D2F7BC00D67987D56",
  "name": "Джо Диспенза - Медитации к Силе подсознания [Александр Шаронов]",
  "variants": [
    {
      "id": "вступление.mp3",
      "length_bytes": 1181881
    },
    {
      "id": "глава 1.mp3",
      "length_bytes": 40623850
    },
    {
      "id": "глава 2.mp3",
      "length_bytes": 42107250
    }
  ],
  "allow_multiple_variants": true,
  "downloader_name": "torrent"
}
```


### `POST /metadata`

As you could've noticed, in previous calls part of the URL was lost.
To work around it, service also supports `POST` requests to `/metadata` endpoint.
In this case, you can pass the URL in the JSON body of the request.

```
$ curl -X POST '/metadata'--data-raw='{"url": "magnet:?xt=urn:btih:FB0B49D5E3E18E29868C680D2F7BC00D67987D56&tr=http%3A%2F%2Fbt.t-ru.org"}'
{
  "url": "magnet:?xt=urn:btih:FB0B49D5E3E18E29868C680D2F7BC00D67987D56\u0026tr=http%3A%2F%2Fbt.t-ru.org",
  "name": "Джо Диспенза - Медитации к Силе подсознания [Александр Шаронов]",
  "variants": [
    {
      "id": "вступление.mp3",
      "length_bytes": 1181881
    },
    {
      "id": "глава 1.mp3",
      "length_bytes": 40623850
    },
    {
      "id": "глава 2.mp3",
      "length_bytes": 42107250
    }
  ],
  "allow_multiple_variants": true,
  "downloader_name": "torrent"
}
```


### `/metadata` - YouTube

The endpoint also supports fetching metadata for YouTube videos.
Note that instead of file paths we get different options of desired formats:
Video, Audio, different qualities, etc.

This will allow you to choose the format you want to download later in the same UI as for torrent files.

Since it does not make sense to concatenate different versions of the same video,
response also will have `'"allow_multiple_files": false`. 
Take this into account while presenting format options to user

```
$ curl -X GET '/metadata?url=https://www.youtube.com/watch?v=kPN-uWB28X8'
{"status": "accepted"}
```


and then later

```
$ curl -X GET '/metadata?url=https://www.youtube.com/watch?v=kPN-uWB28X8'
{
  "url": "https://www.youtube.com/watch?v=kPN-uWB28X8",
  "name": "Miles Davis - Baby won't you please come home",
  "variants": [
    {
      "id": "Video (mp4)"
    },
    {
      "id": "Audio (mp3), High Quality"
    },
    {
      "id": "Audio (mp3), Medium Quality"
    },
    {
      "id": "Audio (mp3), Low Quality"
    }
  ],
  "allow_multiple_variants": false,
  "downloader_name": "ytdl"
}
```


### `/jobs` 

POST to `/jobs` will schedule for background execution a process of downloading, converting/processing and uploading the media.
Only required parameters are `url` and `type`. `type` signifies the type of operation to be performed. 
Each operation can require some additional parameters, passed as `params`. For example, `concatenate` job
requires a list of files to be concatenated and, optionally, an `audioCodec` to be used for the output file.

```
$ curl -X POST '/jobs'--data-raw='{
	"url": "magnet:?xt=urn:btih:58C665647C1A34019A0DC99C9046BD459F006B73&tr=http%3A%2F%2Fbt3.t-ru.org",
	"type": "concatenate",
	"params": {
		"variants": [
			"01-001.mp3",
			"01-002.mp3"
		],
		"audioCodec": "mp3",
		"uploadUrl": "http://localhost:63511/some-bucket/some-path/some-file.some-ext?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Credential=dummy%2F20250328%2Fus-east-1%2Fs3%2Faws4_request&X-Amz-Date=20250328T055057Z&X-Amz-Expires=900&X-Amz-Security-Token=dummy&X-Amz-SignedHeaders=host&x-id=PutObject&X-Amz-Signature=d5b57e052672ea95b24ea792b5191f588a45429e3cafc392774b759cd3f9dcc7"
	}
}'
{"status": "accepted", "id": "2416057c0c287c3d051da22dc0b307ec"}
```


### `/jobs/:id`

Since jobs can run for a long time, job creation api responds immediately with a job ID.
To check the status of the job, you can use the `/jobs/:id` endpoint.

0s after starting the job:

```
$ curl -X GET '/jobs/2416057c0c287c3d051da22dc0b307ec'
{
  "url": "magnet:?xt=urn:btih:58C665647C1A34019A0DC99C9046BD459F006B73\u0026tr=http%3A%2F%2Fbt3.t-ru.org",
  "type": "concatenate",
  "params": {
    "audioCodec": "mp3",
    "uploadUrl": "http://localhost:63511/some-bucket/some-path/some-file.some-ext?X-Amz-Algorithm=AWS4-HMAC-SHA256\u0026X-Amz-Credential=dummy%2F20250328%2Fus-east-1%2Fs3%2Faws4_request\u0026X-Amz-Date=20250328T055057Z\u0026X-Amz-Expires=900\u0026X-Amz-Security-Token=dummy\u0026X-Amz-SignedHeaders=host\u0026x-id=PutObject\u0026X-Amz-Signature=d5b57e052672ea95b24ea792b5191f588a45429e3cafc392774b759cd3f9dcc7",
    "variants": [
      "01-001.mp3",
      "01-002.mp3"
    ]
  },
  "id": "2416057c0c287c3d051da22dc0b307ec",
  "status": "created"
}
```


1s later:

```
$ curl -X GET '/jobs/2416057c0c287c3d051da22dc0b307ec'
{
  "url": "magnet:?xt=urn:btih:58C665647C1A34019A0DC99C9046BD459F006B73\u0026tr=http%3A%2F%2Fbt3.t-ru.org",
  "type": "concatenate",
  "params": {
    "audioCodec": "mp3",
    "uploadUrl": "http://localhost:63511/some-bucket/some-path/some-file.some-ext?X-Amz-Algorithm=AWS4-HMAC-SHA256\u0026X-Amz-Credential=dummy%2F20250328%2Fus-east-1%2Fs3%2Faws4_request\u0026X-Amz-Date=20250328T055057Z\u0026X-Amz-Expires=900\u0026X-Amz-Security-Token=dummy\u0026X-Amz-SignedHeaders=host\u0026x-id=PutObject\u0026X-Amz-Signature=d5b57e052672ea95b24ea792b5191f588a45429e3cafc392774b759cd3f9dcc7",
    "variants": [
      "01-001.mp3",
      "01-002.mp3"
    ]
  },
  "id": "2416057c0c287c3d051da22dc0b307ec",
  "status": "downloading"
}
```


21s later:

```
$ curl -X GET '/jobs/2416057c0c287c3d051da22dc0b307ec'
{
  "url": "magnet:?xt=urn:btih:58C665647C1A34019A0DC99C9046BD459F006B73\u0026tr=http%3A%2F%2Fbt3.t-ru.org",
  "type": "concatenate",
  "params": {
    "audioCodec": "mp3",
    "uploadUrl": "http://localhost:63511/some-bucket/some-path/some-file.some-ext?X-Amz-Algorithm=AWS4-HMAC-SHA256\u0026X-Amz-Credential=dummy%2F20250328%2Fus-east-1%2Fs3%2Faws4_request\u0026X-Amz-Date=20250328T055057Z\u0026X-Amz-Expires=900\u0026X-Amz-Security-Token=dummy\u0026X-Amz-SignedHeaders=host\u0026x-id=PutObject\u0026X-Amz-Signature=d5b57e052672ea95b24ea792b5191f588a45429e3cafc392774b759cd3f9dcc7",
    "variants": [
      "01-001.mp3",
      "01-002.mp3"
    ]
  },
  "id": "2416057c0c287c3d051da22dc0b307ec",
  "status": "processing"
}
```


23s later:

```
$ curl -X GET '/jobs/2416057c0c287c3d051da22dc0b307ec'
{
  "url": "magnet:?xt=urn:btih:58C665647C1A34019A0DC99C9046BD459F006B73\u0026tr=http%3A%2F%2Fbt3.t-ru.org",
  "type": "concatenate",
  "params": {
    "audioCodec": "mp3",
    "uploadUrl": "http://localhost:63511/some-bucket/some-path/some-file.some-ext?X-Amz-Algorithm=AWS4-HMAC-SHA256\u0026X-Amz-Credential=dummy%2F20250328%2Fus-east-1%2Fs3%2Faws4_request\u0026X-Amz-Date=20250328T055057Z\u0026X-Amz-Expires=900\u0026X-Amz-Security-Token=dummy\u0026X-Amz-SignedHeaders=host\u0026x-id=PutObject\u0026X-Amz-Signature=d5b57e052672ea95b24ea792b5191f588a45429e3cafc392774b759cd3f9dcc7",
    "variants": [
      "01-001.mp3",
      "01-002.mp3"
    ]
  },
  "id": "2416057c0c287c3d051da22dc0b307ec",
  "status": "uploading",
  "result_file_bytes": 42580566
}
```


1s later:

```
$ curl -X GET '/jobs/2416057c0c287c3d051da22dc0b307ec'
{
  "url": "magnet:?xt=urn:btih:58C665647C1A34019A0DC99C9046BD459F006B73\u0026tr=http%3A%2F%2Fbt3.t-ru.org",
  "type": "concatenate",
  "params": {
    "audioCodec": "mp3",
    "uploadUrl": "http://localhost:63511/some-bucket/some-path/some-file.some-ext?X-Amz-Algorithm=AWS4-HMAC-SHA256\u0026X-Amz-Credential=dummy%2F20250328%2Fus-east-1%2Fs3%2Faws4_request\u0026X-Amz-Date=20250328T055057Z\u0026X-Amz-Expires=900\u0026X-Amz-Security-Token=dummy\u0026X-Amz-SignedHeaders=host\u0026x-id=PutObject\u0026X-Amz-Signature=d5b57e052672ea95b24ea792b5191f588a45429e3cafc392774b759cd3f9dcc7",
    "variants": [
      "01-001.mp3",
      "01-002.mp3"
    ]
  },
  "id": "2416057c0c287c3d051da22dc0b307ec",
  "status": "complete",
  "result_file_bytes": 42580566
}
```


### Downloading YouTube audio

To download a YouTube video, you need to pass the URL of the video to the `/jobs` endpoint.

```
$ curl -X POST '/jobs'--data-raw='{
	"url": "https://www.youtube.com/watch?v=kPN-uWB28X8",
	"type": "upload_original",
	"params": {
		"variant": "Audio (mp3), Low Quality",
		"uploadUrl": "http://localhost:63511/some-bucket/some-path/some-file.some-ext?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Credential=dummy%2F20250328%2Fus-east-1%2Fs3%2Faws4_request&X-Amz-Date=20250328T055143Z&X-Amz-Expires=900&X-Amz-Security-Token=dummy&X-Amz-SignedHeaders=host&x-id=PutObject&X-Amz-Signature=390a7f86388332735e5889aa4d8fb6f43d9bbac9b18a8f95c3a530e732ea2c67"
	}
}'
{"status": "accepted", "id": "a581ebd433a361a159995bf3f58c5769"}
```

<!-- stop autogenerated samples -->

