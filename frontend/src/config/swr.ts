/**
 * SWR configuration constants for data fetching intervals.
 * All values are in milliseconds.
 */

/** Deduping intervals - prevent duplicate requests within time window */
export const SWR_DEDUPING = {
  /** 1 second - for rapidly changing data (e.g. typing indicators) */
  FAST: 1000,
  /** 5 seconds - for frequently accessed data (e.g. messages, calendar) */
  SHORT: 5000,
  /** 10 seconds - for moderately accessed data (e.g. activity feeds) */
  MEDIUM: 10000,
  /** 15 seconds - for notification counts */
  NOTIFICATIONS: 15000,
  /** 30 seconds - for slowly changing data (e.g. dashboard stats, sync logs) */
  LONG: 30000,
  /** 60 seconds - for rarely changing data (e.g. notification stats) */
  EXTRA_LONG: 60000,
  /** Never dedupe - for static data that should always be fetched fresh */
  NONE: Infinity,
} as const

/** Refresh intervals - auto-polling for fresh data */
export const SWR_REFRESH = {
  /** 5 seconds - for active sync status polling */
  ACTIVE_POLL: 5000,
  /** 30 seconds - for real-time data (e.g. conversations, activity) */
  REALTIME: 30000,
  /** 60 seconds - for dashboard and notification data */
  STANDARD: 60000,
  /** 5 minutes - for slowly changing data (e.g. mood) */
  SLOW: 300000,
} as const
