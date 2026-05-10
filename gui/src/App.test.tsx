import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/react'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import App from './App'

function mockResponse(payload: unknown, ok = true): Response {
  return {
    ok,
    status: ok ? 200 : 500,
    statusText: ok ? 'OK' : 'Error',
    json: async () => payload,
  } as Response
}

describe('App', () => {
  beforeEach(() => {
    vi.stubGlobal('alert', vi.fn())
  })

  afterEach(() => {
    cleanup()
    vi.unstubAllGlobals()
    vi.clearAllMocks()
  })

  it('renders worker nodes from API', async () => {
    const fetchMock = vi.fn().mockImplementation(async (url: string) => {
      if (url.includes('/api/nodes')) {
        return mockResponse([{ id: 'node-a', capacity: 120, status: 'active' }])
      }
      return mockResponse({}, false)
    })
    vi.stubGlobal('fetch', fetchMock)

    render(<App />)

    expect(await screen.findByText('node-a')).toBeInTheDocument()
    expect(screen.getByText('Total Nodes')).toBeInTheDocument()
    expect(fetchMock).toHaveBeenCalledWith('http://localhost:8081/api/nodes')
  })

  it('disables ping for offline workers', async () => {
    const fetchMock = vi.fn().mockImplementation(async (url: string) => {
      if (url.includes('/api/nodes')) {
        return mockResponse([{ id: 'node-offline', capacity: 64, status: 'offline' }])
      }
      return mockResponse({}, false)
    })
    vi.stubGlobal('fetch', fetchMock)

    render(<App />)

    expect(await screen.findByText('node-offline')).toBeInTheDocument()
    const pingButton = screen.getByRole('button', { name: 'Ping Proof' })
    expect(pingButton).toBeDisabled()
  })

  it('sends ping request for active node', async () => {
    const fetchMock = vi.fn().mockImplementation(async (url: string) => {
      if (url.includes('/ping')) {
        return mockResponse({ status: 'queued' })
      }
      if (url.includes('/api/nodes')) {
        return mockResponse([{ id: 'node-a', capacity: 99, status: 'active' }])
      }
      return mockResponse({}, false)
    })
    vi.stubGlobal('fetch', fetchMock)

    render(<App />)

    expect(await screen.findByText('node-a')).toBeInTheDocument()
    fireEvent.click(screen.getByRole('button', { name: 'Ping Proof' }))

    await waitFor(() => {
      expect(fetchMock).toHaveBeenNthCalledWith(
        2,
        'http://localhost:8081/api/nodes/node-a/ping',
        { method: 'POST' },
      )
    })
  })
})
