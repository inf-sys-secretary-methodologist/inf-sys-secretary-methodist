import {
  scheduleLessonsApi,
  scheduleChangesApi,
  classroomsApi,
  studentGroupsApi,
  disciplinesApi,
  semestersApi,
  lessonTypesApi,
} from '../schedule'
import { apiClient } from '../../api'

// Backfill coverage for the schedule reference/lesson API wrappers. These
// unwrap the {success,data} envelope (apiClient does not) and default list
// endpoints to [] on an empty payload. useSchedule.test.ts mocks this module,
// so the wrappers themselves were previously never exercised directly.

jest.mock('../../api', () => ({
  apiClient: { get: jest.fn(), post: jest.fn(), put: jest.fn(), delete: jest.fn() },
}))

const mocked = jest.mocked(apiClient)
const ok = (data: unknown) => ({ success: true, data })

beforeEach(() => jest.clearAllMocks())

describe('scheduleLessonsApi', () => {
  it('list unwraps the envelope with filter params, [] when empty', async () => {
    mocked.get.mockResolvedValueOnce(ok([{ id: 1 }])).mockResolvedValueOnce(ok(null))
    expect(await scheduleLessonsApi.list({ semester_id: 3 })).toEqual([{ id: 1 }])
    expect(mocked.get).toHaveBeenCalledWith('/api/schedule/lessons', { params: { semester_id: 3 } })
    expect(await scheduleLessonsApi.list()).toEqual([])
  })

  it('getTimetable / getById unwrap', async () => {
    mocked.get.mockResolvedValueOnce(ok([{ id: 2 }])).mockResolvedValueOnce(ok({ id: 9 }))
    expect(await scheduleLessonsApi.getTimetable({ group_id: 5 })).toEqual([{ id: 2 }])
    expect(mocked.get).toHaveBeenCalledWith('/api/schedule/lessons/timetable', {
      params: { group_id: 5 },
    })
    expect(await scheduleLessonsApi.getById(9)).toEqual({ id: 9 })
    expect(mocked.get).toHaveBeenCalledWith('/api/schedule/lessons/9')
  })

  it('create / update / delete hit the right endpoints', async () => {
    mocked.post.mockResolvedValue(ok({ id: 10 }))
    mocked.put.mockResolvedValue(ok({ id: 11 }))
    mocked.delete.mockResolvedValue(undefined)
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    expect(await scheduleLessonsApi.create({ semester_id: 1 } as any)).toEqual({ id: 10 })
    expect(mocked.post).toHaveBeenCalledWith('/api/schedule/lessons', { semester_id: 1 })
    expect(await scheduleLessonsApi.update(11, { classroom_id: 2 })).toEqual({ id: 11 })
    expect(mocked.put).toHaveBeenCalledWith('/api/schedule/lessons/11', { classroom_id: 2 })
    await scheduleLessonsApi.delete(12)
    expect(mocked.delete).toHaveBeenCalledWith('/api/schedule/lessons/12')
  })
})

describe('scheduleChangesApi', () => {
  it('create posts, list unwraps and defaults to []', async () => {
    mocked.post.mockResolvedValue(ok({ id: 20 }))
    mocked.get.mockResolvedValue(ok(null))
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    expect(await scheduleChangesApi.create({ lesson_id: 1 } as any)).toEqual({ id: 20 })
    expect(mocked.post).toHaveBeenCalledWith('/api/schedule/changes', { lesson_id: 1 })
    expect(await scheduleChangesApi.list({ lesson_id: 1 })).toEqual([])
  })
})

describe('reference list APIs', () => {
  it.each([
    [classroomsApi, '/api/classrooms'],
    [studentGroupsApi, '/api/student-groups'],
    [disciplinesApi, '/api/disciplines'],
    [semestersApi, '/api/semesters'],
    [lessonTypesApi, '/api/lesson-types'],
  ] as const)('list unwraps %s', async (api, url) => {
    mocked.get.mockResolvedValue(ok([{ id: 1 }]))
    expect(await api.list()).toEqual([{ id: 1 }])
    expect(mocked.get).toHaveBeenCalledWith(url)
  })

  it('falls back to [] on an empty payload', async () => {
    mocked.get.mockResolvedValue(ok(null))
    expect(await semestersApi.list()).toEqual([])
  })
})
