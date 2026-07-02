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
// so the wrappers themselves were previously exercised only via the generate
// endpoints — this suite pins the envelope handling directly.

jest.mock('../../api', () => ({
  apiClient: {
    get: jest.fn(),
    post: jest.fn(),
    put: jest.fn(),
    delete: jest.fn(),
  },
}))

const mocked = jest.mocked(apiClient)

beforeEach(() => jest.clearAllMocks())

describe('scheduleLessonsApi', () => {
  it('list unwraps the envelope and passes filter params', async () => {
    mocked.get.mockResolvedValue({ success: true, data: [{ id: 1 }] })
    const res = await scheduleLessonsApi.list({ semester_id: 3 })
    expect(mocked.get).toHaveBeenCalledWith('/api/schedule/lessons', {
      params: { semester_id: 3 },
    })
    expect(res).toEqual([{ id: 1 }])
  })

  it('list falls back to [] when data is absent', async () => {
    mocked.get.mockResolvedValue({ success: true, data: null })
    expect(await scheduleLessonsApi.list()).toEqual([])
  })

  it('getTimetable unwraps the envelope', async () => {
    mocked.get.mockResolvedValue({ success: true, data: [{ id: 2 }] })
    const res = await scheduleLessonsApi.getTimetable({ group_id: 5 })
    expect(mocked.get).toHaveBeenCalledWith('/api/schedule/lessons/timetable', {
      params: { group_id: 5 },
    })
    expect(res).toEqual([{ id: 2 }])
  })

  it('getById unwraps a single lesson', async () => {
    mocked.get.mockResolvedValue({ success: true, data: { id: 9 } })
    const res = await scheduleLessonsApi.getById(9)
    expect(mocked.get).toHaveBeenCalledWith('/api/schedule/lessons/9')
    expect(res).toEqual({ id: 9 })
  })

  it('create posts the body and unwraps the created lesson', async () => {
    mocked.post.mockResolvedValue({ success: true, data: { id: 10 } })
    const input = {
      semester_id: 1,
      discipline_id: 1,
      lesson_type_id: 1,
      teacher_id: 1,
      group_id: 1,
      classroom_id: 1,
      day_of_week: 1,
      time_start: '09:00',
      time_end: '10:30',
      week_type: 'all' as const,
      date_start: '2026-09-01',
      date_end: '2026-12-31',
    }
    const res = await scheduleLessonsApi.create(input)
    expect(mocked.post).toHaveBeenCalledWith('/api/schedule/lessons', input)
    expect(res).toEqual({ id: 10 })
  })

  it('update puts to the id endpoint and unwraps', async () => {
    mocked.put.mockResolvedValue({ success: true, data: { id: 11 } })
    const res = await scheduleLessonsApi.update(11, { classroom_id: 2 })
    expect(mocked.put).toHaveBeenCalledWith('/api/schedule/lessons/11', { classroom_id: 2 })
    expect(res).toEqual({ id: 11 })
  })

  it('delete calls the id endpoint', async () => {
    mocked.delete.mockResolvedValue(undefined)
    await scheduleLessonsApi.delete(12)
    expect(mocked.delete).toHaveBeenCalledWith('/api/schedule/lessons/12')
  })
})

describe('scheduleChangesApi', () => {
  it('create posts the change and unwraps it', async () => {
    mocked.post.mockResolvedValue({ success: true, data: { id: 20 } })
    const input = {
      lesson_id: 1,
      change_type: 'cancelled' as const,
      original_date: '2026-09-10',
    }
    const res = await scheduleChangesApi.create(input)
    expect(mocked.post).toHaveBeenCalledWith('/api/schedule/changes', input)
    expect(res).toEqual({ id: 20 })
  })

  it('list unwraps and defaults to [] when empty', async () => {
    mocked.get.mockResolvedValue({ success: true, data: null })
    const res = await scheduleChangesApi.list({ lesson_id: 1 })
    expect(mocked.get).toHaveBeenCalledWith('/api/schedule/changes', {
      params: { lesson_id: 1 },
    })
    expect(res).toEqual([])
  })
})

describe('reference list APIs', () => {
  it.each([
    ['classroomsApi', classroomsApi, '/api/classrooms'],
    ['studentGroupsApi', studentGroupsApi, '/api/student-groups'],
    ['disciplinesApi', disciplinesApi, '/api/disciplines'],
    ['semestersApi', semestersApi, '/api/semesters'],
    ['lessonTypesApi', lessonTypesApi, '/api/lesson-types'],
  ] as const)('%s.list unwraps the list envelope', async (_name, api, url) => {
    mocked.get.mockResolvedValue({ success: true, data: [{ id: 1 }] })
    const res = await api.list()
    expect(mocked.get).toHaveBeenCalledWith(url)
    expect(res).toEqual([{ id: 1 }])
  })

  it('reference lists fall back to [] on empty payload', async () => {
    mocked.get.mockResolvedValue({ success: true, data: null })
    expect(await semestersApi.list()).toEqual([])
  })
})
