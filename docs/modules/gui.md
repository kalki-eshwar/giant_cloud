# GUI Module

Path: `gui`

## Responsibilities

- Display network/node status from control plane HTTP API.
- Surface capacity and active node summary.
- Provide manual proof challenge trigger per node.

## Main Files

- `src/App.tsx`: dashboard state management, API polling, table and actions.
- `src/App.css`: dashboard styling.
- `src/main.tsx`: React bootstrap entrypoint.

## API Contract Used

- `GET http://localhost:8081/api/nodes`
  - Returns array of `{ id, capacity, status }`.
- `POST http://localhost:8081/api/nodes/{id}/ping`
  - Triggers proof challenge on active node.

## UX Behavior

- Polls node list every 5 seconds.
- Displays loading state until first fetch resolves.
- Shows warning banner if control plane API is unavailable.
- Disables `Ping Proof` action for offline nodes.
