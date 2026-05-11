// i18n parity test — runs против actual JSON message files (не the
// useTranslations mock) so missing keys in any locale fail the build.
// Mirror к the lesson recorded в memory feedback_i18n_json_load_parity_test:
// the global mock returns keys verbatim and hides namespace bugs.
import ru from '../../../../../messages/ru.json'
import en from '../../../../../messages/en.json'
import fr from '../../../../../messages/fr.json'
import ar from '../../../../../messages/ar.json'

type MessagesShape = {
  nav?: { annualReport?: string }
  reports?: {
    annual?: {
      title?: string
      yearLabel?: string
      downloadButton?: string
      errorBadYear?: string
      errorForbidden?: string
      errorServer?: string
    }
  }
}

const locales: Array<readonly [string, MessagesShape]> = [
  ['ru', ru as MessagesShape],
  ['en', en as MessagesShape],
  ['fr', fr as MessagesShape],
  ['ar', ar as MessagesShape],
]

const annualKeys = [
  'title',
  'yearLabel',
  'downloadButton',
  'errorBadYear',
  'errorForbidden',
  'errorServer',
] as const

describe.each(locales)('messages/%s.json — annual report parity', (name, msgs) => {
  it(`has nav.annualReport`, () => {
    expect(msgs.nav?.annualReport).toBeTruthy()
    expect(typeof msgs.nav?.annualReport).toBe('string')
  })

  it.each(annualKeys)(`has reports.annual.%s`, (key) => {
    const value = msgs.reports?.annual?.[key]
    expect(value).toBeTruthy()
    expect(typeof value).toBe('string')
  })

  // Guard against the global useTranslations mock illusion — if a key is
  // missing in production, the mock returns the key verbatim (e.g.
  // "title") which still passes most component tests. This test reads
  // the raw JSON so any locale lacking the key fails here.
  it(`exposes a translated string (не the key verbatim) for reports.annual.title`, () => {
    expect(msgs.reports?.annual?.title).not.toBe('title')
    expect(msgs.reports?.annual?.title).not.toBe(`reports.annual.title`)
  })
})
