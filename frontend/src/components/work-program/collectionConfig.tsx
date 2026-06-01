// Per-collection descriptors driving the schema-driven CollectionItemDialog
// and DeleteCollectionItemDialog on the РПД detail page. One registry entry
// per inner collection centralizes: the form field schema, the edit-prefill
// mapper (item → string form values), the submit closure (form values →
// typed *Input → add/update hook), the delete closure, and a short item
// label for the delete confirmation. The detail page stays declarative —
// it flips dialog state and looks the behavior up by collection key.
//
// 12c-1 wired goals (proves the generic machinery end-to-end); 12c-2a adds
// competences + topics. assessments + references join in 12c-2b — each is
// just another entry here plus its section wiring.

import {
  addGoal,
  updateGoal,
  deleteGoal,
  addCompetence,
  updateCompetence,
  deleteCompetence,
  addTopic,
  updateTopic,
  deleteTopic,
} from '@/hooks/useWorkPrograms'
import {
  COMPETENCE_TYPES,
  TOPIC_KINDS,
  type CompetenceType,
  type TopicKind,
  type WorkProgram,
  type WorkProgramGoal,
  type WorkProgramCompetence,
  type WorkProgramTopic,
} from '@/types/workProgram'
import type { CollectionField } from './CollectionItemDialog'

export type CollectionKind = 'goals' | 'competences' | 'topics'

// CollectionItem is the union of inner-collection row shapes a config
// operates on. Each entry only ever receives its own row type; the union
// keeps the registry homogeneous for the detail page's dialog state. It
// widens further as assessments/references join the registry in 12c-2b.
export type CollectionItem = WorkProgramGoal | WorkProgramCompetence | WorkProgramTopic

// Generic over the concrete row type so each registry entry stays type-safe
// against its own collection — initialValues/itemLabel receive the concrete
// row, not the union. Entries are widened to the union for homogeneous
// storage via defineCollection (the single controlled cast).
export interface CollectionConfig<T extends CollectionItem = CollectionItem> {
  fields: CollectionField[]
  addTitleKey: string
  editTitleKey: string
  // Prefill string-form values from an existing row (edit mode).
  initialValues: (item: T) => Record<string, string>
  // Build the submit fn: add (itemId=null) or update (itemId set). The
  // returned fn maps raw string form values into the typed *Input and
  // calls the matching hook, returning the updated parent aggregate.
  submit: (
    wpId: number,
    itemId: number | null
  ) => (values: Record<string, string>) => Promise<WorkProgram>
  remove: (wpId: number, itemId: number) => Promise<WorkProgram>
  // Short preview rendered in the delete confirmation body.
  itemLabel: (item: T) => string
}

// defineCollection erases the concrete row type to the union for homogeneous
// storage in COLLECTION_CONFIG. The detail page only ever passes a row that
// matches the looked-up kind, so the widening is sound; the cast lives here
// alone rather than leaking `as` into every entry's mappers.
function defineCollection<T extends CollectionItem>(cfg: CollectionConfig<T>): CollectionConfig {
  return cfg as unknown as CollectionConfig
}

