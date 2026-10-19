# 6CC545 CW1 — Distributed System (Weeks 1–5)

This repository contains the cumulative Go coursework implementation. Each week's work extends the same `source/` module; preserve earlier working functionality and tests when adding a new milestone. Consult the supplied weekly `TASKS.md` and workbook for assessment requirements.

## Implemented functionality

- HTTP REST endpoints, JSON validation and health checks from earlier weeks.
- Peer registration and a concurrency-safe peer registry.
- Local file metadata catalogue and earlier TCP/UDP IPC functionality.
- Week 5 Protocol Buffer contract and generated gRPC bindings.
- gRPC `FileService.ListFiles` server, remote client, and HTTP-to-gRPC adapter.
- Two-node retrieval of **remote** file metadata, with deadline and error handling.

## Safety boundary

**Teaching / loopback only.** Services bind to `127.0.0.1`, use synthetic data, and have **no TLS or authentication**. The gRPC client's insecure transport is intentional for this restricted exercise. Never expose the services to an untrusted network.

## Requirements

Use a teaching-team-approved Go toolchain compatible with the module's Go 1.23 baseline. Week 5 uses pinned gRPC/protobuf dependencies and an approved `protoc` installation. See [`DEPENDENCIES.md`](DEPENDENCIES.md) for versions, generation, and provenance.

## Generate and build

From `source/` on PowerShell:

```powershell
.\scripts\generate.ps1
go mod verify
go build ./...
```

On Bash, use `sh scripts/generate.sh` for generation. Generated bindings live in `api/files/v1/files.pb.go` and `api/files/v1/files_grpc.pb.go` and should be committed with the contract.

## Start two nodes

In separate PowerShell terminals, both located in `source/`:

```powershell
.\scripts\run-node-a.ps1
```

```powershell
.\scripts\run-node-b.ps1
```

For Bash use `sh scripts/run-node-a.sh` and `sh scripts/run-node-b.sh`. The scripts build and run each node in the foreground; press Ctrl+C to stop.

| Node | HTTP | TCP | UDP | gRPC |
| --- | --- | --- | --- | --- |
| A | `127.0.0.1:8081` | `127.0.0.1:7001` | `127.0.0.1:7101` | `127.0.0.1:9091` |
| B | `127.0.0.1:8082` | `127.0.0.1:7002` | `127.0.0.1:7102` | `127.0.0.1:9092` |

## Verify remote file retrieval

In a third terminal from `source/`, register B in A:

```powershell
Invoke-RestMethod -Uri 'http://127.0.0.1:8081/v1/peers' -Method Post -ContentType 'application/json' -InFile '.\testdata\peer-b.json'
go run ./cmd/request -url 'http://127.0.0.1:8081/v1/remote-files?peer_id=node-b'
```

Expected HTTP `200` response: `node_id` is `node-b`, and files include `diagram.txt` and `notes.txt`, **not** A's local `alpha.txt`.

For reverse retrieval, register A in B:

```powershell
$peerA = @{ id='node-a'; http_address='127.0.0.1:8081'; grpc_address='127.0.0.1:9091'; status='unknown' } | ConvertTo-Json
Invoke-RestMethod -Uri 'http://127.0.0.1:8082/v1/peers' -Method Post -ContentType 'application/json' -Body $peerA
go run ./cmd/request -url 'http://127.0.0.1:8082/v1/remote-files?peer_id=node-a'
```

Expected HTTP `200` response: `node_id` is `node-a`, with `alpha.txt`. A newly started node's registry must be populated again.

The HTTP adapter returns `400` for invalid peer IDs, `404` for unknown peers, `503` for unavailable gRPC services and `504` for remote deadlines.

## Test and verification

```powershell
go test ./... -count=1 -timeout 30s
go test -race ./... -count=1 -timeout 60s
go vet ./...
go mod verify
go build ./...
```

On Windows, the race detector requires CGO and a supported C compiler. These commands were reported successful on 19 October 2026; the race test exited `0`. Evidence is stored in `../evidence/week05/`.

## Week 5 source map

| Path | Responsibility |
| --- | --- |
| `api/files/v1/files.proto` | Wire contract for file metadata and `ListFiles` |
| `api/files/v1/files.pb.go` | Generated protobuf messages |
| `api/files/v1/files_grpc.pb.go` | Generated gRPC client/server interfaces |
| `internal/rpcfiles/service.go` | gRPC server implementation |
| `internal/rpcfiles/client.go` | Remote RPC invocation and timeout |
| `internal/httpapi/remote.go` | REST endpoint and gRPC status mapping |
| `internal/app/stage.go` | Service wiring and lifecycle |
| `scripts/generate.ps1`, `scripts/generate.sh` | Reproducible code generation |

The Week 5 checkpoint tag is `week05`. The separate final assessed submission tag is `final-cw1`; do not create or move that final tag until final integration and packaging are complete.
