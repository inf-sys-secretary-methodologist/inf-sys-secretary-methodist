// Task module types matching backend DTO at internal/modules/tasks/application/dto/task_dto.go

export type TaskStatus =
  | 'new'
  | 'assigned'
  | 'in_progress'
  | 'review'
  | 'completed'
  | 'canceled'
  | 'deferred'

export type TaskPriority = 'low' | 'normal' | 'high' | 'urgent'

export type ProjectStatus = 'planning' | 'active' | 'on_hold' | 'completed' | 'canceled'

export const TASK_STATUSES: TaskStatus[] = [
  'new',
  'assigned',
  'in_progress',
  'review',
  'completed',
  'canceled',
  'deferred',
]

export const TASK_PRIORITIES: TaskPriority[] = ['low', 'normal', 'high', 'urgent']

export interface UserSummary {
  id: number
  name: string
  email: string
}

export interface ProjectSummary {
  id: number
  name: string
  status: ProjectStatus
}

export interface TaskChecklistItem {
  id: number
  checklist_id: number
  title: string
  is_completed: boolean
  position: number
  completed_by?: number
  completed_at?: string
  created_at: string
}

export interface TaskChecklist {
  id: number
  task_id: number
  title: string
  position: number
  completion_percentage: number
  created_at: string
  items?: TaskChecklistItem[]
}

export interface Task {
  id: number
  project_id?: number
  title: string
  description?: string
  document_id?: number
  author_id: number
  assignee_id?: number
  status: TaskStatus
  priority: TaskPriority
  due_date?: string
  start_date?: string
  completed_at?: string
  progress: number
  estimated_hours?: number
  actual_hours?: number
  tags?: string[]
  metadata?: Record<string, unknown>
  is_overdue: boolean
  created_at: string
  updated_at: string
  project?: ProjectSummary
  assignee?: UserSummary
  watchers?: UserSummary[]
  checklists?: TaskChecklist[]
}

export interface TaskListResponse {
  tasks: Task[]
  total: number
  limit: number
  offset: number
}

export interface TaskFilterParams {
  project_id?: number
  author_id?: number
  assignee_id?: number
  status?: TaskStatus
  priority?: TaskPriority
  is_overdue?: boolean
  search?: string
  tags?: string[]
  limit?: number
  offset?: number
}

export interface CreateTaskInput {
  project_id?: number
  title: string
  description?: string
  document_id?: number
  assignee_id?: number
  priority?: TaskPriority
  due_date?: string
  start_date?: string
  estimated_hours?: number
  tags?: string[]
  metadata?: Record<string, unknown>
}

export interface UpdateTaskInput {
  title?: string
  description?: string
  project_id?: number
  assignee_id?: number
  priority?: TaskPriority
  due_date?: string
  start_date?: string
  progress?: number
  estimated_hours?: number
  actual_hours?: number
  tags?: string[]
  metadata?: Record<string, unknown>
}
