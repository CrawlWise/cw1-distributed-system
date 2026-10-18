# Implementation log

Record genuine work as it happens. Do not prefill outcomes.

| Date | Task / commit | Decision and reason | Test run and actual result | Feedback / next step | Sources or AI assistance |
|---|---|---|---|---|---|
| | | | | | |


# Go version
* go version go1.27.1 windows/amd64

# Git version
* git version 2.55.0.windows.5

# Go EVN GOMOD
* cw1-distributed-system\source\go.mod


## Task 1.1 — Establish a Reproducible Starting Point

### Objective
Configure the Go development environment, initialise the coursework workspace using the supplied starter files, and establish a baseline for testing.

### Steps Taken
1. Extracted the CW1 Weeks 1–5 wireframe starter pack.
2. Identified the Week 1 project scaffold.
3. Set up the working application inside `source/`.
4. Used the PowerShell terminal to execute the starter's `scripts/check.ps1`.
5. Examined the initial test results to identify unfinished implementation tasks i need to work on.

### Initial Test Results
The test run confirmed that the application compiled and the example package tests passed. However, the functional tests failed because configuration validation and the health endpoint were not yet implemented.

The failures included:
- `TestWeek01_Config`
- `TestWeek01_OccupiedPort`
- `TestWeek01_CancelledLifecycle`
- `TestWeek01_Health`
- `TestWeek01_HealthWrongMethod`

### Observation
The supplied starter is intentionally incomplete. Successful compilation does not guarantee that the application satisfies its functional requirements.

## Task 1.2 — Validate Configuration

### Objective
Implement configuration validation for node identity, HTTP listening address, network timeout and optional network addresses.

### Files Examined
- `internal/config/config.go`
- `internal/config/config_test.go`

### Steps Taken
1. Reviewed the `Config` struct and its fields.
2. Examined the supplied `ValidID()` and `ValidAddress()` helper functions.
3. Reviewed `config_test.go` to understand the expected validation behaviour.
4. Implemented the initial `Validate()` function to check node identity, HTTP address and positive timeout.
5. Executed the configuration package tests.

### Test Command
`go test ./internal/config -v`

### Observed Result
The following tests passed:
- `TestSmoke_ParseHelp`
- `TestWeek01_Config`
- `TestWeek01_UnknownFlag`

### Implementation Decision
Existing helper functions were reused to avoid duplicating validation logic. Configuration errors are returned to the caller instead of terminating the application directly.

### Remaining Work
- Extend timeout validation to enforce the 30-second maximum.
- Validate optional TCP, UDP and gRPC addresses.
- Add an independent table-driven test.
- Rerun the tests and record the final results.

### Reflection
This task demonstrates the importance of validating configuration before starting a network service. It also shows that passing supplied tests does not necessarily cover every requirement specified in the coursework documentation.



## Task 1.2 — Debugging Configuration Validation

Issue encountered: The TestWeek01_Config test failed because the validation function rejected an empty TCP address.

* code: `if c.TCP != " " && !ValidAddress(c.TCP) {
		return fmt.Errorf("invalid TCP address: %q", c.TCP)
	}`

Error message: valid configuration rejected: invalid TCP address: ""

Cause: The TCP address was being validated even when no value was supplied, contrary to the coursework requirement that optional network addresses may be empty.

Resolution: Updated the TCP validation conditions to check whether an address was provided before calling ValidAddress(), and removing the empty space between the quote `""`.

Verification: Pending rerun of go test ./internal/config -v.

Learning outcome: Optional configuration fields require different validation rules from mandatory fields. An empty optional field should be accepted, while a supplied value must satisfy the required format.


## Task 1.3 — Implement a Purposeful Health Endpoint

### Objective
Implement an HTTP health endpoint that confirms a node process is responding and identifies the node.

### Files Examined
- `internal/httpapi/health.go`
- `internal/httpapi/health_test.go`

### Implementation Steps
1. Reviewed the supplied `WriteJSON()` and `ErrorJSON()` helper functions.
2. Examined the existing `Health()` placeholder implementation.
3. Reviewed the tests to identify the expected HTTP methods, status codes and JSON response.
4. Implemented HTTP method validation to accept GET requests only.
5. Added the `Allow: GET` response header for unsupported methods.
6. Implemented a JSON health response containing the node identity and status.
7. Reused the supplied JSON helpers rather than duplicating response-writing logic.

### Implementation Decision
The health endpoint uses the node identity supplied during handler creation. This allows independently configured nodes to return different identities while sharing the same implementation.

