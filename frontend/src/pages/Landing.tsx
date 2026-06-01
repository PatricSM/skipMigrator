import { Link } from 'react-router-dom'
import { ArrowRight, Package, Zap, ShieldCheck, GitBranch } from 'lucide-react'

export default function Landing({ authed }: { authed: boolean }) {
  return (
    <div className="min-h-screen">
      <header className="container flex items-center justify-between py-6">
        <div className="flex items-center gap-2 text-xl font-bold">
          <Package className="size-6 text-primary" />
          <span>Skip Migrator</span>
        </div>
        <nav className="flex items-center gap-4 text-sm">
          <a href="#features" className="text-muted-foreground hover:text-foreground">Features</a>
          <a href="#how" className="text-muted-foreground hover:text-foreground">How it works</a>
          <Link
            to={authed ? '/app' : '/login'}
            className="rounded-lg bg-primary px-4 py-2 font-medium text-primary-foreground hover:opacity-90"
          >
            {authed ? 'Open app' : 'Sign in'}
          </Link>
        </nav>
      </header>

      <section className="container py-24 text-center">
        <h1 className="mx-auto max-w-4xl text-5xl font-extrabold tracking-tight md:text-6xl">
          From <span className="text-primary">Lovable</span> to <span className="text-primary">Skip</span>
          <br />
          in one upload.
        </h1>
        <p className="mx-auto mt-6 max-w-2xl text-lg text-muted-foreground">
          Drop your Lovable project ZIP. We migrate pages, components, hooks, Supabase migrations and edge
          functions onto the modern Skip stack — and run <code className="rounded bg-muted px-1.5 py-0.5 text-sm">pnpm build</code> before
          handing you the ZIP back.
        </p>
        <div className="mt-10 flex items-center justify-center gap-4">
          <Link
            to={authed ? '/app/new' : '/login'}
            className="inline-flex items-center gap-2 rounded-lg bg-primary px-6 py-3 font-semibold text-primary-foreground hover:opacity-90"
          >
            Start migrating <ArrowRight className="size-4" />
          </Link>
          <a
            href="https://github.com/PatricSM/skip-migrator"
            className="rounded-lg border border-border px-6 py-3 font-semibold text-foreground hover:bg-accent"
          >
            View on GitHub
          </a>
        </div>
      </section>

      <section id="features" className="container py-16">
        <div className="grid gap-6 md:grid-cols-3">
          <Feature icon={<Zap className="size-5" />} title="~30 seconds end-to-end" body="Transformations + dependency install + build, all in a single job. Most projects finish before you grab coffee." />
          <Feature icon={<ShieldCheck className="size-5" />} title="Build-validated output" body="We run pnpm install, tsc --noEmit and pnpm build. If anything fails, you get the full log instead of a broken ZIP." />
          <Feature icon={<GitBranch className="size-5" />} title="13 deterministic phases" body="Same migration logic from our reference Lovable → Skip migration. Zod 3→4, React 19 useRef, calendar v8→v9, all handled." />
        </div>
      </section>

      <section id="how" className="container py-16">
        <h2 className="mb-12 text-center text-3xl font-bold">How it works</h2>
        <ol className="mx-auto grid max-w-3xl gap-6">
          <Step n={1} title="Upload your Lovable project ZIP" body="Sign in with email, drop the ZIP (up to 100 MB)." />
          <Step n={2} title="Choose options" body="Pixel-perfect override of components/ui? Run build validation? Stay default and ship." />
          <Step n={3} title="We migrate + validate" body="13 phases run in a sandboxed worker. Output ZIP is built and ready to deploy." />
          <Step n={4} title="Download" body="Signed link to your output ZIP. Full build log included." />
        </ol>
      </section>

      <footer className="container py-12 text-center text-sm text-muted-foreground">
        © {new Date().getFullYear()} Skip Migrator · MIT
      </footer>
    </div>
  )
}

function Feature({ icon, title, body }: { icon: React.ReactNode; title: string; body: string }) {
  return (
    <div className="rounded-xl border border-border bg-card p-6">
      <div className="mb-3 inline-flex items-center justify-center rounded-lg bg-primary/10 p-2 text-primary">{icon}</div>
      <h3 className="mb-2 text-lg font-semibold">{title}</h3>
      <p className="text-sm text-muted-foreground">{body}</p>
    </div>
  )
}

function Step({ n, title, body }: { n: number; title: string; body: string }) {
  return (
    <li className="flex gap-4 rounded-xl border border-border bg-card p-6">
      <div className="grid size-10 shrink-0 place-items-center rounded-full bg-primary text-lg font-bold text-primary-foreground">
        {n}
      </div>
      <div>
        <h4 className="text-lg font-semibold">{title}</h4>
        <p className="text-sm text-muted-foreground">{body}</p>
      </div>
    </li>
  )
}
