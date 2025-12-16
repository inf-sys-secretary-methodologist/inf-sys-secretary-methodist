import type { MetadataRoute } from 'next'

export default function manifest(): MetadataRoute.Manifest {
  return {
    name: 'Секретарь-Методист | Информационная система',
    short_name: 'СМ ИС',
    description:
      'Информационная система для управления документооборотом и автоматизации работы секретаря-методиста учебного заведения',
    start_url: '/',
    display: 'standalone',
    background_color: '#ffffff',
    theme_color: '#0f172a',
    orientation: 'portrait-primary',
    scope: '/',
    lang: 'ru',
    categories: ['education', 'productivity', 'business'],
    icons: [
      {
        src: '/icons/icon-72x72.png',
        sizes: '72x72',
        type: 'image/png',
        purpose: 'maskable',
      },
      {
        src: '/icons/icon-96x96.png',
        sizes: '96x96',
        type: 'image/png',
        purpose: 'maskable',
      },
      {
        src: '/icons/icon-128x128.png',
        sizes: '128x128',
        type: 'image/png',
        purpose: 'maskable',
      },
      {
        src: '/icons/icon-144x144.png',
        sizes: '144x144',
        type: 'image/png',
        purpose: 'maskable',
      },
      {
        src: '/icons/icon-152x152.png',
        sizes: '152x152',
        type: 'image/png',
        purpose: 'maskable',
      },
      {
        src: '/icons/icon-192x192.png',
        sizes: '192x192',
        type: 'image/png',
        purpose: 'any',
      },
      {
        src: '/icons/icon-384x384.png',
        sizes: '384x384',
        type: 'image/png',
        purpose: 'any',
      },
      {
        src: '/icons/icon-512x512.png',
        sizes: '512x512',
        type: 'image/png',
        purpose: 'any',
      },
    ],
    screenshots: [
      {
        src: '/screenshots/desktop.png',
        sizes: '1280x720',
        type: 'image/png',
        form_factor: 'wide',
        label: 'Рабочий стол приложения',
      },
      {
        src: '/screenshots/mobile.png',
        sizes: '750x1334',
        type: 'image/png',
        form_factor: 'narrow',
        label: 'Мобильная версия',
      },
    ],
    shortcuts: [
      {
        name: 'Документы',
        short_name: 'Документы',
        description: 'Перейти к документам',
        url: '/documents',
        icons: [{ src: '/icons/shortcut-documents.png', sizes: '96x96' }],
      },
      {
        name: 'Календарь',
        short_name: 'Календарь',
        description: 'Открыть календарь',
        url: '/calendar',
        icons: [{ src: '/icons/shortcut-calendar.png', sizes: '96x96' }],
      },
      {
        name: 'Уведомления',
        short_name: 'Уведомления',
        description: 'Просмотр уведомлений',
        url: '/notifications',
        icons: [{ src: '/icons/shortcut-notifications.png', sizes: '96x96' }],
      },
    ],
    related_applications: [],
    prefer_related_applications: false,
  }
}
