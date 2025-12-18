import type { Metadata, Viewport } from 'next'
import './globals.css'
import { ThemeProvider } from '@/components/providers/theme-provider'
import { ToasterProvider } from '@/components/providers/toaster-provider'
import { ServiceWorkerRegistration } from '@/components/pwa/service-worker-registration'
import { BackgroundProvider } from '@/components/backgrounds'
import { ScreenReaderAnnouncerProvider } from '@/components/ui/screen-reader-announcer'

export const metadata: Metadata = {
  title: {
    default: 'Секретарь-Методист | Информационная система',
    template: '%s | СМ ИС',
  },
  description:
    'Информационная система для управления документооборотом и автоматизации работы секретаря-методиста учебного заведения',
  keywords: [
    'секретарь',
    'методист',
    'документооборот',
    'учебное заведение',
    'автоматизация',
    'расписание',
  ],
  authors: [{ name: 'Inf-Sys Secretary Methodist Team' }],
  creator: 'Inf-Sys Secretary Methodist Team',
  publisher: 'Inf-Sys Secretary Methodist',
  formatDetection: {
    email: false,
    address: false,
    telephone: false,
  },
  metadataBase: new URL(process.env.NEXT_PUBLIC_APP_URL || 'http://localhost:3000'),
  openGraph: {
    type: 'website',
    locale: 'ru_RU',
    siteName: 'Секретарь-Методист ИС',
    title: 'Секретарь-Методист | Информационная система',
    description:
      'Информационная система для управления документооборотом и автоматизации работы секретаря-методиста',
    images: [
      {
        url: '/og-image.png',
        width: 1200,
        height: 630,
        alt: 'Секретарь-Методист ИС',
      },
    ],
  },
  twitter: {
    card: 'summary_large_image',
    title: 'Секретарь-Методист | Информационная система',
    description: 'Информационная система для управления документооборотом учебного заведения',
    images: ['/og-image.png'],
  },
  robots: {
    index: true,
    follow: true,
    googleBot: {
      index: true,
      follow: true,
    },
  },
  icons: {
    icon: [
      { url: '/icons/icon-32x32.png', sizes: '32x32', type: 'image/png' },
      { url: '/icons/icon-16x16.png', sizes: '16x16', type: 'image/png' },
    ],
    shortcut: '/favicon.ico',
    apple: [{ url: '/icons/apple-touch-icon.png', sizes: '180x180', type: 'image/png' }],
    other: [
      {
        rel: 'mask-icon',
        url: '/icons/safari-pinned-tab.svg',
        color: '#0f172a',
      },
    ],
  },
  manifest: '/manifest.webmanifest',
  appleWebApp: {
    capable: true,
    statusBarStyle: 'default',
    title: 'СМ ИС',
  },
}

export const viewport: Viewport = {
  themeColor: [
    { media: '(prefers-color-scheme: light)', color: '#ffffff' },
    { media: '(prefers-color-scheme: dark)', color: '#0f172a' },
  ],
  width: 'device-width',
  initialScale: 1,
  maximumScale: 5,
  userScalable: true,
  viewportFit: 'cover',
}

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="ru" suppressHydrationWarning>
      <body>
        <ThemeProvider
          attribute="class"
          defaultTheme="system"
          enableSystem
          disableTransitionOnChange
        >
          <ScreenReaderAnnouncerProvider>
            <BackgroundProvider />
            {children}
            <ToasterProvider />
            <ServiceWorkerRegistration />
          </ScreenReaderAnnouncerProvider>
        </ThemeProvider>
      </body>
    </html>
  )
}
