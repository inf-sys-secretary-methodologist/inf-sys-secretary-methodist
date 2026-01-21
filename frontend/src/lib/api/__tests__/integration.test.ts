import { integrationApi, syncApi, employeesApi, studentsApi, conflictsApi } from '../integration'
import { apiClient } from '../../api'

// Mock the API client
jest.mock('../../api', () => ({
  apiClient: {
    get: jest.fn(),
    post: jest.fn(),
    delete: jest.fn(),
  },
}))

const mockedApiClient = jest.mocked(apiClient)

describe('syncApi', () => {
  beforeEach(() => {
    jest.clearAllMocks()
  })

  describe('start', () => {
    it('starts sync for employees', async () => {
      mockedApiClient.post.mockResolvedValue({
        data: { sync_log_id: 1, status: 'in_progress', message: 'Started' },
      })

      await syncApi.start({ entity_type: 'employee', direction: 'import' })

      expect(mockedApiClient.post).toHaveBeenCalledWith('/api/integration/sync/start', {
        entity_type: 'employee',
        direction: 'import',
      })
    })

    it('starts sync for students with bidirectional', async () => {
      mockedApiClient.post.mockResolvedValue({
        data: { sync_log_id: 2, status: 'in_progress' },
      })

      await syncApi.start({ entity_type: 'student', direction: 'bidirectional' })

      expect(mockedApiClient.post).toHaveBeenCalledWith('/api/integration/sync/start', {
        entity_type: 'student',
        direction: 'bidirectional',
      })
    })
  })

  describe('getStatus', () => {
    it('fetches sync status', async () => {
      const mockStatus = { id: 1, status: 'completed' }
      mockedApiClient.get.mockResolvedValue({ data: mockStatus })

      const result = await syncApi.getStatus(1)

      expect(mockedApiClient.get).toHaveBeenCalledWith('/api/integration/sync/status/1')
      expect(result.data).toEqual(mockStatus)
    })
  })

  describe('cancel', () => {
    it('cancels sync operation', async () => {
      mockedApiClient.post.mockResolvedValue({ data: { message: 'Cancelled' } })

      await syncApi.cancel(1)

      expect(mockedApiClient.post).toHaveBeenCalledWith('/api/integration/sync/cancel/1')
    })
  })

  describe('getLogs', () => {
    it('fetches sync logs without filters', async () => {
      mockedApiClient.get.mockResolvedValue({ data: { logs: [], total: 0 } })

      await syncApi.getLogs()

      expect(mockedApiClient.get).toHaveBeenCalledWith('/api/integration/sync/logs')
    })

    it('fetches sync logs with filters', async () => {
      mockedApiClient.get.mockResolvedValue({ data: { logs: [], total: 0 } })

      await syncApi.getLogs({
        entity_type: 'employee',
        status: 'completed',
        limit: 10,
      })

      expect(mockedApiClient.get).toHaveBeenCalledWith(
        expect.stringContaining('entity_type=employee')
      )
    })
  })

  describe('getStats', () => {
    it('fetches sync statistics', async () => {
      const mockStats = { total_syncs: 100, successful_syncs: 95 }
      mockedApiClient.get.mockResolvedValue({ data: mockStats })

      const result = await syncApi.getStats()

      expect(mockedApiClient.get).toHaveBeenCalledWith('/api/integration/sync/stats')
      expect(result.data).toEqual(mockStats)
    })

    it('fetches sync statistics by entity type', async () => {
      mockedApiClient.get.mockResolvedValue({ data: { total_syncs: 50 } })

      await syncApi.getStats('employee')

      expect(mockedApiClient.get).toHaveBeenCalledWith(
        '/api/integration/sync/stats?entity_type=employee'
      )
    })
  })
})

