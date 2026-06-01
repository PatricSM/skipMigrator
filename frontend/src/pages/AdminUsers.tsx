import { useState } from 'react'
import { Navigate } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import { Copy, Trash2, UserPlus, Loader2, Shield } from 'lucide-react'
import { adminListUsers, adminCreateUser, adminDeleteUser } from '@/lib/api'
import { useMe } from '@/hooks/useMe'
import AppHeader from '@/components/AppHeader'

export default function AdminUsers() {
  const me = useMe()
  const qc = useQueryClient()
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [lastCreated, setLastCreated] = useState<{ email: string; password: string } | null>(null)

  const users = useQuery({ queryKey: ['admin-users'], queryFn: adminListUsers, enabled: me?.is_super_admin })

  const createMut = useMutation({
    mutationFn: () => adminCreateUser(email, password || undefined),
    onSuccess: (data) => {
      toast.success(`Conta criada para ${data.user.email}`)
      setLastCreated({ email: data.user.email, password: data.generated_password })
      setEmail('')
      setPassword('')
      qc.invalidateQueries({ queryKey: ['admin-users'] })
    },
    onError: (err: Error) => toast.error(err.message),
  })

  const deleteMut = useMutation({
    mutationFn: (id: string) => adminDeleteUser(id),
    onSuccess: () => {
      toast.success('Usuário removido')
      qc.invalidateQueries({ queryKey: ['admin-users'] })
    },
    onError: (err: Error) => toast.error(err.message),
  })

  if (me === null) return <div className="grid h-screen place-items-center text-muted-foreground">Carregando…</div>
  if (me === undefined) return <Navigate to="/login" replace />
  if (!me.is_super_admin) return <Navigate to="/app" replace />

  return (
    <div className="min-h-screen">
      <AppHeader showAppNav />
      <main className="container max-w-4xl py-8">
        <div className="mb-6 flex items-center gap-3">
          <div className="grid size-10 place-items-center rounded-md bg-primary/15 text-primary">
            <Shield className="size-5" />
          </div>
          <div>
            <h1 className="font-display text-2xl font-bold">Administração de contas</h1>
            <p className="text-sm text-muted-foreground">
              Acesso por convite — só super admins criam contas. Compartilhe credenciais por canal seguro.
            </p>
          </div>
        </div>

        {/* Create form */}
        <form
          onSubmit={(e) => { e.preventDefault(); createMut.mutate() }}
          className="mb-6 grid gap-3 rounded-lg border border-border bg-card p-6 md:grid-cols-[2fr_1fr_auto]"
        >
          <div>
            <label className="mb-1 block text-xs font-semibold uppercase tracking-wider text-muted-foreground">E-mail</label>
            <input
              type="email"
              required
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              placeholder="usuario@empresa.com"
              className="w-full rounded-md border border-input bg-input px-3 py-2.5 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
            />
          </div>
          <div>
            <label className="mb-1 block text-xs font-semibold uppercase tracking-wider text-muted-foreground">
              Senha (opcional)
            </label>
            <input
              type="text"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              placeholder="deixe vazio p/ gerar"
              className="w-full rounded-md border border-input bg-input px-3 py-2.5 font-mono text-sm focus:outline-none focus:ring-2 focus:ring-ring"
            />
          </div>
          <button
            type="submit"
            disabled={createMut.isPending}
            className="inline-flex items-end gap-2 self-end rounded-md bg-primary px-5 py-2.5 text-sm font-semibold text-primary-foreground shadow-glow transition hover:opacity-95 disabled:opacity-50"
          >
            {createMut.isPending ? <Loader2 className="size-4 animate-spin" /> : <UserPlus className="size-4" />}
            Criar conta
          </button>
        </form>

        {/* Last-created callout: surface the generated password ONCE */}
        {lastCreated && (
          <div className="mb-6 rounded-lg border border-warning/30 bg-warning/10 p-4 text-sm">
            <p className="font-semibold text-foreground">
              Conta criada para <span className="font-mono">{lastCreated.email}</span>
            </p>
            <p className="mt-1 text-muted-foreground">
              Senha gerada (mostrada apenas uma vez — copie agora e envie ao usuário):
            </p>
            <div className="mt-2 flex items-center gap-2 rounded-md border border-border bg-card p-2.5">
              <code className="flex-1 font-mono text-sm">{lastCreated.password}</code>
              <button
                onClick={() => { void navigator.clipboard.writeText(lastCreated.password); toast.success('Senha copiada') }}
                className="rounded-md p-1.5 text-muted-foreground hover:bg-secondary hover:text-foreground"
                aria-label="Copiar"
              >
                <Copy className="size-4" />
              </button>
            </div>
          </div>
        )}

        {/* Users list */}
        <div className="rounded-lg border border-border bg-card">
          <div className="border-b border-border p-4 text-sm font-semibold">
            Usuários ({users.data?.length ?? 0})
          </div>
          {users.isLoading && (
            <div className="grid place-items-center p-8 text-muted-foreground">
              <Loader2 className="size-5 animate-spin" />
            </div>
          )}
          {users.data && (
            <ul className="divide-y divide-border">
              {users.data.map((u) => (
                <li key={u.id} className="flex items-center justify-between gap-4 p-4">
                  <div className="min-w-0 flex-1">
                    <p className="truncate text-sm font-medium">{u.email}</p>
                    <p className="text-xs text-muted-foreground">
                      Criado {new Date(u.created_at).toLocaleString('pt-BR')}
                      {u.last_sign_in_at && ` · último login ${new Date(u.last_sign_in_at).toLocaleString('pt-BR')}`}
                    </p>
                  </div>
                  <button
                    onClick={() => {
                      if (u.email === me.email) {
                        toast.error('Não dá pra remover a própria conta.')
                        return
                      }
                      if (confirm(`Remover ${u.email}?`)) deleteMut.mutate(u.id)
                    }}
                    disabled={deleteMut.isPending}
                    className="rounded-md border border-border p-2 text-muted-foreground transition hover:border-destructive hover:text-destructive disabled:opacity-50"
                    aria-label="Remover"
                  >
                    <Trash2 className="size-4" />
                  </button>
                </li>
              ))}
            </ul>
          )}
        </div>
      </main>
    </div>
  )
}
