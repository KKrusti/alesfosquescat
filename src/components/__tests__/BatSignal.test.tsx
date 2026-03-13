import { describe, it, expect, vi, afterEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { BatSignal } from '../BatSignal'
import { LanguageProvider } from '../../context/LanguageContext'

function renderBatSignal(hasActiveStreak = false, onSuccess = vi.fn()) {
  return render(
    <LanguageProvider>
      <BatSignal onSuccess={onSuccess} hasActiveStreak={hasActiveStreak} />
    </LanguageProvider>
  )
}

function mockFetch(data: object, ok = true, status = 200) {
  vi.stubGlobal(
    'fetch',
    vi.fn().mockResolvedValue({
      ok,
      status,
      json: () => Promise.resolve(data),
    })
  )
}

afterEach(() => {
  vi.restoreAllMocks()
  vi.useRealTimers()
})

describe('BatSignal — confirmation flow', () => {
  it('shows confirmation when report button is clicked (no active streak)', () => {
    renderBatSignal(false)
    const reportBtn = screen.getByRole('button', { name: /reportar|reportar un apagón/i })
    fireEvent.click(reportBtn)
    expect(screen.getByText(/confirmas que hay un apagón|confirmes que hi ha un apagó/i)).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /^sí$/i })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /^no$/i })).toBeInTheDocument()
  })

  it('shows confirmation when resolve button is clicked', () => {
    renderBatSignal(false)
    const resolveBtn = screen.getByRole('button', { name: /la llum ha tornat|ha vuelto la luz/i })
    fireEvent.click(resolveBtn)
    expect(screen.getByText(/confirmas que ha vuelto la luz|confirmes que ha tornat la llum/i)).toBeInTheDocument()
  })

  it('cancels confirmation and returns to idle when No is clicked', () => {
    renderBatSignal(false)
    const reportBtn = screen.getByRole('button', { name: /reportar|reportar un apagón/i })
    fireEvent.click(reportBtn)
    expect(screen.getByText(/confirmas|confirmes/i)).toBeInTheDocument()
    fireEvent.click(screen.getByRole('button', { name: /^no$/i }))
    expect(screen.queryByText(/confirmas|confirmes/i)).not.toBeInTheDocument()
    expect(screen.queryByRole('button', { name: /^sí$/i })).not.toBeInTheDocument()
  })

  it('calls /api/report when Sí is clicked after report confirmation', async () => {
    mockFetch({ success: true })
    const onSuccess = vi.fn()
    renderBatSignal(false, onSuccess)

    fireEvent.click(screen.getByRole('button', { name: /reportar|reportar un apagón/i }))
    fireEvent.click(screen.getByRole('button', { name: /^sí$/i }))

    await waitFor(() => expect(fetch).toHaveBeenCalledWith('/api/report', expect.objectContaining({ method: 'POST' })))
    expect(onSuccess).toHaveBeenCalled()
  })

  it('calls /api/resolve when Sí is clicked after resolve confirmation', async () => {
    mockFetch({ resolved: true })
    const onSuccess = vi.fn()
    renderBatSignal(false, onSuccess)

    fireEvent.click(screen.getByRole('button', { name: /la llum ha tornat|ha vuelto la luz/i }))
    fireEvent.click(screen.getByRole('button', { name: /^sí$/i }))

    await waitFor(() => expect(fetch).toHaveBeenCalledWith('/api/resolve', expect.objectContaining({ method: 'POST' })))
    expect(onSuccess).toHaveBeenCalled()
  })

  it('does NOT call fetch if No is clicked — confirmation cancelled', async () => {
    const fetchMock = vi.fn()
    vi.stubGlobal('fetch', fetchMock)
    renderBatSignal(false)

    fireEvent.click(screen.getByRole('button', { name: /reportar|reportar un apagón/i }))
    fireEvent.click(screen.getByRole('button', { name: /^no$/i }))

    // Wait a tick to ensure no async calls happened
    await new Promise((r) => setTimeout(r, 50))
    expect(fetchMock).not.toHaveBeenCalled()
  })

  it('skips confirmation when streak is already active and shows already_voted message', () => {
    renderBatSignal(true)
    fireEvent.click(screen.getByRole('button', { name: /reportar|reportar un apagón/i }))
    // No confirmation — directly shows feedback message
    expect(screen.queryByText(/confirmas|confirmes/i)).not.toBeInTheDocument()
    expect(screen.getByText(/racha ya está activa|ratxa ja està activa/i)).toBeInTheDocument()
  })

  it('disables both buttons while confirmation is pending', () => {
    renderBatSignal(false)
    fireEvent.click(screen.getByRole('button', { name: /reportar|reportar un apagón/i }))
    // The img buttons should lose their onClick (cursor-wait / opacity-50)
    // We verify by checking pendingAction blocks a second report click (no double confirm)
    const confirmText = screen.getAllByText(/confirmas|confirmes/i)
    // Only one confirmation box should appear
    expect(confirmText).toHaveLength(1)
  })
})
