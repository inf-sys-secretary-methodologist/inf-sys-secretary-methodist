import { apiClient } from '../api'

// annualReportApi wraps the GET /api/reports/annual endpoint that streams
// a DOCX file. Authentication / role gating happen server-side; the
// client only needs to forward the year query parameter и treat the
// response as a binary Blob (responseType: 'blob' bypasses JSON parsing).
export const annualReportApi = {
  async download(year: number): Promise<Blob> {
    const response = await apiClient.get('/api/reports/annual', {
      params: { year },
      responseType: 'blob',
    })
    return (response as { data: Blob }).data
  },
}
