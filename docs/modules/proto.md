# Proto Module

Path: `proto`

## Responsibilities

- Define canonical API contracts shared by control plane and worker.
- Generate Go message and gRPC service code.

## Source of Truth

- `cloud.proto` defines:
  - `WorkerService`: register, heartbeat stream, proof submission.
  - `ClientService`: upload manifest, node allocation.
  - all request and response message types.

## Generated Artifacts

- `cloud/cloud.pb.go`: message types and serialization runtime.
- `cloud/cloud_grpc.pb.go`: gRPC service interfaces and client stubs.

## Regeneration

Run from repository root when `cloud.proto` changes:

```bash
protoc --go_out=. --go-grpc_out=. proto/cloud.proto
```

Ensure `protoc`, `protoc-gen-go`, and `protoc-gen-go-grpc` are installed.
