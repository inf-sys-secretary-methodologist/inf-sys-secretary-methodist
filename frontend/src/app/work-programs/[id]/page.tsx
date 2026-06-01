'use client'

import { useMemo, useState, type ReactNode } from 'react'
import { useParams } from 'next/navigation'
import Link from 'next/link'
import { useTranslations } from 'next-intl'
import {
  Archive,
  ArrowLeft,
  BookMarked,
  Calendar,
  CheckCircle2,
  FilePlus2,
  GraduationCap,
  Loader2,
  Pencil,
  Plus,
  Send,
  Sparkles,
  Trash2,
  XCircle,
} from 'lucide-react'

import { AppLayout } from '@/components/layout'
import { Button } from '@/components/ui/button'
import { useWorkProgram } from '@/hooks/useWorkPrograms'
import { useAuthCheck } from '@/hooks/useAuth'
import { canApproveWorkProgram, canCreateWorkProgram } from '@/lib/auth/permissions'
import { GenerateWorkProgramDialog } from '@/components/work-program/GenerateWorkProgramDialog'
import { SubmitWorkProgramDialog } from '@/components/work-program/SubmitWorkProgramDialog'
import { DiscardWorkProgramDialog } from '@/components/work-program/DiscardWorkProgramDialog'
import { ApproveWorkProgramDialog } from '@/components/work-program/ApproveWorkProgramDialog'
import { RejectWorkProgramDialog } from '@/components/work-program/RejectWorkProgramDialog'
import { CreateRevisionDialog } from '@/components/work-program/CreateRevisionDialog'
import { SubmitRevisionDialog } from '@/components/work-program/SubmitRevisionDialog'
import { ApproveRevisionDialog } from '@/components/work-program/ApproveRevisionDialog'
import { RejectRevisionDialog } from '@/components/work-program/RejectRevisionDialog'
import { CollectionItemDialog } from '@/components/work-program/CollectionItemDialog'
import { DeleteCollectionItemDialog } from '@/components/work-program/DeleteCollectionItemDialog'
import {
  COLLECTION_CONFIG,
  type CollectionKind,
  type CollectionItem,
} from '@/components/work-program/collectionConfig'
import { STATUS_STYLES, statusKey, revisionStatusKey } from '@/components/work-program/status'
import type { WorkProgram, WorkProgramStatus } from '@/types/workProgram'
import { cn } from '@/lib/utils'

// WorkProgramDetailPage — full РПД view with all six inner collections.
// Author-side draft actions (submit / discard) and approver actions
// (approve / reject) are wired by role + status. Visible to all
// authenticated roles (no student redirect — 273-ФЗ ст. 29); the backend
// scopes what each role can fetch.
export default function WorkProgramDetailPage() {
  const params = useParams<{ id: string }>()
  const id = useMemo(() => {
    const raw = params?.id
    const parsed = typeof raw === 'string' ? Number(raw) : NaN
    return Number.isInteger(parsed) && parsed > 0 ? parsed : null
  }, [params])

  const { user, isAuthenticated, isLoading: authLoading } = useAuthCheck()
  const t = useTranslations('workProgram')

  const enabled = !authLoading && isAuthenticated && id !== null
  const {
    workProgram: wp,
    isLoading: detailLoading,
    error,
    mutate,
  } = useWorkProgram(id, { enabled })

  if (authLoading || !isAuthenticated) {
    return (
      <AppLayout>
        <div className="flex items-center justify-center py-16">
          <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
        </div>
      </AppLayout>
    )
  }

  return (
    <AppLayout>
      <div className="max-w-4xl mx-auto space-y-6">
        <Link
          href="/work-programs"
          className="inline-flex items-center gap-1.5 text-sm text-muted-foreground hover:text-foreground"
        >
          <ArrowLeft className="h-4 w-4" />
          {t('detail.backToList')}
        </Link>

        {id == null ? (
          <div className="rounded-xl border border-border bg-card p-6 text-center">
            <p className="font-medium">{t('detail.notFound')}</p>
          </div>
        ) : detailLoading ? (
          <div className="flex items-center justify-center py-16">
            <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
          </div>
        ) : error || !wp ? (
          <div className="rounded-xl border border-destructive/30 bg-destructive/5 p-6 text-center">
            <p className="font-medium text-destructive">{t('detail.loadFailed')}</p>
          </div>
        ) : (
          <WorkProgramDetail wp={wp} t={t} role={user?.role} userId={user?.id} onMutate={mutate} />
        )}
      </div>
    </AppLayout>
  )
}

