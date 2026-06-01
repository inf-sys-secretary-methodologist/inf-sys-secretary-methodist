import { COLLECTION_CONFIG } from '../collectionConfig'
import type {
  WorkProgramCompetence,
  WorkProgramTopic,
  WorkProgramAssessment,
  WorkProgramReference,
} from '@/types/workProgram'

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
const addAssessment = jest.fn()
const updateAssessment = jest.fn()
const deleteAssessment = jest.fn()
const addReference = jest.fn()
const updateReference = jest.fn()
const deleteReference = jest.fn()

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
  addAssessment: (...a: unknown[]) => addAssessment(...a),
  updateAssessment: (...a: unknown[]) => updateAssessment(...a),
  deleteAssessment: (...a: unknown[]) => deleteAssessment(...a),
  addReference: (...a: unknown[]) => addReference(...a),
  updateReference: (...a: unknown[]) => updateReference(...a),
  deleteReference: (...a: unknown[]) => deleteReference(...a),
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

describe('COLLECTION_CONFIG.assessments', () => {
  const cfg = COLLECTION_CONFIG.assessments
  const assessment: WorkProgramAssessment = {
    id: 15,
    type: 'current',
    description: 'Контрольная работа',
    max_score: 20,
    example_questions: ['Что такое нормализация?', 'Опишите 3НФ'],
  }

  it('maps values to AssessmentInput, parsing max_score and splitting ФОС by line', async () => {
    addAssessment.mockResolvedValue({ id: 99 })
    await cfg.submit(
      99,
      null
    )({
      type: 'final',
      description: '  Экзамен  ',
      max_score: '40',
      // ФОС textarea: trimmed, blank lines dropped
      example_questions: '  Вопрос 1  \n\nВопрос 2\n   ',
    })
    expect(addAssessment).toHaveBeenCalledWith(99, {
      type: 'final',
      description: 'Экзамен',
      max_score: 40,
      example_questions: ['Вопрос 1', 'Вопрос 2'],
    })
  })

  it('maps an empty ФОС textarea to an empty array', async () => {
    addAssessment.mockResolvedValue({ id: 99 })
    await cfg.submit(
      99,
      null
    )({
      type: 'intermediate',
      description: 'Зачёт',
      max_score: '',
      example_questions: '',
    })
    expect(addAssessment).toHaveBeenCalledWith(99, {
      type: 'intermediate',
      description: 'Зачёт',
      max_score: 0,
      example_questions: [],
    })
  })

  it('dispatches to updateAssessment on edit', async () => {
    updateAssessment.mockResolvedValue({ id: 99 })
    await cfg.submit(
      99,
      15
    )({
      type: 'current',
      description: 'КР',
      max_score: '10',
      example_questions: 'A',
    })
    expect(updateAssessment).toHaveBeenCalledWith(99, 15, {
      type: 'current',
      description: 'КР',
      max_score: 10,
      example_questions: ['A'],
    })
  })

  it('prefills initialValues joining ФОС by newline', () => {
    expect(cfg.initialValues(assessment)).toEqual({
      type: 'current',
      description: 'Контрольная работа',
      max_score: '20',
      example_questions: 'Что такое нормализация?\nОпишите 3НФ',
    })
  })

  it('removes via deleteAssessment and labels by description', () => {
    cfg.remove(99, 15)
    expect(deleteAssessment).toHaveBeenCalledWith(99, 15)
    expect(cfg.itemLabel(assessment)).toBe('Контрольная работа')
  })
})

describe('COLLECTION_CONFIG.references', () => {
  const cfg = COLLECTION_CONFIG.references
  const reference: WorkProgramReference = {
    id: 21,
    kind: 'main',
    citation: 'Дейт К. Дж. Введение в системы баз данных',
    year: 2005,
    isbn: '978-5-8459-0788-2',
    url: 'https://example.org/date',
    order_index: 2,
  }

  it('maps values to ReferenceInput, parsing year and preserving order_index', async () => {
    addReference.mockResolvedValue({ id: 99 })
    await cfg.submit(
      99,
      null
    )({
      kind: 'additional',
      citation: '  Гарсиа-Молина  ',
      year: '2002',
      isbn: '  111  ',
      url: '  https://x.test  ',
      order_index: '4',
    })
    expect(addReference).toHaveBeenCalledWith(99, {
      kind: 'additional',
      citation: 'Гарсиа-Молина',
      year: 2002,
      isbn: '111',
      url: 'https://x.test',
      order_index: 4,
    })
  })

  it('maps empty optional year/isbn/url to null/empty (order_index defaults 0)', async () => {
    addReference.mockResolvedValue({ id: 99 })
    await cfg.submit(
      99,
      null
    )({
      kind: 'electronic',
      citation: 'ГОСТ 7.32',
      year: '',
      isbn: '',
      url: '',
      order_index: '',
    })
    expect(addReference).toHaveBeenCalledWith(99, {
      kind: 'electronic',
      citation: 'ГОСТ 7.32',
      year: null,
      isbn: '',
      url: '',
      order_index: 0,
    })
  })

  it('dispatches to updateReference on edit', async () => {
    updateReference.mockResolvedValue({ id: 99 })
    await cfg.submit(
      99,
      21
    )({
      kind: 'main',
      citation: 'Дейт',
      year: '2005',
      isbn: 'x',
      url: '',
      order_index: '2',
    })
    expect(updateReference).toHaveBeenCalledWith(99, 21, {
      kind: 'main',
      citation: 'Дейт',
      year: 2005,
      isbn: 'x',
      url: '',
      order_index: 2,
    })
  })

  it('prefills initialValues, blanking absent year/isbn/url', () => {
    expect(cfg.initialValues(reference)).toEqual({
      kind: 'main',
      citation: 'Дейт К. Дж. Введение в системы баз данных',
      year: '2005',
      isbn: '978-5-8459-0788-2',
      url: 'https://example.org/date',
      order_index: '2',
    })
    expect(
      cfg.initialValues({ ...reference, year: null, isbn: undefined, url: undefined })
    ).toMatchObject({ year: '', isbn: '', url: '' })
  })

  it('removes via deleteReference and labels by citation', () => {
    cfg.remove(99, 21)
    expect(deleteReference).toHaveBeenCalledWith(99, 21)
    expect(cfg.itemLabel(reference)).toBe('Дейт К. Дж. Введение в системы баз данных')
  })
})