### Testing
**Command:** `go test ./internal/httpapi -v`

Expected result: The health endpoint tests pass for valid GET requests and unsupported HTTP methods.

Actual result: All test pass successfully.

### Learning Outcome
The health endpoint demonstrates how a Go HTTP handler processes requests and produces structured JSON responses. A successful health response confirms that the local process is responding but does not establish the availability of other nodes or the readiness of future distributed services.

### Task 1.3 — Full Regression Test Results

Command: `go test ./... -count=1 -timeout 20s`

Result: PASS

All tested packages passed:
- `examples/firstfunction`
- `internal/app`
- `internal/config`
- `internal/httpapi`

The `cmd/node` and `cmd/request` packages compiled successfully but contained no test files.

Observation: The previously failing configuration, health endpoint and application lifecycle tests now pass.

Conclusion: The supplied automated tests confirm that the implemented Week 1 functionality satisfies the behaviours covered by the test suite. Manual process-level verification and checkpoint evidence are still required.

## Task 1.4 — HTTP Connection Refused During Node Testing

- Objective: Verify that Node A responds to HTTP requests on port 8081.

- Error encountered: dial tcp 127.0.0.1:8081: connectex: No connection could be made because the target machine actively refused it.

- Observation: The HTTP client could not establish a connection to Node A.

- Possible cause: Node A was not running, had failed during startup, or was not listening on the expected address.

- Troubleshooting: Check the Node A startup output and verify that port 8081 is listening.

Resolution: Manually start the .\scripts\run-node-a.ps1 for both node-a and node-b.

Learning outcome: Passing automated tests does not guarantee that a server process is currently running. A client can communicate with the server only when the correct network listener is active.

### Task 1.4 — Node A Startup Verification

Date: 16 October 2026
Command: `.\scripts\run-node-a.ps1`

# Observed output:

`2026/10/16 13:30:18 node=node-a http=127.0.0.1:8081 ready (loopback teaching service; no authentication)`

Result: Node A reported successful startup and readiness on its configured HTTP address.

Previous issue: An HTTP request initially failed with a connection-refused error.

Current status: Node A is reporting readiness. Successful HTTP communication remains to be verified.

Learning outcome: A Go node must start its HTTP listener before a client can connect. Starting the application and successfully handling a client request are separate verification steps.

## Task 1.4 — Run Two Independent Nodes


### Objective
Run two independent Go node processes using different identities and HTTP ports, then verify that each responds correctly to health requests.

### Commands Executed

Node A:
`.\scripts\run-node-a.ps1`

Node B:
`.\scripts\run-node-b.ps1`

HTTP client:
`go run ./cmd/request -url http://127.0.0.1:8081/healthz`

`go run ./cmd/request -url http://127.0.0.1:8082/healthz`

### Actual Results

Node A
- HTTP status: 200
- Content-Type: application/json; charset=utf-8
- Response: `{"node_id":"node-a","status":"ok"}`

Node B
- HTTP status: 200
- Content-Type: application/json; charset=utf-8
- Response: `{"node_id":"node-b","status":"ok"}`

### Initial Issue
The first attempt to connect to Node A returned a connection-refused error. Node A was subsequently started successfully, and its health endpoint returned the expected response.

### Outcome
PASS — Both independently running nodes returned the correct identities and healthy statuses.

### Learning Outcome
Independent processes can use the same application code while maintaining different configurations, identities and listening ports. Successful HTTP responses demonstrate that each process is reachable through its configured local endpoint.



## Task 1.5 — Test 1: Occupied HTTP Port

Date: 16 October 2026

Objective: Verify that a node reports a clear startup error when another process already owns its configured HTTP port.

Test command:

go run ./cmd/node -id node-conflict -http 127.0.0.1:8081

Observed error:

node stopped: listen HTTP 127.0.0.1:8081: listen tcp 127.0.0.1:8081: bind: Only one usage of each socket address (protocol/network address/port) is normally permitted.

The command terminated with exit status 1.

Verification command:

go run ./cmd/request -url http://127.0.0.1:8081/healthz

Observed response:

HTTP 200

{"node_id":"node-a","status":"ok"}

Result: PASS.

Conclusion: The application detected the occupied port, reported the failure and exited without disrupting the existing Node A process.

Learning outcome: A TCP listener owns its configured address and port. A second process cannot normally bind to the same address and port while the first listener remains active.


### Task 1.5 — Test 2: Invalid Startup Configuration


Objective: Verify that the application rejects invalid configuration values before starting the HTTP server.

