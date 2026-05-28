// Contract pins for the workProgram type layer. The runtime-observable
// parts of the module are the enum const arrays + the error-code list;
// these mirror the backend DTO (work_program_handler.go) + the domain
// enums, so a drift here is a wire-contract drift. The TS interfaces
// themselves are compile-time checked by tsc, not asserted here.

import {
  WORK_PROGRAM_STATUSES,
  COMPETENCE_TYPES,
  TOPIC_KINDS,
  ASSESSMENT_TYPES,
  REFERENCE_KINDS,
  WORK_PROGRAM_ERROR_CODES,
} from '../workProgram'

describe('workProgram type contracts', () => {
  it('lists the 5 lifecycle statuses in FSM order', () => {
    expect(WORK_PROGRAM_STATUSES).toEqual([
      'draft',
      'pending_approval',
      'approved',
      'needs_revision',
      'archived',
    ])
  })

  it('lists the 3 competence types (ПК / ОК / УК)', () => {
    expect(COMPETENCE_TYPES).toEqual(['pk', 'ok', 'uk'])
  })

  it('lists the 4 topic kinds', () => {
    expect(TOPIC_KINDS).toEqual(['lecture', 'practice', 'lab', 'self_study'])
  })

  it('lists the 3 assessment types (ФОС)', () => {
    expect(ASSESSMENT_TYPES).toEqual(['current', 'intermediate', 'final'])
  })

  it('lists the 3 reference kinds', () => {
    expect(REFERENCE_KINDS).toEqual(['main', 'additional', 'electronic'])
  })

  it('lists the 8 error codes mirrored from mapWorkProgramError', () => {
    expect(WORK_PROGRAM_ERROR_CODES).toEqual([
      'IDENTITY_EXISTS',
      'VERSION_CONFLICT',
      'INVALID_TRANSITION',
      'REJECT_REASON_REQUIRED',
      'INVALID_WORK_PROGRAM',
      'FORBIDDEN',
      'NOT_FOUND',
      'GENERIC',
    ])
  })
})
