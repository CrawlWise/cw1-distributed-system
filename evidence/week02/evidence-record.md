## Week 02 — Evidence Record

**Date:** 18 October 2026

| Evidence ID | Description | Result |
|---|---|---|
| W02-E01 | Successful project build using `go build ./...` | PASS |
| W02-E02 | Week 2 REST API tests — 13 supplied subtests | PASS |
| W02-E03 | Full project regression tests | PASS |
| W02-E04 | Live GET and POST HTTP requests | Pending |
| W02-E05 | Invalid request preserves previous state | Pending live verification |
| W02-E06 | Independent boundary test | Pending |
| W02-E07 | Static analysis using `go vet ./...` | Pending |

### Evidence locations

- `evidence/week02/rest-tests.txt`
- `evidence/week02/full-test-results.txt`
- `evidence/week02/live-api-tests.txt`

The first two test results have been observed in the terminal output. Their evidence files still need to be saved.

### Verification notes

The supplied Week 2 REST test suite passed all 13 subtests. The full project test suite also completed successfully.

Live HTTP testing and independent boundary testing remain outstanding.

The `week02` Git checkpoint has not yet been created.


### W02-E04 — Live REST API Verification

**Date:** 18 October 2026  
**Node:** node-a  
**Address:** `http://127.0.0.1:8081`

**Result: PASS**

Six live HTTP requests were executed successfully:

1. GET `/v1/status` returned `healthy`.
2. POST `/v1/status` changed the status to `degraded`.
3. GET `/v1/status` confirmed the updated status persisted.
4. GET `/healthz` returned `{"node_id":"node-a","status":"ok"}`.
5. POST malformed JSON returned HTTP 400 with `{"error":"invalid_json"}`.
6. GET `/v1/status` confirmed the status remained `degraded`.

**Conclusion:** The API successfully processes valid status updates and rejects malformed JSON without modifying the previously stored status.

**Evidence file:** `evidence/week02/live-api-tests.txt`

**Remaining verification:** Independent boundary test and Node B isolation test.