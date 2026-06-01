import { Link } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { useEffect } from 'react'
import { Plus, CircleCheck, CircleAlert, Loader2, Clock, Download } from 'lucide-react'
import { listMigrations, downloadURL, type Migration } from '@/lib/api'
import { supabase } from '@/lib/supabase'
import AppHeader from '@/components/AppHeader'

const STATUS_META: Record<Migration['Status'], { label: string; color: string; icon: React.ReactNode }> = {
  queued:  { label: 'Na fila',     color: 'text-muted-foreground', icon: <Clock className="size-4" /> },
  running: { label: 'Processando', color: 'text-primary',          icon: <Loader2 className="size-4 animate-spin" /> },
  success: { label: 'Concluída',   color: 'text-success',          icon: <CircleCheck className="size-4" /> },
  failed:  { label: 'Falhou',      color: 'text-destructive',      icon: <CircleAlert className="size-4" /> },
}

export default function Dashboard() {
  const { data: rows = [], refetch } = useQuery({
    queryKey: ['migrations'],
    queryFn: listMigrations,
    refetchInterval: (q) => {
      const arr = q.state.data as Migration[] | undefined
      return arr?.some((m) => m.Status === 'queued' || m.Status === 'running') ? 3000 : false
    },
  })

  useEffect(() => {
    const ch = supabase
      .channel('migrations-changes')
      .on('postgres_changes', { event: '*', schema: 'public', table: 'migrations' }, () => refetch())
      .subscribe()
    return () => { void supabase.removeChannel(ch) }
  }, [refetch])

  return (
    <div className="min-h-screen">
      <AppHeader showAppNav />
      <main className="container py-8">
        <h1 className="mb-6 font-display text-2xl font-bold">Suas migrações</h1>
        {rows.length === 0 ? (
          <div className="rounded-lg border border-dashed border-border bg-card p-12 text-center">
            <p className="text-muted-foreground">Nenhuma migração por aqui ainda.</p>
            <Link
              to="/app/new"
              className="mt-4 inline-flex items-center gap-2 rounded-md bg-primary px-4 py-2 text-sm font-semibold text-primary-foreground shadow-glow transition hover:opacity-95"
            >
              <Plus className="size-4" /> Iniciar primeira migração
            </Link>
          </div>
        ) : (
          <ul className="space-y-3">
            {rows.map((m) => (
              <li key={m.ID} className="rounded-lg border border-border bg-card p-5 transition hover:border-skip-glow">
                <div className="flex items-center justify-between gap-4">
                  <Link to={`/app/m/${m.ID}`} className="min-w-0 flex-1">
                    <div className="flex items-center gap-2">
                      <span className={`inline-flex items-center gap-1.5 text-sm font-semibold ${STATUS_META[m.Status].color}`}>
                        {STATUS_META[m.Status].icon} {STATUS_META[m.Status].label}
                      </span>
                      <span className="truncate text-sm text-muted-foreground">{m.SourceZipPath.split('/').pop()}</span>
                    </div>
                    <p className="mt-1 text-xs text-muted-foreground">
                      {new Date(m.CreatedAt).toLocaleString('pt-BR')}
                      {m.PixelPerfect && ' · pixel-perfect'}
                      {m.Validate && ' · validada'}
                    </p>
                  </Link>
                  {m.Status === 'success' && m.OutputZipPath && (
                    <a
                      href={downloadURL(m.ID)}
                      className="inline-flex items-center gap-2 rounded-md border border-border bg-card px-3 py-1.5 text-sm hover:border-skip-glow hover:bg-secondary"
                    >
                      <Download className="size-4" /> Baixar
                    </a>
                  )}
                </div>
              </li>
            ))}
          </ul>
        )}
      </main>
    </div>
  )
}
