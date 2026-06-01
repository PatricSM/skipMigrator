import { useState } from 'react'
import { Navigate } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import { Copy, Trash2, UserPlus, Loader2, Shield, ShieldOff, User } from 'lucide-react'
import { adminListUsers, adminCreateUser, adminDeleteUser, adminUpdateRole, type UserRole } from '@/lib/api'
import { useMe } from '@/hooks/useMe'
import AppHeader from '@/components/AppHeader'

export default function AdminUsers() {
  const me = useMe()
  const qc = useQueryClient()
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [role, setRole] = useState<UserRole>('user')
  const [lastCreated, setLastCreated] = useState<{ email: string; password: string; role: UserRole } | null>(null)

  const users = useQuery({ queryKey: ['admin-users'], queryFn: adminListUsers, enabled: me?.is_super_admin })

  const createMut = useMutation({
    mutationFn: () => adminCreateUser(email, password || undefined, role),
    onSuccess: (data) => {
      toast.success(`Conta criada para ${data.user.email}`)
      setLastCreated({ email: data.user.email, password: data.generated_password, role: data.user.role })
      setEmail('')
      setPassword('')
      setRole('user')
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

  const roleMut = useMutation({
    mutationFn: ({ id, role }: { id: string; role: UserRole }) => adminUpdateRole(id, role),
    onSuccess: (_d, vars) => {
      toast.success(vars.role === 'super_admin' ? 'Promovido a super admin' : 'Rebaixado para usuário')
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
              Acesso por convite. Compartilhe credenciais por canal seguro. Super admins podem criar contas e promover outros.
            </p>
          </div>
        </div>

        {/* Create form */}
        <form
          onSubmit={(e) => { e.preventDefault(); createMut.mutate() }}
          data-tour="create-user-form"
          className="mb-6 grid gap-3 rounded-lg border border-border bg-card p-6 md:grid-cols-[2fr_1fr_1fr_auto]"
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
          <div>
            <label className="mb-1 block text-xs font-semibold uppercase tracking-wider text-muted-foreground">
              Papel
            </label>
            <select
              value={role}
              onChange={(e) => setRole(e.target.value as UserRole)}
              className="w-full rounded-md border border-input bg-input px-3 py-2.5 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
            >
              <option value="user">Usuário</option>
              <option value="super_admin">Super admin</option>
            </select>
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

        {/* Last-created callout */}
        {lastCreated && (
          <div className="mb-6 rounded-lg border border-warning/30 bg-warning/10 p-4 text-sm">
            <p className="font-semibold text-foreground">
              Conta criada para <span className="font-mono">{lastCreated.email}</span> como{' '}
              <RoleBadge role={lastCreated.role} />
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
              {users.data.map((u) => {
                const isSelf = u.email === me.email
                return (
                  <li key={u.id} className="flex items-center justify-between gap-4 p-4">
                    <div className="min-w-0 flex-1">
                      <div className="flex items-center gap-2">
                        <p className="truncate text-sm font-medium">{u.email}</p>
                        <RoleBadge role={u.role} />
                      </div>
                      <p className="text-xs text-muted-foreground">
                        Criado {new Date(u.created_at).toLocaleString('pt-BR')}
                        {u.last_sign_in_at && ` · último login ${new Date(u.last_sign_in_at).toLocaleString('pt-BR')}`}
                      </p>
                    </div>
                    <div className="flex items-center gap-2">
                      {!isSelf && (
                        u.role === 'super_admin' ? (
                          <button
                            onClick={() => {
                              if (confirm(`Rebaixar ${u.email} para usuário comum?`)) roleMut.mutate({ id: u.id, role: 'user' })
                            }}
                            disabled={roleMut.isPending}
                            className="inline-flex items-center gap-1.5 rounded-md border border-border bg-card px-3 py-1.5 text-xs font-medium text-muted-foreground transition hover:border-skip-glow hover:text-foreground disabled:opacity-50"
                          >
                            <ShieldOff className="size-3.5" /> Rebaixar
                          </button>
                        ) : (
                          <button
                            onClick={() => {
                              if (confirm(`Promover ${u.email} a super admin?`)) roleMut.mutate({ id: u.id, role: 'super_admin' })
                            }}
                            disabled={roleMut.isPending}
                            className="inline-flex items-center gap-1.5 rounded-md border border-border bg-card px-3 py-1.5 text-xs font-medium text-primary transition hover:border-skip-glow disabled:opacity-50"
                          >
                            <Shield className="size-3.5" /> Promover
                          </button>
                        )
                      )}
                      <button
                        onClick={() => {
                          if (isSelf) { toast.error('Não dá pra remover a própria conta.'); return }
                          if (confirm(`Remover ${u.email}?`)) deleteMut.mutate(u.id)
                        }}
                        disabled={deleteMut.isPending}
                        className="rounded-md border border-border p-2 text-muted-foreground transition hover:border-destructive hover:text-destructive disabled:opacity-50"
                        aria-label="Remover"
                      >
                        <Trash2 className="size-4" />
                      </button>
                    </div>
                  </li>
                )
              })}
            </ul>
          )}
        </div>

        <p className="mt-4 text-xs text-muted-foreground">
          <strong>Nota:</strong> mudanças de papel só refletem no JWT após o usuário fazer logout e login novamente.
        </p>
      </main>
    </div>
  )
}

function RoleBadge({ role }: { role: UserRole }) {
  if (role === 'super_admin') {
    return (
      <span className="inline-flex items-center gap-1 rounded-full bg-primary/15 px-2 py-0.5 text-xs font-semibold text-primary">
        <Shield className="size-3" /> Super admin
      </span>
    )
  }
  return (
    <span className="inline-flex items-center gap-1 rounded-full bg-secondary px-2 py-0.5 text-xs font-medium text-muted-foreground">
      <User className="size-3" /> Usuário
    </span>
  )
}
