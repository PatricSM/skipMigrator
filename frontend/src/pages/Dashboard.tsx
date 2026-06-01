import { Link, useNavigate } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { useEffect } from 'react'
import { Package, Plus, LogOut, CircleCheck, CircleAlert, Loader2, Clock, Download } from 'lucide-react'
import { listMigrations, downloadURL, type Migration } from '@/lib/api'
import { supabase } from '@/lib/supabase'

const STATUS_META: Record<Migration['Status'], { label: string; color: string; icon: React.ReactNode }> = {
  queued:  { label: 'Queued',  color: 'text-muted-foreground', icon: <Clock className="size-4" /> },
  running: { label: 'Running', color: 'text-primary',          icon: <Loader2 className="size-4 animate-spin" /> },
  success: { label: 'Success', color: 'text-emerald-400',      icon: <CircleCheck className="size-4" /> },
  failed:  { label: 'Failed',  color: 'text-destructive',      icon: <CircleAlert className="size-4" /> },
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

  // Realtime subscription for fast updates
  useEffect(() => {
    const ch = supabase
      .channel('migrations-changes')
      .on('postgres_changes', { event: '*', schema: 'public', table: 'migrations' }, () => refetch())
      .subscribe()
    return () => { void supabase.removeChannel(ch) }
  }, [refetch])

  return (
    <div className="min-h-screen">
      <header className="container flex items-center justify-between py-6">
        <Link to="/" className="flex items-center gap-2 text-lg font-bold">
          <Package className="size-5 text-primary" />
          Skip Migrator
        </Link>
        <div className="flex items-center gap-3">
          <Link
            to="/app/new"
            className="inline-flex items-center gap-2 rounded-lg bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:opacity-90"
          >
            <Plus className="size-4" /> New migration
          </Link>
          <button
            onClick={async () => { await supabase.auth.signOut(); navigate('/') }}
            className="rounded-lg border border-border p-2 text-muted-foreground hover:text-foreground"
            aria-label="Sign out"
          >
            <LogOut className="size-4" />
          </button>
        </div>
      </header>

      <main className="container py-8">
        <h1 className="mb-6 text-2xl font-bold">Your migrations</h1>
        {rows.length === 0 ? (
          <div className="rounded-xl border border-dashed border-border p-12 text-center">
            <p className="text-muted-foreground">No migrations yet.</p>
            <Link
              to="/app/new"
              className="mt-4 inline-flex items-center gap-2 rounded-lg bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:opacity-90"
            >
              <Plus className="size-4" /> Start your first migration
            </Link>
          </div>
        ) : (
          <ul className="space-y-3">
            {rows.map((m) => (
              <li key={m.ID} className="rounded-xl border border-border bg-card p-4">
                <div className="flex items-center justify-between gap-4">
                  <Link to={`/app/m/${m.ID}`} className="flex-1 min-w-0">
                    <div className="flex items-center gap-2">
                      <span className={`inline-flex items-center gap-1 text-sm font-medium ${STATUS_META[m.Status].color}`}>
                        {STATUS_META[m.Status].icon} {STATUS_META[m.Status].label}
                      </span>
                      <span className="text-sm text-muted-foreground truncate">{m.SourceZipPath.split('/').pop()}</span>
                    </div>
                    <p className="mt-1 text-xs text-muted-foreground">
                      {new Date(m.CreatedAt).toLocaleString()}
                      {m.PixelPerfect && ' · pixel-perfect'}
                      {m.Validate && ' · validated'}
                    </p>
                  </Link>
                  {m.Status === 'success' && m.OutputZipPath && (
                    <a
                      href={downloadURL(m.ID)}
                      className="inline-flex items-center gap-2 rounded-lg border border-border px-3 py-1.5 text-sm hover:bg-accent"
                    >
                      <Download className="size-4" /> Download
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
