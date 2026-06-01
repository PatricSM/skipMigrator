import { Link, useParams } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { useEffect } from 'react'
import { ArrowLeft, Package, Download, Loader2, CircleCheck, CircleAlert, Clock } from 'lucide-react'
import { getMigration, downloadURL, type Migration } from '@/lib/api'
import { supabase } from '@/lib/supabase'

const META: Record<Migration['Status'], { label: string; color: string; icon: React.ReactNode }> = {
  queued:  { label: 'Queued — waiting for a worker',   color: 'text-muted-foreground', icon: <Clock className="size-5" /> },
  running: { label: 'Running — typically 30s to 3min', color: 'text-primary',          icon: <Loader2 className="size-5 animate-spin" /> },
  success: { label: 'Success — your ZIP is ready',     color: 'text-emerald-400',      icon: <CircleCheck className="size-5" /> },
  failed:  { label: 'Failed',                          color: 'text-destructive',      icon: <CircleAlert className="size-5" /> },
}

export default function MigrationDetail() {
  const { id = '' } = useParams()
  const { data: m, refetch } = useQuery({
    queryKey: ['migration', id],
    queryFn: () => getMigration(id),
    refetchInterval: (q) => {
      const d = q.state.data as Migration | undefined
      return d?.Status === 'queued' || d?.Status === 'running' ? 2000 : false
    },
  })

  useEffect(() => {
    const ch = supabase
      .channel(`migration-${id}`)
      .on('postgres_changes',
        { event: 'UPDATE', schema: 'public', table: 'migrations', filter: `id=eq.${id}` },
        () => refetch())
      .subscribe()
    return () => { void supabase.removeChannel(ch) }
  }, [id, refetch])

  if (!m) return <div className="grid h-screen place-items-center text-muted-foreground">Loading…</div>

  const meta = META[m.Status]

  return (
    <div className="min-h-screen">
      <header className="container flex items-center justify-between py-6">
        <Link to="/app" className="flex items-center gap-2 text-sm text-muted-foreground hover:text-foreground">
          <ArrowLeft className="size-4" /> Dashboard
        </Link>
        <Link to="/" className="flex items-center gap-2 text-lg font-bold">
          <Package className="size-5 text-primary" /> Skip Migrator
        </Link>
      </header>

      <main className="container max-w-3xl py-8">
        <div className={`mb-6 flex items-center gap-3 rounded-xl border border-border bg-card p-6 ${meta.color}`}>
          {meta.icon}
          <div className="flex-1">
            <h1 className="text-xl font-bold">{meta.label}</h1>
            <p className="text-sm text-muted-foreground">
              Migration <span className="font-mono">{m.ID.slice(0, 8)}</span> · {new Date(m.CreatedAt).toLocaleString()}
            </p>
          </div>
          {m.Status === 'success' && m.OutputZipPath && (
            <a
              href={downloadURL(m.ID)}
              className="inline-flex items-center gap-2 rounded-lg bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:opacity-90"
            >
              <Download className="size-4" /> Download ZIP
            </a>
          )}
        </div>

        <div className="grid gap-3 rounded-xl border border-border bg-card p-6 text-sm">
          <Row k="Source ZIP"          v={m.SourceZipPath.split('/').pop() ?? ''} />
          <Row k="Pixel-perfect"       v={m.PixelPerfect ? 'yes' : 'no'} />
          <Row k="Build validated"     v={m.Validate ? 'yes' : 'no'} />
          <Row k="Supabase strategy"   v={m.SupabaseStrategy} />
          {m.ErrorMessage && <Row k="Error" v={m.ErrorMessage} />}
        </div>

        {m.BuildLog && (
          <details className="mt-6 rounded-xl border border-border bg-card">
            <summary className="cursor-pointer p-4 text-sm font-medium">Build log</summary>
            <pre className="max-h-[480px] overflow-auto p-4 pt-0 text-xs text-muted-foreground">{m.BuildLog}</pre>
          </details>
        )}
      </main>
    </div>
  )
}

function Row({ k, v }: { k: string; v: string }) {
  return (
    <div className="flex justify-between gap-4">
      <span className="text-muted-foreground">{k}</span>
      <span className="font-medium text-right break-all">{v}</span>
    </div>
  )
}
