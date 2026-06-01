import { useState } from 'react'
import { useNavigate, Link } from 'react-router-dom'
import { toast } from 'sonner'
import { ArrowLeft, Upload, Loader2 } from 'lucide-react'
import { createMigration } from '@/lib/api'

export default function NewMigration() {
  const [file, setFile] = useState<File | null>(null)
  const [pixelPerfect, setPixelPerfect] = useState(false)
  const [validate, setValidate] = useState(true)
  const [busy, setBusy] = useState(false)
  const navigate = useNavigate()

  const submit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!file) {
      toast.error('Selecione um arquivo ZIP primeiro')
      return
    }
    setBusy(true)
    try {
      const r = await createMigration(file, { pixelPerfect, validate })
      toast.success('Migração na fila')
      navigate(`/app/m/${r.id}`)
    } catch (err: unknown) {
      toast.error(err instanceof Error ? err.message : 'Falha no upload')
    } finally {
      setBusy(false)
    }
  }

  return (
    <div className="min-h-screen bg-gradient-subtle">
      <header className="container flex items-center justify-between py-6">
        <Link to="/app" className="flex items-center gap-2 text-sm text-muted-foreground hover:text-foreground">
          <ArrowLeft className="size-4" /> Voltar para o painel
        </Link>
        <Link to="/" className="flex items-center gap-2 font-display text-lg font-bold">
          <img src="/logoskip.png" alt="Skip" className="size-10 rounded-xl shadow-card" />
          Skip Migrator
        </Link>
      </header>

      <main className="container max-w-2xl py-8">
        <h1 className="mb-1 font-display text-2xl font-bold">Nova migração</h1>
        <p className="mb-8 text-sm text-muted-foreground">
          Envie o ZIP do seu projeto Lovable. Máximo de 100 MB. Garanta que <code className="rounded bg-muted px-1.5 py-0.5 font-mono text-xs">node_modules</code> e <code className="rounded bg-muted px-1.5 py-0.5 font-mono text-xs">dist</code> estão excluídos.
        </p>

        <form onSubmit={submit} className="space-y-6 rounded-2xl border border-border bg-card p-6 shadow-soft">
          <label className="block">
            <span className="mb-2 block text-sm font-semibold">Arquivo ZIP do projeto</span>
            <div className="relative grid place-items-center rounded-xl border-2 border-dashed border-border bg-background p-12 transition-colors hover:bg-accent">
              <input
                type="file"
                accept=".zip,application/zip"
                onChange={(e) => setFile(e.target.files?.[0] ?? null)}
                className="absolute inset-0 cursor-pointer opacity-0"
                required
              />
              <Upload className="size-8 text-muted-foreground" />
              <p className="mt-3 text-sm font-semibold">
                {file ? file.name : 'Arraste o ZIP aqui ou clique para selecionar'}
              </p>
              {file && (
                <p className="mt-1 text-xs text-muted-foreground">
                  {(file.size / 1024 / 1024).toFixed(1)} MB
                </p>
              )}
            </div>
          </label>

          <div className="space-y-3">
            <label className="flex items-start gap-3 rounded-xl border border-border bg-background p-4">
              <input
                type="checkbox"
                checked={validate}
                onChange={(e) => setValidate(e.target.checked)}
                className="mt-1 size-4 accent-primary"
              />
              <div>
                <div className="font-semibold">Validar build</div>
                <p className="text-sm text-muted-foreground">
                  Roda <code className="rounded bg-muted px-1.5 py-0.5 font-mono text-xs">pnpm install + tsc + pnpm build</code> antes de devolver o ZIP. Adiciona ~3 minutos mas garante que o resultado funciona.
                </p>
              </div>
            </label>

            <label className="flex items-start gap-3 rounded-xl border border-border bg-background p-4">
              <input
                type="checkbox"
                checked={pixelPerfect}
                onChange={(e) => setPixelPerfect(e.target.checked)}
                className="mt-1 size-4 accent-primary"
              />
              <div>
                <div className="font-semibold">Pixel-perfect</div>
                <p className="text-sm text-muted-foreground">
                  Sobrescreve <code className="rounded bg-muted px-1.5 py-0.5 font-mono text-xs">components/ui/*</code> do Skip com as versões do seu projeto (adapta calendar v8→v9). Saída ~99% idêntica ao original.
                </p>
              </div>
            </label>
          </div>

          <button
            type="submit"
            disabled={busy || !file}
            className="inline-flex w-full items-center justify-center gap-2 rounded-xl bg-primary px-4 py-3.5 font-semibold text-primary-foreground shadow-card transition hover:opacity-95 disabled:opacity-50"
          >
            {busy ? <Loader2 className="size-4 animate-spin" /> : <Upload className="size-4" />}
            {busy ? 'Enviando…' : 'Iniciar migração'}
          </button>
        </form>
      </main>
    </div>
  )
}
