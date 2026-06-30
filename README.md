# smplmsgbrd

A minimal shared message board written in Go. Intended as a self-contained training artefact for packaging Go applications in a Artifact Registry and deploying them on K8s.

The server holds messages in memory and exposes a small JSON API. The browser page polls for updates without requiring WebSockets or a framework.

## Features

- Holds up to the last 2048 messages; oldest are evicted automatically when full
- Three-endpoint JSON API (`send_msg`, `peek_msgs`, `fetch_msgs`)
- Browser polls every 3 s and appends new messages without a full page reload; shows an eviction notice banner when old messages were dropped
- Messages are limited to 500 characters
- Single binary, no external Go dependencies
- HTML is served via `text/template` with CSS and JS injected at render time; all three are constants at the bottom of `main.go` — no separate asset pipeline

## Requirements

- Go 1.22 or later

## Build

```sh
go build -o smplmsgbrd .
```

## Run

```sh
./smplmsgbrd                  # listens on :8080
./smplmsgbrd -port 9090       # custom port
```

Open `http://localhost:8080` in a browser. Open the same URL in a second tab to see messages appear in real time.

## API

All responses are `application/json`.

### `GET /api/peek_msgs`

Returns the current `head` and `tail` counters. `head == tail` means there are no messages; `tail - head` is the count (at most 2048).

```json
{ "head": 0, "tail": 42 }
```

### `GET /api/fetch_msgs?start=N&end=N`

Returns messages in the half-open range `[start, end)`. `start` and `end` must be within the `[head, tail]` window from `peek_msgs`.

```json
{ "messages": ["hello", "world"] }
```

Pass `head` as `start` and `tail` as `end` to retrieve all current messages.

### `POST /api/send_msg`

Stores a message. Request body must be JSON.

```json
{ "message": "your text here" }
```

Returns `204 No Content` on success.

## Tests

```sh
go test ./...
```

The test suite covers the empty buffer, basic send/fetch, eviction when full, and wrap-around fetches.

## Packaging for K8s (outline)

1. Write a `Dockerfile` — a two-stage build (Go builder → `gcr.io/distroless/static`) produces a minimal image.
2. Push to Google Artifact Registry:
   ```sh
   docker build -t REGION-docker.pkg.dev/PROJECT/REPO/smplmsgbrd:TAG .
   docker push REGION-docker.pkg.dev/PROJECT/REPO/smplmsgbrd:TAG
   ```
3. Deploy to K8s with a `Deployment` + `Service` manifest, setting `containerPort: 8080` and passing `-port 8080` as the container command argument.

## License

GNU Lesser General Public License v2.1 — see [LICENSE](LICENSE).