describe('employeesApi', () => {
  beforeEach(() => {
    jest.clearAllMocks()
  })

  describe('list', () => {
    it('fetches external employees without filters', async () => {
      mockedApiClient.get.mockResolvedValue({ data: { employees: [], total: 0 } })

      await employeesApi.list()

      expect(mockedApiClient.get).toHaveBeenCalledWith('/api/integration/employees')
    })

    it('fetches external employees with filters', async () => {
      mockedApiClient.get.mockResolvedValue({ data: { employees: [], total: 0 } })

      await employeesApi.list({ is_active: true, department: 'IT' })

      expect(mockedApiClient.get).toHaveBeenCalledWith(expect.stringContaining('is_active=true'))
    })
  })

  describe('getById', () => {
    it('fetches single employee', async () => {
      const mockEmployee = { id: 1, first_name: 'John' }
      mockedApiClient.get.mockResolvedValue({ data: mockEmployee })

      const result = await employeesApi.getById(1)

      expect(mockedApiClient.get).toHaveBeenCalledWith('/api/integration/employees/1')
      expect(result.data).toEqual(mockEmployee)
    })
  })

  describe('getByExternalId', () => {
    it('fetches employee by external ID', async () => {
      const mockEmployee = { id: 1, external_id: 'EXT001' }
      mockedApiClient.get.mockResolvedValue({ data: mockEmployee })

      const result = await employeesApi.getByExternalId('EXT001')

      expect(mockedApiClient.get).toHaveBeenCalledWith('/api/integration/employees/external/EXT001')
      expect(result.data).toEqual(mockEmployee)
    })
  })

  describe('getStats', () => {
    it('fetches employee statistics', async () => {
      const mockStats = { total: 100, active: 90, inactive: 10 }
      mockedApiClient.get.mockResolvedValue({ data: mockStats })

      const result = await employeesApi.getStats()

      expect(mockedApiClient.get).toHaveBeenCalledWith('/api/integration/employees/stats')
      expect(result.data).toEqual(mockStats)
    })
  })
})

describe('studentsApi', () => {
  beforeEach(() => {
    jest.clearAllMocks()
  })

  describe('list', () => {
    it('fetches external students without filters', async () => {
      mockedApiClient.get.mockResolvedValue({ data: { students: [], total: 0 } })

      await studentsApi.list()

      expect(mockedApiClient.get).toHaveBeenCalledWith('/api/integration/students')
    })

    it('fetches external students with filters', async () => {
      mockedApiClient.get.mockResolvedValue({ data: { students: [], total: 0 } })

      await studentsApi.list({ faculty: 'IT', course: 3 })

      expect(mockedApiClient.get).toHaveBeenCalledWith(expect.stringContaining('faculty=IT'))
    })
  })

  describe('getById', () => {
    it('fetches single student', async () => {
      const mockStudent = { id: 1, first_name: 'Jane' }
      mockedApiClient.get.mockResolvedValue({ data: mockStudent })

      const result = await studentsApi.getById(1)

      expect(mockedApiClient.get).toHaveBeenCalledWith('/api/integration/students/1')
      expect(result.data).toEqual(mockStudent)
    })
  })

  describe('getByExternalId', () => {
    it('fetches student by external ID', async () => {
      const mockStudent = { id: 1, external_id: 'STU001' }
      mockedApiClient.get.mockResolvedValue({ data: mockStudent })

      const result = await studentsApi.getByExternalId('STU001')

      expect(mockedApiClient.get).toHaveBeenCalledWith('/api/integration/students/external/STU001')
      expect(result.data).toEqual(mockStudent)
    })
  })

  describe('getStats', () => {
    it('fetches student statistics', async () => {
      const mockStats = { total: 500, active: 450, inactive: 50 }
      mockedApiClient.get.mockResolvedValue({ data: mockStats })

      const result = await studentsApi.getStats()

      expect(mockedApiClient.get).toHaveBeenCalledWith('/api/integration/students/stats')
      expect(result.data).toEqual(mockStats)
    })
  })
})

