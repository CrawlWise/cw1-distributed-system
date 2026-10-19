# CW1 Distributed System — Weeks 1–5 Verification Guide

Run the commands below in **Windows PowerShell**. All networking examples use `127.0.0.1` (loopback only); the coursework services have no TLS or authentication and must not be exposed publicly.

## 1. Project setup and automated checks

```powershell
cd C:\workspace\cw1-distributed-system\source

go version
go mod verify
go build ./...
go test ./... -count=1 -timeout 30s
go test -race ./... -count=1 -timeout 60s
go vet ./...
```

**Expected:** Dependency verification succeeds, the build and vet commands return no errors, and every package containing tests reports `ok`. The race detector should report no races during the run. On Windows, `-race` requires CGO and a supported C compiler.

## 2. Start both nodes

Use separate PowerShell terminals, both opened in the `source` directory. Leave them running during the live checks.

**Terminal 1 — Node A**

```powershell
.\scripts\run-node-a.ps1
```

Node A: HTTP `127.0.0.1:8081`, TCP `127.0.0.1:7001`, UDP `127.0.0.1:7101`, gRPC `127.0.0.1:9091`.

**Terminal 2 — Node B**

```powershell
.\scripts\run-node-b.ps1
```

Node B: HTTP `127.0.0.1:8082`, TCP `127.0.0.1:7002`, UDP `127.0.0.1:7102`, gRPC `127.0.0.1:9092`.

Use a **third PowerShell terminal** in the `source` directory for the following commands.

## 3. Week 1 — HTTP health endpoints

```powershell
Invoke-RestMethod http://127.0.0.1:8081/healthz
Invoke-RestMethod http://127.0.0.1:8082/healthz

go test ./internal/httpapi -run '^TestWeek01_' -v -count=1
```

**Expected:** Both health endpoints respond successfully, and the Week 1 tests pass. To verify unsupported-method handling, use:

```powershell
go run ./cmd/request -method POST -url 'http://127.0.0.1:8081/healthz'
```

The CLI method flag is usable only if your `cmd/request` implementation supports `-method`; the automated Week 1 tests also cover wrong-method handling.

## 4. Week 2 — REST and JSON validation

```powershell
go test ./internal/httpapi -run '^TestWeek02_' -v -count=1
```

**Expected:** Tests pass for REST requests, JSON validation, unsupported media types, malformed payloads and method handling.

## 5. Week 3 — Peer registry

Register Node B with Node A:

```powershell
Invoke-RestMethod `
    -Uri 'http://127.0.0.1:8081/v1/peers' `
    -Method Post `
    -ContentType 'application/json' `
    -InFile '.\testdata\peer-b.json'
```

If you already registered this peer in the current node process, a duplicate-peer error is expected. Check the registry:

```powershell
Invoke-RestMethod 'http://127.0.0.1:8081/v1/peers'
go test ./internal/httpapi -run '^TestWeek03_' -v -count=1
```

**Expected:** Node B is listed with HTTP address `127.0.0.1:8082` and gRPC address `127.0.0.1:9092`; automated tests pass.

## 6. Week 4 — TCP and UDP IPC

Check TCP listeners:

```powershell
Test-NetConnection 127.0.0.1 -Port 7001
Test-NetConnection 127.0.0.1 -Port 7002
```

**Expected:** `TcpTestSucceeded : True` for both. This verifies TCP reachability, **not** UDP communication.

Run IPC and application tests:

```powershell
go test ./internal/ipc -v -count=1
go test ./internal/app -v -count=1
```

**Expected:** All tests pass, including the IPC scenarios exercised by the suite.

## 7. Week 5 — gRPC and remote file metadata

Check gRPC listeners:

```powershell
Test-NetConnection 127.0.0.1 -Port 9091
Test-NetConnection 127.0.0.1 -Port 9092
```

**Expected:** `TcpTestSucceeded : True` for both. The next HTTP requests verify actual gRPC functionality.

### Node A retrieves Node B's catalogue

```powershell
go run ./cmd/request -url 'http://127.0.0.1:8081/v1/remote-files?peer_id=node-b'
```

**Expected:** HTTP `200`, `node_id` = `node-b`, files `diagram.txt` and `notes.txt`. If Node B is unknown, register it using the Week 3 command above.

### Node B retrieves Node A's catalogue

Register Node A with Node B if not already registered:

```powershell
$peerA = @{
    id = 'node-a'
    http_address = '127.0.0.1:8081'
    grpc_address = '127.0.0.1:9091'
    status = 'unknown'
} | ConvertTo-Json

Invoke-RestMethod `
    -Uri 'http://127.0.0.1:8082/v1/peers' `
    -Method Post `
    -ContentType 'application/json' `
    -Body $peerA
```

Then request Node A's files:

```powershell
go run ./cmd/request -url 'http://127.0.0.1:8082/v1/remote-files?peer_id=node-a'
```

**Expected:** HTTP `200`, `node_id` = `node-a`, file `alpha.txt`.

### Check error responses

```powershell
# Missing peer ID: HTTP 400
go run ./cmd/request -url 'http://127.0.0.1:8081/v1/remote-files'

# Invalid peer ID: HTTP 400
go run ./cmd/request -url 'http://127.0.0.1:8081/v1/remote-files?peer_id=invalid%20id'

# Unknown peer ID: HTTP 404
go run ./cmd/request -url 'http://127.0.0.1:8081/v1/remote-files?peer_id=node-unknown'
```

Run Week 5 automated tests:

```powershell
go test ./internal/rpcfiles -run '^TestWeek05_' -v -count=1
go test ./internal/httpapi -run '^TestWeek05_' -v -count=1
```

**Expected:** All tests pass, including cancellation, unavailable remote server, validation, success, empty catalogue and HTTP error mapping.

## 8. Git tags and working tree

```powershell
git tag -n
git status
git ls-remote --tags origin
```

**Expected:** Tags `week01`, `week02`, `week03`, `week04`, and `week05` appear locally and remotely (provided you renamed and pushed Week 4's tag). A clean working tree is expected after committing all intended changes.

## 9. Quick routine verification

For a repeatable automated check without starting nodes manually:

```powershell
cd C:\workspace\cw1-distributed-system\source
go mod verify
go build ./...
go test ./... -count=1 -timeout 30s
go test -race ./... -count=1 -timeout 60s
go vet ./...
```

For **full end-to-end** confidence, also run both nodes and repeat the bidirectional remote-files requests from Section 7.

> **Scope:** Passing tests confirms the exercised scenarios, not every possible network or deployment failure. The system is a loopback-only coursework service.
