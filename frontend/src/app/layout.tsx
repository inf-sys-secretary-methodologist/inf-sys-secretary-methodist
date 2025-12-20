import type { Metadata, Viewport } from 'next'
import { NextIntlClientProvider } from 'next-intl'
import { getLocale, getMessages, getTranslations } from 'next-intl/server'
import './globals.css'
import { ThemeProvider } from '@/components/providers/theme-provider'
import { ToasterProvider } from '@/components/providers/toaster-provider'
import { ServiceWorkerRegistration } from '@/components/pwa/service-worker-registration'
import { BackgroundProvider } from '@/components/backgrounds'
import { ScreenReaderAnnouncerProvider } from '@/components/ui/screen-reader-announcer'
import { isRtlLocale, type Locale } from '@/i18n/config'

const ogLocaleMap: Record<string, string> = {
  ru: 'ru_RU',
  en: 'en_US',
  fr: 'fr_FR',
  ar: 'ar_SA',
}

function getOgLocale(locale: string): string {
  return ogLocaleMap[locale] || 'en_US'
}

export async function generateMetadata(): Promise<Metadata> {
  const locale = await getLocale()
  const t = await getTranslations('metadata')

  return {
    title: {
      default: t('title'),
      template: t('titleTemplate'),
    },
    description: t('description'),
    keywords: t('keywords').split(', '),
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
      locale: getOgLocale(locale),
      siteName: t('siteName'),
      title: t('title'),
      description: t('ogDescription'),
      images: [
        {
          url: '/og-image.png',
          width: 1200,
          height: 630,
          alt: t('ogImageAlt'),
        },
      ],
    },
    twitter: {
      card: 'summary_large_image',
      title: t('title'),
      description: t('ogDescription'),
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
      title: t('shortName'),
    },
  }
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

export default async function RootLayout({ children }: { children: React.ReactNode }) {
  const locale = (await getLocale()) as Locale
  const messages = await getMessages()
  const isRtl = isRtlLocale(locale)

  return (
    <html lang={locale} dir={isRtl ? 'rtl' : 'ltr'} suppressHydrationWarning>
      <body>
        <NextIntlClientProvider messages={messages}>
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
        </NextIntlClientProvider>
      </body>
    </html>
  )
}
