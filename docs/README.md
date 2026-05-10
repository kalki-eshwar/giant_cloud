# Giant Cloud Documentation

This folder contains project-level documentation for the Giant Cloud distributed storage MVP.

## Contents

- `architecture.md`: high-level architecture, ports, data flow, and runtime dependencies.
- `modules/control_plane.md`: control plane services, APIs, and operational notes.
- `modules/worker_node.md`: worker daemon and job pipeline behavior.
- `modules/proto.md`: protobuf contracts and regeneration workflow.
- `modules/gui.md`: frontend dashboard behavior and API contract.
- `testing.md`: test matrix and commands for all modules.

## Quick Start

1. Start infrastructure (PostgreSQL, Redis, MinIO).
2. Start control plane (`cd control_plane && go run .`).
3. Start worker node (`cd worker_node && go run .`).
4. Start GUI (`cd gui && npm install && npm run dev`).
