'use client'

import useSWR from 'swr'
import { SWR_DEDUPING } from '@/config/swr'
import {
  scheduleLessonsApi,
  scheduleChangesApi,
  classroomsApi,
  studentGroupsApi,
  disciplinesApi,
  semestersApi,
  lessonTypesApi,
} from '@/lib/api/schedule'
import type {
  Lesson,
  Classroom,
  StudentGroup,
  Discipline,
  LessonTypeInfo,
  Semester,
  LessonFilterParams,
  CreateLessonInput,
  UpdateLessonInput,
  CreateChangeInput,
} from '@/types/schedule'

// Fetchers that return data directly (SWR expects Promise<T>)
const timetableFetcher = (key: string): Promise<Lesson[]> => {
  const params = JSON.parse(key.split('::')[1] || '{}') as LessonFilterParams
  return scheduleLessonsApi.getTimetable(params)
}

const classroomsFetcher = (): Promise<Classroom[]> => classroomsApi.list()
const studentGroupsFetcher = (): Promise<StudentGroup[]> => studentGroupsApi.list()
const disciplinesFetcher = (): Promise<Discipline[]> => disciplinesApi.list()
const semestersFetcher = (): Promise<Semester[]> => semestersApi.list()
const lessonTypesFetcher = (): Promise<LessonTypeInfo[]> => lessonTypesApi.list()

/**
 * Fetch timetable lessons with optional filters.
 */
export function useScheduleTimetable(params?: LessonFilterParams) {
  const key = `schedule-timetable::${JSON.stringify(params || {})}`

  const { data, error, isLoading, mutate } = useSWR<Lesson[]>(key, timetableFetcher, {
    revalidateOnFocus: false,
    dedupingInterval: SWR_DEDUPING.SHORT,
  })

  return {
    lessons: data || [],
    isLoading,
    error,
    mutate,
  }
}

/**
 * Fetch all classrooms.
 */
export function useClassrooms() {
  const { data, error, isLoading } = useSWR<Classroom[]>(
    'classrooms',
    classroomsFetcher,
    { revalidateOnFocus: false, dedupingInterval: SWR_DEDUPING.LONG }
  )

  return { classrooms: data || [], isLoading, error }
}

/**
 * Fetch all student groups.
 */
export function useStudentGroups() {
  const { data, error, isLoading } = useSWR<StudentGroup[]>(
    'student-groups',
    studentGroupsFetcher,
    { revalidateOnFocus: false, dedupingInterval: SWR_DEDUPING.LONG }
  )

  return { groups: data || [], isLoading, error }
}

/**
 * Fetch all disciplines.
 */
export function useDisciplines() {
  const { data, error, isLoading } = useSWR<Discipline[]>(
    'disciplines',
    disciplinesFetcher,
    { revalidateOnFocus: false, dedupingInterval: SWR_DEDUPING.LONG }
  )

  return { disciplines: data || [], isLoading, error }
}

/**
 * Fetch all semesters.
 */
export function useSemesters() {
  const { data, error, isLoading } = useSWR<Semester[]>(
    'semesters',
    semestersFetcher,
    { revalidateOnFocus: false, dedupingInterval: SWR_DEDUPING.LONG }
  )

  return { semesters: data || [], isLoading, error }
}

/**
 * Fetch all lesson types.
 */
export function useLessonTypes() {
  const { data, error, isLoading } = useSWR<LessonTypeInfo[]>(
    'lesson-types',
    lessonTypesFetcher,
    { revalidateOnFocus: false, dedupingInterval: SWR_DEDUPING.LONG }
  )

  return { lessonTypes: data || [], isLoading, error }
}

// --- Mutation functions (not hooks) ---

export async function createLesson(input: CreateLessonInput): Promise<Lesson> {
  return scheduleLessonsApi.create(input)
}

export async function updateLesson(id: number, input: UpdateLessonInput): Promise<Lesson> {
  return scheduleLessonsApi.update(id, input)
}

export async function deleteLesson(id: number): Promise<void> {
  return scheduleLessonsApi.delete(id)
}

export async function createScheduleChange(input: CreateChangeInput): Promise<void> {
  await scheduleChangesApi.create(input)
}
