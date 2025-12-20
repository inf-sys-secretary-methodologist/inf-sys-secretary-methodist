import type { Metadata } from 'next'
import { getTranslations } from 'next-intl/server'

export async function generateMetadata(): Promise<Metadata> {
  const t = await getTranslations('errorPages.offline')
  const tNav = await getTranslations('navigation')

  return {
    title: `${t('title')} | ${tNav('appName')}`,
    description: t('description'),
  }
}

export default function OfflineLayout({ children }: { children: React.ReactNode }) {
  return children
}