export const COLLECTION_CONFIG: Record<CollectionKind, CollectionConfig> = {
  goals: defineCollection<WorkProgramGoal>({
    fields: [
      {
        name: 'text',
        labelKey: 'collectionDialog.goals.text',
        type: 'textarea',
        required: true,
        placeholderKey: 'collectionDialog.goals.textPlaceholder',
      },
    ],
    addTitleKey: 'collectionDialog.goals.addTitle',
    editTitleKey: 'collectionDialog.goals.editTitle',
    initialValues: (item) => ({
      text: item.text,
      order_index: String(item.order_index),
    }),
    submit: (wpId, itemId) => (values) => {
      const input = {
        text: (values.text ?? '').trim(),
        order_index: Number(values.order_index ?? 0) || 0,
      }
      return itemId == null ? addGoal(wpId, input) : updateGoal(wpId, itemId, input)
    },
    remove: deleteGoal,
    itemLabel: (item) => item.text,
  }),

  competences: defineCollection<WorkProgramCompetence>({
    fields: [
      {
        name: 'code',
        labelKey: 'collectionDialog.competences.code',
        type: 'text',
        required: true,
        placeholderKey: 'collectionDialog.competences.codePlaceholder',
      },
      {
        name: 'type',
        labelKey: 'collectionDialog.competences.type',
        type: 'select',
        required: true,
        // Reuse the detail-page competence-type labels — no new i18n keys.
        options: COMPETENCE_TYPES.map((v) => ({
          value: v,
          labelKey: `detail.competenceType.${v}`,
        })),
      },
      {
        name: 'description',
        labelKey: 'collectionDialog.competences.description',
        type: 'textarea',
        required: true,
        placeholderKey: 'collectionDialog.competences.descriptionPlaceholder',
      },
    ],
    addTitleKey: 'collectionDialog.competences.addTitle',
    editTitleKey: 'collectionDialog.competences.editTitle',
    initialValues: (item) => ({
      code: item.code,
      type: item.type,
      description: item.description,
    }),
    submit: (wpId, itemId) => (values) => {
      const input = {
        code: (values.code ?? '').trim(),
        type: values.type as CompetenceType,
        description: (values.description ?? '').trim(),
      }
      return itemId == null ? addCompetence(wpId, input) : updateCompetence(wpId, itemId, input)
    },
    remove: deleteCompetence,
    itemLabel: (item) => item.code,
  }),

  topics: defineCollection<WorkProgramTopic>({
    fields: [
      {
        name: 'kind',
        labelKey: 'collectionDialog.topics.kind',
        type: 'select',
        required: true,
        // Reuse the detail-page topic-kind labels — no new i18n keys.
        options: TOPIC_KINDS.map((v) => ({ value: v, labelKey: `detail.topicKind.${v}` })),
      },
      {
        name: 'title',
        labelKey: 'collectionDialog.topics.title',
        type: 'text',
        required: true,
        placeholderKey: 'collectionDialog.topics.titlePlaceholder',
      },
      { name: 'hours', labelKey: 'collectionDialog.topics.hours', type: 'number' },
      { name: 'week_number', labelKey: 'collectionDialog.topics.weekNumber', type: 'number' },
      {
        name: 'learning_outcomes',
        labelKey: 'collectionDialog.topics.learningOutcomes',
        type: 'textarea',
        placeholderKey: 'collectionDialog.topics.learningOutcomesPlaceholder',
      },
    ],
    addTitleKey: 'collectionDialog.topics.addTitle',
    editTitleKey: 'collectionDialog.topics.editTitle',
    initialValues: (item) => ({
      kind: item.kind,
      title: item.title,
      hours: String(item.hours),
      // week_number is optional in the domain — blank when the topic is not
      // pinned to a teaching week, so the form leaves the field empty.
      week_number: item.week_number != null ? String(item.week_number) : '',
      learning_outcomes: item.learning_outcomes,
      // order_index is preserved (not surfaced as a field) so editing a
      // topic does not reset its position in the syllabus.
      order_index: String(item.order_index),
    }),
    submit: (wpId, itemId) => (values) => {
      const week = (values.week_number ?? '').trim()
      const input = {
        kind: values.kind as TopicKind,
        title: (values.title ?? '').trim(),
        hours: Number(values.hours ?? 0) || 0,
        week_number: week ? Number(week) : null,
        learning_outcomes: (values.learning_outcomes ?? '').trim(),
        order_index: Number(values.order_index ?? 0) || 0,
      }
      return itemId == null ? addTopic(wpId, input) : updateTopic(wpId, itemId, input)
    },
    remove: deleteTopic,
    itemLabel: (item) => item.title,
  }),
}
