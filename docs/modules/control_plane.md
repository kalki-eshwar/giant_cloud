# Control Plane Module

Path: `control_plane`

## Responsibilities

- Expose worker-facing gRPC API.
- Expose client-facing gRPC API.
- Expose dashboard HTTP API.
- Maintain active worker streams.
- Persist node and manifest metadata.
- Poll Redis for proof challenge events.

## Main Files

- `main.go`: process bootstrap, DB/Redis init, gRPC start, HTTP start, Redis event loop.
- `server.go`: gRPC service handlers and HTTP endpoints.
- `db.go`: PostgreSQL connection setup and schema management.

## gRPC APIs

Worker service:

- `Register(RegisterRequest)`: adds worker and updates DB status.
- `Heartbeat(stream)`: long-lived bidirectional stream for liveness and push tasks.
- `SubmitProof(ProofRequest)`: receives proof responses.

Client service:

- `UploadManifest(ManifestRequest)`: records upload metadata.
- `AllocateNodes(AllocateRequest)`: round-robin assignment across active workers.

## HTTP APIs

- `GET /api/nodes`: list nodes and current active/offline state.
- `POST /api/nodes/{id}/ping`: enqueue immediate proof challenge for a specific active node.

## Runtime Notes

- Starts gRPC at `:50051` and HTTP at `:8081`.
- CORS is enabled for dashboard integration.
- DB and Redis failures are logged as warnings where possible to keep local development moving.
