import { apiClient } from '../api'

// Risk level types
export type RiskLevel = 'low' | 'medium' | 'high' | 'critical'

// Risk factors breakdown
export interface RiskFactors {
  attendance_weight: number
  grade_weight: number
  activity_weight: number
  attendance_score: number
  grade_score: number
  activity_score: number
}

// Student risk response
export interface StudentRiskInfo {
  student_id: number
  student_name: string
  group_name?: string
  attendance_rate?: number
  grade_average?: number
  risk_level: RiskLevel
  risk_score: number
  risk_factors?: RiskFactors
}

// At-risk students list response
export interface AtRiskStudentsResponse {
  students: StudentRiskInfo[]
  total: number
  page: number
  page_size: number
}

// Risk distribution in a group
export interface RiskDistribution {
  critical: number
  high: number
  medium: number
  low: number
}

// Group summary response
export interface GroupSummaryInfo {
  group_name: string
  total_students: number
  avg_attendance_rate: number
  avg_grade: number
  risk_distribution: RiskDistribution
  at_risk_percentage: number
}

// All groups summary response
export interface AllGroupsSummaryResponse {
  groups: GroupSummaryInfo[]
  total: number
}

// Monthly attendance trend
export interface MonthlyTrendInfo {
  month: string // Format: "2024-01"
  unique_students: number
  total_records: number
  present_count: number
  absent_count: number
  attendance_rate: number
}

// Attendance trend response
export interface AttendanceTrendResponse {
  trends: MonthlyTrendInfo[]
  months: number
}

// Attendance record
export interface AttendanceRecord {
  id: number
  student_id: number
  lesson_id: number
  lesson_date: string
  status: 'present' | 'absent' | 'late' | 'excused'
  marked_by?: number
  notes?: string
  created_at: string
}

// Lesson attendance summary
export interface LessonAttendanceSummary {
  total: number
  present: number
  absent: number
  late: number
  excused: number
}

// Lesson attendance response
export interface LessonAttendanceResponse {
  lesson_id: number
  lesson_date: string
  records: AttendanceRecord[]
  summary: LessonAttendanceSummary
}

// Mark attendance request
export interface MarkAttendanceRequest {
  student_id: number
  lesson_id: number
  lesson_date: string // Format: "2024-01-15"
  status: 'present' | 'absent' | 'late' | 'excused'
  notes?: string
}

// Bulk attendance record
export interface BulkAttendanceRecord {
  student_id: number
  status: 'present' | 'absent' | 'late' | 'excused'
  notes?: string
}

// Bulk mark attendance request
export interface BulkMarkAttendanceRequest {
  lesson_id: number
  lesson_date: string
  records: BulkAttendanceRecord[]
}

// Risk weight configuration
export interface RiskWeightConfig {
  attendance_weight: number
  grade_weight: number
  submission_weight: number
  inactivity_weight: number
  high_risk_threshold: number
  critical_risk_threshold: number
  updated_at: string
}

// Risk history entry
export interface RiskHistoryEntry {
  risk_score: number
  risk_level: string
  attendance_rate?: number
  grade_average?: number
  submission_rate?: number
  calculated_at: string
}

// Risk history response
export interface RiskHistoryResponse {
  student_id: number
  history: RiskHistoryEntry[]
  total: number
}

// API response wrapper
interface ApiResponse<T> {
  success: boolean
  data: T
  meta?: {
    timestamp: string
  }
}

