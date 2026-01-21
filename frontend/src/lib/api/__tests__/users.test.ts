import { usersApi, departmentsApi, positionsApi } from '../users'
import { apiClient } from '../../api'

// Mock the API client
jest.mock('../../api', () => ({
  apiClient: {
    get: jest.fn(),
    post: jest.fn(),
    put: jest.fn(),
    delete: jest.fn(),
  },
}))

const mockedApiClient = jest.mocked(apiClient)

describe('usersApi', () => {
  beforeEach(() => {
    jest.clearAllMocks()
  })

  describe('getAll', () => {
    it('fetches all users', async () => {
      const mockUsers = [
        { id: 1, name: 'User 1', email: 'user1@test.com' },
        { id: 2, name: 'User 2', email: 'user2@test.com' },
      ]
      mockedApiClient.get.mockResolvedValue({ data: { users: mockUsers } })

      const result = await usersApi.getAll()

      expect(mockedApiClient.get).toHaveBeenCalledWith('/api/users?limit=1000')
      expect(result).toEqual(mockUsers)
    })

    it('returns empty array when no users', async () => {
      mockedApiClient.get.mockResolvedValue({ data: {} })

      const result = await usersApi.getAll()

      expect(result).toEqual([])
    })
  })

  describe('list', () => {
    it('fetches users with filters', async () => {
      const mockResponse = { data: { users: [], total: 0, page: 1, limit: 10, total_pages: 0 } }
      mockedApiClient.get.mockResolvedValue(mockResponse)

      await usersApi.list({
        department_id: 1,
        position_id: 2,
        role: 'student',
        status: 'active',
        search: 'test',
        page: 1,
        limit: 10,
      })

      expect(mockedApiClient.get).toHaveBeenCalledWith(expect.stringContaining('/api/users?'))
    })

    it('fetches users without filters', async () => {
      const mockResponse = { data: { users: [], total: 0, page: 1, limit: 10, total_pages: 0 } }
      mockedApiClient.get.mockResolvedValue(mockResponse)

      await usersApi.list()

      expect(mockedApiClient.get).toHaveBeenCalledWith('/api/users')
    })
  })

  describe('getById', () => {
    it('fetches single user', async () => {
      const mockUser = { id: 1, name: 'Test User', email: 'test@test.com' }
      mockedApiClient.get.mockResolvedValue({ data: mockUser })

      const result = await usersApi.getById(1)

      expect(mockedApiClient.get).toHaveBeenCalledWith('/api/users/1')
      expect(result.data).toEqual(mockUser)
    })
  })

  describe('updateProfile', () => {
    it('updates user profile', async () => {
      mockedApiClient.put.mockResolvedValue({ data: { message: 'Updated' } })

      const updateData = { department_id: 1, phone: '123456789' }
      await usersApi.updateProfile(1, updateData)

      expect(mockedApiClient.put).toHaveBeenCalledWith('/api/users/1/profile', updateData)
    })
  })

  describe('uploadAvatar', () => {
    it('uploads avatar file', async () => {
      mockedApiClient.post.mockResolvedValue({ data: { avatar_url: '/uploads/avatar.jpg' } })

      const file = new File([''], 'avatar.jpg', { type: 'image/jpeg' })
      await usersApi.uploadAvatar(1, file)

      expect(mockedApiClient.post).toHaveBeenCalledWith(
        '/api/users/1/avatar',
        expect.any(FormData),
        { headers: { 'Content-Type': 'multipart/form-data' } }
      )
    })
  })

  describe('deleteAvatar', () => {
    it('deletes avatar', async () => {
      mockedApiClient.delete.mockResolvedValue({ data: { message: 'Deleted' } })

      await usersApi.deleteAvatar(1)

      expect(mockedApiClient.delete).toHaveBeenCalledWith('/api/users/1/avatar')
    })
  })

  describe('updateRole', () => {
    it('updates user role', async () => {
      mockedApiClient.put.mockResolvedValue({ data: { message: 'Updated' } })

      await usersApi.updateRole(1, 'teacher')

      expect(mockedApiClient.put).toHaveBeenCalledWith('/api/users/1/role', { role: 'teacher' })
    })
  })

  describe('updateStatus', () => {
    it('updates user status', async () => {
      mockedApiClient.put.mockResolvedValue({ data: { message: 'Updated' } })

      await usersApi.updateStatus(1, 'inactive')

      expect(mockedApiClient.put).toHaveBeenCalledWith('/api/users/1/status', {
        status: 'inactive',
      })
    })
  })

  describe('delete', () => {
    it('deletes user', async () => {
      mockedApiClient.delete.mockResolvedValue({ data: { message: 'Deleted' } })

      await usersApi.delete(1)

      expect(mockedApiClient.delete).toHaveBeenCalledWith('/api/users/1')
    })
  })

  describe('bulkUpdateDepartment', () => {
    it('updates department for multiple users', async () => {
      mockedApiClient.post.mockResolvedValue({ data: { message: 'Updated' } })

      await usersApi.bulkUpdateDepartment([1, 2, 3], 5)

      expect(mockedApiClient.post).toHaveBeenCalledWith('/api/users/bulk/department', {
        user_ids: [1, 2, 3],
        department_id: 5,
      })
    })
  })

  describe('bulkUpdatePosition', () => {
    it('updates position for multiple users', async () => {
      mockedApiClient.post.mockResolvedValue({ data: { message: 'Updated' } })

      await usersApi.bulkUpdatePosition([1, 2], 3)

      expect(mockedApiClient.post).toHaveBeenCalledWith('/api/users/bulk/position', {
        user_ids: [1, 2],
        position_id: 3,
      })
    })
  })

  describe('getByDepartment', () => {
    it('fetches users by department', async () => {
      mockedApiClient.get.mockResolvedValue({ data: { users: [] } })

      await usersApi.getByDepartment(1)

      expect(mockedApiClient.get).toHaveBeenCalledWith('/api/users/by-department/1')
    })
  })

  describe('getByPosition', () => {
    it('fetches users by position', async () => {
      mockedApiClient.get.mockResolvedValue({ data: { users: [] } })

      await usersApi.getByPosition(1)

      expect(mockedApiClient.get).toHaveBeenCalledWith('/api/users/by-position/1')
    })
  })
})

