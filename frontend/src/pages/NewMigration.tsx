import { useState } from 'react'
import { useNavigate, Link } from 'react-router-dom'
import { toast } from 'sonner'
import { ArrowLeft, Upload, Package, Loader2 } from 'lucide-react'
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
      toast.error('Select a ZIP file first')
      return
    }
    setBusy(true)
    try {
      const r = await createMigration(file, { pixelPerfect, validate })
      toast.success('Migration queued')
      navigate(`/app/m/${r.id}`)
    } catch (err: unknown) {
      toast.error(err instanceof Error ? err.message : 'upload failed')
    } finally {
      setBusy(false)
    }
  }

  return (
    <div className="min-h-screen">
      <header className="container flex items-center justify-between py-6">
        <Link to="/app" className="flex items-center gap-2 text-sm text-muted-foreground hover:text-foreground">
          <ArrowLeft className="size-4" /> Back to dashboard
        </Link>
        <Link to="/" className="flex items-center gap-2 text-lg font-bold">
          <Package className="size-5 text-primary" /> Skip Migrator
        </Link>
      </header>

      <main className="container max-w-2xl py-8">
        <h1 className="mb-1 text-2xl font-bold">New migration</h1>
        <p className="mb-8 text-sm text-muted-foreground">
          Upload a Lovable project ZIP. Max 100 MB. Make sure node_modules and dist are excluded.
        </p>

        <form onSubmit={submit} className="space-y-6 rounded-xl border border-border bg-card p-6">
          <label className="block">
            <span className="mb-2 block text-sm font-medium">Project ZIP</span>
            <div className="relative grid place-items-center rounded-lg border-2 border-dashed border-border p-12 transition-colors hover:bg-accent">
              <input
                type="file"
                accept=".zip,application/zip"
                onChange={(e) => setFile(e.target.files?.[0] ?? null)}
                className="absolute inset-0 cursor-pointer opacity-0"
                required
              />
              <Upload className="size-8 text-muted-foreground" />
              <p className="mt-3 text-sm font-medium">
                {file ? file.name : 'Drop ZIP here or click to browse'}
              </p>
              {file && (
                <p className="mt-1 text-xs text-muted-foreground">
                  {(file.size / 1024 / 1024).toFixed(1)} MB
                </p>
              )}
            </div>
          </label>

          <div className="space-y-3">
            <label className="flex items-start gap-3 rounded-lg border border-border p-3">
              <input
                type="checkbox"
                checked={validate}
                onChange={(e) => setValidate(e.target.checked)}
                className="mt-1 size-4 accent-primary"
              />
              <div>
                <div className="font-medium">Validate build</div>
                <p className="text-sm text-muted-foreground">
                  Run <code className="rounded bg-muted px-1.5 py-0.5 text-xs">pnpm install + tsc + pnpm build</code> before
                  returning the ZIP. Adds ~3 minutes but guarantees the output works.
                </p>
              </div>
            </label>

            <label className="flex items-start gap-3 rounded-lg border border-border p-3">
              <input
                type="checkbox"
                checked={pixelPerfect}
                onChange={(e) => setPixelPerfect(e.target.checked)}
                className="mt-1 size-4 accent-primary"
              />
              <div>
                <div className="font-medium">Pixel-perfect</div>
                <p className="text-sm text-muted-foreground">
                  Override Skip's <code className="rounded bg-muted px-1.5 py-0.5 text-xs">components/ui/*</code> with
                  your project's versions (adapts calendar v8→v9). ~99% identical to original render.
                </p>
              </div>
            </label>
          </div>

          <button
            type="submit"
            disabled={busy || !file}
            className="inline-flex w-full items-center justify-center gap-2 rounded-lg bg-primary px-4 py-3 font-semibold text-primary-foreground hover:opacity-90 disabled:opacity-50"
          >
            {busy ? <Loader2 className="size-4 animate-spin" /> : <Upload className="size-4" />}
            {busy ? 'Uploading…' : 'Start migration'}
          </button>
        </form>
      </main>
    </div>
  )
}
