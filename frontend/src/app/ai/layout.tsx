import { Metadata } from 'next'

export const metadata: Metadata = {
  title: 'AI Assistant',
  description: 'AI-powered assistant for document search and analysis',
}

export default function AILayout({ children }: { children: React.ReactNode }) {
  return children
}
