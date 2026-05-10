# Testing Guide

This repository now includes baseline unit tests for each module.

## Test Matrix

- Control Plane (`control_plane`): gRPC service logic and allocation behavior.
- Worker Node (`worker_node`): job execution and storage utility behavior.
- Proto (`proto`): protobuf message serialization round-trip checks.
- GUI (`gui`): dashboard rendering and API interaction tests with mocked `fetch`.

## Commands

From repository root:

```bash
go test ./...
```

For GUI:

```bash
cd gui
npm install
npm run test
```

## Notes

- Go unit tests do not require live Postgres, Redis, or MinIO services.
- GUI tests run in `jsdom` using Vitest and React Testing Library.
- Integration and end-to-end tests are not yet included; current scope is module-level unit coverage.
