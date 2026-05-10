# Architecture

Giant Cloud is a multi-component system for distributed encrypted chunk storage and proof-of-storage tasks.

## Components

- Control Plane (Go): orchestrates worker registration, heartbeat streams, node allocation, and HTTP API for dashboard views.
- Worker Node (Go): registers itself, keeps a heartbeat stream open, stores chunks, and submits proof responses.
- Proto Module (Go): shared gRPC and message contracts generated from `proto/cloud.proto`.
- GUI (React + Vite): dashboard for network health, node listing, and manual proof challenge trigger.

## Ports and Endpoints

- Control Plane gRPC: `localhost:50051`
- Control Plane HTTP API: `localhost:8081`
- Worker local HTTP upload endpoint: `localhost:8080/chunk/upload`
- MinIO/S3: `localhost:9000`
- PostgreSQL: `localhost:5432`
- Redis: `localhost:6379`

## Control Flow

1. Worker starts and calls `Register` on control plane gRPC.
2. Worker opens `Heartbeat` bidirectional stream and sends periodic status heartbeats.
3. Control plane tracks active streams and allocates nodes for upload chunks.
4. Control plane can push proof challenges through heartbeat stream.
5. Worker receives challenge and submits `SubmitProof` response.
6. GUI polls HTTP API for node states and allows manual `ping` challenge trigger.

## Persistence and Events

- PostgreSQL stores persistent node and manifest metadata.
- Redis list `proof_tasks` is polled by control plane event loop for challenge scheduling.
- MinIO stores chunk payloads under object paths by job/chunk.
