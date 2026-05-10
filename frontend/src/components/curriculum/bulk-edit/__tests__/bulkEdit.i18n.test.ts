/**
 * Guards the i18n contract BulkEditTable + BulkEditPanel + curriculum
 * detail page consume. The jest.setup.ts mocks `useTranslations` to
 * return the key verbatim, so component-level tests pass even когда
 * the JSON file lacks the path. This test loads actual ru/en/fr/ar
 * locale files and asserts every key v0.128.4 references resolves к
 * a non-empty string + all 4 locales expose the same set.
 *
 * Mirror к MFAVerifyLoginStep.i18n.test.ts pattern.
 */
import fs from 'fs'
import path from 'path'

const LOCALES = ['ru', 'en', 'fr', 'ar'] as const

// Keys consumed by BulkEditTable + BulkEditPanel + curriculum/[id]/page.
// Namespace = 'curriculum'. Keep in sync с component code.
const REQUIRED_KEYS = [
  'detail.sections.heading',
  'detail.sections.empty',
  'disciplineItems.controlForm.zachet',
  'disciplineItems.controlForm.exam',
  'disciplineItems.controlForm.course_project',
  'disciplineItems.controlForm.differential_zachet',
  'disciplineItems.bulkEdit.loading',
  'disciplineItems.bulkEdit.empty',
  'disciplineItems.bulkEdit.addRow',
  'disciplineItems.bulkEdit.removeRow',
  'disciplineItems.bulkEdit.submit',
  'disciplineItems.bulkEdit.cancel',
  'disciplineItems.bulkEdit.successToast',
  'disciplineItems.bulkEdit.errorEmptyBulk',
  'disciplineItems.bulkEdit.errorCrossSection',
  'disciplineItems.bulkEdit.errorNotEditable',
  'disciplineItems.bulkEdit.errorInvalidInput',
  'disciplineItems.bulkEdit.errorNotFound',
  'disciplineItems.bulkEdit.errorForbidden',
  'disciplineItems.bulkEdit.errorGeneric',
  'disciplineItems.bulkEdit.columns.title',
  'disciplineItems.bulkEdit.columns.hoursLectures',
  'disciplineItems.bulkEdit.columns.hoursPractice',
  'disciplineItems.bulkEdit.columns.hoursLab',
  'disciplineItems.bulkEdit.columns.hoursSelf',
  'disciplineItems.bulkEdit.columns.controlForm',
  'disciplineItems.bulkEdit.columns.credits',
  'disciplineItems.bulkEdit.columns.semester',
  'disciplineItems.bulkEdit.columns.order',
  'disciplineItems.bulkEdit.cancelDialog.title',
  'disciplineItems.bulkEdit.cancelDialog.description',
  'disciplineItems.bulkEdit.cancelDialog.confirm',
  'disciplineItems.bulkEdit.cancelDialog.keepEditing',
  'disciplineItems.bulkEdit.conflictBanner.heading',
  'disciplineItems.bulkEdit.conflictBanner.message',
  'disciplineItems.bulkEdit.conflictBanner.applyServer',
  // v0.128.7 P4 — ARIA labels на cell inputs / selects / delete toggle.
  'disciplineItems.bulkEdit.aria.titleInput',
  'disciplineItems.bulkEdit.aria.hoursLecturesInput',
  'disciplineItems.bulkEdit.aria.hoursPracticeInput',
  'disciplineItems.bulkEdit.aria.hoursLabInput',
  'disciplineItems.bulkEdit.aria.hoursSelfInput',
  'disciplineItems.bulkEdit.aria.controlFormSelect',
  'disciplineItems.bulkEdit.aria.creditsInput',
  'disciplineItems.bulkEdit.aria.semesterInput',
  'disciplineItems.bulkEdit.aria.orderIndexInput',
  'disciplineItems.bulkEdit.aria.deleteToggle',
  // v0.128.8 T2-1 — column header (was hardcoded English "select" в v0.128.4).
  'disciplineItems.bulkEdit.aria.deleteColumnHeader',
] as const

function loadLocale(locale: string): Record<string, unknown> {
  const file = path.join(process.cwd(), 'messages', `${locale}.json`)
  return JSON.parse(fs.readFileSync(file, 'utf-8'))
}

function resolvePath(obj: Record<string, unknown>, dottedPath: string): unknown {
  return dottedPath.split('.').reduce<unknown>((acc, segment) => {
    if (acc && typeof acc === 'object' && segment in (acc as object)) {
      return (acc as Record<string, unknown>)[segment]
    }
    return undefined
  }, obj)
}

describe('v0.128.4 bulk-edit i18n contract', () => {
  it.each(LOCALES)(
    '%s.json: every curriculum.disciplineItems.bulkEdit.* + sections heading key resolves to non-empty string',
    (locale) => {
      const data = loadLocale(locale)
      for (const k of REQUIRED_KEYS) {
        const value = resolvePath(data, `curriculum.${k}`)
        expect({ key: k, locale, value }).toEqual(
          expect.objectContaining({ value: expect.any(String) })
        )
        expect(typeof value).toBe('string')
        expect((value as string).trim().length).toBeGreaterThan(0)
      }
    }
  )

  it('all 4 locales expose identical curriculum.disciplineItems.bulkEdit key sets', () => {
    function flatKeys(obj: Record<string, unknown>, prefix = ''): string[] {
      const out: string[] = []
      for (const [k, v] of Object.entries(obj)) {
        const full = prefix ? `${prefix}.${k}` : k
        if (v && typeof v === 'object' && !Array.isArray(v)) {
          out.push(...flatKeys(v as Record<string, unknown>, full))
        } else {
          out.push(full)
        }
      }
      return out.sort()
    }

    const sets = LOCALES.map((loc) => {
      const data = loadLocale(loc)
      const node = resolvePath(data, 'curriculum.disciplineItems.bulkEdit')
      expect(node).toBeDefined()
      return flatKeys(node as Record<string, unknown>)
    })

    for (let i = 1; i < sets.length; i += 1) {
      expect(sets[i]).toEqual(sets[0])
    }
  })

  it('all 4 locales expose identical curriculum.disciplineItems.controlForm key sets', () => {
    function flatKeys(obj: Record<string, unknown>): string[] {
      return Object.keys(obj).sort()
    }
    const sets = LOCALES.map((loc) => {
      const data = loadLocale(loc)
      const node = resolvePath(data, 'curriculum.disciplineItems.controlForm')
      expect(node).toBeDefined()
      return flatKeys(node as Record<string, unknown>)
    })
    for (let i = 1; i < sets.length; i += 1) {
      expect(sets[i]).toEqual(sets[0])
    }
  })
})
