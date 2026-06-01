# skip-migrator — E2E tests

Playwright suites driving the deployed app like a real browser session.

## Layout

- `tests/landing.spec.ts` — landing page, public surface
- `tests/auth.spec.ts` — header on /login, invalid creds, successful login
- `tests/admin.spec.ts` — `/admin/users` CRUD + self-deletion guard
- `tests/migration-flow.spec.ts` — full upload → success → download + Supabase checklist
- `fixtures/lovable-sample.zip` — small real Lovable project used by the migration test

## Local run

```bash
cd e2e
pnpm install   # or npm install / yarn
npx playwright install chromium   # first time
pnpm test
```

Defaults to `https://migrator.170.84.141.15.sslip.io`. Point elsewhere with:

```bash
TEST_BASE_URL=http://localhost:5173 pnpm test
```

Super-admin creds are read from env, with safe defaults for the live deploy:

```bash
E2E_SUPER_EMAIL=super@skip.dev
E2E_SUPER_PASSWORD=GoSkip@123
```

## What gets created in the target Supabase project

- Each `admin.spec.ts` run creates one user `e2e-test-<random>@e2e.skip.dev` and removes it at the end.
- Each `migration-flow.spec.ts` run creates one row in `public.migrations` plus two ZIP objects in the `source-zips` / `output-zips` buckets. These are not auto-cleaned; periodic GC is on the roadmap.

## CI

See `.github/workflows/e2e.yml` — runs on push to `main` and on PRs.
