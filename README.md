# Giant Cloud - Distributed Encrypted Storage MVP

This repository contains the MVP for a distributed encrypted storage marketplace. It has been significantly upgraded from its initial prototype to a robust, scalable architecture using **Go**, **gRPC**, **PostgreSQL**, **Redis**, and **MinIO (S3)**.

## Architecture

The system consists of three main components:

1.  **Control Plane (`/control_plane`)**: Written in Go. Exposes a gRPC server for worker node communication (registration, bidirectional heartbeats) and an HTTP REST server for the frontend GUI. It uses PostgreSQL for state management and Redis for event-driven task distribution.
2.  **Worker Node (`/worker_node`)**: Written in Go. A daemon that registers with the Control Plane via gRPC, maintains a continuous heartbeat stream, and executes proof-of-storage tasks. It abstracts storage to an S3-compatible backend (MinIO) making the workers stateless and scalable.
3.  **GUI Dashboard (`/gui`)**: A modern React application built with Vite that provides a visual overview of network health, active nodes, capacity, and allows manual triggering of proof challenges.

## Prerequisites

To run the full stack locally, you need the following infrastructure running on their default ports:

*   **PostgreSQL**: `localhost:5432` (Database: `cp`, User: `user`, Pass: `pass`)
*   **Redis**: `localhost:6379`
*   **MinIO**: `localhost:9000` (Access Key: `minioadmin`, Secret Key: `minioadmin`)

*(You can easily spin these up using a standard `docker-compose.yml` if you have Docker installed).*

## How to Run

### 1. Start the Control Plane

The Control Plane must be run from the `control_plane` directory.

```bash
cd control_plane
go run .
```
*This starts the gRPC server on `:50051` and the HTTP REST API on `:8081`.*

### 2. Start a Worker Node

The Worker Node must be run from the `worker_node` directory.

```bash
cd worker_node
go run .
```
*This connects to the Control Plane and initializes the S3 storage bucket.*

### 3. Start the GUI Dashboard

The GUI is a React application located in the `gui` directory.

```bash
cd gui
npm install
npm run dev
```
*This starts the Vite development server. Open the provided Local URL (usually `http://localhost:5173`) in your browser to view the dashboard.*

## Development Notes

*   **gRPC/Protobufs**: The `.proto` files are located in `/proto`. If you modify them, you must regenerate the Go code using `protoc` from the project root.
*   **Go Workspace**: The project uses a `go.work` file to manage the local modules (`control_plane`, `worker_node`, `proto`) together.