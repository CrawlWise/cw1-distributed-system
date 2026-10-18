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