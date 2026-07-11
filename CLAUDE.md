# mediary

Go service that downloads, converts, and uploads media (magnet/YouTube → mp3 → S3).
Used by tg-podcastotron. Remote: `git@github.com:dir01/mediary.git`.

## Dependency upgrades — do NOT run `go get -u ./...`

A blanket `go get -u ./...` **breaks the build**. It drags the `anacrolix/*` transitive
dependencies (generics, missinggo, dht, etc.) forward past what `github.com/anacrolix/torrent`
supports. Those newer anacrolix versions have moved to `github.com/RoaringBitmap/roaring/v2`,
while `torrent` still uses roaring **v1**, so the two collide inside torrent's own code:

```
github.com/anacrolix/torrent/peerconn.go: cannot use cn.t._completedPieces.Clone()
(value of type *roaring.Bitmap) as *roaring/v2.Bitmap value in struct literal
```

`anacrolix/torrent` is a **direct** dependency (`downloader/torrent`, `tests/local_seeder.go`),
and `v1.61.0` is already the latest release — there is nothing newer to bump *to*. The whole
`anacrolix/*` + `RoaringBitmap/roaring` + `pion/*` cluster is tightly coupled and must move in
lockstep with `torrent`, which can't move. Treat that cluster as **pinned**.

**To upgrade dependencies:** upgrade everything *except* the torrent cluster with a targeted
list, then tidy. Leave `github.com/anacrolix/*`, `github.com/RoaringBitmap/roaring*`, and
`github.com/pion/*` untouched. Example (the deps worth refreshing — otel, aws, testcontainers,
sqlite, and their trees; `golang.org/x/*` come along transitively and torrent tolerates them):

```
go get -u \
  github.com/aws/aws-sdk-go-v2 github.com/aws/aws-sdk-go-v2/config \
  github.com/aws/aws-sdk-go-v2/credentials github.com/aws/aws-sdk-go-v2/service/s3 \
  github.com/testcontainers/testcontainers-go modernc.org/sqlite \
  github.com/bogem/id3v2/v2 github.com/gojuno/minimock/v3 github.com/google/uuid \
  github.com/joho/godotenv github.com/samber/oops github.com/dir01/sqlq \
  go.opentelemetry.io/otel go.opentelemetry.io/otel/... \
  go.opentelemetry.io/contrib/bridges/otelslog \
  go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp
go mod tidy
```

After any upgrade, verify with `go build ./...` before committing. If the build fails on
`peerconn.go`/roaring, an anacrolix module slipped through — `git checkout -- go.mod go.sum`
and redo the upgrade without it.

## Go toolchain

`go.mod` pins `go 1.26.5` + `toolchain go1.26.5`; `GOTOOLCHAIN=auto` fetches it on demand.

## Build & test

- `go build ./...`, `go vet ./...`
- Tests need **`CGO_ENABLED=1`** (sqlite driver).
- The `tests/` e2e suite is behind a build tag: `CGO_ENABLED=1 go test -tags gen_docs ./tests/...`.
  It spins up containers via **testcontainers**, so Docker must be running. It drives the real
  HTTP mux (`mux.ServeHTTP`), so it's the end-to-end check for routing changes.

## HTTP routing

The router (`http/http.go`) uses Go 1.22+ method-based `ServeMux` patterns
(`POST /jobs`, `GET /jobs/{id}`, `DELETE /jobs/{id}`, `GET|POST /metadata`, …) with
`req.PathValue("id")`. Span names come from `req.Pattern` (the low-cardinality route
template like `GET /jobs/{id}`) — never bake concrete IDs into span names.