Test A — Invalid Node Identity

Command: go run ./cmd/node -id "invalid node!" -http 127.0.0.1:8090

Observed: configuration: invalid node ID: "invalid node!"

Exit status: 2 — PASS.

Test B — Invalid HTTP Address

Command: go run ./cmd/node -id node-test -http 0.0.0.0:8090

Observed: configuration: invalid HTTP address: "0.0.0.0:8090"

Exit status: 2 — PASS.

Test C — Unknown CLI Flag

Command: go run ./cmd/node -id node-test -http 127.0.0.1:8090 -unknown value

Observed: flag provided but not defined: -unknown

The application displayed the supported command-line flags.

Exit status: 2 — PASS.

Test D — Invalid Timeout

Command: go run ./cmd/node -id node-test -http 127.0.0.1:8090 -timeout 31s

Observed: configuration: timeout must be greater than zero and at most 30 seconds

Exit status: 2 — PASS.

Conclusion: All four invalid startup configurations were correctly rejected. The application provided meaningful diagnostic messages and terminated without reporting server readiness.

Learning outcome: Early configuration validation prevents invalid settings from reaching the network-listener startup stage and makes operational failures easier to diagnose.


### Task 1.6 — Independent HTTP Handler Test

**Objective:** Add an independent automated test that verifies the HTTP health handler's method restrictions and response format.

**File:** `internal/httpapi/health_test.go`

**Test name:** `TestWeek01_HealthAllowHeader`

**Test coverage:**
- Unsupported POST requests return HTTP 405.
- The `Allow` response header contains `GET`.
- The response body contains a valid JSON error.
- The error message is `method not allowed`.

**Design explanation:** An HTTP request is received by the application's listener and routed to the registered health handler. The handler validates the request method and writes a JSON response through the supplied helper functions. The application lifecycle owns the listener and shutdown process. The handler should not call `os.Exit`, because doing so would terminate the entire process instead of returning an HTTP response and allowing the application to manage its resources.

**Test result:** The test executed successfully.

### Task 1.6 — Test Discovery Issue

**Command:** `go test ./internal/httpapi -run "^TestWeek01_HealthAllowHeader$" -v`

**Observed result:** `testing: warning: no tests to run`

**Explanation:** Go did not discover a test matching the specified function name. Although the command returned PASS, the intended test was not executed.

**Troubleshooting:** Inspect the test file, verify the function name, save the changes and use `go test ./internal/httpapi -list .` to confirm test discovery.

**Resolution:** I added the test to the `/health_test.go` file but forget to save the file that was the reason for the observed result.

### Task 1.6 — Independent HTTP Handler Test

**Date:** 16 October 2026

**Test:** `TestWeek01_HealthAllowHeader`

**Command:** `go test ./internal/httpapi -run "^TestWeek01_HealthAllowHeader$" -v`

**Actual result:** PASS

The test verified that an unsupported HTTP method returns status 405, includes the `Allow: GET` response header, and produces the expected JSON error response.

**Issue encountered:** The first execution reported `no tests to run`. After correcting test discovery, the test was successfully executed.

**Conclusion:** The independent test passed, extending coverage beyond the supplied starter tests.


### Week 1 — Design Note

The Week 1 implementation establishes the foundation for a distributed node application using Go. Each node operates as an independent process with a unique identity and configurable HTTP listening address. Configuration validation ensures that node identifiers, network addresses and timeout values meet the required constraints before startup. The `/healthz` endpoint provides a lightweight mechanism for checking whether a node is responding and confirming its identity. HTTP handlers are responsible for processing requests and returning structured responses, while the application lifecycle manages listeners, cancellation and shutdown. Handlers should not call `os.Exit`, because terminating the process would bypass normal application-level cleanup. Running two nodes on separate ports demonstrated process independence, while occupied-port and shutdown tests verified error handling and resource management. These capabilities provide a foundation for later distributed communication.


## Week 1 — Final Verification

**Date:** 16 October 2026

**Verification command:** `.\scripts\check.ps1`

**Go version:** go1.27.1 windows/amd64

**Result:** PASS

The verification script completed successfully, reporting passing checks for the example, application, configuration and HTTP API packages.

### Completed Implementation

- Implemented configuration validation for node identity, loopback addresses and timeout constraints.
- Implemented the `/healthz` endpoint with node identity and HTTP method validation.
- Successfully started two independent nodes on ports 8081 and 8082.
- Verified both nodes returned HTTP 200 and their respective identities.
- Confirmed occupied-port errors were handled without disrupting the existing node.
- Verified invalid startup configurations were rejected.
- Tested node shutdown, listener cleanup and successful restart.
- Added and passed an independent HTTP handler test.

