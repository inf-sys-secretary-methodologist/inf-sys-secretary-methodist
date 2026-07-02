// CalendarSubscription mirrors the backend CalendarSubscriptionOutput: the
// user's iCal feed subscription state and its secret URL (empty when none).
export interface CalendarSubscription {
  subscribed: boolean
  url?: string
}
