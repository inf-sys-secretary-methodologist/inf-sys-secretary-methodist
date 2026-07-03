import { scheduleGenerateApi } from '../schedule'
import { apiClient } from '../../api'
import type { GenerateScheduleRequest } from '@/types/schedule'

jest.mock('../../api', () => ({
  apiClient: {
    get: jest.fn(),
    post: jest.fn(),
    put: jest.fn(),
    delete: jest.fn(),
  },
}))

const mocked = jest.mocked(apiClient)

const req: GenerateScheduleRequest = { semester_id: 3, days: [1, 2, 3] }

describe('scheduleGenerateApi.preview', () => {
  beforeEach(() => jest.clearAllMocks())

  it('posts the request to /generate and unwraps the envelope', async () => {
    mocked.post.mockResolvedValue({
      success: true,
      data: {
        lessons: [{ load_id: 1, group_name: 'IS-21' }],
        unplaced: [],
        total_requested: 1,
        placed_count: 1,
        unplaced_count: 0,
      },
    })

    const res = await scheduleGenerateApi.preview(req)

    expect(mocked.post).toHaveBeenCalledWith('/api/schedule/generate', req)
    expect(res.placed_count).toBe(1)
    expect(res.lessons[0].group_name).toBe('IS-21')
  })
})

describe('scheduleGenerateApi.apply', () => {
  beforeEach(() => jest.clearAllMocks())

  it('posts the request to /generate/apply and unwraps the envelope', async () => {
    mocked.post.mockResolvedValue({
      success: true,
      data: { created: 12, unplaced: 3 },
    })

    const res = await scheduleGenerateApi.apply(req)

    expect(mocked.post).toHaveBeenCalledWith('/api/schedule/generate/apply', req)
    expect(res.created).toBe(12)
    expect(res.unplaced).toBe(3)
  })
})
