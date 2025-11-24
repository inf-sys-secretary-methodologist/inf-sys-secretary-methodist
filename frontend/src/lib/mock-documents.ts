/**
 * Mock Documents Data
 *
 * Provides sample documents for development and testing.
 * This will be replaced with actual API calls in production.
 */

import { Document, DocumentCategory, DocumentStatus } from '@/types/document'

export const mockDocuments: Document[] = [
  {
    id: '1',
    name: 'Учебный план 2024-2025.pdf',
    category: DocumentCategory.SYLLABUS,
    status: DocumentStatus.READY,
    metadata: {
      size: 2457600,
      mimeType: 'application/pdf',
      uploadedBy: 'Иванов И.И.',
      uploadedAt: new Date('2024-01-15T10:30:00'),
      modifiedAt: new Date('2024-01-20T14:15:00'),
      version: 2,
    },
    url: 'https://www.w3.org/WAI/ER/tests/xhtml/testfiles/resources/pdf/dummy.pdf',
    description: 'Утвержденный учебный план на 2024-2025 учебный год',
    tags: ['план', '2024-2025', 'утвержденный'],
  },
  {
    id: '2',
    name: 'Посещаемость январь.xlsx',
    category: DocumentCategory.ATTENDANCE,
    status: DocumentStatus.READY,
    metadata: {
      size: 524288,
      mimeType: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet',
      uploadedBy: 'Петрова А.С.',
      uploadedAt: new Date('2024-02-01T09:00:00'),
    },
    description: 'Сводная таблица посещаемости студентов за январь',
    tags: ['посещаемость', 'январь', 'отчет'],
  },
  {
    id: '3',
    name: 'Оценки за семестр.xlsx',
    category: DocumentCategory.GRADES,
    status: DocumentStatus.READY,
    metadata: {
      size: 1048576,
      mimeType: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet',
      uploadedBy: 'Сидоров П.В.',
      uploadedAt: new Date('2024-01-28T16:45:00'),
      modifiedAt: new Date('2024-02-05T11:20:00'),
      version: 3,
    },
    description: 'Итоговые оценки студентов за осенний семестр',
    tags: ['оценки', 'семестр', 'итоговые'],
  },
  {
    id: '4',
    name: 'Отчет о практике.docx',
    category: DocumentCategory.REPORT,
    status: DocumentStatus.PROCESSING,
    metadata: {
      size: 786432,
      mimeType: 'application/vnd.openxmlformats-officedocument.wordprocessingml.document',
      uploadedBy: 'Козлов В.М.',
      uploadedAt: new Date('2024-02-10T13:30:00'),
    },
    description: 'Отчет о прохождении производственной практики',
    tags: ['практика', 'отчет'],
  },
  {
    id: '5',
    name: 'Задание на курсовую работу.pdf',
    category: DocumentCategory.ASSIGNMENT,
    status: DocumentStatus.READY,
    metadata: {
      size: 327680,
      mimeType: 'application/pdf',
      uploadedBy: 'Морозова Е.А.',
      uploadedAt: new Date('2024-02-08T10:15:00'),
    },
    url: 'https://www.w3.org/WAI/ER/tests/xhtml/testfiles/resources/pdf/dummy.pdf',
    description: 'Методические указания и задание на курсовую работу',
    tags: ['курсовая', 'задание', 'методичка'],
  },
  {
    id: '6',
    name: 'Экзаменационные билеты.docx',
    category: DocumentCategory.EXAM,
    status: DocumentStatus.READY,
    metadata: {
      size: 458752,
      mimeType: 'application/vnd.openxmlformats-officedocument.wordprocessingml.document',
      uploadedBy: 'Новиков К.Л.',
      uploadedAt: new Date('2024-01-25T12:00:00'),
      version: 1,
    },
    description: 'Утвержденные экзаменационные билеты по математике',
    tags: ['экзамен', 'математика', 'билеты'],
  },
  {
    id: '7',
    name: 'График защиты дипломов.pdf',
    category: DocumentCategory.OTHER,
    status: DocumentStatus.READY,
    metadata: {
      size: 204800,
      mimeType: 'application/pdf',
      uploadedBy: 'Романова О.И.',
      uploadedAt: new Date('2024-02-12T15:20:00'),
    },
    url: 'https://www.w3.org/WAI/ER/tests/xhtml/testfiles/resources/pdf/dummy.pdf',
    description: 'График защиты выпускных квалификационных работ',
    tags: ['диплом', 'график', 'защита'],
  },
  {
    id: '8',
    name: 'Список студентов.xlsx',
    category: DocumentCategory.OTHER,
    status: DocumentStatus.UPLOADING,
    metadata: {
      size: 163840,
      mimeType: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet',
      uploadedBy: 'Волкова Н.С.',
      uploadedAt: new Date('2024-02-15T09:45:00'),
    },
    description: 'Актуальный список студентов всех курсов',
    tags: ['студенты', 'список'],
  },
  {
    id: '9',
    name: 'Протокол заседания.docx',
    category: DocumentCategory.REPORT,
    status: DocumentStatus.ERROR,
    metadata: {
      size: 245760,
      mimeType: 'application/vnd.openxmlformats-officedocument.wordprocessingml.document',
      uploadedBy: 'Лебедев А.Г.',
      uploadedAt: new Date('2024-02-14T11:30:00'),
    },
    description: 'Протокол заседания кафедры от 14.02.2024',
    tags: ['протокол', 'кафедра'],
  },
  {
    id: '10',
    name: 'Учебные материалы.zip',
    category: DocumentCategory.OTHER,
    status: DocumentStatus.READY,
    metadata: {
      size: 15728640,
      mimeType: 'application/zip',
      uploadedBy: 'Семенов Д.Р.',
      uploadedAt: new Date('2024-02-11T14:00:00'),
    },
    description: 'Архив с учебными материалами для студентов',
    tags: ['материалы', 'архив', 'обучение'],
  },
]

