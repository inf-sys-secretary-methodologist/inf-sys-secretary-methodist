import React from 'react'
import { renderHook, waitFor } from '@testing-library/react'
import { SWRConfig } from 'swr'
import {
  useTeachingLoads,
  createTeachingLoad,
  updateTeachingLoad,
  deleteTeachingLoad,
  buildTeachingLoadUrl,
  TEACHING_LOAD_URL,
} from '../useTeachingLoad'
import { apiClient } from '@/lib/api'

jest.mock('@/lib/api', () => ({
  apiClient: {
    get: jest.fn(),
    post: jest.fn(),
    put: jest.fn(),
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

describe('buildTeachingLoadUrl', () => {
  it('returns base URL with no filter', () => {
    expect(buildTeachingLoadUrl()).toBe(TEACHING_LOAD_URL)
  })

  it('appends only defined filter params', () => {
    const url = buildTeachingLoadUrl({ semester_id: 1, teacher_id: 4 })
    expect(url).toContain('semester_id=1')
    expect(url).toContain('teacher_id=4')
    expect(url).not.toContain('group_id')
  })
})

describe('teaching load mutations', () => {
  beforeEach(() => jest.clearAllMocks())

  const input = {
    semester_id: 1,
    group_id: 2,
    discipline_id: 3,
    teacher_id: 4,
    lesson_type_id: 5,
    pairs_per_week: 2,
    week_type: 'all' as const,
  }

  it('createTeachingLoad posts and unwraps the envelope', async () => {
    mocked.post.mockResolvedValue({ success: true, data: { id: 9, ...input } })
    const res = await createTeachingLoad(input)
    expect(mocked.post).toHaveBeenCalledWith(TEACHING_LOAD_URL, input)
    expect(res.id).toBe(9)
  })

  it('updateTeachingLoad puts to the id endpoint and unwraps', async () => {
    mocked.put.mockResolvedValue({ success: true, data: { id: 7, ...input } })
    const res = await updateTeachingLoad(7, input)
    expect(mocked.put).toHaveBeenCalledWith(`${TEACHING_LOAD_URL}/7`, input)
    expect(res.id).toBe(7)
  })

  it('deleteTeachingLoad calls delete on the id endpoint', async () => {
    mocked.delete.mockResolvedValue(undefined)
    await deleteTeachingLoad(5)
    expect(mocked.delete).toHaveBeenCalledWith(`${TEACHING_LOAD_URL}/5`)
  })
})

describe('useTeachingLoads', () => {
  beforeEach(() => jest.clearAllMocks())

  it('unwraps the list envelope and applies the filter to the key', async () => {
    mocked.get.mockResolvedValue({
      success: true,
      data: { teaching_loads: [{ id: 1, group_name: 'IS-21' }] },
    })

    const { result } = renderHook(() => useTeachingLoads({ semester_id: 3 }), { wrapper })

    await waitFor(() => expect(result.current.items.length).toBe(1))
    expect(mocked.get).toHaveBeenCalledWith(`${TEACHING_LOAD_URL}?semester_id=3`)
    expect(result.current.items[0].group_name).toBe('IS-21')
  })
})
