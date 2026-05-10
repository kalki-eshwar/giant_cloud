# Worker Node Module

Path: `worker_node`

## Responsibilities

- Register with control plane.
- Maintain heartbeat stream and receive push tasks.
- Execute storage and compute job abstractions.
- Submit proof responses when challenged.
- Expose an HTTP endpoint for chunk payload upload.

## Main Files

- `main.go`: bootstrap node, initialize S3 storage client, start daemon, host `/chunk/upload` endpoint.
- `daemon/daemon.go`: gRPC connection lifecycle, register + heartbeat loop, proof submission on challenge.
- `jobs/interface.go`: shared job types and S3-backed storage job implementation.
- `jobs/storage.go`: local file persistence utility.
- `jobs/compute.go`: placeholder for future compute orchestration.

## Data Handling

- Upload endpoint accepts `POST /chunk/upload` and forwards body payload to `StorageJob`.
- `StorageJob` writes chunks to MinIO bucket `chunks` with object key `{jobID}/{chunkID}`.

## Runtime Notes

- Worker local HTTP server binds to `:8080`.
- Daemon connects to control plane gRPC at `localhost:50051`.
- Proof responses are simulated with dummy hash in current MVP.
