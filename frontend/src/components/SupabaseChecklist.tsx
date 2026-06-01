import { ExternalLink, CheckCircle2, AlertCircle, Database } from 'lucide-react'

interface Props {
  projectRef: string
}

export default function SupabaseChecklist({ projectRef }: Props) {
  const base = `https://supabase.com/dashboard/project/${projectRef}`

  return (
    <section className="mt-6 rounded-lg border border-border bg-card p-6">
      <div className="mb-4 flex items-center gap-3">
        <div className="grid size-10 place-items-center rounded-md bg-primary/15 text-primary">
          <Database className="size-5" />
        </div>
        <div>
          <h2 className="font-display text-lg font-bold">Checklist Supabase</h2>
          <p className="text-sm text-muted-foreground">
            Projeto <span className="font-mono">{projectRef.slice(0, 8)}…</span> · plug-and-play na maioria dos casos
          </p>
        </div>
      </div>

      <div className="rounded-md border border-success/30 bg-success/10 p-4">
        <div className="flex items-start gap-2">
          <CheckCircle2 className="mt-0.5 size-4 shrink-0 text-success" />
          <div className="flex-1 text-sm">
            <p className="font-semibold text-foreground">Não precisa mudar nada se você usar o mesmo Supabase</p>
            <ul className="mt-2 grid gap-1 text-muted-foreground">
              <li>• Schema, RLS, triggers, functions, enums — já valem</li>
              <li>• Storage buckets — já existem</li>
              <li>• Edge functions — já deployadas pelo Lovable</li>
              <li>• Sessões e JWTs ativos — continuam válidos</li>
            </ul>
          </div>
        </div>
      </div>

      <div className="mt-4 rounded-md border border-warning/30 bg-warning/10 p-4">
        <div className="flex items-start gap-2">
          <AlertCircle className="mt-0.5 size-4 shrink-0 text-warning" />
          <div className="flex-1 text-sm">
            <p className="font-semibold text-foreground">
              Verifique APENAS se mudar o domínio público (ex: de <code className="rounded bg-secondary px-1.5 py-0.5 font-mono text-xs">*.lovable.app</code> para um próprio)
            </p>
            <ul className="mt-3 grid gap-2">
              <ChecklistLink href={`${base}/auth/url-configuration`} title="Site URL" subtitle="E-mails de reset/confirmação apontam pra cá" />
              <ChecklistLink href={`${base}/auth/url-configuration`} title="Redirect URLs" subtitle="Adicione a nova URL pra OAuth, magic link, password reset funcionarem" />
              <ChecklistLink href={`${base}/auth/providers`} title="OAuth callbacks (Google, GitHub, etc)" subtitle="Cada provider tem callback URL que precisa bater com a nova" />
              <ChecklistLink href={`${base}/settings/api`} title="CORS allowed origins" subtitle="Só se você já tiver restringido (default permite tudo)" />
            </ul>
          </div>
        </div>
      </div>

      <div className="mt-4 rounded-md border border-border bg-input p-4 text-sm">
        <p className="mb-2 font-semibold">Regenerar <code className="rounded bg-secondary px-1.5 py-0.5 font-mono text-xs">types.ts</code> (só se alterar schema no futuro)</p>
        <pre className="overflow-x-auto rounded-md bg-background/60 p-3 text-xs text-muted-foreground">{`npx supabase gen types typescript --project-id ${projectRef} > src/integrations/supabase/types.ts`}</pre>
      </div>

      <a
        href={base}
        target="_blank"
        rel="noreferrer"
        className="mt-4 inline-flex items-center gap-2 text-sm font-semibold text-primary hover:underline"
      >
        Abrir painel do projeto <ExternalLink className="size-3.5" />
      </a>
    </section>
  )
}

function ChecklistLink({ href, title, subtitle }: { href: string; title: string; subtitle: string }) {
  return (
    <li>
      <a
        href={href}
        target="_blank"
        rel="noreferrer"
        className="group flex items-start gap-2 rounded-md border border-border bg-card p-2.5 transition hover:border-skip-glow"
      >
        <ExternalLink className="mt-1 size-3.5 shrink-0 text-muted-foreground group-hover:text-foreground" />
        <div className="flex-1">
          <div className="font-semibold text-foreground">{title}</div>
          <div className="text-xs text-muted-foreground">{subtitle}</div>
        </div>
      </a>
    </li>
  )
}
