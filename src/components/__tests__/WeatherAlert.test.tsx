import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import { WeatherAlert } from '../WeatherAlert'
import { LanguageProvider } from '../../context/LanguageContext'

// Wrap with LanguageProvider (provides 'ca' by default via browser detection fallback)
function renderAlert() {
  return render(
    <LanguageProvider>
      <WeatherAlert />
    </LanguageProvider>
  )
}

function mockFetch(data: object | null, ok = true) {
  vi.stubGlobal(
    'fetch',
    vi.fn().mockResolvedValue({
      ok,
      json: () => Promise.resolve(data),
    })
  )
}

beforeEach(() => {
  // Reset location.search to empty (no ?mock=1)
  Object.defineProperty(window, 'location', {
    writable: true,
    value: { ...window.location, search: '' },
  })
})

afterEach(() => {
  vi.restoreAllMocks()
})

describe('WeatherAlert', () => {
  it('renders nothing when alert is false', async () => {
    mockFetch({ alert: false, days_until: 0, mm: 0, prob: 0 })
    const { container } = renderAlert()
    await waitFor(() => expect(fetch).toHaveBeenCalledWith('/api/weather'))
    expect(container.firstChild).toBeNull()
  })

  it('renders nothing when fetch fails', async () => {
    vi.stubGlobal('fetch', vi.fn().mockRejectedValue(new Error('network')))
    const { container } = renderAlert()
    // Wait a tick for the effect to settle
    await waitFor(() => expect(fetch).toHaveBeenCalled())
    expect(container.firstChild).toBeNull()
  })

  it('renders nothing when response is not ok', async () => {
    mockFetch(null, false)
    const { container } = renderAlert()
    await waitFor(() => expect(fetch).toHaveBeenCalled())
    expect(container.firstChild).toBeNull()
  })

  it('renders banner when alert is true with days_until > 0', async () => {
    mockFetch({ alert: true, days_until: 3, mm: 18.5, prob: 75 })
    renderAlert()
    // Expect the alert role to appear
    const banner = await screen.findByRole('alert')
    expect(banner).toBeInTheDocument()
    // Check days appear in the text
    expect(banner).toHaveTextContent('3')
    // Check prob and mm
    expect(banner).toHaveTextContent('75%')
    expect(banner).toHaveTextContent('18.5 mm')
  })

  it('renders "today" message when days_until is 0 (no day number in text)', async () => {
    mockFetch({ alert: true, days_until: 0, mm: 8.2, prob: 82 })
    renderAlert()
    const banner = await screen.findByRole('alert')
    // Neither "en 0 dies" nor "en 0 días" should appear — the "today" variant is used
    expect(banner).not.toHaveTextContent('en 0')
    // The banner must still show prob and mm
    expect(banner).toHaveTextContent('82%')
    expect(banner).toHaveTextContent('8.2 mm')
  })

  it('uses mock data when ?mock=1 is in the URL without fetching', async () => {
    Object.defineProperty(window, 'location', {
      writable: true,
      value: { ...window.location, search: '?mock=1' },
    })
    const fetchSpy = vi.fn()
    vi.stubGlobal('fetch', fetchSpy)
    renderAlert()
    const banner = await screen.findByRole('alert')
    expect(banner).toBeInTheDocument()
    // fetch should NOT have been called
    expect(fetchSpy).not.toHaveBeenCalled()
    // Mock data: days_until=3, prob=75
    expect(banner).toHaveTextContent('3')
    expect(banner).toHaveTextContent('75%')
  })

  it('mm is formatted to one decimal place', async () => {
    mockFetch({ alert: true, days_until: 1, mm: 12, prob: 70 })
    renderAlert()
    const banner = await screen.findByRole('alert')
    expect(banner).toHaveTextContent('12.0 mm')
  })
})
