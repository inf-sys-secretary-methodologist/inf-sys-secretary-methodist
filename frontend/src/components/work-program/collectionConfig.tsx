// Per-collection descriptors driving the schema-driven CollectionItemDialog
// and DeleteCollectionItemDialog on the РПД detail page. One registry entry
// per inner collection centralizes: the form field schema, the edit-prefill
// mapper (item → string form values), the submit closure (form values →
// typed *Input → add/update hook), the delete closure, and a short item
// label for the delete confirmation. The detail page stays declarative —
// it flips dialog state and looks the behavior up by collection key.
//
// 12c-1 wires goals (proves the generic machinery end-to-end);
// competences / topics / assessments / references join the registry in
// 12c-2 — each is just another entry here plus its section wiring.

import { addGoal, updateGoal, deleteGoal } from '@/hooks/useWorkPrograms'
import type { WorkProgram, WorkProgramGoal } from '@/types/workProgram'
import type { CollectionField } from './CollectionItemDialog'

export type CollectionKind = 'goals'

// CollectionItem is the union of inner-collection row shapes a config
// operates on. Each entry only ever receives its own row type; the union
// keeps the registry homogeneous for the detail page's dialog state. It
// widens as collections join the registry in 12c-2.
export type CollectionItem = WorkProgramGoal

export interface CollectionConfig {
  fields: CollectionField[]
  addTitleKey: string
  editTitleKey: string
  // Prefill string-form values from an existing row (edit mode).
  initialValues: (item: CollectionItem) => Record<string, string>
  // Build the submit fn: add (itemId=null) or update (itemId set). The
  // returned fn maps raw string form values into the typed *Input and
  // calls the matching hook, returning the updated parent aggregate.
  submit: (
    wpId: number,
    itemId: number | null
  ) => (values: Record<string, string>) => Promise<WorkProgram>
  remove: (wpId: number, itemId: number) => Promise<WorkProgram>
  // Short preview rendered in the delete confirmation body.
  itemLabel: (item: CollectionItem) => string
}

export const COLLECTION_CONFIG: Record<CollectionKind, CollectionConfig> = {
  goals: {
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
  },
}
