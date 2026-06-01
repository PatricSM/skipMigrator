import { Link, useNavigate } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { useEffect } from 'react'
import { Package, Plus, LogOut, CircleCheck, CircleAlert, Loader2, Clock, Download } from 'lucide-react'
import { listMigrations, downloadURL, type Migration } from '@/lib/api'
import { supabase } from '@/lib/supabase'

const STATUS_META: Record<Migration['Status'], { label: string; color: string; icon: React.ReactNode }> = {
  queued:  { label: 'Na fila',      color: 'text-muted-foreground',     icon: <Clock className="size-4" /> },
  running: { label: 'Processando',  color: 'text-primary',              icon: <Loader2 className="size-4 animate-spin" /> },
  success: { label: 'Concluída',    color: 'text-emerald-600',          icon: <CircleCheck className="size-4" /> },
  failed:  { label: 'Falhou',       color: 'text-destructive',          icon: <CircleAlert className="size-4" /> },
}

export default function Dashboard() {
  const navigate = useNavigate()
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
    <div className="min-h-screen bg-gradient-subtle">
      <header className="container flex items-center justify-between py-6">
        <Link to="/" className="flex items-center gap-2 font-display text-lg font-bold">
          <span className="inline-flex size-9 items-center justify-center rounded-full bg-primary text-primary-foreground shadow-card">
            <Package className="size-5" />
          </span>
          Skip Migrator
        </Link>
        <div className="flex items-center gap-3">
          <Link
            to="/app/new"
            className="inline-flex items-center gap-2 rounded-full bg-primary px-5 py-2 text-sm font-semibold text-primary-foreground shadow-card transition hover:opacity-95"
          >
            <Plus className="size-4" /> Nova migração
          </Link>
          <button
            onClick={async () => { await supabase.auth.signOut(); navigate('/') }}
            className="rounded-full border border-border bg-card p-2.5 text-muted-foreground shadow-soft transition hover:text-foreground"
            aria-label="Sair"
          >
            <LogOut className="size-4" />
          </button>
        </div>
      </header>

      <main className="container py-8">
        <h1 className="mb-6 font-display text-2xl font-bold">Suas migrações</h1>
        {rows.length === 0 ? (
          <div className="rounded-2xl border border-dashed border-border bg-card p-12 text-center shadow-soft">
            <p className="text-muted-foreground">Nenhuma migração por aqui ainda.</p>
            <Link
              to="/app/new"
              className="mt-4 inline-flex items-center gap-2 rounded-full bg-primary px-5 py-2 text-sm font-semibold text-primary-foreground shadow-card transition hover:opacity-95"
            >
              <Plus className="size-4" /> Iniciar primeira migração
            </Link>
          </div>
        ) : (
          <ul className="space-y-3">
            {rows.map((m) => (
              <li key={m.ID} className="rounded-2xl border border-border bg-card p-5 shadow-soft transition hover:shadow-card">
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
                      className="inline-flex items-center gap-2 rounded-full border border-border bg-card px-4 py-2 text-sm font-semibold shadow-soft transition hover:bg-accent"
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
