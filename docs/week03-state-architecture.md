### Week 3 — Process-Local State Ownership

The CW1 distributed system uses process-local state ownership. Each node initializes and maintains its own runtime components through application composition.

**Application components**

- `internal/config`: Validates node configuration and addresses.
- `internal/model`: Defines peer and file metadata structures with validation methods.
- `internal/state/registry.go`: Stores registered peers using a concurrency-safe map protected by `sync.RWMutex`.
- `internal/state/catalog.go`: Stores file metadata using a concurrency-safe collection.
- `internal/httpapi/peers.go`: Exposes the peer registry and file catalogue through HTTP endpoints.
- `internal/app/stage.go`: Composes the runtime state and registers the HTTP handlers.

**State ownership**

Each node maintains independent in-memory instances of its peer registry and file catalogue. Node A cannot directly read or modify Node B's memory.

The file catalogue is initialized from node-specific JSON metadata. The registry is updated through the local node's HTTP API.

**Concurrency and isolation**

The registry uses synchronization to protect peer insertion and reads. The catalogue uses synchronization to protect replacement and snapshot operations.

Both components return independent snapshots rather than exposing mutable internal collections.

**Scope limitation**

Week 3 implements local state modelling and HTTP access. Automatic state replication, file transfer and distributed consensus are not implemented at this stage.