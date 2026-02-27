export type MoodState =
  | 'happy'
  | 'content'
  | 'worried'
  | 'stressed'
  | 'panicking'
  | 'relaxed'
  | 'inspired'

export interface MoodResponse {
  state: MoodState
  intensity: number
  reason: string
  message: string
  greeting: string
  fun_fact?: string
  overdue_documents: number
  at_risk_students: number
  computed_at: string
}
