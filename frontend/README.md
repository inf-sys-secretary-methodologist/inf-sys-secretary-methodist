# Информационная система секретаря-методиста

[![Frontend CI](https://github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/actions/workflows/frontend-ci.yml/badge.svg)](https://github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/actions/workflows/frontend-ci.yml)

Система управления учебной частью с автоматизацией документооборота и управления расписанием.

## 🚀 Технологический стек

### Frontend

- **Next.js 15.1.0** - React фреймворк
- **TypeScript 5.7** - Типизация
- **Tailwind CSS 4.1** - Стилизация
- **Zustand** - Управление состоянием
- **React Hook Form** + **Zod** - Формы и валидация
- **Playwright** - E2E тестирование
- **Jest** + **React Testing Library** - Unit тестирование
- **@paper-design/shaders-react** - Анимированные шейдерные фоны
- **next-themes** - Управление темами (light/dark/system)
- **next-intl** - Интернационализация (ru, en, fr, ar с RTL)
- **PWA** - Service Worker, офлайн-режим, установка как приложение

### Backend

- **Go 1.25** - Основной язык
- **Gin** - HTTP фреймворк
- **PostgreSQL** - База данных
- **Redis** - Кэширование
- **JWT** - Аутентификация

## 📦 Структура проекта

\`\`\`
.
├── frontend/ # Next.js приложение
├── cmd/server/ # Backend сервер
├── internal/ # Внутренние пакеты Go
├── migrations/ # SQL миграции
└── .github/ # CI/CD workflows
\`\`\`

## ✨ Основные возможности

- **🎨 Настройки внешнего вида** - Светлая/тёмная тема, анимированные фоны
- **🌍 Мультиязычность** - Русский, английский, французский, арабский (RTL)
- **📱 PWA** - Установка как приложение, офлайн-режим
- **♿ Доступность** - WCAG 2.1 AA, ARIA-атрибуты, клавиатурная навигация
- **🔔 Уведомления** - Email и Telegram интеграция
- **📄 Документооборот** - Создание, версионирование, шаринг документов
- **📅 Расписание** - Управление учебным расписанием
- **🏢 Интеграция с 1С** - Синхронизация сотрудников и студентов

## 🛠️ Разработка

### Frontend

\`\`\`bash
cd frontend
npm install
npm run dev # Development сервер
npm run build # Production сборка
npm run lint # Линтинг
npm run lint:fix # Автоисправление
npm run type-check # TypeScript проверка
npm run test # Unit тесты
npm run test:e2e # E2E тесты
\`\`\`

### Backend

\`\`\`bash

# Установка зависимостей

go mod download

# Запуск сервера

go run cmd/server/main.go

# Тесты

go test ./...
\`\`\`

## ✅ CI/CD

Frontend CI автоматически проверяет:

- ✅ ESLint & Prettier
- ✅ TypeScript type checking
- ✅ Unit тесты (coverage >= 70%)
- ✅ E2E тесты (Playwright)
- ✅ Production build
- ✅ Bundle size analysis

## 📄 Лицензия

Proprietary

## 👥 Команда

VDV001 и контрибьюторы
