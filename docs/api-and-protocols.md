## Week 02 — REST API Design Rationale

### Resource-oriented design

The `/v1/status` endpoint represents the simulated status of a running node.

Using a resource-oriented endpoint allows the same resource to support different operations through standard HTTP methods instead of introducing an arbitrary endpoint that performs unrelated actions.

### HTTP method selection

**GET** retrieves the current status without modifying it.

**POST** submits a JSON representation containing the desired status and updates the existing resource after validation.

### HTTP response codes

- **200 OK:** Successful retrieval or update of the existing status resource.
- **400 Bad Request:** Malformed JSON, unsupported status values or invalid request structure.
- **405 Method Not Allowed:** Unsupported HTTP method. The response includes the `Allow: GET, POST` header.
- **413 Content Too Large:** Request body exceeds 4096 bytes.
- **415 Unsupported Media Type:** POST request does not specify a supported JSON content type.

HTTP 200 is used for a successful POST because the operation updates and returns an existing resource rather than creating a new resource. HTTP 201 would be more appropriate for an operation that creates a new resource.

### Request validation and state preservation

The handler distinguishes three categories of failure:

1. **HTTP request errors:** Unsupported methods and media types.
2. **JSON decoding errors:** Malformed input, unknown fields, oversized bodies and multiple JSON values.
3. **Domain validation errors:** Status values other than `healthy` or `degraded`.

The application decodes incoming JSON into a temporary request object before calling the status store's `Set()` method.

This ensures that invalid requests are rejected before modifying application state.

### Status versus liveness

`/healthz` reports whether the node process is running and responding to HTTP requests.

`/v1/status` exposes a simulated status that can be deliberately changed through POST requests.

A simulated `degraded` status does not imply that the process is unavailable.

### Testing evidence

The implementation passed all 13 supplied REST subtests, an independent empty-body boundary test, and the complete project regression suite.

Live HTTP tests confirmed successful updates, rejection without mutation, continued `/healthz` availability and independent status values for Node A and Node B.

### Known limitation

The JSON decoder rejects unknown fields and multiple JSON values, but the current implementation does not explicitly reject duplicate known JSON keys. This limitation is documented rather than presented as complete JSON validation.