import { COLLECTION_CONFIG } from '../collectionConfig'
import type { WorkProgramCompetence, WorkProgramTopic } from '@/types/workProgram'

// The registry's value→*Input submit closures and edit-prefill mappers are
// the only logic in collectionConfig worth testing: the dialog (rendering,
// validation, toast) is covered by CollectionItemDialog.test.tsx. We mock
// the mutation hooks and assert each closure maps the raw string form values
// into the correctly typed payload and dispatches to the matching hook.

const addGoal = jest.fn()
const updateGoal = jest.fn()
const deleteGoal = jest.fn()
const addCompetence = jest.fn()
const updateCompetence = jest.fn()
const deleteCompetence = jest.fn()
const addTopic = jest.fn()
const updateTopic = jest.fn()
const deleteTopic = jest.fn()

jest.mock('@/hooks/useWorkPrograms', () => ({
  addGoal: (...a: unknown[]) => addGoal(...a),
  updateGoal: (...a: unknown[]) => updateGoal(...a),
  deleteGoal: (...a: unknown[]) => deleteGoal(...a),
  addCompetence: (...a: unknown[]) => addCompetence(...a),
  updateCompetence: (...a: unknown[]) => updateCompetence(...a),
  deleteCompetence: (...a: unknown[]) => deleteCompetence(...a),
  addTopic: (...a: unknown[]) => addTopic(...a),
  updateTopic: (...a: unknown[]) => updateTopic(...a),
  deleteTopic: (...a: unknown[]) => deleteTopic(...a),
}))

beforeEach(() => jest.clearAllMocks())

describe('COLLECTION_CONFIG.competences', () => {
  const cfg = COLLECTION_CONFIG.competences
  const competence: WorkProgramCompetence = {
    id: 8,
    code: 'ПК-1',
    type: 'pk',
    description: 'Способен проектировать БД',
  }

  it('maps trimmed string values to CompetenceInput on add (itemId=null)', async () => {
    addCompetence.mockResolvedValue({ id: 99 })
    const result = await cfg.submit(
      99,
      null
    )({
      code: '  ПК-1  ',
      type: 'pk',
      description: '  Способен проектировать БД  ',
    })
    expect(addCompetence).toHaveBeenCalledWith(99, {
      code: 'ПК-1',
      type: 'pk',
      description: 'Способен проектировать БД',
    })
    expect(updateCompetence).not.toHaveBeenCalled()
    expect(result).toEqual({ id: 99 })
  })

  it('dispatches to updateCompetence on edit (itemId set)', async () => {
    updateCompetence.mockResolvedValue({ id: 99 })
    await cfg.submit(99, 8)({ code: 'ОК-2', type: 'ok', description: 'desc' })
    expect(updateCompetence).toHaveBeenCalledWith(99, 8, {
      code: 'ОК-2',
      type: 'ok',
      description: 'desc',
    })
    expect(addCompetence).not.toHaveBeenCalled()
  })

  it('prefills initialValues from an existing row', () => {
    expect(cfg.initialValues(competence)).toEqual({
      code: 'ПК-1',
      type: 'pk',
      description: 'Способен проектировать БД',
    })
  })

  it('removes via deleteCompetence and labels by code', () => {
    cfg.remove(99, 8)
    expect(deleteCompetence).toHaveBeenCalledWith(99, 8)
    expect(cfg.itemLabel(competence)).toBe('ПК-1')
  })
})

describe('COLLECTION_CONFIG.topics', () => {
  const cfg = COLLECTION_CONFIG.topics
  const topic: WorkProgramTopic = {
    id: 12,
    kind: 'lecture',
    title: 'Реляционная модель',
    hours: 4,
    week_number: 2,
    learning_outcomes: 'Знает нормальные формы',
    order_index: 1,
  }

  it('maps values to TopicInput, parsing numbers and preserving order_index', async () => {
    addTopic.mockResolvedValue({ id: 99 })
    await cfg.submit(
      99,
      null
    )({
      kind: 'practice',
      title: '  SQL  ',
      hours: '6',
      week_number: '3',
      learning_outcomes: '  Пишет запросы  ',
      order_index: '5',
    })
    expect(addTopic).toHaveBeenCalledWith(99, {
      kind: 'practice',
      title: 'SQL',
      hours: 6,
      week_number: 3,
      learning_outcomes: 'Пишет запросы',
      order_index: 5,
    })
  })

  it('maps an empty week_number to null (optional field)', async () => {
    addTopic.mockResolvedValue({ id: 99 })
    await cfg.submit(
      99,
      null
    )({
      kind: 'self_study',
      title: 'Эссе',
      hours: '0',
      week_number: '',
      learning_outcomes: '',
      order_index: '',
    })
    expect(addTopic).toHaveBeenCalledWith(99, {
      kind: 'self_study',
      title: 'Эссе',
      hours: 0,
      week_number: null,
      learning_outcomes: '',
      order_index: 0,
    })
  })

  it('dispatches to updateTopic on edit', async () => {
    updateTopic.mockResolvedValue({ id: 99 })
    await cfg.submit(
      99,
      12
    )({
      kind: 'lab',
      title: 'ER',
      hours: '2',
      week_number: '1',
      learning_outcomes: 'x',
      order_index: '0',
    })
    expect(updateTopic).toHaveBeenCalledWith(99, 12, {
      kind: 'lab',
      title: 'ER',
      hours: 2,
      week_number: 1,
      learning_outcomes: 'x',
      order_index: 0,
    })
  })

  it('prefills initialValues including blank week_number when absent', () => {
    expect(cfg.initialValues(topic)).toEqual({
      kind: 'lecture',
      title: 'Реляционная модель',
      hours: '4',
      week_number: '2',
      learning_outcomes: 'Знает нормальные формы',
      order_index: '1',
    })
    expect(cfg.initialValues({ ...topic, week_number: null })).toMatchObject({
      week_number: '',
    })
  })

  it('removes via deleteTopic and labels by title', () => {
    cfg.remove(99, 12)
    expect(deleteTopic).toHaveBeenCalledWith(99, 12)
    expect(cfg.itemLabel(topic)).toBe('Реляционная модель')
  })
})
