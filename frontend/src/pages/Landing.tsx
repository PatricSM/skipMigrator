import { Link } from 'react-router-dom'
import { ArrowRight, Zap, ShieldCheck, GitBranch, CheckCircle2 } from 'lucide-react'

export default function Landing({ authed }: { authed: boolean }) {
  return (
    <div className="min-h-screen bg-gradient-subtle">
      <header className="container flex items-center justify-between py-6">
        <div className="flex items-center gap-2 font-display text-xl font-bold text-foreground">
          <img src="/logoskip.png" alt="Skip" className="size-10 rounded-xl shadow-card" />
          Skip Migrator
        </div>
        <nav className="flex items-center gap-6 text-sm">
          <a href="#recursos" className="text-muted-foreground hover:text-foreground">Recursos</a>
          <a href="#como-funciona" className="text-muted-foreground hover:text-foreground">Como funciona</a>
          <Link
            to={authed ? '/app' : '/login'}
            className="rounded-full bg-primary px-5 py-2 font-semibold text-primary-foreground shadow-card transition hover:opacity-90"
          >
            {authed ? 'Abrir painel' : 'Entrar'}
          </Link>
        </nav>
      </header>

      <section className="container relative overflow-hidden py-24 text-center">
        <div className="pointer-events-none absolute inset-x-0 top-12 mx-auto h-72 max-w-3xl rounded-full bg-primary/10 blur-3xl" />
        <div className="relative">
          <span className="mb-6 inline-flex items-center gap-2 rounded-full border border-border bg-card px-4 py-1.5 text-xs font-semibold uppercase tracking-wider text-primary shadow-soft">
            <CheckCircle2 className="size-3.5" /> Migração validada com pnpm build
          </span>
          <h1 className="mx-auto max-w-4xl font-display text-5xl font-extrabold leading-tight tracking-tight text-foreground md:text-6xl">
            Do <span className="text-gradient-gold">Lovable</span> para o <span className="text-gradient-gold">Skip</span>
            <br />
            em um único upload.
          </h1>
          <p className="mx-auto mt-6 max-w-2xl text-lg text-muted-foreground">
            Envie o ZIP do seu projeto Lovable. Migramos páginas, componentes, hooks, migrations e edge
            functions do Supabase para a stack moderna do Skip — e rodamos{' '}
            <code className="rounded bg-muted px-2 py-0.5 font-mono text-sm">pnpm build</code> antes de
            devolver o ZIP pronto.
          </p>
          <div className="mt-10 flex items-center justify-center gap-4">
            <Link
              to={authed ? '/app/new' : '/login'}
              className="inline-flex items-center gap-2 rounded-full bg-primary px-7 py-3.5 font-semibold text-primary-foreground shadow-card transition hover:translate-y-[-1px] hover:opacity-95"
            >
              Começar migração <ArrowRight className="size-4" />
            </Link>
            <a
              href="https://github.com/PatricSM/skipMigrator"
              className="rounded-full border border-border bg-card px-7 py-3.5 font-semibold text-foreground shadow-soft transition hover:bg-accent"
            >
              Ver no GitHub
            </a>
          </div>
        </div>
      </section>

      <section id="recursos" className="container py-16">
        <div className="grid gap-6 md:grid-cols-3">
          <Feature
            icon={<Zap className="size-5" />}
            title="~30 segundos de ponta a ponta"
            body="Transformações + instalação de dependências + build, tudo em um único job. A maioria dos projetos termina antes do seu café esfriar."
          />
          <Feature
            icon={<ShieldCheck className="size-5" />}
            title="Saída validada pelo build"
            body="Rodamos pnpm install, tsc --noEmit e pnpm build. Se algo quebrar, você recebe o log completo em vez de um ZIP quebrado."
          />
          <Feature
            icon={<GitBranch className="size-5" />}
            title="13 fases determinísticas"
            body="Mesma lógica da nossa migração de referência Lovable → Skip. Zod 3→4, useRef do React 19, calendar v8→v9, tudo já tratado."
          />
        </div>
      </section>

      <section id="como-funciona" className="container py-16">
        <h2 className="mb-12 text-center font-display text-3xl font-bold">Como funciona</h2>
        <ol className="mx-auto grid max-w-3xl gap-4">
          <Step n={1} title="Envie o ZIP do projeto Lovable" body="Cadastre-se com e-mail, arraste o ZIP (até 100 MB)." />
          <Step n={2} title="Escolha as opções" body="Pixel-perfect? Validar build? Os padrões já cobrem 90% dos casos." />
          <Step n={3} title="Nós migramos + validamos" body="As 13 fases rodam em um worker isolado. O ZIP final é buildado e pronto para deploy." />
          <Step n={4} title="Baixe o resultado" body="Link assinado para o ZIP de saída, com o log completo do build em anexo." />
        </ol>
      </section>

      <footer className="container py-12 text-center text-sm text-muted-foreground">
        © {new Date().getFullYear()} Skip Migrator · Licença MIT
      </footer>
    </div>
  )
}

function Feature({ icon, title, body }: { icon: React.ReactNode; title: string; body: string }) {
  return (
    <div className="rounded-2xl border border-border bg-card p-6 shadow-soft transition hover:shadow-card">
      <div className="mb-4 inline-flex items-center justify-center rounded-xl bg-accent p-2.5 text-primary">{icon}</div>
      <h3 className="mb-2 font-display text-lg font-semibold">{title}</h3>
      <p className="text-sm leading-relaxed text-muted-foreground">{body}</p>
    </div>
  )
}

function Step({ n, title, body }: { n: number; title: string; body: string }) {
  return (
    <li className="flex gap-5 rounded-2xl border border-border bg-card p-6 shadow-soft transition hover:shadow-card">
      <div className="grid size-11 shrink-0 place-items-center rounded-full bg-primary font-display text-lg font-bold text-primary-foreground">
        {n}
      </div>
      <div>
        <h4 className="font-display text-lg font-semibold">{title}</h4>
        <p className="mt-1 text-sm leading-relaxed text-muted-foreground">{body}</p>
      </div>
    </li>
  )
}
