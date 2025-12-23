/**
 * Report Builder Types and Interfaces
 *
 * Defines the structure for custom report builder
 */

export type DataSourceType = 'documents' | 'users' | 'events' | 'tasks' | 'students'

export interface ReportField {
  id: string
  name: string
  label: string
  type: 'string' | 'number' | 'date' | 'boolean' | 'enum'
  source: DataSourceType
  enumValues?: string[]
}

export interface SelectedField {
  field: ReportField
  order: number
  alias?: string
  aggregation?: 'count' | 'sum' | 'avg' | 'min' | 'max'
}

export type FilterOperator =
  | 'equals'
  | 'not_equals'
  | 'contains'
  | 'not_contains'
  | 'starts_with'
  | 'ends_with'
  | 'greater_than'
  | 'less_than'
  | 'greater_or_equal'
  | 'less_or_equal'
  | 'between'
  | 'in'
  | 'not_in'
  | 'is_null'
  | 'is_not_null'

export interface ReportFilter {
  id: string
  field: ReportField
  operator: FilterOperator
  value: string | number | boolean | Date | null
  value2?: string | number | Date // For 'between' operator
}

export interface ReportGrouping {
  field: ReportField
  order: 'asc' | 'desc'
}

export interface ReportSorting {
  field: ReportField
  order: 'asc' | 'desc'
}

export interface ReportTemplate {
  id: string
  name: string
  description?: string
  dataSource: DataSourceType
  fields: SelectedField[]
  filters: ReportFilter[]
  groupings: ReportGrouping[]
  sortings: ReportSorting[]
  createdAt: Date
  updatedAt: Date
  createdBy: number
  isPublic: boolean
}

export interface ReportPreviewData {
  columns: { key: string; label: string }[]
  rows: Record<string, unknown>[]
  totalCount: number
}

export interface ReportExportOptions {
  format: 'pdf' | 'xlsx' | 'csv'
  includeHeaders: boolean
  pageSize?: 'A4' | 'A3' | 'Letter'
  orientation?: 'portrait' | 'landscape'
}

// Available fields for each data source
export const AVAILABLE_FIELDS: Record<DataSourceType, ReportField[]> = {
  documents: [
    { id: 'doc_id', name: 'id', label: 'ID', type: 'string', source: 'documents' },
    { id: 'doc_name', name: 'name', label: 'Name', type: 'string', source: 'documents' },
    {
      id: 'doc_category',
      name: 'category',
      label: 'Category',
      type: 'enum',
      source: 'documents',
      enumValues: ['educational', 'hr', 'administrative', 'methodical', 'financial', 'archive'],
    },
    {
      id: 'doc_status',
      name: 'status',
      label: 'Status',
      type: 'enum',
      source: 'documents',
      enumValues: ['uploading', 'processing', 'ready', 'error'],
    },
    { id: 'doc_size', name: 'size', label: 'Size', type: 'number', source: 'documents' },
    {
      id: 'doc_created',
      name: 'created_at',
      label: 'Created At',
      type: 'date',
      source: 'documents',
    },
    {
      id: 'doc_updated',
      name: 'updated_at',
      label: 'Updated At',
      type: 'date',
      source: 'documents',
    },
    { id: 'doc_author', name: 'author_name', label: 'Author', type: 'string', source: 'documents' },
    { id: 'doc_tags', name: 'tags', label: 'Tags', type: 'string', source: 'documents' },
  ],
  users: [
    { id: 'user_id', name: 'id', label: 'ID', type: 'number', source: 'users' },
    { id: 'user_name', name: 'name', label: 'Name', type: 'string', source: 'users' },
    { id: 'user_email', name: 'email', label: 'Email', type: 'string', source: 'users' },
    {
      id: 'user_role',
      name: 'role',
      label: 'Role',
      type: 'enum',
      source: 'users',
      enumValues: ['admin', 'methodist', 'secretary', 'teacher', 'student'],
    },
    {
      id: 'user_department',
      name: 'department',
      label: 'Department',
      type: 'string',
      source: 'users',
    },
    { id: 'user_created', name: 'created_at', label: 'Created At', type: 'date', source: 'users' },
    { id: 'user_active', name: 'is_active', label: 'Is Active', type: 'boolean', source: 'users' },
  ],
  events: [
    { id: 'event_id', name: 'id', label: 'ID', type: 'number', source: 'events' },
    { id: 'event_title', name: 'title', label: 'Title', type: 'string', source: 'events' },
    {
      id: 'event_type',
      name: 'type',
      label: 'Type',
      type: 'enum',
      source: 'events',
      enumValues: ['lecture', 'seminar', 'exam', 'meeting', 'other'],
    },
    { id: 'event_start', name: 'start_time', label: 'Start Time', type: 'date', source: 'events' },
    { id: 'event_end', name: 'end_time', label: 'End Time', type: 'date', source: 'events' },
    { id: 'event_location', name: 'location', label: 'Location', type: 'string', source: 'events' },
    {
      id: 'event_organizer',
      name: 'organizer',
      label: 'Organizer',
      type: 'string',
      source: 'events',
    },
  ],
  tasks: [
    { id: 'task_id', name: 'id', label: 'ID', type: 'number', source: 'tasks' },
    { id: 'task_title', name: 'title', label: 'Title', type: 'string', source: 'tasks' },
    {
      id: 'task_status',
      name: 'status',
      label: 'Status',
      type: 'enum',
      source: 'tasks',
      enumValues: ['pending', 'in_progress', 'completed', 'cancelled'],
    },
    {
      id: 'task_priority',
      name: 'priority',
      label: 'Priority',
      type: 'enum',
      source: 'tasks',
      enumValues: ['low', 'medium', 'high', 'urgent'],
    },
    { id: 'task_due', name: 'due_date', label: 'Due Date', type: 'date', source: 'tasks' },
    { id: 'task_assignee', name: 'assignee', label: 'Assignee', type: 'string', source: 'tasks' },
    { id: 'task_created', name: 'created_at', label: 'Created At', type: 'date', source: 'tasks' },
  ],
  students: [
    { id: 'student_id', name: 'id', label: 'ID', type: 'number', source: 'students' },
    { id: 'student_name', name: 'name', label: 'Name', type: 'string', source: 'students' },
    { id: 'student_group', name: 'group', label: 'Group', type: 'string', source: 'students' },
    { id: 'student_course', name: 'course', label: 'Course', type: 'number', source: 'students' },
    {
      id: 'student_faculty',
      name: 'faculty',
      label: 'Faculty',
      type: 'string',
      source: 'students',
    },
    {
      id: 'student_status',
      name: 'status',
      label: 'Status',
      type: 'enum',
      source: 'students',
      enumValues: ['active', 'academic_leave', 'expelled', 'graduated'],
    },
    {
      id: 'student_enrolled',
      name: 'enrolled_at',
      label: 'Enrolled At',
      type: 'date',
      source: 'students',
    },
  ],
}