// Helper function to filter documents
export function filterDocuments(
  documents: Document[],
  filters: {
    search?: string
    category?: DocumentCategory
    status?: DocumentStatus
    tags?: string[]
  }
): Document[] {
  let filtered = [...documents]

  if (filters.search) {
    const searchLower = filters.search.toLowerCase()
    filtered = filtered.filter(
      (doc) =>
        doc.name.toLowerCase().includes(searchLower) ||
        doc.description?.toLowerCase().includes(searchLower) ||
        doc.tags?.some((tag) => tag.toLowerCase().includes(searchLower))
    )
  }

  if (filters.category) {
    filtered = filtered.filter((doc) => doc.category === filters.category)
  }

  if (filters.status) {
    filtered = filtered.filter((doc) => doc.status === filters.status)
  }

  if (filters.tags && filters.tags.length > 0) {
    filtered = filtered.filter((doc) =>
      filters.tags!.some((filterTag) =>
        doc.tags?.some((docTag) => docTag.toLowerCase().includes(filterTag.toLowerCase()))
      )
    )
  }

  return filtered
}

// Helper function to sort documents
export function sortDocuments(
  documents: Document[],
  field: 'name' | 'uploadedAt' | 'modifiedAt' | 'size',
  order: 'asc' | 'desc'
): Document[] {
  const sorted = [...documents]

  sorted.sort((a, b) => {
    let comparison = 0

    switch (field) {
      case 'name':
        comparison = a.name.localeCompare(b.name)
        break
      case 'uploadedAt':
        comparison =
          new Date(a.metadata.uploadedAt).getTime() - new Date(b.metadata.uploadedAt).getTime()
        break
      case 'modifiedAt':
        const aModified = a.metadata.modifiedAt || a.metadata.uploadedAt
        const bModified = b.metadata.modifiedAt || b.metadata.uploadedAt
        comparison = new Date(aModified).getTime() - new Date(bModified).getTime()
        break
      case 'size':
        comparison = a.metadata.size - b.metadata.size
        break
    }

    return order === 'asc' ? comparison : -comparison
  })

  return sorted
}
