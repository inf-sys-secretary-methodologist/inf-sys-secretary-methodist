import { studentDebtsApi } from '../studentDebts'
import { apiClient } from '../../api'

jest.mock('../../api', () => ({
  apiClient: {
    get: jest.fn(),
    post: jest.fn(),
  },
}))

const mockedApiClient = jest.mocked(apiClient)

describe('studentDebtsApi.import', () => {
  beforeEach(() => jest.clearAllMocks())

  it('POSTs a multipart FormData with the file to the import endpoint', async () => {
    mockedApiClient.post.mockResolvedValue({
      data: { created: 2, updated: 1, skipped: 0, errors: [] },
    })
    const file = new File(['xlsx-bytes'], 'debts.xlsx', {
      type: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet',
    })

    const result = await studentDebtsApi.import(file)

    expect(mockedApiClient.post).toHaveBeenCalledTimes(1)
    const [url, body, config] = mockedApiClient.post.mock.calls[0] as [
      string,
      FormData,
      { headers?: Record<string, string> },
    ]
    expect(url).toBe('/api/student-debts/import')
    expect(body).toBeInstanceOf(FormData)
    expect((body as FormData).get('file')).toBe(file)
    expect(config?.headers?.['Content-Type']).toBe('multipart/form-data')
    expect(result.created).toBe(2)
    expect(result.updated).toBe(1)
  })
})

describe('studentDebtsApi.import1C', () => {
  beforeEach(() => jest.clearAllMocks())

  it('POSTs to the 1С import endpoint with no body and returns the result', async () => {
    mockedApiClient.post.mockResolvedValue({
      data: { created: 3, updated: 1, skipped: 0, errors: [] },
    })

    const result = await studentDebtsApi.import1C()

    expect(mockedApiClient.post).toHaveBeenCalledTimes(1)
    const [url, body] = mockedApiClient.post.mock.calls[0] as [string, unknown]
    expect(url).toBe('/api/student-debts/import-1c')
    expect(body).toBeUndefined()
    expect(result.created).toBe(3)
    expect(result.updated).toBe(1)
  })
})

describe('studentDebtsApi.export', () => {
  beforeEach(() => jest.clearAllMocks())

  it('GETs the export endpoint as a blob and returns it', async () => {
    const blob = new Blob(['xlsx'], {
      type: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet',
    })
    mockedApiClient.get.mockResolvedValue(blob)

    const result = await studentDebtsApi.export({ group_name: 'ИВТ-21', status: 'open' })

    expect(mockedApiClient.get).toHaveBeenCalledTimes(1)
    const [url, config] = mockedApiClient.get.mock.calls[0] as [
      string,
      { responseType?: string; params?: Record<string, unknown> },
    ]
    expect(url).toBe('/api/student-debts/export')
    expect(config?.responseType).toBe('blob')
    expect(config?.params).toMatchObject({ group_name: 'ИВТ-21', status: 'open' })
    expect(result).toBe(blob)
  })

  it('omits undefined filter params', async () => {
    mockedApiClient.get.mockResolvedValue(new Blob(['x']))
    await studentDebtsApi.export({
      status: 'commission',
      group_name: undefined,
      semester: undefined,
    })
    const [, config] = mockedApiClient.get.mock.calls[0] as [
      string,
      { params?: Record<string, unknown> },
    ]
    expect(config?.params).toEqual({ status: 'commission' })
  })
})
