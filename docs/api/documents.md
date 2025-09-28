# 📄 Document Management API

## 📋 Обзор

Сервис управления документами обеспечивает полный жизненный цикл документооборота: создание, редактирование, версионирование, поиск и интеграция с workflow-процессами.

## 🌐 Base URL
```
https://api.inf-sys.example.com/documents
```

## 📚 Типы документов

- **curriculum** - Учебные планы
- **syllabus** - Силлабусы
- **methodology** - Методические материалы
- **report** - Отчеты
- **regulation** - Положения и регламенты
- **schedule** - Расписание
- **form** - Формы и бланки

---

## 🚀 Endpoints

### GET `/documents`
Получение списка документов с фильтрацией и поиском

**Query Parameters:**
```
?type=curriculum&status=published&author_id=123&department=ИТ
&created_after=2025-01-01&search=математика
&tags=анализ,базовый&page=1&limit=20
&sort=created_at&order=desc
```

**Response (200):**
```json
{
  "documents": [
    {
      "id": "doc-12345",
      "title": "Учебный план по математическому анализу",
      "type": "curriculum",
      "status": "published",
      "author": {
        "id": "user-123",
        "name": "Анна Петрова",
        "department": "Математический факультет"
      },
      "version": "2.1",
      "created_at": "2025-01-01T10:00:00Z",
      "updated_at": "2025-01-15T14:30:00Z",
      "published_at": "2025-01-10T12:00:00Z",
      "tags": ["математика", "анализ", "базовый_курс"],
      "metadata": {
        "department": "Математический факультет",
        "academic_year": "2024-2025",
        "semester": 1,
        "credits": 6,
        "hours": 180
      },
      "workflow": {
        "status": "approved",
        "current_step": null,
        "approved_by": "user-456",
        "approved_at": "2025-01-10T11:45:00Z"
      },
      "permissions": {
        "can_edit": true,
        "can_delete": false,
        "can_publish": true
      }
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 89,
    "pages": 5,
    "has_next": true,
    "has_prev": false
  },
  "filters_applied": {
    "type": "curriculum",
    "status": "published",
    "department": "ИТ"
  }
}
```

### POST `/documents`
Создание нового документа

**Request:**
```json
{
  "title": "Новый учебный план по программированию",
  "type": "curriculum",
  "content": {
    "description": "Учебный план для изучения основ программирования",
    "objectives": [
      "Изучить основы алгоритмизации",
      "Освоить язык программирования Python",
      "Изучить структуры данных"
    ],
    "subjects": [
      {
        "name": "Основы программирования",
        "code": "PROG-101",
        "hours": 120,
        "credits": 4,
        "type": "lecture",
        "semester": 1
      },
      {
        "name": "Практикум по программированию",
        "code": "PROG-102",
        "hours": 60,
        "credits": 2,
        "type": "practice",
        "semester": 1
      }
    ],
    "requirements": [
      "Базовые знания математики",
      "Логическое мышление"
    ],
    "assessment": {
      "methods": ["exam", "coursework", "labs"],
      "weights": {
        "exam": 40,
        "coursework": 30,
        "labs": 30
      }
    }
  },
  "metadata": {
    "department": "Информационных технологий",
    "academic_year": "2024-2025",
    "semester": 1,
    "level": "bachelor",
    "language": "ru"
  },
  "tags": ["программирование", "python", "алгоритмы"],
  "workflow_template": "curriculum_approval"
}
```

**Response (201):**
```json
{
  "id": "doc-67890",
  "title": "Новый учебный план по программированию",
  "type": "curriculum",
  "status": "draft",
  "author": {
    "id": "user-123",
    "name": "Анна Петрова"
  },
  "version": "1.0",
  "created_at": "2025-01-15T16:00:00Z",
  "workflow": {
    "id": "wf-123",
    "status": "in_progress",
    "current_step": "author_review"
  }
}
```

### GET `/documents/{id}`
Получение документа по ID