describe('departmentsApi', () => {
  beforeEach(() => {
    jest.clearAllMocks()
  })

  describe('list', () => {
    it('fetches departments with pagination', async () => {
      mockedApiClient.get.mockResolvedValue({ data: { departments: [], total: 0 } })

      await departmentsApi.list(1, 10, false)

      expect(mockedApiClient.get).toHaveBeenCalledWith(
        '/api/departments?page=1&limit=10&active_only=false'
      )
    })

    it('fetches only active departments', async () => {
      mockedApiClient.get.mockResolvedValue({ data: { departments: [], total: 0 } })

      await departmentsApi.list(1, 10, true)

      expect(mockedApiClient.get).toHaveBeenCalledWith(
        '/api/departments?page=1&limit=10&active_only=true'
      )
    })
  })

  describe('getById', () => {
    it('fetches single department', async () => {
      const mockDept = { id: 1, name: 'IT', code: 'IT001' }
      mockedApiClient.get.mockResolvedValue({ data: mockDept })

      const result = await departmentsApi.getById(1)

      expect(mockedApiClient.get).toHaveBeenCalledWith('/api/departments/1')
      expect(result.data).toEqual(mockDept)
    })
  })

  describe('create', () => {
    it('creates department', async () => {
      const newDept = { name: 'HR', code: 'HR001', description: 'Human Resources' }
      mockedApiClient.post.mockResolvedValue({ data: { id: 1, ...newDept } })

      await departmentsApi.create(newDept)

      expect(mockedApiClient.post).toHaveBeenCalledWith('/api/departments', newDept)
    })
  })

  describe('update', () => {
    it('updates department', async () => {
      const updateData = { name: 'IT Updated', code: 'IT001', is_active: true }
      mockedApiClient.put.mockResolvedValue({ data: { id: 1, ...updateData } })

      await departmentsApi.update(1, updateData)

      expect(mockedApiClient.put).toHaveBeenCalledWith('/api/departments/1', updateData)
    })
  })

  describe('delete', () => {
    it('deletes department', async () => {
      mockedApiClient.delete.mockResolvedValue({ data: { message: 'Deleted' } })

      await departmentsApi.delete(1)

      expect(mockedApiClient.delete).toHaveBeenCalledWith('/api/departments/1')
    })
  })

  describe('getChildren', () => {
    it('fetches child departments', async () => {
      mockedApiClient.get.mockResolvedValue({ data: { departments: [] } })

      await departmentsApi.getChildren(1)

      expect(mockedApiClient.get).toHaveBeenCalledWith('/api/departments/1/children')
    })
  })
})

describe('positionsApi', () => {
  beforeEach(() => {
    jest.clearAllMocks()
  })

  describe('list', () => {
    it('fetches positions with pagination', async () => {
      mockedApiClient.get.mockResolvedValue({ data: { positions: [], total: 0 } })

      await positionsApi.list(1, 10, false)

      expect(mockedApiClient.get).toHaveBeenCalledWith(
        '/api/positions?page=1&limit=10&active_only=false'
      )
    })
  })

  describe('getById', () => {
    it('fetches single position', async () => {
      const mockPos = { id: 1, name: 'Developer', code: 'DEV001' }
      mockedApiClient.get.mockResolvedValue({ data: mockPos })

      const result = await positionsApi.getById(1)

      expect(mockedApiClient.get).toHaveBeenCalledWith('/api/positions/1')
      expect(result.data).toEqual(mockPos)
    })
  })

  describe('create', () => {
    it('creates position', async () => {
      const newPos = { name: 'Manager', code: 'MGR001', level: 2 }
      mockedApiClient.post.mockResolvedValue({ data: { id: 1, ...newPos } })

      await positionsApi.create(newPos)

      expect(mockedApiClient.post).toHaveBeenCalledWith('/api/positions', newPos)
    })
  })

  describe('update', () => {
    it('updates position', async () => {
      const updateData = { name: 'Senior Developer', code: 'SDEV001', level: 3, is_active: true }
      mockedApiClient.put.mockResolvedValue({ data: { id: 1, ...updateData } })

      await positionsApi.update(1, updateData)

      expect(mockedApiClient.put).toHaveBeenCalledWith('/api/positions/1', updateData)
    })
  })

  describe('delete', () => {
    it('deletes position', async () => {
      mockedApiClient.delete.mockResolvedValue({ data: { message: 'Deleted' } })

      await positionsApi.delete(1)

      expect(mockedApiClient.delete).toHaveBeenCalledWith('/api/positions/1')
    })
  })
})
