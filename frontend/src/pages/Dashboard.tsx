import { Link, useNavigate } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { useEffect } from 'react'
import { Plus, CircleCheck, CircleAlert, Loader2, Clock, Download, Upload, Package, FileDown, Sparkles } from 'lucide-react'
import { listMigrations, downloadURL, type Migration } from '@/lib/api'
import { supabase } from '@/lib/supabase'
import AppHeader from '@/components/AppHeader'
import { useMe } from '@/hooks/useMe'
import { startTour } from '@/lib/tour'

const STATUS_META: Record<Migration['Status'], { label: string; color: string; icon: React.ReactNode }> = {
  queued:  { label: 'Na fila',     color: 'text-muted-foreground', icon: <Clock className="size-4" /> },
  running: { label: 'Processando', color: 'text-primary',          icon: <Loader2 className="size-4 animate-spin" /> },
  success: { label: 'Concluída',   color: 'text-success',          icon: <CircleCheck className="size-4" /> },
  failed:  { label: 'Falhou',      color: 'text-destructive',      icon: <CircleAlert className="size-4" /> },
}

export default function Dashboard() {
  const me = useMe()
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
    <div className="min-h-screen">
      <AppHeader showAppNav />
      <main className="container py-8" data-tour="dashboard-main">
        <h1 className="mb-6 font-display text-2xl font-bold">Suas migrações</h1>
        {rows.length === 0 ? (
          <EmptyState
            onStartTour={() =>
              me && startTour({ navigate: (p) => navigate(p), includeAdmin: !!me.is_super_admin })
            }
          />
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

function EmptyState({ onStartTour }: { onStartTour: () => void }) {
  return (
    <section className="rounded-2xl border border-skip-glow bg-skip-glow p-8">
      <div className="flex items-center gap-3">
        <div className="grid size-12 place-items-center rounded-md bg-primary/15 text-primary">
          <Sparkles className="size-6" />
        </div>
        <div>
          <h2 className="font-display text-xl font-bold">Bem-vindo ao Skip Migrator</h2>
          <p className="text-sm text-muted-foreground">
            3 passos para migrar seu primeiro projeto Lovable em ~90 segundos.
          </p>
        </div>
      </div>

      <ol className="mt-6 grid gap-4 md:grid-cols-3">
        <Step
          n={1}
          icon={<Package className="size-5" />}
          title="Exporte do Lovable"
          body="Baixe o ZIP do seu projeto pela UI do Lovable."
        />
        <Step
          n={2}
          icon={<Upload className="size-5" />}
          title="Suba aqui"
          body="Arraste no campo de upload e configure as opções."
        />
        <Step
          n={3}
          icon={<FileDown className="size-5" />}
          title="Baixe o resultado"
          body="Pronto: ZIP migrado para a stack Skip, com build validado."
        />
      </ol>

      <div className="mt-8 flex flex-wrap gap-3">
        <Link
          to="/app/new"
          data-tour="empty-start-btn"
          className="inline-flex items-center gap-2 rounded-md bg-primary px-5 py-2.5 text-sm font-semibold text-primary-foreground shadow-glow transition hover:opacity-95"
        >
          <Plus className="size-4" /> Iniciar primeira migração
        </Link>
        <button
          onClick={onStartTour}
          className="inline-flex items-center gap-2 rounded-md border border-border bg-card px-5 py-2.5 text-sm font-semibold text-foreground transition hover:border-skip-glow hover:bg-secondary"
        >
          <Sparkles className="size-4" /> Fazer tour guiado
        </button>
      </div>
    </section>
  )
}

function Step({ n, icon, title, body }: { n: number; icon: React.ReactNode; title: string; body: string }) {
  return (
    <li className="rounded-md border border-border bg-card p-5">
      <div className="mb-3 flex items-center gap-2">
        <span className="grid size-7 place-items-center rounded-full bg-primary text-xs font-bold text-primary-foreground">
          {n}
        </span>
        <span className="text-primary">{icon}</span>
      </div>
      <h3 className="font-display text-base font-semibold">{title}</h3>
      <p className="mt-1 text-sm text-muted-foreground">{body}</p>
    </li>
  )
}
