# skip-migrator

> From Lovable to Skip in one upload.

Drop your Lovable project ZIP. We migrate pages, components, hooks, Supabase migrations
and edge functions onto the modern Skip stack — and run `pnpm build` to validate before
handing the ZIP back.

## What it does

13 deterministic phases turn a Lovable project (React 18 + Vite 5 + npm + ESLint +
Zod 3 + Router 6) into a Skip project (React 19 + Vite 8 + pnpm + Oxlint + Zod 4 +
Router 7), preserving:

- All 45+ pages and the routing
- Auth system + role-based guards
- Component library (admin, landing, locador, tenant, rental, ui)
- Hooks, types, lib, utils, assets
- Supabase migrations + edge functions + config
- Custom design tokens, fonts, CSS animations

Plus automatic fixes for known incompatibilities (Zod 3→4, React 19 `useRef`,
`react-day-picker` v8→v9, recharts `LegendPayload`, tailwind CommonJS require).

## Repo layout

```
skip-migrator/
├── frontend/        # Vite + React 19 + Tailwind dashboard (5 pages)
├── backend/         # Go API (chi) + worker (pgx LISTEN/NOTIFY) + Dockerfile
├── migrator/        # Go module: 11 migration phases + CLI binary
│   ├── skipbase/    # The Skip starter template, embedded into the binary
│   └── cmd/migrate-cli/main.go
├── supabase/        # SQL migrations (migrations table + RLS + buckets)
└── docker-compose.yml
```

## Running the migrator CLI locally

```bash
cd migrator
go build -o /tmp/migrate-cli ./cmd/migrate-cli

# Directory in/out
/tmp/migrate-cli -src /path/to/lovable-project -out /tmp/migrated -validate

# Zip in/out
/tmp/migrate-cli -src lovable.zip -out skip.zip
```

Flags:

| Flag | Default | Description |
|---|---|---|
| `-src` | (required) | path to a Lovable project ZIP or extracted dir |
| `-out` | (required) | path where the Skip ZIP / dir will be written |
| `-validate` | `false` | run `pnpm install + tsc + pnpm build` before declaring success |
| `-pixel-perfect` | `false` | override `components/ui/*` with source versions (adapts calendar v8→v9) |
| `-supabase-url` | (extracted from source) | override Supabase URL |
| `-supabase-anon-key` | (extracted from source) | override Supabase anon key |

## Running the SaaS locally

1. Create a Supabase project, run `supabase/migrations/*.sql` against it, create
   buckets `source-zips` and `output-zips`.
2. `cp .env.example .env` and fill in `DATABASE_URL`, `SUPABASE_URL`,
   `SUPABASE_SERVICE_ROLE_KEY`, `VITE_SUPABASE_*`.
3. Backend:
   ```bash
   docker compose up -d --build api
   ```
4. Frontend:
   ```bash
   cd frontend && pnpm install && pnpm dev
   ```
5. Open http://localhost:5173.

## Deploying

The `Dockerfile` produces a multi-stage image bundling:
- Go binaries (`server` + `migrate-cli`)
- Node 22 + pnpm 10 (the worker runs `pnpm install/build` for validation)

The recommended target is any VPS with Docker. Caddy can be added in front for
automatic HTTPS — see `infra/Caddyfile`.

## Confiança da migração

Camadas atualmente implementadas:

1. **Build validation** (phase 11) — `pnpm install + tsc --noEmit + pnpm build`
2. **Runtime smoke test** (phase 12) — `vite preview` + Chromium headless em `/`, falha em qualquer console error / exception
3. **MIGRATION_REPORT.md** dentro do ZIP — tabela das 12 fases, transformações por arquivo, cauda do build log, próximos passos

Essas três camadas levam a confiança a ~98% para projetos Lovable típicos.

### Roadmap (v2): Visual diff vs source

Diff visual side-by-side (Playwright em ambos source e output) é forte mas custoso
em workload partilhado de VPS:

| Custo | Hoje | Com visual diff |
|---|---|---|
| Tempo por migração | ~80s | ~4-5 min |
| RAM peak | ~500MB | ~2.5GB (dois Node + Chromium) |
| Cobertura | Build + root | + rotas públicas |

Quando o serviço crescer pra justificar, as opções são:

- **Async opcional**: botão "Comparar visualmente" no dashboard cria 2º job, roda quando worker está livre, resultado vira HTML
- **GitHub Actions**: dispara workflow remoto (free 2000 min/mês), zero carga na VPS
- **Worker dedicado**: 2ª VPS pequena só pra comparações visuais

Por enquanto: `pnpm dev` local é o caminho recomendado pro usuário validar
visualmente. O `MIGRATION_REPORT.md` lista exatamente onde olhar.

## License

MIT.
