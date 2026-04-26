import { renderHook, waitFor } from '@testing-library/react'
import { SWRConfig } from 'swr'
import React from 'react'
import {
  useFiles,
  useFileVersions,
  uploadFile,
  deleteFile,
  downloadFile,
  createFileVersion,
} from '../useFiles'
import { apiClient } from '@/lib/api'

jest.mock('@/lib/api', () => ({
  apiClient: {
    get: jest.fn(),
    post: jest.fn(),
    put: jest.fn(),
    delete: jest.fn(),
  },
}))

const mockedApiClient = jest.mocked(apiClient)

const wrapper = ({ children }: { children: React.ReactNode }) =>
  React.createElement(
    SWRConfig,
    { value: { dedupingInterval: 0, provider: () => new Map() } },
    children
  )

describe('useFiles hooks', () => {
  beforeEach(() => {
    jest.clearAllMocks()
  })

  describe('useFiles', () => {
    it('returns files list from API', async () => {
      const mockResponse = {
        files: [
          {
            id: 1,
            original_name: 'report.pdf',
            size: 1024,
            mime_type: 'application/pdf',
            checksum: 'abc123',
            uploaded_by: 1,
            is_temporary: false,
            created_at: '2026-04-25T10:00:00Z',
            updated_at: '2026-04-25T10:00:00Z',
          },
        ],
        total: 1,
        page: 1,
        limit: 20,
        total_pages: 1,
      }

      mockedApiClient.get.mockResolvedValue({ data: mockResponse })

      const { result } = renderHook(() => useFiles(), { wrapper })

      await waitFor(() => {
        expect(result.current.files).toHaveLength(1)
      })
      expect(result.current.files[0].original_name).toBe('report.pdf')
      expect(result.current.total).toBe(1)
      expect(result.current.totalPages).toBe(1)
    })

    it('passes filter params as query string', async () => {
      mockedApiClient.get.mockResolvedValue({
        data: { files: [], total: 0, page: 1, limit: 20, total_pages: 0 },
      })

      renderHook(() => useFiles({ page: 2, limit: 10 }), { wrapper })

      await waitFor(() => {
        expect(mockedApiClient.get).toHaveBeenCalled()
      })

      const callUrl = mockedApiClient.get.mock.calls[0][0] as string
      expect(callUrl).toContain('page=2')
      expect(callUrl).toContain('limit=10')
    })

    it('returns empty array on error', async () => {
      mockedApiClient.get.mockRejectedValue(new Error('Network error'))

      const { result } = renderHook(() => useFiles(), { wrapper })

      await waitFor(() => {
        expect(result.current.error).toBeTruthy()
      })
      expect(result.current.files).toEqual([])
    })
  })

  describe('useFileVersions', () => {
    it('returns versions for a file', async () => {
      const mockVersions = [
        {
          id: 1,
          version_number: 1,
          size: 1024,
          checksum: 'abc',
          created_by: 1,
          created_at: '2026-04-25T10:00:00Z',
        },
        {
          id: 2,
          version_number: 2,
          size: 2048,
          checksum: 'def',
          comment: 'Updated',
          created_by: 1,
          created_at: '2026-04-25T11:00:00Z',
        },
      ]

      mockedApiClient.get.mockResolvedValue({ data: mockVersions })

      const { result } = renderHook(() => useFileVersions(1), { wrapper })

      await waitFor(() => {
        expect(result.current.versions).toHaveLength(2)
      })
      expect(result.current.versions[1].comment).toBe('Updated')
    })

    it('skips fetch when id is null', () => {
      renderHook(() => useFileVersions(null), { wrapper })

      expect(mockedApiClient.get).not.toHaveBeenCalled()
    })
  })

  describe('mutation functions', () => {
    it('uploadFile sends FormData via POST', async () => {
      const mockFile = new File(['content'], 'test.txt', { type: 'text/plain' })
      const mockResponse = {
        file_id: 1,
        original_name: 'test.txt',
        size: 7,
        mime_type: 'text/plain',
        checksum: 'abc',
      }
      mockedApiClient.post.mockResolvedValue(mockResponse)

      const result = await uploadFile(mockFile)

      expect(mockedApiClient.post).toHaveBeenCalledWith(
        '/api/files/upload',
        expect.any(FormData),
        { headers: { 'Content-Type': 'multipart/form-data' } }
      )
      expect(result).toEqual(mockResponse)
    })

    it('deleteFile sends DELETE request', async () => {
      mockedApiClient.delete.mockResolvedValue(undefined)

      await deleteFile(42)

      expect(mockedApiClient.delete).toHaveBeenCalledWith('/api/files/42')
    })

    it('downloadFile returns presigned URL', async () => {
      const mockResponse = {
        presigned_url: 'https://minio/file',
        file_name: 'report.pdf',
        mime_type: 'application/pdf',
        size: 1024,
      }
      mockedApiClient.get.mockResolvedValue(mockResponse)

      const result = await downloadFile(1)

      expect(mockedApiClient.get).toHaveBeenCalledWith('/api/files/1/download')
      expect(result.presigned_url).toBe('https://minio/file')
    })

    it('createFileVersion sends file + comment', async () => {
      const mockFile = new File(['v2'], 'test.txt', { type: 'text/plain' })
      mockedApiClient.post.mockResolvedValue({ id: 2, version_number: 2 })

      await createFileVersion(1, mockFile, 'Updated content')

      expect(mockedApiClient.post).toHaveBeenCalledWith(
        '/api/files/1/versions',
        expect.any(FormData),
        { headers: { 'Content-Type': 'multipart/form-data' } }
      )
    })
  })
})
