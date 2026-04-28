import { apiClient } from '../api'
import type {
  Lesson,
  Classroom,
  StudentGroup,
  Discipline,
  LessonTypeInfo,
  Semester,
  ScheduleChange,
  CreateLessonInput,
  UpdateLessonInput,
  LessonFilterParams,
  CreateChangeInput,
} from '@/types/schedule'

// Backend API response wrappers
interface ApiResponse<T> {
  success: boolean
  data: T
  error?: { code: string; message: string }
}

interface ListResponse<T> {
  success: boolean
  data: T[]
}

export const scheduleLessonsApi = {
  async list(params?: LessonFilterParams): Promise<Lesson[]> {
    const response = await apiClient.get<ApiResponse<Lesson[]>>(
      '/api/schedule/lessons',
      { params }
    )
    return response.data || []
  },

  async getTimetable(params?: LessonFilterParams): Promise<Lesson[]> {
    const response = await apiClient.get<ApiResponse<Lesson[]>>(
      '/api/schedule/lessons/timetable',
      { params }
    )
    return response.data || []
  },

  async getById(id: number): Promise<Lesson> {
    const response = await apiClient.get<ApiResponse<Lesson>>(
      `/api/schedule/lessons/${id}`
    )
    return response.data
  },

  async create(input: CreateLessonInput): Promise<Lesson> {
    const response = await apiClient.post<ApiResponse<Lesson>>(
      '/api/schedule/lessons',
      input
    )
    return response.data
  },

  async update(id: number, input: UpdateLessonInput): Promise<Lesson> {
    const response = await apiClient.put<ApiResponse<Lesson>>(
      `/api/schedule/lessons/${id}`,
      input
    )
    return response.data
  },

  async delete(id: number): Promise<void> {
    await apiClient.delete('/api/schedule/lessons/' + id)
  },
}

export const scheduleChangesApi = {
  async create(input: CreateChangeInput): Promise<ScheduleChange> {
    const response = await apiClient.post<ApiResponse<ScheduleChange>>(
      '/api/schedule/changes',
      input
    )
    return response.data
  },

  async list(params?: { lesson_id?: number }): Promise<ScheduleChange[]> {
    const response = await apiClient.get<ListResponse<ScheduleChange>>(
      '/api/schedule/changes',
      { params }
    )
    return response.data || []
  },
}

export const classroomsApi = {
  async list(): Promise<Classroom[]> {
    const response = await apiClient.get<ApiResponse<Classroom[]>>('/api/classrooms')
    return response.data || []
  },
}

export const studentGroupsApi = {
  async list(): Promise<StudentGroup[]> {
    const response = await apiClient.get<ApiResponse<StudentGroup[]>>('/api/student-groups')
    return response.data || []
  },
}

export const disciplinesApi = {
  async list(): Promise<Discipline[]> {
    const response = await apiClient.get<ApiResponse<Discipline[]>>('/api/disciplines')
    return response.data || []
  },
}

export const semestersApi = {
  async list(): Promise<Semester[]> {
    const response = await apiClient.get<ApiResponse<Semester[]>>('/api/semesters')
    return response.data || []
  },
}

export const lessonTypesApi = {
  async list(): Promise<LessonTypeInfo[]> {
    const response = await apiClient.get<ApiResponse<LessonTypeInfo[]>>('/api/lesson-types')
    return response.data || []
  },
}
