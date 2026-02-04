import { apiClient } from '../api'
import { DocumentInfo, DocumentResponse } from './documents'

// Template variable definition
export interface TemplateVariable {
  name: string
  description?: string
  default_value?: string
  required: boolean
  variable_type: 'text' | 'date' | 'number' | 'select'
  options?: string[] // For select type
}

// Template response from API
export interface TemplateInfo {
  id: number
  name: string
  code: string
  description?: string
  template_content?: string
  template_variables?: TemplateVariable[]
  has_template: boolean
}

export interface TemplateListResponse {
  templates: TemplateInfo[]
  total: number
}

export interface PreviewTemplateRequest {
  variables: Record<string, string>
}

export interface PreviewTemplateResponse {
  content: string
}

export interface CreateFromTemplateRequest {
  title: string
  variables: Record<string, string>
  category_id?: number
}

export interface UpdateTemplateRequest {
  template_content?: string
  template_variables?: TemplateVariable[]
}

export const templatesApi = {
  /**
   * Get all available document templates
   */
  async getAll(): Promise<TemplateInfo[]> {
    const response = await apiClient.get<TemplateListResponse>('/api/templates')
    return response.templates || []
  },

  /**
   * Get a specific template by ID
   */
  async getById(id: number | string): Promise<TemplateInfo> {
    return apiClient.get<TemplateInfo>(`/api/templates/${id}`)
  },

  /**
   * Preview template with variables (renders without creating document)
   */
  async preview(id: number | string, variables: Record<string, string>): Promise<string> {
    const response = await apiClient.post<PreviewTemplateResponse>(`/api/templates/${id}/preview`, {
      variables,
    })
    return response.content
  },

  /**
   * Create a new document from a template
   */
  async createDocument(
    templateId: number | string,
    params: CreateFromTemplateRequest
  ): Promise<DocumentInfo> {
    const response = await apiClient.post<DocumentResponse>(
      `/api/templates/${templateId}/create`,
      params
    )
    return response.data
  },

  /**
   * Update template content and variables (admin/secretary only)
   */
  async update(id: number | string, params: UpdateTemplateRequest): Promise<void> {
    await apiClient.put(`/api/templates/${id}`, params)
  },
}
