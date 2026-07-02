import {
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

describe('calendar subscription actions', () => {
  beforeEach(() => {
    jest.clearAllMocks()
  })

  it('createCalendarSubscription posts to the base URL and returns the subscription', async () => {
    mocked.post.mockResolvedValue({
      subscribed: true,
      url: 'https://h/api/public/calendar/tok/feed.ics',
    })
    const res = await createCalendarSubscription()
    expect(mocked.post).toHaveBeenCalledWith(CALENDAR_SUBSCRIPTION_URL)
    expect(res).toEqual({ subscribed: true, url: 'https://h/api/public/calendar/tok/feed.ics' })
  })

  it('rotateCalendarSubscription posts to the rotate endpoint', async () => {
    mocked.post.mockResolvedValue({
      subscribed: true,
      url: 'https://h/api/public/calendar/new/feed.ics',
    })
    const res = await rotateCalendarSubscription()
    expect(mocked.post).toHaveBeenCalledWith(`${CALENDAR_SUBSCRIPTION_URL}/rotate`)
    expect(res.url).toContain('/new/feed.ics')
  })

  it('deleteCalendarSubscription calls delete on the base URL', async () => {
    mocked.delete.mockResolvedValue(undefined)
    await deleteCalendarSubscription()
    expect(mocked.delete).toHaveBeenCalledWith(CALENDAR_SUBSCRIPTION_URL)
  })
})