describe('conflictsApi', () => {
  beforeEach(() => {
    jest.clearAllMocks()
  })

  describe('list', () => {
    it('fetches conflicts without filters', async () => {
      mockedApiClient.get.mockResolvedValue({ data: { conflicts: [], total: 0 } })

      await conflictsApi.list()

      expect(mockedApiClient.get).toHaveBeenCalledWith('/api/integration/conflicts')
    })

    it('fetches conflicts with filters', async () => {
      mockedApiClient.get.mockResolvedValue({ data: { conflicts: [], total: 0 } })

      await conflictsApi.list({
        resolution: 'pending',
        entity_type: 'employee',
      })

      expect(mockedApiClient.get).toHaveBeenCalledWith(
        expect.stringContaining('resolution=pending')
      )
    })
  })

  describe('getPending', () => {
    it('fetches pending conflicts', async () => {
      mockedApiClient.get.mockResolvedValue({ data: { conflicts: [], total: 0 } })

      await conflictsApi.getPending(10, 0)

      expect(mockedApiClient.get).toHaveBeenCalledWith(
        '/api/integration/conflicts/pending?limit=10&offset=0'
      )
    })
  })

  describe('getById', () => {
    it('fetches single conflict', async () => {
      const mockConflict = { id: 1, entity_type: 'employee' }
      mockedApiClient.get.mockResolvedValue({ data: mockConflict })

      const result = await conflictsApi.getById(1)

      expect(mockedApiClient.get).toHaveBeenCalledWith('/api/integration/conflicts/1')
      expect(result.data).toEqual(mockConflict)
    })
  })

  describe('resolve', () => {
    it('resolves conflict with use_local', async () => {
      mockedApiClient.post.mockResolvedValue({ data: { message: 'Resolved' } })

      await conflictsApi.resolve(1, { resolution: 'use_local' })

      expect(mockedApiClient.post).toHaveBeenCalledWith('/api/integration/conflicts/1/resolve', {
        resolution: 'use_local',
      })
    })

    it('resolves conflict with merge and custom data', async () => {
      mockedApiClient.post.mockResolvedValue({ data: { message: 'Resolved' } })

      await conflictsApi.resolve(1, {
        resolution: 'merge',
        resolved_data: JSON.stringify({ name: 'Merged Name' }),
        notes: 'Manual merge',
      })

      expect(mockedApiClient.post).toHaveBeenCalledWith('/api/integration/conflicts/1/resolve', {
        resolution: 'merge',
        resolved_data: JSON.stringify({ name: 'Merged Name' }),
        notes: 'Manual merge',
      })
    })
  })

  describe('bulkResolve', () => {
    it('resolves multiple conflicts', async () => {
      mockedApiClient.post.mockResolvedValue({
        data: { message: 'Resolved 3 conflicts', count: 3 },
      })

      await conflictsApi.bulkResolve({
        ids: [1, 2, 3],
        resolution: 'use_external',
      })

      expect(mockedApiClient.post).toHaveBeenCalledWith('/api/integration/conflicts/bulk-resolve', {
        ids: [1, 2, 3],
        resolution: 'use_external',
      })
    })
  })

  describe('delete', () => {
    it('deletes conflict', async () => {
      mockedApiClient.delete.mockResolvedValue({ data: { message: 'Deleted' } })

      await conflictsApi.delete(1)

      expect(mockedApiClient.delete).toHaveBeenCalledWith('/api/integration/conflicts/1')
    })
  })

  describe('getStats', () => {
    it('fetches conflict statistics', async () => {
      const mockStats = { total_conflicts: 10, pending_conflicts: 5 }
      mockedApiClient.get.mockResolvedValue({ data: mockStats })

      const result = await conflictsApi.getStats()

      expect(mockedApiClient.get).toHaveBeenCalledWith('/api/integration/conflicts/stats')
      expect(result.data).toEqual(mockStats)
    })
  })
})

describe('integrationApi', () => {
  it('exposes sync API', () => {
    expect(integrationApi.sync).toBe(syncApi)
  })

  it('exposes employees API', () => {
    expect(integrationApi.employees).toBe(employeesApi)
  })

  it('exposes students API', () => {
    expect(integrationApi.students).toBe(studentsApi)
  })

  it('exposes conflicts API', () => {
    expect(integrationApi.conflicts).toBe(conflictsApi)
  })
})
