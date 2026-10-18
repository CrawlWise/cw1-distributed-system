# Weekly evidence record

Student number:
Week:
Repository URL:
Immutable tag:
Commit hash:
Go version / OS:

## Commands actually executed

## Observed results (do not paste expected examples as results)

## Failure case and diagnosis

## Design decision and alternative

## Remaining limitations

## Sources and AI-use disclosure


# CW1 — Week 1 Evidence Record

## Task 1.1 — Initial Baseline

**Command:** `.\scripts\check.ps1`

**Observed result:** Functional tests failed because `W01_CONFIG` and `W01_HEALTH` were incomplete.

**Evidence:** Preserve the original terminal output in `evidence/week01/initial-test-results.txt`.

## Task 1.2 — Configuration Validation

**Command:** `go test ./internal/config -v`

**Observed result:** All three supplied configuration package tests reported PASS.

**Evidence:** Save the complete test output in `evidence/week01/config-test-results.txt`.

**Remaining verification:** Full configuration contract, independent tests and cumulative regression tests.

## Task 1.3 — Health Endpoint

**Status:** Not started.

## Task 1.4 — Two Independent Nodes

**Status:** Not started.

## Task 1.5 — Startup and Shutdown

**Status:** Not started.

## Task 1.6 — Independent Test and Explanation

**Status:** Not started.



### W02-E08 — Independent Node State Verification

**Date:** 18 October 2026

**Objective:** Verify that two independently running Go nodes maintain separate simulated status values.

**Test results:**

- Node A (`127.0.0.1:8081`) returned HTTP 200 with `{"status":"degraded"}`.
- Node B (`127.0.0.1:8082`) returned HTTP 200 with `{"status":"healthy"}`.
- A subsequent GET request to Node A confirmed its status remained `degraded`.

**Result: PASS**

**Conclusion:** The two running nodes maintain independent local status stores. Accessing Node B did not modify Node A's status.

**Evidence:** `evidence/week02/node-isolation-tests.txt`


### Week 3 — Task 3.2: Model Validation Evidence

**Command executed:**

```powershell
go test ./internal/model -run "^TestWeek03_MetadataValidation$" -v -count=1
```

**Actual result:**

```text
=== RUN   TestWeek03_MetadataValidation
--- PASS: TestWeek03_MetadataValidation (0.00s)
PASS
ok      example.edu/6cc545/cw1/internal/model   1.632s
```

**Result:** PASS

The valid file metadata record was accepted, while the supplied invalid metadata cases were rejected.

**Additional test evidence:**

- `TestWeek03_PeerValidation` — PASS (earlier run)
- `TestWeek03_IPv6PeerValidation` — PASS (earlier run)

**Combined model suite:** Pending final execution.


### Week 3 — Task 3.3: Race Detector Setup
The initial race-detector execution failed because CGO was disabled. A subsequent check showed that GCC was not available through the Windows PowerShell PATH.
The four ordinary registry tests passed, including concurrent insertion and duplicate-ID detection.
Race-detector verification remains pending installation and configuration of a compatible C compiler.


### Week 3 — Task 3.3: Race Detector Configuration

**Environment:** Windows, PowerShell

**Compiler:** GCC 16.2.0 (MSYS2 UCRT64)

**CGO configuration:**
- `CGO_ENABLED=1`
- `CC=gcc`

**Initial issue:** Race-detector execution was blocked because CGO was disabled and GCC was unavailable through PATH.

**Resolution:** Configured the MSYS2 GCC compiler and enabled CGO for the PowerShell session.

**Race-test result:** Pending execution.

**Verification scope:** Registry insertion, duplicate detection, snapshot behavior and concurrent access.

**Limitation:** A successful race-detector run only confirms that no data races were detected in the execution paths exercised by the tests. It does not prove the absence of every possible concurrency defect.

-----------------------------------------------------------------------------
### Week 3 — Task 3.3: Race Detector Verification

**Environment:** Windows PowerShell, MSYS2 GCC 16.2.0, CGO enabled.

**Command executed:**

```powershell
go test -race ./internal/state -run "^TestWeek03_(Registry|ConcurrentRegistry|InvalidPeer|ConcurrentDuplicateID)$" -v -count=1 -timeout 60s
```

**Results:**

- `TestWeek03_Registry` — PASS
- `TestWeek03_ConcurrentRegistry` — PASS
- `TestWeek03_InvalidPeer` — PASS
- `TestWeek03_ConcurrentDuplicateID` — PASS

**Package result:** PASS

**Exit code:** 0

**Evidence file:** `evidence/week03/registry-race-tests.txt`

**Conclusion:** The peer registry passed its functional and concurrent tests. The race detector reported no data races during this execution. The result does not prove that every possible concurrent execution is race-free.

**Task 3.3 status:** Completed.

### Week 3 — Task 3.4: Catalogue Verification

**Test environment:** Go on Windows PowerShell.

**Catalogue test command:**

`go test ./internal/state -run "^TestWeek03_Catalog" -v -count=1`

**Observed results:**

- `TestWeek03_Catalog` — PASS
- `TestWeek03_CatalogDuplicateNames` — PASS
- `TestWeek03_CatalogEmpty` — PASS

**Exit code:** 0

**Evidence:** `evidence/week03/catalog-tests.txt`

**Complete state test command:**

`go test ./internal/state -run "^TestWeek03_" -v -count=1`

**Observed result:** All seven tests passed.

**Race detector command:**

`go test -race ./internal/state -run "^TestWeek03_" -count=1 -timeout 60s`

**Observed result:** PASS (`1.585s`).

**Conclusion:** The catalogue and registry passed the supplied and additional state tests. The race detector reported no data races during the tested execution paths.

**Limitation:** These tests do not establish correctness for all possible concurrent executions.


### Week 3 — Task 3.7: Full Race Detector Verification

**Environment:** Windows PowerShell, MSYS2 GCC 16.2.0, CGO enabled.

**Command executed:**

```powershell
go test -race ./... -count=1 -timeout 60s
```

**Observed results:**

| Package | Result |
|---|---|
| `cmd/node` | No test files |
| `cmd/request` | No test files |
| `examples/firstfunction` | PASS |
| `examples/jsonmodel` | No test files |
| `internal/app` | PASS |
| `internal/config` | PASS |
| `internal/httpapi` | PASS |
| `internal/model` | PASS |
| `internal/state` | PASS |

**Exit code:** 0

**Evidence file:** `evidence/week03/full-race-results.txt`

**Conclusion:** The full race-enabled Go test suite completed successfully. No data races were detected in the execution paths exercised by the tests.

**Limitation:** Race detection does not guarantee the absence of all concurrency defects.