**Response (200):**
```json
{
  "id": "doc-12345",
  "title": "Учебный план по математическому анализу",
  "type": "curriculum",
  "status": "published",
  "content": {
    "description": "Полный курс математического анализа для первого курса",
    "objectives": [
      "Изучить основы дифференциального исчисления",
      "Освоить интегральное исчисление",
      "Изучить ряды и их сходимость"
    ],
    "subjects": [
      {
        "name": "Математический анализ I",
        "code": "MATH-101",
        "hours": 120,
        "credits": 4,
        "type": "lecture",
        "semester": 1
      }
    ]
  },
  "author": {
    "id": "user-123",
    "name": "Анна Петрова",
    "department": "Математический факультет",
    "position": "Старший методист"
  },
  "version": "2.1",
  "created_at": "2025-01-01T10:00:00Z",
  "updated_at": "2025-01-15T14:30:00Z",
  "published_at": "2025-01-10T12:00:00Z",
  "tags": ["математика", "анализ", "базовый_курс"],
  "metadata": {
    "department": "Математический факультет",
    "academic_year": "2024-2025",
    "semester": 1,
    "credits": 6,
    "hours": 180
  },
  "workflow": {
    "id": "wf-456",
    "status": "approved",
    "current_step": null,
    "approved_by": "user-456",
    "approved_at": "2025-01-10T11:45:00Z",
    "history": [
      {
        "step": "creation",
        "status": "completed",
        "user": "user-123",
        "timestamp": "2025-01-01T10:00:00Z"
      }
    ]
  },
  "permissions": {
    "can_edit": true,
    "can_delete": false,
    "can_publish": true,
    "can_archive": true
  },
  "analytics": {
    "views": 156,
    "downloads": 23,
    "last_accessed": "2025-01-15T09:30:00Z"
  }
}
```

### PUT `/documents/{id}`
Обновление документа

**Request:**
```json
{
  "title": "Обновленный учебный план по математическому анализу",
  "content": {
    "description": "Расширенный курс с дополнительными темами",
    "objectives": [
      "Изучить основы дифференциального исчисления",
      "Освоить интегральное исчисление",
      "Изучить ряды и их сходимость",
      "Изучить функции многих переменных"
    ]
  },
  "tags": ["математика", "анализ", "расширенный_курс"],
  "metadata": {
    "credits": 8,
    "hours": 240
  }
}
```

**Response (200):**
```json
{
  "id": "doc-12345",
  "version": "2.2",
  "updated_at": "2025-01-15T17:00:00Z",
  "status": "draft",
  "workflow": {
    "id": "wf-789",
    "status": "in_progress",
    "current_step": "internal_review"
  }
}
```

### DELETE `/documents/{id}`
Архивирование документа (мягкое удаление)

**Response (200):**
```json
{
  "message": "Document archived successfully",
  "archived_at": "2025-01-15T17:30:00Z"
}
```

---

## 📝 Версионирование

### GET `/documents/{id}/versions`
Получение истории версий документа

**Response (200):**
```json
{
  "versions": [
    {
      "version": "2.1",
      "created_at": "2025-01-15T14:30:00Z",
      "author": "user-123",
      "status": "published",
      "changes_summary": "Добавлены новые темы и обновлены часы",
      "is_current": true
    },
    {
      "version": "2.0",
      "created_at": "2025-01-10T12:00:00Z",
      "author": "user-123",
      "status": "published",
      "changes_summary": "Крупное обновление структуры курса",
      "is_current": false
    },
    {
      "version": "1.0",
      "created_at": "2025-01-01T10:00:00Z",
      "author": "user-123",
      "status": "archived",
      "changes_summary": "Первоначальная версия",
      "is_current": false
    }
  ],
  "pagination": {
    "page": 1,
    "total": 3
  }
}
```

### POST `/documents/{id}/versions`
Создание новой версии документа

**Request:**
```json
{
  "changes_summary": "Обновлена программа практических занятий",
  "auto_increment": true
}
```

**Response (201):**
```json
{
  "version": "2.2",
  "created_at": "2025-01-15T18:00:00Z",
  "status": "draft"
}
```

### GET `/documents/{id}/versions/{version}`
Получение конкретной версии документа

---

## 🔍 Поиск и фильтрация

### GET `/documents/search`
Расширенный поиск документов

**Query Parameters:**
```
?query=математический анализ
&type[]=curriculum&type[]=syllabus
&department[]=Математический факультет
&tags[]=анализ&tags[]=математика
&status[]=published&status[]=draft
&author_id[]=123&author_id[]=456
&created_from=2025-01-01&created_to=2025-01-31
&academic_year=2024-2025
&semester=1
&fuzzy=true
&include_content=true
```

