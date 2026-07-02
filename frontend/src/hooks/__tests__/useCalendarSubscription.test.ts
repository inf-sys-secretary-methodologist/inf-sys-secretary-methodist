import React from 'react'
import { renderHook, waitFor } from '@testing-library/react'
import { SWRConfig } from 'swr'
import {
  useCalendarSubscription,
  createCalendarSubscription,
  rotateCalendarSubscription,
  deleteCalendarSubscription,
  CALENDAR_SUBSCRIPTION_URL,
} from '../useCalendarSubscription'
import { apiClient } from '@/lib/api'

jest.mock('@/lib/api', () => ({
  apiClient: {
    get: jest.fn(),
    post: jest.fn(),
    delete: jest.fn(),
  },
}))

const mocked = jest.mocked(apiClient)

const wrapper = ({ children }: { children: React.ReactNode }) =>
  React.createElement(
    SWRConfig,
    { value: { dedupingInterval: 0, provider: () => new Map() } },
    children
  )

describe('calendar subscription actions', () => {
  beforeEach(() => {
    jest.clearAllMocks()
  })

  it('createCalendarSubscription posts to the base URL', async () => {
    mocked.post.mockResolvedValue(undefined)
    await createCalendarSubscription()
    expect(mocked.post).toHaveBeenCalledWith(CALENDAR_SUBSCRIPTION_URL)
  })

  it('rotateCalendarSubscription posts to the rotate endpoint', async () => {
    mocked.post.mockResolvedValue(undefined)
    await rotateCalendarSubscription()
    expect(mocked.post).toHaveBeenCalledWith(`${CALENDAR_SUBSCRIPTION_URL}/rotate`)
  })

  it('deleteCalendarSubscription calls delete on the base URL', async () => {
    mocked.delete.mockResolvedValue(undefined)
    await deleteCalendarSubscription()
    expect(mocked.delete).toHaveBeenCalledWith(CALENDAR_SUBSCRIPTION_URL)
  })
})

describe('useCalendarSubscription', () => {
  beforeEach(() => {
    jest.clearAllMocks()
  })

  it('unwraps the API envelope and exposes the subscription', async () => {
    // swrFetcher calls apiClient.get and unwraps { success, data }.
    mocked.get.mockResolvedValue({
      success: true,
      data: { subscribed: true, url: 'https://h/api/public/calendar/tok/feed.ics' },
    })

    const { result } = renderHook(() => useCalendarSubscription(), { wrapper })

    await waitFor(() => expect(result.current.subscription).toBeDefined())
    expect(mocked.get).toHaveBeenCalledWith(CALENDAR_SUBSCRIPTION_URL)
    expect(result.current.subscription).toEqual({
      subscribed: true,
      url: 'https://h/api/public/calendar/tok/feed.ics',
    })
  })
})
