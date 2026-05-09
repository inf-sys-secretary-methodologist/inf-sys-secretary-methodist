import { pickBulkEditErrorKey, type BulkEditErrorKey } from '../pickBulkEditErrorKey'

interface Case {
  name: string
  status: number | undefined
  errorCode: string | undefined
  expected: BulkEditErrorKey
}

// Table-driven (≥3-variant gate per CLAUDE.md). Covers:
//  - 422 sub-cases distinguished by error code (4 sentinels per handler).
//  - 404 (single key — handler emits NOT_FOUND for both section + item).
//  - 403 (FORBIDDEN — ScopeForbidden sentinel only path on bulk endpoint).
//  - 500 + unknown statuses + missing/null status → errorGeneric (default-deny).
const cases: Case[] = [
  {
    name: '422 EMPTY_BULK_INPUT → errorEmptyBulk',
    status: 422,
    errorCode: 'EMPTY_BULK_INPUT',
    expected: 'errorEmptyBulk',
  },
  {
    name: '422 CROSS_SECTION_BULK_EDIT → errorCrossSection',
    status: 422,
    errorCode: 'CROSS_SECTION_BULK_EDIT',
    expected: 'errorCrossSection',
  },
  {
    name: '422 NOT_EDITABLE → errorNotEditable',
    status: 422,
    errorCode: 'NOT_EDITABLE',
    expected: 'errorNotEditable',
  },
  {
    name: '422 INVALID_INPUT → errorInvalidInput',
    status: 422,
    errorCode: 'INVALID_INPUT',
    expected: 'errorInvalidInput',
  },
  {
    name: '404 NOT_FOUND (section gone) → errorNotFound',
    status: 404,
    errorCode: 'NOT_FOUND',
    expected: 'errorNotFound',
  },
  {
    name: '404 без code → errorNotFound (status-only mapping)',
    status: 404,
    errorCode: undefined,
    expected: 'errorNotFound',
  },
  {
    name: '403 FORBIDDEN → errorForbidden',
    status: 403,
    errorCode: 'FORBIDDEN',
    expected: 'errorForbidden',
  },
  {
    name: '422 unknown code → errorGeneric (default-deny)',
    status: 422,
    errorCode: 'WHO_KNOWS',
    expected: 'errorGeneric',
  },
  {
    name: '500 INTERNAL_ERROR → errorGeneric',
    status: 500,
    errorCode: 'INTERNAL_ERROR',
    expected: 'errorGeneric',
  },
  {
    name: '418 (unknown) → errorGeneric',
    status: 418,
    errorCode: 'UNKNOWN',
    expected: 'errorGeneric',
  },
  {
    name: 'undefined status (network error) → errorGeneric',
    status: undefined,
    errorCode: undefined,
    expected: 'errorGeneric',
  },
]

describe('pickBulkEditErrorKey', () => {
  it.each(cases)('$name', ({ status, errorCode, expected }) => {
    expect(pickBulkEditErrorKey(status, errorCode)).toBe(expected)
  })
})
