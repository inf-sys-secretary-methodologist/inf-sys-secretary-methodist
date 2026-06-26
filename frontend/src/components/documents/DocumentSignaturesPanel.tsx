'use client'

import { useState } from 'react'
import { useTranslations } from 'next-intl'
import { ShieldCheck, ShieldAlert, ShieldX, Loader2 } from 'lucide-react'

import { Button } from '@/components/ui/button'
import { useDocumentSignatures, verifySignature } from '@/hooks/useDocumentSignatures'
import type { SignatureVerification } from '@/types/document'

interface DocumentSignaturesPanelProps {
  documentId: number
  enabled?: boolean
}

type VerdictState = Record<number, SignatureVerification | 'loading'>

// DocumentSignaturesPanel lists a document's cryptographic signatures and lets
// the viewer verify each against the document's current state (#140). The
// verdict badge reflects whether the body was modified after signing.
export function DocumentSignaturesPanel({
  documentId,
  enabled = true,
}: DocumentSignaturesPanelProps) {
  const t = useTranslations('documentSignatures')
  const { signatures, isLoading } = useDocumentSignatures(documentId, { enabled })
  const [verdicts, setVerdicts] = useState<VerdictState>({})

  const handleVerify = async (sigId: number) => {
    setVerdicts((v) => ({ ...v, [sigId]: 'loading' }))
    try {
      const verdict = await verifySignature(documentId, sigId)
      setVerdicts((v) => ({ ...v, [sigId]: verdict }))
    } catch {
      setVerdicts((v) => {
        const next = { ...v }
        delete next[sigId]
        return next
      })
    }
  }

  if (isLoading) {
    return <p className="text-sm text-muted-foreground">{t('panel.loading')}</p>
  }

  if (signatures.length === 0) {
    return <p className="text-sm text-muted-foreground">{t('panel.empty')}</p>
  }

  return (
    <ul className="space-y-2" data-testid="document-signatures-list">
      {signatures.map((sig) => {
        const verdict = verdicts[sig.id]
        return (
          <li key={sig.id} className="rounded-md border p-3 text-sm">
            <div className="flex items-center justify-between gap-3">
              <div className="min-w-0">
                <div className="font-medium">{sig.signer_name}</div>
                <div className="truncate text-xs text-muted-foreground">
                  {new Date(sig.signed_at).toLocaleString()} · {sig.algorithm} · v
                  {sig.document_version}
                </div>
              </div>
              <div className="flex shrink-0 items-center gap-2">
                {verdict && verdict !== 'loading' && <VerdictBadge verdict={verdict} />}
                <Button
                  size="sm"
                  variant="outline"
                  onClick={() => handleVerify(sig.id)}
                  disabled={verdict === 'loading'}
                >
                  {verdict === 'loading' ? (
                    <Loader2 className="h-4 w-4 animate-spin" />
                  ) : (
                    t('actions.verify')
                  )}
                </Button>
              </div>
            </div>
          </li>
        )
      })}
    </ul>
  )
}

function VerdictBadge({ verdict }: { verdict: SignatureVerification }) {
  const t = useTranslations('documentSignatures')
  const base = 'inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-xs font-medium'

  if (verdict.valid) {
    return (
      <span className={`${base} bg-green-100 text-green-800`}>
        <ShieldCheck className="h-3.5 w-3.5" />
        {t('verdicts.valid')}
      </span>
    )
  }
  if (!verdict.digest_match) {
    return (
      <span className={`${base} bg-amber-100 text-amber-800`}>
        <ShieldAlert className="h-3.5 w-3.5" />
        {t('verdicts.document_modified')}
      </span>
    )
  }
  return (
    <span className={`${base} bg-red-100 text-red-800`}>
      <ShieldX className="h-3.5 w-3.5" />
      {t('verdicts.crypto_invalid')}
    </span>
  )
}
