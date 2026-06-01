import { test, expect } from '@playwright/test'
import path from 'node:path'
import { fileURLToPath } from 'node:url'
import { loginAsSuperAdmin } from './helpers'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const FIXTURE_ZIP = path.resolve(__dirname, '..', 'fixtures', 'lovable-sample.zip')

test.describe('Migration flow (end-to-end)', () => {
  test.beforeEach(async ({ page }) => {
    await loginAsSuperAdmin(page)
  })

  // ~90s on the real worker; the suite timeout (5m) covers it.
  test('upload ZIP → queued → success → checklist + download', async ({ page }) => {
    test.slow() // gives this test 3× the configured timeout (15m max)

    // Open the new-migration form
    await page.goto('/app/new')
    await expect(page.getByRole('heading', { name: 'Nova migração' })).toBeVisible()

    // Upload fixture
    const input = page.locator('input[type="file"]')
    await input.setInputFiles(FIXTURE_ZIP)
    await expect(page.getByText(/lovable-sample\.zip/)).toBeVisible()

    // Keep defaults (validate=true, pixelPerfect=false)
    await page.getByRole('button', { name: /Iniciar migração/ }).click()

    // Should land on the migration detail page
    await expect(page).toHaveURL(/\/app\/m\/[0-9a-f-]{36}$/, { timeout: 15_000 })

    // Status progresses queued → running → success (poll the heading)
    const statusHeading = page.getByRole('heading', { level: 1 })
    await expect(statusHeading).toContainText(/Na fila|Processando/, { timeout: 10_000 })
    await expect(statusHeading).toContainText('Concluída', { timeout: 4 * 60 * 1000 })

    // Once success, the download button + Supabase checklist must appear
    await expect(page.getByRole('link', { name: /Baixar ZIP/ })).toBeVisible()
    await expect(page.getByRole('heading', { name: 'Checklist Supabase' })).toBeVisible()
    await expect(page.getByText(/plug-and-play na maioria dos casos/)).toBeVisible()

    // Build log is collapsed but present
    await expect(page.getByText('Log do build')).toBeVisible()
  })
})
