import { ExternalLink, CheckCircle2, AlertCircle, Database } from 'lucide-react'

interface Props {
  projectRef: string
}

/**
 * Renderiza o checklist pós-migração contextualizado pro projeto Supabase
 * do cliente. Mostrado após o ZIP de saída ficar pronto.
 *
 * Premissa: o cliente vai conectar a app migrada no MESMO projeto Supabase
 * que o Lovable já usava (schema, RLS, edge functions já estão lá).
 * Logo, na maior parte dos casos é plug-and-play. As exceções aparecem só
 * se ele trocar o domínio público.
 */
export default function SupabaseChecklist({ projectRef }: Props) {
  const base = `https://supabase.com/dashboard/project/${projectRef}`

  return (
    <section className="mt-6 rounded-2xl border border-border bg-card p-6 shadow-soft">
      <div className="mb-4 flex items-center gap-3">
        <div className="grid size-10 place-items-center rounded-xl bg-accent text-primary">
          <Database className="size-5" />
        </div>
        <div>
          <h2 className="font-display text-lg font-bold">Checklist Supabase</h2>
          <p className="text-sm text-muted-foreground">
            Projeto <span className="font-mono">{projectRef.slice(0, 8)}…</span> · plug-and-play na maioria dos casos
          </p>
        </div>
      </div>

      <div className="rounded-xl border border-emerald-200 bg-emerald-50 p-4 dark:border-emerald-900 dark:bg-emerald-950/30">
        <div className="flex items-start gap-2">
          <CheckCircle2 className="mt-0.5 size-4 shrink-0 text-emerald-600" />
          <div className="flex-1 text-sm">
            <p className="font-semibold text-emerald-900 dark:text-emerald-100">Não precisa mudar nada se você usar o mesmo Supabase</p>
            <ul className="mt-2 grid gap-1 text-emerald-800 dark:text-emerald-200">
              <li>• Schema, RLS, triggers, functions, enums — já valem</li>
              <li>• Storage buckets — já existem</li>
              <li>• Edge functions — já deployadas pelo Lovable</li>
              <li>• Sessões e JWTs ativos — continuam válidos</li>
            </ul>
          </div>
        </div>
      </div>

      <div className="mt-4 rounded-xl border border-amber-200 bg-amber-50 p-4 dark:border-amber-900 dark:bg-amber-950/30">
        <div className="flex items-start gap-2">
          <AlertCircle className="mt-0.5 size-4 shrink-0 text-amber-600" />
          <div className="flex-1 text-sm">
            <p className="font-semibold text-amber-900 dark:text-amber-100">
              Verifique APENAS se mudar o domínio público (ex: de <code className="rounded bg-amber-100 px-1.5 py-0.5 font-mono text-xs dark:bg-amber-900">*.lovable.app</code> para um próprio)
            </p>
            <ul className="mt-3 grid gap-2 text-amber-900 dark:text-amber-100">
              <ChecklistLink href={`${base}/auth/url-configuration`} title="Site URL" subtitle="E-mails de reset/confirmação apontam pra cá" />
              <ChecklistLink href={`${base}/auth/url-configuration`} title="Redirect URLs" subtitle="Adicione a nova URL pra OAuth, magic link, password reset funcionarem" />
              <ChecklistLink href={`${base}/auth/providers`} title="OAuth callbacks (Google, GitHub, etc)" subtitle="Cada provider tem callback URL que precisa bater com a nova" />
              <ChecklistLink href={`${base}/settings/api`} title="CORS allowed origins" subtitle="Só se você já tiver restringido (default permite tudo)" />
            </ul>
          </div>
        </div>
      </div>

      <div className="mt-4 rounded-xl border border-border bg-background p-4 text-sm">
        <p className="mb-2 font-semibold">Regenerar <code className="rounded bg-muted px-1.5 py-0.5 font-mono text-xs">types.ts</code> (só se alterar schema no futuro)</p>
        <pre className="overflow-x-auto rounded-lg bg-muted/50 p-3 text-xs">{`npx supabase gen types typescript --project-id ${projectRef} > src/integrations/supabase/types.ts`}</pre>
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
        className="group flex items-start gap-2 rounded-lg bg-amber-100/60 p-2.5 transition hover:bg-amber-100 dark:bg-amber-900/30 dark:hover:bg-amber-900/50"
      >
        <ExternalLink className="mt-1 size-3.5 shrink-0 opacity-60 group-hover:opacity-100" />
        <div className="flex-1">
          <div className="font-semibold">{title}</div>
          <div className="text-xs opacity-75">{subtitle}</div>
        </div>
      </a>
    </li>
  )
}