**Response (200):**
```json
{
  "documents": [
    {
      "id": "doc-12345",
      "title": "Учебный план по <em>математическому анализу</em>",
      "type": "curriculum",
      "excerpt": "Полный курс <em>математического анализа</em> для первого курса...",
      "relevance_score": 0.95,
      "match_fields": ["title", "content.description"],
      "author": {
        "name": "Анна Петрова"
      }
    }
  ],
  "facets": {
    "types": {
      "curriculum": 45,
      "syllabus": 23,
      "methodology": 12
    },
    "departments": {
      "Математический факультет": 34,
      "ИТ факультет": 21
    },
    "authors": {
      "Анна Петрова": 15,
      "Иван Иванов": 12
    }
  },
  "suggestions": [
    "математическое моделирование",
    "математическая статистика"
  ]
}
```

---

## 🔄 Workflow Integration

### GET `/documents/{id}/workflow`
Получение информации о workflow документа

**Response (200):**
```json
{
  "workflow": {
    "id": "wf-123",
    "template": "curriculum_approval",
    "status": "in_progress",
    "current_step": {
      "name": "methodical_review",
      "title": "Методическая экспертиза",
      "assignees": [
        {
          "id": "user-456",
          "name": "Экспертная комиссия",
          "department": "Методический совет"
        }
      ],
      "deadline": "2025-01-20T17:00:00Z",
      "started_at": "2025-01-15T10:00:00Z"
    },
    "completed_steps": [
      {
        "name": "creation",
        "completed_at": "2025-01-01T10:00:00Z",
        "completed_by": "user-123"
      },
      {
        "name": "internal_review",
        "completed_at": "2025-01-10T15:00:00Z",
        "completed_by": "user-124"
      }
    ],
    "remaining_steps": [
      "methodical_review",
      "final_approval",
      "publication"
    ]
  }
}
```

### POST `/documents/{id}/workflow/advance`
Продвижение документа по workflow

**Request:**
```json
{
  "action": "approve",
  "comment": "Документ соответствует требованиям",
  "attachments": ["file-123", "file-456"]
}
```

### POST `/documents/{id}/workflow/reject`
Отклонение документа в workflow

**Request:**
```json
{
  "reason": "Требуется доработка раздела практических занятий",
  "return_to_step": "author_review",
  "attachments": ["feedback-file-123"]
}
```

---

## 📊 Аналитика и статистика

### GET `/documents/{id}/analytics`
Получение аналитики по документу

**Response (200):**
```json
{
  "views": {
    "total": 156,
    "unique": 87,
    "last_30_days": 45
  },
  "downloads": {
    "total": 23,
    "last_30_days": 8,
    "by_format": {
      "pdf": 18,
      "docx": 5
    }
  },
  "engagement": {
    "average_time_spent": 420,
    "bounce_rate": 0.12,
    "return_visits": 23
  },
  "geographic": [
    {
      "region": "Москва",
      "views": 45
    },
    {
      "region": "Санкт-Петербург",
      "views": 23
    }
  ]
}
```

### GET `/documents/analytics/summary`
Общая аналитика по документам

**Query Parameters:**
```
?period=last_30_days&department=ИТ&type=curriculum
```

---

## 📁 Управление файлами

### POST `/documents/{id}/attachments`
Добавление вложения к документу

**Request (multipart/form-data):**
```
file: <binary data>
metadata: {
  "type": "supporting_material",
  "description": "Дополнительные материалы к курсу"
}
```

**Response (201):**
```json
{
  "attachment": {
    "id": "att-123",
    "filename": "supplementary_materials.pdf",
    "size": 2048576,
    "type": "application/pdf",
    "uploaded_at": "2025-01-15T16:00:00Z",
    "url": "https://files.inf-sys.example.com/att-123"
  }
}
```

### GET `/documents/{id}/attachments`
Получение списка вложений документа

### DELETE `/documents/{id}/attachments/{attachment_id}`
Удаление вложения

---

## 🔄 Экспорт и импорт

### POST `/documents/{id}/export`
Экспорт документа в различных форматах

**Request:**
```json
{
  "format": "pdf|docx|html|json",
  "options": {
    "include_metadata": true,
    "include_attachments": true,
    "template": "official"
  }
}
```

