// Per-collection descriptors driving the schema-driven CollectionItemDialog
// and DeleteCollectionItemDialog on the РПД detail page. One registry entry
// per inner collection centralizes: the form field schema, the edit-prefill
// mapper (item → string form values), the submit closure (form values →
// typed *Input → add/update hook), the delete closure, and a short item
// label for the delete confirmation. The detail page stays declarative —
// it flips dialog state and looks the behavior up by collection key.
//
// 12c-1 wired goals (proves the generic machinery end-to-end); 12c-2a added
// competences + topics; 12c-2b completes the set with assessments (ФОС) and
// references — each is just another entry here plus its section wiring.

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
  addAssessment,
  updateAssessment,
  deleteAssessment,
  addReference,
  updateReference,
  deleteReference,
} from '@/hooks/useWorkPrograms'
import {
  COMPETENCE_TYPES,
  TOPIC_KINDS,
  ASSESSMENT_TYPES,
  REFERENCE_KINDS,
  type CompetenceType,
  type TopicKind,
  type AssessmentType,
  type ReferenceKind,
  type WorkProgram,
  type WorkProgramGoal,
  type WorkProgramCompetence,
  type WorkProgramTopic,
  type WorkProgramAssessment,
  type WorkProgramReference,
} from '@/types/workProgram'
import type { CollectionField } from './CollectionItemDialog'

export type CollectionKind = 'goals' | 'competences' | 'topics' | 'assessments' | 'references'

// CollectionItem is the union of inner-collection row shapes a config
// operates on. Each entry only ever receives its own row type; the union
// keeps the registry homogeneous for the detail page's dialog state.
export type CollectionItem =
  | WorkProgramGoal
  | WorkProgramCompetence
  | WorkProgramTopic
  | WorkProgramAssessment
  | WorkProgramReference

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

  assessments: defineCollection<WorkProgramAssessment>({
    fields: [
      {
        name: 'type',
        labelKey: 'collectionDialog.assessments.type',
        type: 'select',
        required: true,
        // Reuse the detail-page assessment-type labels — no new i18n keys.
        options: ASSESSMENT_TYPES.map((v) => ({
          value: v,
          labelKey: `detail.assessmentType.${v}`,
        })),
      },
      {
        name: 'description',
        labelKey: 'collectionDialog.assessments.description',
        type: 'textarea',
        required: true,
        placeholderKey: 'collectionDialog.assessments.descriptionPlaceholder',
      },
      { name: 'max_score', labelKey: 'collectionDialog.assessments.maxScore', type: 'number' },
      {
        name: 'example_questions',
        labelKey: 'collectionDialog.assessments.exampleQuestions',
        type: 'textarea',
        placeholderKey: 'collectionDialog.assessments.exampleQuestionsPlaceholder',
      },
    ],
    addTitleKey: 'collectionDialog.assessments.addTitle',
    editTitleKey: 'collectionDialog.assessments.editTitle',
    initialValues: (item) => ({
      type: item.type,
      description: item.description,
      max_score: String(item.max_score),
      // ФОС is edited as one textarea, one question per line.
      example_questions: item.example_questions.join('\n'),
    }),
    submit: (wpId, itemId) => (values) => {
      const input = {
        type: values.type as AssessmentType,
        description: (values.description ?? '').trim(),
        max_score: Number(values.max_score ?? 0) || 0,
        // Split the textarea into questions, trimming each and dropping
        // blank lines so stray newlines do not create empty ФОС items.
        example_questions: (values.example_questions ?? '')
          .split('\n')
          .map((q) => q.trim())
          .filter((q) => q.length > 0),
      }
      return itemId == null ? addAssessment(wpId, input) : updateAssessment(wpId, itemId, input)
    },
    remove: deleteAssessment,
    itemLabel: (item) => item.description,
  }),

  references: defineCollection<WorkProgramReference>({
    fields: [
      {
        name: 'kind',
        labelKey: 'collectionDialog.references.kind',
        type: 'select',
        required: true,
        // Reuse the detail-page reference-kind labels — no new i18n keys.
        options: REFERENCE_KINDS.map((v) => ({
          value: v,
          labelKey: `detail.referenceType.${v}`,
        })),
      },
      {
        name: 'citation',
        labelKey: 'collectionDialog.references.citation',
        type: 'textarea',
        required: true,
        placeholderKey: 'collectionDialog.references.citationPlaceholder',
      },
      { name: 'year', labelKey: 'collectionDialog.references.year', type: 'number' },
      { name: 'isbn', labelKey: 'collectionDialog.references.isbn', type: 'text' },
      { name: 'url', labelKey: 'collectionDialog.references.url', type: 'text' },
    ],
    addTitleKey: 'collectionDialog.references.addTitle',
    editTitleKey: 'collectionDialog.references.editTitle',
    initialValues: (item) => ({
      kind: item.kind,
      citation: item.citation,
      // year is optional — blank when the source has none.
      year: item.year != null ? String(item.year) : '',
      isbn: item.isbn ?? '',
      url: item.url ?? '',
      // order_index is preserved (not surfaced as a field) so editing a
      // reference does not reset its position in the bibliography.
      order_index: String(item.order_index),
    }),
    submit: (wpId, itemId) => (values) => {
      const year = (values.year ?? '').trim()
      const input = {
        kind: values.kind as ReferenceKind,
        citation: (values.citation ?? '').trim(),
        year: year ? Number(year) : null,
        isbn: (values.isbn ?? '').trim(),
        url: (values.url ?? '').trim(),
        order_index: Number(values.order_index ?? 0) || 0,
      }
      return itemId == null ? addReference(wpId, input) : updateReference(wpId, itemId, input)
    },
    remove: deleteReference,
    itemLabel: (item) => item.citation,
  }),
}
