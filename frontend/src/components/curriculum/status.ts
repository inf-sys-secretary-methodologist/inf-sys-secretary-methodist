import { Archive, CheckCircle2, Clock, PenLine } from 'lucide-react'

import type { CurriculumStatus } from '@/types/curriculum'

// statusKey collapses the wire-format 'pending_approval' to the
// shorter UI key 'pending'. UI translation keys live under
// `curriculum.card.status.*` / `curriculum.filters.statusOptions.*` /
// `curriculum.detail.statusHint.*` and use the short form for
// brevity. Wire format remains backend-canonical (no string trim
// in the type layer) — translation through this explicit mapper.
export function statusKey(status: CurriculumStatus): string {
  return status === 'pending_approval' ? 'pending' : status
}

// STATUS_STYLES is the single source of truth for curriculum status
// presentation: background tint, foreground text, and the lucide
// icon paired with each lifecycle stage. Consumed by CurriculumCard
// (list pill) and the detail page (header pill + status hint
// panel). Centralising here ensures a recolor / icon change touches
// one site, not two.
export const STATUS_STYLES: Record<
  CurriculumStatus,
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
  archived: {
    bg: 'bg-zinc-100 dark:bg-zinc-800/40',
    text: 'text-zinc-600 dark:text-zinc-400',
    Icon: Archive,
  },
}
