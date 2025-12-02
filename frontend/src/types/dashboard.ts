export interface StatItem {
  total: number
  change: number
  period: string
}

export interface DashboardStats {
  documents: StatItem
  students: StatItem
  events: StatItem
  reports: StatItem
  tasks: StatItem
}

export interface TrendPoint {
  date: string
  value: number
}

export interface DashboardTrends {
  documents_trend: TrendPoint[]
  reports_trend: TrendPoint[]
  tasks_trend: TrendPoint[]
  events_trend: TrendPoint[]
}

export interface ActivityItem {
  id: number
  type: 'document' | 'report' | 'task' | 'event' | 'announcement'
  action: string
  title: string
  description?: string
  user_id: number
  user_name: string
  created_at: string
}

export interface DashboardActivity {
  activities: ActivityItem[]
  total: number
}

export interface ExportInput {
  format: 'pdf' | 'xlsx'
  start_date?: string
  end_date?: string
  sections?: string[]
}

export interface ExportOutput {
  file_url: string
  file_name: string
  file_size: number
  expires_at: string
}
