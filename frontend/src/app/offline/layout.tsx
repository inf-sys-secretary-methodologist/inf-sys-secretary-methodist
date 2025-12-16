import type { Metadata } from 'next'

export const metadata: Metadata = {
  title: 'Офлайн | СМ ИС',
  description: 'Нет подключения к интернету',
}

export default function OfflineLayout({ children }: { children: React.ReactNode }) {
  return children
}
