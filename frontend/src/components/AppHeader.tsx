import { Link, useNavigate } from 'react-router-dom'
import { LogOut, Plus, Shield } from 'lucide-react'
import { supabase } from '@/lib/supabase'
import { useMe } from '@/hooks/useMe'

interface Props {
  /** Show the "Nova migração" + sign-out controls (true on authenticated pages). */
  showAppNav?: boolean
}

/**
 * Reusable top bar. Renders the Skip logo on every page (including /login)
 * and adapts the right-side controls based on auth state and role.
 */
export default function AppHeader({ showAppNav = false }: Props) {
  const me = useMe()
  const navigate = useNavigate()

  return (
    <header className="container flex items-center justify-between py-6">
      <Link to="/" className="flex items-center gap-2 font-display text-lg font-bold">
        <img src="/logoskip.png" alt="Skip" className="size-10 rounded-xl" />
        Skip Migrator
      </Link>
      <div className="flex items-center gap-3">
        {showAppNav && (
          <>
            {me?.is_super_admin && (
              <Link
                to="/admin/users"
                className="inline-flex items-center gap-1.5 rounded-md border border-border bg-card px-4 py-2 text-sm font-semibold text-foreground transition hover:border-skip-glow"
              >
                <Shield className="size-4" /> Admin
              </Link>
            )}
            <Link
              to="/app/new"
              className="inline-flex items-center gap-2 rounded-md bg-primary px-4 py-2 text-sm font-semibold text-primary-foreground shadow-glow transition hover:opacity-95"
            >
              <Plus className="size-4" /> Nova migração
            </Link>
            <button
              onClick={async () => { await supabase.auth.signOut(); navigate('/') }}
              className="rounded-md border border-border bg-card p-2.5 text-muted-foreground transition hover:border-skip-glow hover:text-foreground"
              aria-label="Sair"
            >
              <LogOut className="size-4" />
            </button>
          </>
        )}
        {!showAppNav && !me && (
          <Link
            to="/login"
            className="rounded-md bg-primary px-5 py-2 text-sm font-semibold text-primary-foreground shadow-glow transition hover:opacity-95"
          >
            Entrar
          </Link>
        )}
        {!showAppNav && me && (
          <Link
            to="/app"
            className="rounded-md bg-primary px-5 py-2 text-sm font-semibold text-primary-foreground shadow-glow transition hover:opacity-95"
          >
            Abrir painel
          </Link>
        )}
      </div>
    </header>
  )
}
