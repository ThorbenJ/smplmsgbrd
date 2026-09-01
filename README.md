# smplmsgbrd

A minimal shared message board written in Go. Intended as a self-contained training artefact for packaging Go applications in Artifact Registry and deploying them on K8s.

Some branches contain intentional bugs for guided exercises. If you are working through an exercise, check out the relevant branch and read the README there for instructions.

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

## Packaging

### Pre-built image

A multi-arch image (`linux/amd64`, `linux/arm64`) is published automatically to GitHub Container Registry on every `v*` tag push:

```
ghcr.io/thorbenj/smplmsgbrd:latest
ghcr.io/thorbenj/smplmsgbrd:v2026.06.30   # version tag
ghcr.io/thorbenj/smplmsgbrd:<commit-sha>  # exact commit
```

See [`.github/workflows/docker.yml`](.github/workflows/docker.yml) for the CI definition.

### Build it yourself

The repo includes a `Dockerfile` (two-stage build: Go builder → `alpine:3.21`):

```sh
docker build -t smplmsgbrd:local .
```

### Push to your own registry

To push to Google Artifact Registry (or any other registry), tag and push after building:

```sh
docker build -t REGION-docker.pkg.dev/PROJECT/REPO/smplmsgbrd:TAG .
docker push REGION-docker.pkg.dev/PROJECT/REPO/smplmsgbrd:TAG
```

### Deploy to K8s

Ready-to-use manifests live in [`k8s/`](k8s/):

- `deployment.yaml` — Deployment + ClusterIP Service (port 80 → 8080)
- `ingress.yaml` — example Ingress (nginx, TLS placeholder; needs customisation before use)
- `kustomization.yaml` — example Kustomize overlay

Note: the in-memory buffer is per-process; multiple replicas will have independent message histories.

## License

GNU Lesser General Public License v2.1 — see [LICENSE](LICENSE).
