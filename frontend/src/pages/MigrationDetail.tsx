import { Link, useParams } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { useEffect } from 'react'
import { ArrowLeft, Download, Loader2, CircleCheck, CircleAlert, Clock } from 'lucide-react'
import { getMigration, downloadURL, type Migration } from '@/lib/api'
import { supabase } from '@/lib/supabase'
import SupabaseChecklist from '@/components/SupabaseChecklist'
import AppHeader from '@/components/AppHeader'

const META: Record<Migration['Status'], { label: string; color: string; icon: React.ReactNode }> = {
  queued:  { label: 'Na fila — aguardando um worker',      color: 'text-muted-foreground', icon: <Clock className="size-5" /> },
  running: { label: 'Processando — geralmente 30s a 3min', color: 'text-primary',          icon: <Loader2 className="size-5 animate-spin" /> },
  success: { label: 'Concluída — seu ZIP está pronto',     color: 'text-success',          icon: <CircleCheck className="size-5" /> },
  failed:  { label: 'Falhou',                              color: 'text-destructive',      icon: <CircleAlert className="size-5" /> },
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

  if (!m) return <div className="grid h-screen place-items-center text-muted-foreground">Carregando…</div>

  const meta = META[m.Status]

  return (
    <div className="min-h-screen">
      <AppHeader showAppNav />
      <main className="container max-w-3xl py-8">
        <Link to="/app" className="mb-4 inline-flex items-center gap-2 text-sm text-muted-foreground hover:text-foreground">
          <ArrowLeft className="size-4" /> Painel
        </Link>

        <div className={`mb-6 flex items-center gap-4 rounded-lg border border-border bg-card p-6 ${meta.color}`}>
          {meta.icon}
          <div className="flex-1">
            <h1 className="font-display text-xl font-bold">{meta.label}</h1>
            <p className="text-sm text-muted-foreground">
              Migração <span className="font-mono">{m.ID.slice(0, 8)}</span> · {new Date(m.CreatedAt).toLocaleString('pt-BR')}
            </p>
          </div>
          {m.Status === 'success' && m.OutputZipPath && (
            <a
              href={downloadURL(m.ID)}
              className="inline-flex items-center gap-2 rounded-md bg-primary px-5 py-2.5 text-sm font-semibold text-primary-foreground shadow-glow transition hover:opacity-95"
            >
              <Download className="size-4" /> Baixar ZIP
            </a>
          )}
        </div>

        <div className="grid gap-3 rounded-lg border border-border bg-card p-6 text-sm">
          <Row k="ZIP de origem"        v={m.SourceZipPath.split('/').pop() ?? ''} />
          <Row k="Pixel-perfect"        v={m.PixelPerfect ? 'sim' : 'não'} />
          <Row k="Build validado"       v={m.Validate ? 'sim' : 'não'} />
          <Row k="Estratégia Supabase"  v={m.SupabaseStrategy === 'extract' ? 'extrair do source' : 'novo projeto'} />
          {m.ErrorMessage && <Row k="Erro" v={m.ErrorMessage} />}
        </div>

        {m.Status === 'success' && m.SupabaseProjectRef && (
          <SupabaseChecklist projectRef={m.SupabaseProjectRef} />
        )}

        {m.BuildLog && (
          <details className="mt-6 rounded-lg border border-border bg-card">
            <summary className="cursor-pointer p-4 text-sm font-semibold">Log do build</summary>
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
      <span className="break-all text-right font-medium">{v}</span>
    </div>
  )
}
