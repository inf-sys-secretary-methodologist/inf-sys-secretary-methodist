import {
  analyticsApi,
  StudentRiskInfo,
  AtRiskStudentsResponse,
  GroupSummaryInfo,
  MonthlyTrendInfo,
  AttendanceRecord,
  LessonAttendanceResponse,
} from '../analytics'
import { apiClient } from '../../api'

// Mock apiClient
jest.mock('../../api', () => ({
  apiClient: {
    get: jest.fn(),
    post: jest.fn(),
  },
}))

const mockApiClient = apiClient as jest.Mocked<typeof apiClient>

describe('analyticsApi', () => {
  beforeEach(() => {
    jest.clearAllMocks()
  })

  describe('getAtRiskStudents', () => {
    it('returns paginated at-risk students', async () => {
      const mockResponse: AtRiskStudentsResponse = {
        students: [
          {
            student_id: 1,
            student_name: 'John Doe',
            group_name: 'CS-101',
            attendance_rate: 65,
            grade_average: 3.2,
            risk_level: 'high',
            risk_score: 75,
          },
        ],
        total: 1,
        page: 1,
        page_size: 20,
      }

      mockApiClient.get.mockResolvedValueOnce({ data: mockResponse })

      const result = await analyticsApi.getAtRiskStudents()

      expect(mockApiClient.get).toHaveBeenCalledWith('/api/analytics/at-risk-students', {
        params: { page: 1, page_size: 20 },
      })
      expect(result).toEqual(mockResponse)
    })

    it('passes custom page and pageSize', async () => {
      const mockResponse: AtRiskStudentsResponse = {
        students: [],
        total: 0,
        page: 2,
        page_size: 10,
      }

      mockApiClient.get.mockResolvedValueOnce({ data: mockResponse })

      await analyticsApi.getAtRiskStudents(2, 10)

      expect(mockApiClient.get).toHaveBeenCalledWith('/api/analytics/at-risk-students', {
        params: { page: 2, page_size: 10 },
      })
    })
  })

  describe('getStudentRisk', () => {
    it('returns risk info for specific student', async () => {
      const mockStudent: StudentRiskInfo = {
        student_id: 1,
        student_name: 'John Doe',
        group_name: 'CS-101',
        attendance_rate: 70,
        grade_average: 3.5,
        risk_level: 'medium',
        risk_score: 55,
        risk_factors: {
          attendance_weight: 0.4,
          grade_weight: 0.4,
          activity_weight: 0.2,
          attendance_score: 60,
          grade_score: 50,
          activity_score: 55,
        },
      }

      mockApiClient.get.mockResolvedValueOnce({ data: mockStudent })

      const result = await analyticsApi.getStudentRisk(1)

      expect(mockApiClient.get).toHaveBeenCalledWith('/api/analytics/students/1/risk')
      expect(result).toEqual(mockStudent)
    })
  })

  describe('getGroupSummary', () => {
    it('returns summary for specific group', async () => {
      const mockSummary: GroupSummaryInfo = {
        group_name: 'CS-101',
        total_students: 30,
        avg_attendance_rate: 85,
        avg_grade: 3.7,
        risk_distribution: {
          critical: 1,
          high: 3,
          medium: 5,
          low: 21,
        },
        at_risk_percentage: 30,
      }

      mockApiClient.get.mockResolvedValueOnce({ data: mockSummary })

      const result = await analyticsApi.getGroupSummary('CS-101')

      expect(mockApiClient.get).toHaveBeenCalledWith('/api/analytics/groups/CS-101/summary')
      expect(result).toEqual(mockSummary)
    })

    it('encodes group name with special characters', async () => {
      const mockSummary: GroupSummaryInfo = {
        group_name: 'CS 101/A',
        total_students: 15,
        avg_attendance_rate: 90,
        avg_grade: 3.8,
        risk_distribution: { critical: 0, high: 1, medium: 2, low: 12 },
        at_risk_percentage: 20,
      }

      mockApiClient.get.mockResolvedValueOnce({ data: mockSummary })

      await analyticsApi.getGroupSummary('CS 101/A')

      expect(mockApiClient.get).toHaveBeenCalledWith('/api/analytics/groups/CS%20101%2FA/summary')
    })
  })

  describe('getAllGroupsSummary', () => {
    it('returns summary for all groups', async () => {
      const mockGroups: GroupSummaryInfo[] = [
        {
          group_name: 'CS-101',
          total_students: 30,
          avg_attendance_rate: 85,
          avg_grade: 3.7,
          risk_distribution: { critical: 1, high: 3, medium: 5, low: 21 },
          at_risk_percentage: 30,
        },
        {
          group_name: 'CS-102',
          total_students: 25,
          avg_attendance_rate: 90,
          avg_grade: 3.9,
          risk_distribution: { critical: 0, high: 2, medium: 3, low: 20 },
          at_risk_percentage: 20,
        },
      ]

      mockApiClient.get.mockResolvedValueOnce({ data: { groups: mockGroups } })

      const result = await analyticsApi.getAllGroupsSummary()

      expect(mockApiClient.get).toHaveBeenCalledWith('/api/analytics/groups/summary')
      expect(result).toEqual(mockGroups)
    })

    it('returns empty array when no groups', async () => {
      mockApiClient.get.mockResolvedValueOnce({ data: { groups: undefined } })

      const result = await analyticsApi.getAllGroupsSummary()

      expect(result).toEqual([])
    })
  })

  describe('getStudentsByRiskLevel', () => {
    it('returns students filtered by risk level', async () => {
      const mockResponse: AtRiskStudentsResponse = {
        students: [
          {
            student_id: 2,
            student_name: 'Jane Smith',
            risk_level: 'critical',
            risk_score: 90,
          },
        ],
        total: 1,
        page: 1,
        page_size: 20,
      }

      mockApiClient.get.mockResolvedValueOnce({ data: mockResponse })

      const result = await analyticsApi.getStudentsByRiskLevel('critical')

      expect(mockApiClient.get).toHaveBeenCalledWith('/api/analytics/risk-level/critical', {
        params: { page: 1, page_size: 20 },
      })
      expect(result).toEqual(mockResponse)
    })

    it('passes custom pagination params', async () => {
      mockApiClient.get.mockResolvedValueOnce({
        data: { students: [], total: 0, page: 3, page_size: 5 },
      })

      await analyticsApi.getStudentsByRiskLevel('high', 3, 5)

      expect(mockApiClient.get).toHaveBeenCalledWith('/api/analytics/risk-level/high', {
        params: { page: 3, page_size: 5 },
      })
    })
  })

  describe('getAttendanceTrend', () => {
    it('returns monthly attendance trends', async () => {
      const mockTrends: MonthlyTrendInfo[] = [
        {
          month: '2024-01',
          unique_students: 100,
          total_records: 500,
          present_count: 450,
          absent_count: 50,
          attendance_rate: 90,
        },
        {
          month: '2024-02',
          unique_students: 100,
          total_records: 480,
          present_count: 420,
          absent_count: 60,
          attendance_rate: 87.5,
        },
      ]

      mockApiClient.get.mockResolvedValueOnce({ data: { trends: mockTrends } })

      const result = await analyticsApi.getAttendanceTrend()

      expect(mockApiClient.get).toHaveBeenCalledWith('/api/analytics/attendance-trend', {
        params: { months: 6 },
      })
      expect(result).toEqual(mockTrends)
    })

    it('passes custom months parameter', async () => {
      mockApiClient.get.mockResolvedValueOnce({ data: { trends: [] } })

      await analyticsApi.getAttendanceTrend(12)

      expect(mockApiClient.get).toHaveBeenCalledWith('/api/analytics/attendance-trend', {
        params: { months: 12 },
      })
    })

    it('returns empty array when no trends', async () => {
      mockApiClient.get.mockResolvedValueOnce({ data: { trends: undefined } })

      const result = await analyticsApi.getAttendanceTrend()

      expect(result).toEqual([])
    })
  })

  describe('markAttendance', () => {
    it('marks attendance for a student', async () => {
      const mockRecord: AttendanceRecord = {
        id: 1,
        student_id: 1,
        lesson_id: 10,
        lesson_date: '2024-01-15',
        status: 'present',
        marked_by: 5,
        created_at: '2024-01-15T10:00:00Z',
      }

      mockApiClient.post.mockResolvedValueOnce({ data: mockRecord })

      const result = await analyticsApi.markAttendance({
        student_id: 1,
        lesson_id: 10,
        lesson_date: '2024-01-15',
        status: 'present',
      })

      expect(mockApiClient.post).toHaveBeenCalledWith('/api/analytics/attendance', {
        student_id: 1,
        lesson_id: 10,
        lesson_date: '2024-01-15',
        status: 'present',
      })
      expect(result).toEqual(mockRecord)
    })

    it('marks attendance with notes', async () => {
      const mockRecord: AttendanceRecord = {
        id: 2,
        student_id: 2,
        lesson_id: 10,
        lesson_date: '2024-01-15',
        status: 'excused',
        notes: 'Medical leave',
        created_at: '2024-01-15T10:00:00Z',
      }

      mockApiClient.post.mockResolvedValueOnce({ data: mockRecord })

      const result = await analyticsApi.markAttendance({
        student_id: 2,
        lesson_id: 10,
        lesson_date: '2024-01-15',
        status: 'excused',
        notes: 'Medical leave',
      })

      expect(mockApiClient.post).toHaveBeenCalledWith('/api/analytics/attendance', {
        student_id: 2,
        lesson_id: 10,
        lesson_date: '2024-01-15',
        status: 'excused',
        notes: 'Medical leave',
      })
      expect(result).toEqual(mockRecord)
    })
  })

  describe('bulkMarkAttendance', () => {
    it('marks attendance for multiple students', async () => {
      const mockRecords: AttendanceRecord[] = [
        {
          id: 1,
          student_id: 1,
          lesson_id: 10,
          lesson_date: '2024-01-15',
          status: 'present',
          created_at: '2024-01-15T10:00:00Z',
        },
        {
          id: 2,
          student_id: 2,
          lesson_id: 10,
          lesson_date: '2024-01-15',
          status: 'absent',
          created_at: '2024-01-15T10:00:00Z',
        },
      ]

      mockApiClient.post.mockResolvedValueOnce({ data: mockRecords })

      const result = await analyticsApi.bulkMarkAttendance({
        lesson_id: 10,
        lesson_date: '2024-01-15',
        records: [
          { student_id: 1, status: 'present' },
          { student_id: 2, status: 'absent' },
        ],
      })

      expect(mockApiClient.post).toHaveBeenCalledWith('/api/analytics/attendance/bulk', {
        lesson_id: 10,
        lesson_date: '2024-01-15',
        records: [
          { student_id: 1, status: 'present' },
          { student_id: 2, status: 'absent' },
        ],
      })
      expect(result).toEqual(mockRecords)
    })
  })

  describe('getLessonAttendance', () => {
    it('returns attendance records for a lesson', async () => {
      const mockResponse: LessonAttendanceResponse = {
        lesson_id: 10,
        lesson_date: '2024-01-15',
        records: [
          {
            id: 1,
            student_id: 1,
            lesson_id: 10,
            lesson_date: '2024-01-15',
            status: 'present',
            created_at: '2024-01-15T10:00:00Z',
          },
          {
            id: 2,
            student_id: 2,
            lesson_id: 10,
            lesson_date: '2024-01-15',
            status: 'absent',
            created_at: '2024-01-15T10:00:00Z',
          },
        ],
        summary: {
          total: 2,
          present: 1,
          absent: 1,
          late: 0,
          excused: 0,
        },
      }

      mockApiClient.get.mockResolvedValueOnce({ data: mockResponse })

      const result = await analyticsApi.getLessonAttendance(10, '2024-01-15')

      expect(mockApiClient.get).toHaveBeenCalledWith('/api/analytics/lessons/10/attendance', {
        params: { date: '2024-01-15' },
      })
      expect(result).toEqual(mockResponse)
    })
  })
})
