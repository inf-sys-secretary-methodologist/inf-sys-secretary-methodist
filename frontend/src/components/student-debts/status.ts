import { AlertCircle, CalendarClock, CheckCircle2, Users, XCircle } from 'lucide-react'

import type { ControlForm, ResitResult, StudentDebtStatus } from '@/types/studentDebts'

// statusKey collapses the wire-format snake_case debt statuses to the
// camelCase UI keys used under studentDebts.card.status.* /
// studentDebts.filters.statusOptions.* / studentDebts.detail.statusHint.*.
// The wire format stays backend-canonical (no munging in the type layer);
// translation goes through this explicit table.
export function statusKey(status: StudentDebtStatus): string {
  switch (status) {
    case 'resit_scheduled':
      return 'resitScheduled'
    case 'closed_passed':
      return 'closedPassed'
    case 'closed_failed':
      return 'closedFailed'
    default:
      // open / commission map 1:1
      return status
  }
}

// controlFormKey maps the wire control-form enum to its camelCase i18n key
// under studentDebts.card.controlForm.* / studentDebts.detail (control form
// shown on the detail header).
export function controlFormKey(form: ControlForm): string {
  switch (form) {
    case 'course_project':
      return 'courseProject'
    case 'differential_zachet':
      return 'differentialZachet'
    default:
      // zachet / exam map 1:1
      return form
  }
}

// resitResultKey maps the wire resit-result enum to its camelCase i18n key
// under studentDebts.detail.resitResult.* (the attempt timeline) and
// studentDebts.recordDialog.resultOptions.*.
export function resitResultKey(result: ResitResult): string {
  return result === 'no_show' ? 'noShow' : result
}

// STATUS_STYLES is the single source of truth for debt-status presentation:
// background tint, foreground text, and the lucide icon per lifecycle stage.
// Consumed by StudentDebtCard (list pill) and the detail page (header pill +
// status hint panel). closed_passed (emerald + check) and closed_failed
// (rose + x) are kept visually distinct — they are different terminal states
// of the FSM, not a single "closed".
export const STATUS_STYLES: Record<
  StudentDebtStatus,
  { bg: string; text: string; Icon: typeof AlertCircle }
> = {
  open: {
    bg: 'bg-amber-50 dark:bg-amber-950/30',
    text: 'text-amber-700 dark:text-amber-300',
    Icon: AlertCircle,
  },
  resit_scheduled: {
    bg: 'bg-sky-50 dark:bg-sky-950/30',
    text: 'text-sky-700 dark:text-sky-300',
    Icon: CalendarClock,
  },
  commission: {
    bg: 'bg-violet-50 dark:bg-violet-950/30',
    text: 'text-violet-700 dark:text-violet-300',
    Icon: Users,
  },
  closed_passed: {
    bg: 'bg-emerald-50 dark:bg-emerald-950/30',
    text: 'text-emerald-700 dark:text-emerald-300',
    Icon: CheckCircle2,
  },
  closed_failed: {
    bg: 'bg-rose-50 dark:bg-rose-950/30',
    text: 'text-rose-700 dark:text-rose-300',
    Icon: XCircle,
  },
}
