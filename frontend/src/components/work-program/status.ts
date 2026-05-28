import { Archive, AlertTriangle, CheckCircle2, Clock, PenLine } from 'lucide-react'

import type { RevisionStatus, WorkProgramStatus } from '@/types/workProgram'

// statusKey collapses the wire-format multi-token statuses to the
// shorter camelCase UI keys used under workProgram.card.status.* /
// workProgram.filters.statusOptions.* / workProgram.detail.statusHint.*.
// Wire format remains backend-canonical (no string munging in the type
// layer) — translation through this explicit mapper.
export function statusKey(status: WorkProgramStatus): string {
  if (status === 'pending_approval') return 'pending'
  if (status === 'needs_revision') return 'needsRevision'
  return status
}

// revisionStatusKey is the analogue of statusKey for the revision child
// FSM (draft / pending_approval / approved / rejected). Only the
// pending_approval wire value needs collapsing to the short UI key
// `pending`; the rest map 1:1 to their workProgram.detail.revisionStatus
// i18n keys. Without this the detail page would interpolate
// `pending_approval` and render a raw missing key.
export function revisionStatusKey(status: RevisionStatus): string {
  return status === 'pending_approval' ? 'pending' : status
}

// STATUS_STYLES is the single source of truth for РПД status
// presentation: background tint, foreground text, and the lucide icon
// per lifecycle stage. Consumed by WorkProgramCard (list pill) and the
// detail page (header pill + status hint panel). needs_revision uses a
// distinct warning palette (orange + AlertTriangle) so it reads apart
// from pending_approval (amber + Clock) — they are different states in
// the FSM (ADR-2).
export const STATUS_STYLES: Record<
  WorkProgramStatus,
  { bg: string; text: string; Icon: typeof Clock }
> = {
  draft: {
    bg: 'bg-slate-100 dark:bg-slate-800/40',
    text: 'text-slate-700 dark:text-slate-300',
    Icon: PenLine,
  },
  pending_approval: {
    bg: 'bg-amber-50 dark:bg-amber-950/30',
    text: 'text-amber-700 dark:text-amber-300',
    Icon: Clock,
  },
  approved: {
    bg: 'bg-emerald-50 dark:bg-emerald-950/30',
    text: 'text-emerald-700 dark:text-emerald-300',
    Icon: CheckCircle2,
  },
  needs_revision: {
    bg: 'bg-orange-50 dark:bg-orange-950/30',
    text: 'text-orange-700 dark:text-orange-300',
    Icon: AlertTriangle,
  },
  archived: {
    bg: 'bg-zinc-100 dark:bg-zinc-800/40',
    text: 'text-zinc-600 dark:text-zinc-400',
    Icon: Archive,
  },
}