**Response (200):**
```json
{
  "export_id": "exp-123",
  "status": "processing",
  "estimated_completion": "2025-01-15T16:05:00Z"
}
```

### GET `/documents/export/{export_id}`
Получение статуса экспорта

### POST `/documents/import`
Импорт документа

**Request (multipart/form-data):**
```
file: <document file>
options: {
  "type": "curriculum",
  "department": "ИТ",
  "auto_workflow": true
}
```

---

## 🔧 Frontend Integration

### React Hook Example
```typescript
import { useState, useEffect } from 'react';

interface Document {
  id: string;
  title: string;
  type: string;
  status: string;
  author: {
    id: string;
    name: string;
  };
  version: string;
  created_at: string;
  updated_at: string;
}

export const useDocuments = () => {
  const [documents, setDocuments] = useState<Document[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const fetchDocuments = async (filters = {}) => {
    setLoading(true);
    setError(null);

    try {
      const params = new URLSearchParams(filters);
      const response = await fetch(`/api/documents?${params}`, {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('access_token')}`
        }
      });

      if (!response.ok) {
        throw new Error('Failed to fetch documents');
      }

      const data = await response.json();
      setDocuments(data.documents);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unknown error');
    } finally {
      setLoading(false);
    }
  };

  const createDocument = async (documentData: Partial<Document>) => {
    try {
      const response = await fetch('/api/documents', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${localStorage.getItem('access_token')}`
        },
        body: JSON.stringify(documentData)
      });

      if (!response.ok) {
        throw new Error('Failed to create document');
      }

      const newDocument = await response.json();
      setDocuments(prev => [newDocument, ...prev]);
      return newDocument;
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unknown error');
      throw err;
    }
  };

  const updateDocument = async (id: string, updates: Partial<Document>) => {
    try {
      const response = await fetch(`/api/documents/${id}`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${localStorage.getItem('access_token')}`
        },
        body: JSON.stringify(updates)
      });

      if (!response.ok) {
        throw new Error('Failed to update document');
      }

      const updatedDocument = await response.json();
      setDocuments(prev =>
        prev.map(doc => doc.id === id ? { ...doc, ...updatedDocument } : doc)
      );
      return updatedDocument;
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unknown error');
      throw err;
    }
  };

  return {
    documents,
    loading,
    error,
    fetchDocuments,
    createDocument,
    updateDocument
  };
};
```

---

## 🧪 Testing

### Go Backend Tests
```go
func TestCreateDocument(t *testing.T) {
    router := setupTestRouter()

    document := map[string]interface{}{
        "title": "Test Document",
        "type":  "curriculum",
        "content": map[string]interface{}{
            "description": "Test description",
        },
        "metadata": map[string]interface{}{
            "department": "ИТ",
        },
        "tags": []string{"test"},
    }

    jsonData, _ := json.Marshal(document)
    req, _ := http.NewRequest("POST", "/documents", bytes.NewBuffer(jsonData))
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", "Bearer "+testToken)

    w := httptest.NewRecorder()
    router.ServeHTTP(w, req)

    assert.Equal(t, 201, w.Code)

    var response map[string]interface{}
    json.Unmarshal(w.Body.Bytes(), &response)

    assert.Equal(t, "Test Document", response["title"])
    assert.Equal(t, "curriculum", response["type"])
}
```

---

## ⚡ Performance

### Метрики производительности
- **Создание документа**: < 300ms (95th percentile)
- **Поиск документов**: < 100ms (99th percentile)
- **Загрузка документа**: < 200ms (95th percentile)
- **Экспорт в PDF**: < 2s (95th percentile)

### Кэширование
- Метаданные документов кэшируются на 5 минут
- Результаты поиска кэшируются на 2 минуты
- Статистика кэшируется на 1 час

---

## 🚨 Error Codes

| Code | Message | Description |
|------|---------|-------------|
| DOC_001 | Document not found | Документ не найден |
| DOC_002 | Invalid document type | Недопустимый тип документа |
| DOC_003 | Document locked | Документ заблокирован для редактирования |
| DOC_004 | Version conflict | Конфликт версий |
| DOC_005 | Workflow violation | Нарушение процесса workflow |
| DOC_006 | File too large | Файл слишком большой |
| DOC_007 | Invalid format | Недопустимый формат |
| DOC_008 | Insufficient permissions | Недостаточно прав для операции |