### Conclusion

The Week 1 implementation and testing objectives have been achieved. The verification script passed, although this does not establish complete test coverage. The remaining administrative activities are final evidence organization and creation of the Week 1 Git checkpoint.



## Week 02 — REST API and JSON Status Resource

**Date:** 18 October 2026  
**Module:** 6CC545 — Distributed Applications and Web Services  
**Milestone:** Week 02 — Design and test a JSON REST resource

### Objective

Extend the existing Week 1 Go node application with a REST API that supports retrieving and updating a simulated node status through HTTP GET and POST operations.

The implementation must validate incoming JSON, enforce a maximum request body size of 4096 bytes, return appropriate HTTP status codes and prevent invalid requests from modifying existing application state.

### Implementation completed

**1. Week 2 integration**

Integrated the new Week 2 files into the existing project:

- `internal/httpapi/rest.go`
- `internal/httpapi/rest_test.go`
- `internal/state/status.go`
- `examples/jsonmodel/main.go`
- JSON request fixtures in `testdata/`

Updated `internal/app/stage.go` to register the `/v1/status` endpoint while preserving the Week 1 `/healthz` functionality.

**2. JSON request decoder — `DecodeOne()`**

Implemented the `W02_DECODE` requirement in `internal/httpapi/rest.go`.

The decoder:

- Validates the `Content-Type` header.
- Accepts `application/json` with supported parameters.
- Enforces a 4096-byte request body limit.
- Rejects malformed JSON, null values and arrays.
- Rejects unknown JSON fields.
- Rejects multiple JSON values and trailing non-whitespace content.
- Returns structured errors without modifying application state.

**3. REST status handler — `StatusResource()`**

Implemented the `W02_STATUS` requirement.

The handler supports:

- `GET /v1/status` — retrieves the current simulated status.
- `POST /v1/status` — validates and updates the simulated status.
- Unsupported HTTP methods — returns HTTP 405 with an `Allow` header.

Only `healthy` and `degraded` are permitted status values.

**4. Status management**

Used the supplied `internal/state/status.go` implementation.

The status store uses `sync.RWMutex` to synchronize reads and writes. Status updates are performed only after request decoding and validation.

### Tests executed

**Test 1 — Build verification**

Command:

`go build ./...`

Result: **PASS**

The project compiled successfully without reported errors.

**Test 2 — Week 2 REST API tests**

Command:

`go test ./internal/httpapi -run "^TestWeek02_" -v -count=1`

Result: **PASS — 13 of 13 supplied subtests passed.**

Verified scenarios included successful GET and POST requests, JSON charset parameters, malformed requests, unknown fields, invalid values, multiple JSON objects, null and array inputs, unsupported media types, oversized bodies and unsupported methods.

**Test 3 — Full project regression tests**

Command:

`go test ./... -count=1 -timeout 20s`

Result: **PASS**

All packages with tests passed. No failures were reported.

### Problems encountered and resolution

**Problem:** All 13 Week 2 REST subtests initially failed with HTTP 501.

**Cause:** `StatusResource()` still contained the supplied `TODO W02_STATUS` placeholder.

**Resolution:** Implemented GET, POST and unsupported-method handling. Reran the tests and confirmed all 13 subtests passed.

### Current status

The REST API implementation passes the supplied automated tests and the complete project test suite.

### Remaining work

- Document the complete REST API contract.
- Perform live HTTP testing against running nodes.
- Add an independent request-boundary test.
- Run `go vet ./...`.
- Save actual test transcripts.
- Document HTTP design rationale.
- Create the Week 2 Git checkpoint after completing verification.


### Task 2.6 — Independent Boundary Test

**Date:** 18 October 2026

**File created:** `internal/httpapi/rest_boundary_test.go`

Implemented `TestWeek02_EmptyBodyPreservesState` to independently verify that an empty JSON POST request is rejected without changing the existing node status.

**Test procedure:**

1. Initialize a new status store.
2. Set its status to `degraded`.
3. Send a POST request to `/v1/status` with `Content-Type: application/json` and an empty body.
4. Assert that the server returns HTTP 400.
5. Assert that the stored status remains `degraded`.

**Results:**

- Independent boundary test: PASS
- Full project regression suite: PASS
- Static analysis (`go vet ./...`): No errors reported in the provided output

**Conclusion:** The implementation rejects an empty JSON request and preserves the previously stored valid state.