# CW1 Week 5 — Dependencies and code generation

## Toolchain and pinned versions

The module uses Go 1.23 as its **language compatibility baseline**; the installed Go toolchain may be newer if approved by the teaching team.

| Component | Version / requirement |
| --- | --- |
| Go language compatibility | 1.23 |
| gRPC Go runtime (`google.golang.org/grpc`) | v1.75.1 |
| Protobuf Go runtime (`google.golang.org/protobuf`) | v1.36.6 |
| `protoc-gen-go` | v1.36.6 |
| `protoc-gen-go-grpc` | v1.5.1 |
| `protoc` compiler | Teaching-team-approved local installation; record actual version below |
| GCC / C compiler | Required on Windows when running Go's race detector with CGO |

These versions are the Week 5 coursework pins, not a statement that they are the newest or production-approved releases. Verify the actual resolved versions against `go.mod`.

## Local version record

Run these commands and enter the observed versions in your submission notes:

```powershell
go version
protoc --version
go env CGO_ENABLED CC
```

**Actual `go version`:** go1.27.1
**Actual `protoc --version`:** libprotoc 36.2

## Generate bindings

Run from the `source` directory:

```powershell
.\scripts\generate.ps1
```

On Bash:

```bash
sh scripts/generate.sh
```

The generation process uses the contract `api/files/v1/files.proto` and produces:

- `api/files/v1/files.pb.go` — generated message types and serialization bindings.
- `api/files/v1/files_grpc.pb.go` — generated `FileService` client and server interfaces.

The scripts also resolve Go module dependencies. After generation:

```powershell
go mod tidy
go mod verify
go build ./...
```

Commit the `.proto` contract, both generated `.go` files, `go.mod`, and `go.sum`. Do not edit generated code manually or commit compiler binaries, tool caches, or secrets.

## Runtime architecture

- Node A HTTP `127.0.0.1:8081`, gRPC `127.0.0.1:9091`.
- Node B HTTP `127.0.0.1:8082`, gRPC `127.0.0.1:9092`.
- `internal/rpcfiles/service.go` implements `ListFiles` using the shared catalogue snapshot.
- `internal/rpcfiles/client.go` uses `grpc.NewClient`, a bounded context and maximum send/receive message sizes.
- `internal/httpapi/remote.go` exposes `GET /v1/remote-files?peer_id=<id>` and translates gRPC errors into HTTP responses.

The transport uses `insecure.NewCredentials()` **only on loopback for coursework**. There is no TLS or authentication. `requester_id` identifies the caller but is not verified proof of identity. Do not expose these services to public or untrusted networks.

## Verification

The following checks were reported successful on 19 October 2026:

```powershell
go test ./... -count=1 -timeout 30s
go test -race ./... -count=1 -timeout 60s
go vet ./...
go mod verify
go build ./...
```

The race run exited with code `0` and did not report a data race. Bidirectional HTTP-to-gRPC retrieval was observed: A returned B's `diagram.txt` and `notes.txt`; B returned A's `alpha.txt`.

## Contract compatibility

Treat protobuf field numbers as stable wire identifiers. Do not reuse existing numbers for different meanings. Add compatible fields using new numbers; reserve removed field numbers/names when changing the schema. Regenerate both Go bindings from the `.proto` contract and rerun the complete test suite after contract changes.