type T = ReturnType<typeof useTranslations>

function WorkProgramDetail({
  wp,
  t,
  role,
  userId,
  onMutate,
}: {
  wp: WorkProgram
  t: T
  role?: string
  userId?: number
  onMutate: () => void
}) {
  const [generateOpen, setGenerateOpen] = useState(false)
  const [submitOpen, setSubmitOpen] = useState(false)
  const [discardOpen, setDiscardOpen] = useState(false)
  const [approveOpen, setApproveOpen] = useState(false)
  const [rejectOpen, setRejectOpen] = useState(false)
  const [createRevisionOpen, setCreateRevisionOpen] = useState(false)
  // Which revision row (if any) has its submit / approve / reject dialog
  // open. Null = closed. Tracking the id (not a boolean) lets one dialog
  // instance serve every row in the листы-актуализации list.
  const [submitRevisionId, setSubmitRevisionId] = useState<number | null>(null)
  const [approveRevisionId, setApproveRevisionId] = useState<number | null>(null)
  const [rejectRevisionId, setRejectRevisionId] = useState<number | null>(null)

  // Draft author actions (submit / discard) are gated by role + status:
  // the create-capable roles (teacher / methodist / admin per ADR-5) on a
  // draft. The backend is the real gate (it scopes fetches + collapses
  // unauthorized rows to 404), so this only decides button visibility.
  const canDraftActions = wp.status === 'draft' && canCreateWorkProgram(role)

  // Approver actions (approve / reject) are gated to the approver roles
  // (methodist / admin per ADR-5) on a pending_approval programme. Draft
  // and pending are disjoint statuses, so at most one action set shows.
  const canApproveActions = wp.status === 'pending_approval' && canApproveWorkProgram(role)

  // Revision (лист актуализации) gates mirror the backend's author-scoping
  // (create/submit are isAuthorOrSystemAdmin, NOT role-only): creating a
  // revision on an approved / needs_revision programme belongs to its author
  // (or admin override). Backend stays the real gate; this only hides a
  // button a methodist (approver, not author) would otherwise 404 on.
  const isAdmin = role === 'system_admin'
  const isAuthor = userId != null && wp.author_id === userId
  const canCreateRevision =
    (wp.status === 'approved' || wp.status === 'needs_revision') && (isAuthor || isAdmin)

  // Per-row revision approve/reject belongs to the approver roles
  // (methodist / admin) — mirrors the backend approve/reject_revision
  // isApprover scoping. Status gating (pending_approval only) is decided
  // per row below.
  const canApproveRevisions = canApproveWorkProgram(role)

  // Manual collection editing (slice 12c) mirrors the backend author-scoping
  // (isAuthorOrSystemAdmin) + status gate (draft / needs_revision): the РПД
  // author (or admin) hand-edits goals / competences / … while the program
  // is still editable. Backend stays the real gate; this only decides
  // whether the add / edit / delete affordances render.
  const canEditCollections =
    (wp.status === 'draft' || wp.status === 'needs_revision') && (isAuthor || isAdmin)

  // Active collection dialog state. editState=null → closed; item=null → add
  // mode; item set → edit mode. deleteState drives the confirm modal.
  const [editState, setEditState] = useState<{
    kind: CollectionKind
    item: CollectionItem | null
  } | null>(null)
  const [deleteState, setDeleteState] = useState<{
    kind: CollectionKind
    item: CollectionItem
  } | null>(null)

  // Section-header "add" affordance — rendered only when editing is allowed.
  const addButton = (kind: CollectionKind) =>
    canEditCollections ? (
      <Button size="sm" variant="outline" onClick={() => setEditState({ kind, item: null })}>
        <Plus className="h-3.5 w-3.5 mr-1.5" />
        {t('collectionDialog.add')}
      </Button>
    ) : null

  // Per-row edit / delete affordances — mirror the same gate.
  const itemActions = (kind: CollectionKind, item: CollectionItem) =>
    canEditCollections ? (
      <span className="ml-auto flex shrink-0 gap-0.5">
        <Button
          size="sm"
          variant="ghost"
          className="h-7 w-7 p-0"
          aria-label={t('collectionDialog.editAria')}
          onClick={() => setEditState({ kind, item })}
        >
          <Pencil className="h-3.5 w-3.5" />
        </Button>
        <Button
          size="sm"
          variant="ghost"
          className="h-7 w-7 p-0"
          aria-label={t('collectionDialog.deleteAria')}
          onClick={() => setDeleteState({ kind, item })}
        >
          <Trash2 className="h-3.5 w-3.5 text-destructive" />
        </Button>
      </span>
    ) : null

  return (
    <>
      <header className="space-y-3">
        <div className="flex flex-wrap items-start justify-between gap-3">
          <h1 className="text-2xl font-bold">{wp.title}</h1>
          <StatusPill status={wp.status} t={t} />
        </div>
        <dl className="flex flex-wrap gap-x-6 gap-y-1.5 text-sm text-muted-foreground">
          <div className="inline-flex items-center gap-1.5">
            <BookMarked className="h-3.5 w-3.5" />
            <span>{t('card.discipline', { id: wp.discipline_id })}</span>
          </div>
          <div className="inline-flex items-center gap-1.5">
            <GraduationCap className="h-3.5 w-3.5" />
            <span>{wp.specialty_code}</span>
          </div>
          <div className="inline-flex items-center gap-1.5">
            <Calendar className="h-3.5 w-3.5" />
            <span>{wp.applicable_from_year}</span>
          </div>
        </dl>
      </header>

      {canDraftActions ? (
        <section className="flex flex-wrap gap-2">
          <Button onClick={() => setGenerateOpen(true)}>
            <Sparkles className="h-4 w-4 mr-2" />
            {t('detail.actions.generate')}
          </Button>
          <Button onClick={() => setSubmitOpen(true)}>
            <Send className="h-4 w-4 mr-2" />
            {t('detail.actions.submit')}
          </Button>
          <Button onClick={() => setDiscardOpen(true)} variant="outline">
            <Archive className="h-4 w-4 mr-2" />
            {t('detail.actions.discard')}
          </Button>
        </section>
      ) : null}

      {canApproveActions ? (
        <section className="flex flex-wrap gap-2">
          <Button onClick={() => setApproveOpen(true)}>
            <CheckCircle2 className="h-4 w-4 mr-2" />
            {t('detail.actions.approve')}
          </Button>
          <Button onClick={() => setRejectOpen(true)} variant="destructive">
            <XCircle className="h-4 w-4 mr-2" />
            {t('detail.actions.reject')}
          </Button>
        </section>
      ) : null}

      {canCreateRevision ? (
        <section className="flex flex-wrap gap-2">
          <Button onClick={() => setCreateRevisionOpen(true)} variant="outline">
            <FilePlus2 className="h-4 w-4 mr-2" />
            {t('detail.actions.createRevision')}
          </Button>
        </section>
      ) : null}

      <section
        className={cn(
          'rounded-xl border p-4 text-sm',
          STATUS_STYLES[wp.status].bg,
          STATUS_STYLES[wp.status].text
        )}
      >
        {t(`detail.statusHint.${statusKey(wp.status)}`)}
      </section>

      {wp.reject_reason ? (
        <section className="rounded-xl border border-orange-300/40 bg-orange-50 dark:bg-orange-950/20 p-4 text-sm">
          <p className="font-medium">{t('detail.fields.rejectReason')}</p>
          <p className="mt-1 whitespace-pre-wrap">{wp.reject_reason}</p>
        </section>
      ) : null}

      {wp.annotation ? (
        <Section title={t('detail.sections.annotation')} count={1} t={t}>
          <p className="text-sm whitespace-pre-wrap">{wp.annotation}</p>
        </Section>
      ) : null}

      <Section
        title={t('detail.sections.goals')}
        count={wp.goals.length}
        t={t}
        action={addButton('goals')}
      >
        <ol className="list-decimal space-y-1.5 pl-5 text-sm">
          {wp.goals.map((g) => (
            <li key={g.id}>
              <div className="flex items-start justify-between gap-2">
                <span>{g.text}</span>
                {itemActions('goals', g)}
              </div>
            </li>
          ))}
        </ol>
      </Section>

      <Section
        title={t('detail.sections.competences')}
        count={wp.competences.length}
        t={t}
        action={addButton('competences')}
      >
        <ul className="space-y-2 text-sm">
          {wp.competences.map((c) => (
            <li key={c.id} className="flex flex-wrap items-baseline gap-x-2">
              <span className="font-mono font-medium">{c.code}</span>
              <span className="rounded bg-muted px-1.5 py-0.5 text-xs text-muted-foreground">
                {t(`detail.competenceType.${c.type}`)}
              </span>
              <span>{c.description}</span>
              {itemActions('competences', c)}
            </li>
          ))}
        </ul>
      </Section>

      <Section
        title={t('detail.sections.topics')}
        count={wp.topics.length}
        t={t}
        action={addButton('topics')}
      >
        <ul className="space-y-2 text-sm">
          {wp.topics.map((tp) => (
            <li key={tp.id} className="flex flex-wrap items-baseline gap-x-2">
              <span className="rounded bg-muted px-1.5 py-0.5 text-xs text-muted-foreground">
                {t(`detail.topicKind.${tp.kind}`)}
              </span>
              <span className="font-medium">{tp.title}</span>
              <span className="text-muted-foreground">
                {t('detail.topicHours', { hours: tp.hours })}
              </span>
              {typeof tp.week_number === 'number' ? (
                <span className="text-muted-foreground">
                  {t('detail.topicWeek', { week: tp.week_number })}
                </span>
              ) : null}
              {itemActions('topics', tp)}
            </li>
          ))}
        </ul>
      </Section>

      <Section
        title={t('detail.sections.assessments')}
        count={wp.assessments.length}
        t={t}
        action={addButton('assessments')}
      >
        <ul className="space-y-2 text-sm">
          {wp.assessments.map((a) => (
            <li key={a.id} className="flex flex-wrap items-baseline gap-x-2">
              <span className="rounded bg-muted px-1.5 py-0.5 text-xs text-muted-foreground">
                {t(`detail.assessmentType.${a.type}`)}
              </span>
              <span>{a.description}</span>
              <span className="text-muted-foreground">
                {t('detail.maxScore', { score: a.max_score })}
              </span>
              {itemActions('assessments', a)}
            </li>
          ))}
        </ul>
      </Section>

      <Section
        title={t('detail.sections.references')}
        count={wp.references.length}
        t={t}
        action={addButton('references')}
      >
        <ul className="space-y-2 text-sm">
          {wp.references.map((r) => (
            <li key={r.id} className="flex flex-wrap items-baseline gap-x-2">
              <span className="rounded bg-muted px-1.5 py-0.5 text-xs text-muted-foreground">
                {t(`detail.referenceType.${r.kind}`)}
              </span>
              <span>{r.citation}</span>
              {typeof r.year === 'number' ? (
                <span className="text-muted-foreground">{r.year}</span>
              ) : null}
              {itemActions('references', r)}
            </li>
          ))}
        </ul>
      </Section>

      <Section title={t('detail.sections.revisions')} count={wp.revisions.length} t={t}>
        <ul className="space-y-2 text-sm">
          {wp.revisions.map((rev) => {
            // A draft revision can be submitted for approval by its own
            // author (or admin override) — mirrors the backend's
            // submit_revision author-scoping. Other statuses / non-authors
            // see no action.
            const canSubmitRevision =
              rev.status === 'draft' && ((userId != null && rev.author_id === userId) || isAdmin)
            // A pending revision is approved/rejected by an approver
            // (methodist / admin) — mirrors approve/reject_revision scoping.
            const canDecideRevision = rev.status === 'pending_approval' && canApproveRevisions
            return (
              <li key={rev.id} className="flex flex-wrap items-baseline gap-x-2">
                <span className="font-medium">#{rev.revision_number}</span>
                <span className="rounded bg-muted px-1.5 py-0.5 text-xs text-muted-foreground">
                  {t(`detail.revisionChangeType.${rev.change_type}`)}
                </span>
                <span>{rev.change_summary}</span>
                <span className="rounded bg-muted px-1.5 py-0.5 text-xs text-muted-foreground">
                  {t(`detail.revisionStatus.${revisionStatusKey(rev.status)}`)}
                </span>
                {canSubmitRevision ? (
                  <Button
                    size="sm"
                    variant="outline"
                    className="ml-auto"
                    onClick={() => setSubmitRevisionId(rev.id)}
                  >
                    <Send className="h-3.5 w-3.5 mr-1.5" />
                    {t('detail.revisionActions.submit')}
                  </Button>
                ) : null}
                {canDecideRevision ? (
                  <span className="ml-auto flex gap-2">
                    <Button size="sm" onClick={() => setApproveRevisionId(rev.id)}>
                      <CheckCircle2 className="h-3.5 w-3.5 mr-1.5" />
                      {t('detail.revisionActions.approve')}
                    </Button>
                    <Button
                      size="sm"
                      variant="destructive"
                      onClick={() => setRejectRevisionId(rev.id)}
                    >
                      <XCircle className="h-3.5 w-3.5 mr-1.5" />
                      {t('detail.revisionActions.reject')}
                    </Button>
                  </span>
                ) : null}
              </li>
            )
          })}
        </ul>
      </Section>

      <GenerateWorkProgramDialog
        workProgramId={wp.id}
        open={generateOpen}
        onClose={() => setGenerateOpen(false)}
        onGenerated={onMutate}
      />
      <SubmitWorkProgramDialog
        workProgramId={wp.id}
        open={submitOpen}
        onClose={() => setSubmitOpen(false)}
        onSubmitted={onMutate}
      />
      <DiscardWorkProgramDialog
        workProgramId={wp.id}
        open={discardOpen}
        onClose={() => setDiscardOpen(false)}
        onDiscarded={onMutate}
      />
      <ApproveWorkProgramDialog
        workProgramId={wp.id}
        open={approveOpen}
        onClose={() => setApproveOpen(false)}
        onApproved={onMutate}
      />
      <RejectWorkProgramDialog
        workProgramId={wp.id}
        open={rejectOpen}
        onClose={() => setRejectOpen(false)}
        onRejected={onMutate}
      />
      <CreateRevisionDialog
        workProgramId={wp.id}
        open={createRevisionOpen}
        onClose={() => setCreateRevisionOpen(false)}
        onCreated={onMutate}
      />
      {submitRevisionId != null ? (
        <SubmitRevisionDialog
          workProgramId={wp.id}
          revisionId={submitRevisionId}
          open={true}
          onClose={() => setSubmitRevisionId(null)}
          onSubmitted={onMutate}
        />
      ) : null}
      {approveRevisionId != null ? (
        <ApproveRevisionDialog
          workProgramId={wp.id}
          revisionId={approveRevisionId}
          open={true}
          onClose={() => setApproveRevisionId(null)}
          onApproved={onMutate}
        />
      ) : null}
      {rejectRevisionId != null ? (
        <RejectRevisionDialog
          workProgramId={wp.id}
          revisionId={rejectRevisionId}
          open={true}
          onClose={() => setRejectRevisionId(null)}
          onRejected={onMutate}
        />
      ) : null}

      {/* Collection add/edit — one schema-driven dialog for every section. */}
      {editState != null
        ? (() => {
            const cfg = COLLECTION_CONFIG[editState.kind]
            const item = editState.item
            return (
              <CollectionItemDialog
                // Keyed by kind+row so switching the edited item forces a
                // remount → fresh form state seeded from the new
                // initialValues (no stale values leaking between rows).
                key={`${editState.kind}-${item ? item.id : 'add'}`}
                open={true}
                mode={item ? 'edit' : 'add'}
                titleKey={item ? cfg.editTitleKey : cfg.addTitleKey}
                fields={cfg.fields}
                initialValues={item ? cfg.initialValues(item) : {}}
                onSubmit={cfg.submit(wp.id, item ? item.id : null)}
                onDone={onMutate}
                onClose={() => setEditState(null)}
              />
            )
          })()
        : null}

      {/* Collection delete — one generic confirm for every section. */}
      {deleteState != null
        ? (() => {
            const cfg = COLLECTION_CONFIG[deleteState.kind]
            const item = deleteState.item
            return (
              <DeleteCollectionItemDialog
                open={true}
                itemLabel={cfg.itemLabel(item)}
                onConfirm={() => cfg.remove(wp.id, item.id)}
                onDone={onMutate}
                onClose={() => setDeleteState(null)}
              />
            )
          })()
        : null}
    </>
  )
}

// Section renders a titled block, falling back to the shared empty
// label when the collection is empty so the document structure stays
// visible even for a sparsely filled draft.
function Section({
  title,
  count,
  t,
  action,
  children,
}: {
  title: string
  count: number
  t: T
  action?: ReactNode
  children: ReactNode
}) {
  return (
    <section className="space-y-2 rounded-xl border border-border bg-card p-4">
      <div className="flex items-center justify-between gap-2">
        <h2 className="text-base font-semibold">{title}</h2>
        {action}
      </div>
      {count === 0 ? (
        <p className="text-sm text-muted-foreground">{t('detail.sections.empty')}</p>
      ) : (
        children
      )}
    </section>
  )
}

function StatusPill({ status, t }: { status: WorkProgramStatus; t: T }) {
  const styles = STATUS_STYLES[status]
  const Icon = styles.Icon
  return (
    <div
      className={cn(
        'inline-flex items-center gap-1.5 rounded-full px-3 py-1 text-xs font-medium',
        styles.bg,
        styles.text
      )}
    >
      <Icon className="h-3.5 w-3.5" />
      {t(`card.status.${statusKey(status)}`)}
    </div>
  )
}
