import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { toast } from 'sonner'
import { supabase } from '@/lib/supabase'
import AppHeader from '@/components/AppHeader'

export default function Login() {
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [busy, setBusy] = useState(false)
  const navigate = useNavigate()

  const submit = async (e: React.FormEvent) => {
    e.preventDefault()
    setBusy(true)
    try {
      const { error } = await supabase.auth.signInWithPassword({ email, password })
      if (error) throw error
      navigate('/app')
    } catch (err: unknown) {
      toast.error(err instanceof Error ? err.message : 'Falha na autenticação')
    } finally {
      setBusy(false)
    }
  }

  return (
    <div className="min-h-screen">
      <AppHeader />
      <main className="grid place-items-center px-4 py-16">
        <div className="w-full max-w-sm rounded-lg border border-skip-glow bg-skip-glow p-8">
          <h1 className="mb-1 font-display text-2xl font-bold">Entrar</h1>
          <p className="mb-6 text-sm text-muted-foreground">
            Acesse seu histórico de migrações.
          </p>
          <form onSubmit={submit} className="space-y-4">
            <div>
              <label className="mb-1 block text-sm font-medium">E-mail</label>
              <input
                type="email"
                required
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                className="w-full rounded-md border border-input bg-input px-3 py-2.5 text-sm text-foreground focus:outline-none focus:ring-2 focus:ring-ring"
                placeholder="voce@email.com"
              />
            </div>
            <div>
              <label className="mb-1 block text-sm font-medium">Senha</label>
              <input
                type="password"
                required
                minLength={8}
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                className="w-full rounded-md border border-input bg-input px-3 py-2.5 text-sm text-foreground focus:outline-none focus:ring-2 focus:ring-ring"
                placeholder="Mínimo 8 caracteres"
              />
            </div>
            <button
              type="submit"
              disabled={busy}
              className="w-full rounded-md bg-primary px-4 py-2.5 font-semibold text-primary-foreground shadow-glow transition hover:opacity-95 disabled:opacity-50"
            >
              {busy ? 'Aguarde…' : 'Entrar'}
            </button>
          </form>
          <p className="mt-4 text-center text-xs text-muted-foreground">
            Acesso por convite — peça suas credenciais ao administrador.
          </p>
        </div>
      </main>
    </div>
  )
}