export const analyticsApi = {
  /**
   * Get students at risk (paginated)
   */
  async getAtRiskStudents(
    page: number = 1,
    pageSize: number = 20
  ): Promise<AtRiskStudentsResponse> {
    const response = await apiClient.get<ApiResponse<AtRiskStudentsResponse>>(
      '/api/analytics/at-risk-students',
      { params: { page, page_size: pageSize } }
    )
    return response.data
  },

  /**
   * Get risk assessment for a specific student
   */
  async getStudentRisk(studentId: number): Promise<StudentRiskInfo> {
    const response = await apiClient.get<ApiResponse<StudentRiskInfo>>(
      `/api/analytics/students/${studentId}/risk`
    )
    return response.data
  },

  /**
   * Get analytics summary for a specific group
   */
  async getGroupSummary(groupName: string): Promise<GroupSummaryInfo> {
    const response = await apiClient.get<ApiResponse<GroupSummaryInfo>>(
      `/api/analytics/groups/${encodeURIComponent(groupName)}/summary`
    )
    return response.data
  },

  /**
   * Get analytics summary for all groups
   */
  async getAllGroupsSummary(): Promise<GroupSummaryInfo[]> {
    const response = await apiClient.get<ApiResponse<AllGroupsSummaryResponse>>(
      '/api/analytics/groups/summary'
    )
    return response.data.groups || []
  },

  /**
   * Get students filtered by risk level
   */
  async getStudentsByRiskLevel(
    level: RiskLevel,
    page: number = 1,
    pageSize: number = 20
  ): Promise<AtRiskStudentsResponse> {
    const response = await apiClient.get<ApiResponse<AtRiskStudentsResponse>>(
      `/api/analytics/risk-level/${level}`,
      { params: { page, page_size: pageSize } }
    )
    return response.data
  },

  /**
   * Get monthly attendance trend
   */
  async getAttendanceTrend(months: number = 6): Promise<MonthlyTrendInfo[]> {
    const response = await apiClient.get<ApiResponse<AttendanceTrendResponse>>(
      '/api/analytics/attendance-trend',
      { params: { months } }
    )
    return response.data.trends || []
  },

  /**
   * Mark attendance for a student
   */
  async markAttendance(params: MarkAttendanceRequest): Promise<AttendanceRecord> {
    const response = await apiClient.post<ApiResponse<AttendanceRecord>>(
      '/api/analytics/attendance',
      params
    )
    return response.data
  },

  /**
   * Mark attendance for multiple students at once
   */
  async bulkMarkAttendance(params: BulkMarkAttendanceRequest): Promise<AttendanceRecord[]> {
    const response = await apiClient.post<ApiResponse<AttendanceRecord[]>>(
      '/api/analytics/attendance/bulk',
      params
    )
    return response.data
  },

  /**
   * Get attendance records for a specific lesson
   */
  async getLessonAttendance(
    lessonId: number,
    lessonDate: string
  ): Promise<LessonAttendanceResponse> {
    const response = await apiClient.get<ApiResponse<LessonAttendanceResponse>>(
      `/api/analytics/lessons/${lessonId}/attendance`,
      { params: { date: lessonDate } }
    )
    return response.data
  },

  /**
   * Get risk score history for a student
   */
  async getStudentRiskHistory(
    studentId: number,
    limit: number = 90
  ): Promise<RiskHistoryResponse> {
    const response = await apiClient.get<ApiResponse<RiskHistoryResponse>>(
      `/api/analytics/students/${studentId}/risk-history`,
      { params: { limit } }
    )
    return response.data
  },

  /**
   * Get risk weight configuration
   */
  async getRiskWeightConfig(): Promise<RiskWeightConfig> {
    const response = await apiClient.get<ApiResponse<RiskWeightConfig>>(
      '/api/analytics/config/weights'
    )
    return response.data
  },

  /**
   * Update risk weight configuration (admin only)
   */
  async updateRiskWeightConfig(config: Omit<RiskWeightConfig, 'updated_at'>): Promise<void> {
    await apiClient.put('/api/analytics/config/weights', config)
  },

  /**
   * Export at-risk students as CSV or XLSX
   */
  async exportAtRiskStudents(format: 'csv' | 'xlsx' = 'csv'): Promise<Blob> {
    const response = await apiClient.get(`/api/analytics/export`, {
      params: { format },
      responseType: 'blob',
    })
    return (response as { data: Blob }).data
  },
}
