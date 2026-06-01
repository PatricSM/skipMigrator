import { test, expect } from '@playwright/test'
import { loginAsSuperAdmin } from './helpers'

test.describe('Admin — gestão de contas', () => {
  test.beforeEach(async ({ page }) => {
    await loginAsSuperAdmin(page)
  })

  test('admin link visible in header for super admin', async ({ page }) => {
    // Sanity: the authenticated header is present (Nova migração always renders when authed)
    await expect(page.locator('header').getByRole('link', { name: /Nova migração/ })).toBeVisible({ timeout: 15_000 })
    // useMe() resolves after a /api/me round-trip → Admin link then renders
    await expect(page.locator('header').getByRole('link', { name: /^Admin$/ })).toBeVisible({ timeout: 15_000 })
  })

  test('admin users page lists existing accounts', async ({ page }) => {
    await page.goto('/admin/users')
    await expect(page.getByRole('heading', { name: 'Administração de contas' })).toBeVisible()
    // The super admin we logged in with must appear in the list
    await expect(page.getByText('super@skip.dev')).toBeVisible({ timeout: 15_000 })
  })

  test('creates a user and surfaces the generated password once', async ({ page }) => {
    const uniq = Date.now().toString(36)
    const email = `e2e-test-${uniq}@e2e.skip.dev`

    await page.goto('/admin/users')
    await page.getByPlaceholder('usuario@empresa.com').fill(email)
    await page.getByRole('button', { name: /Criar conta/ }).click()

    // Callout with generated password appears once (toast + inline message both match)
    await expect(page.getByText(/Conta criada para/).first()).toBeVisible({ timeout: 15_000 })
    await expect(page.getByText(email).first()).toBeVisible()
    await expect(page.getByRole('button', { name: 'Copiar' })).toBeVisible()

    // The new user should now appear in the list
    await expect(page.getByText(email).first()).toBeVisible()

    // Cleanup: remove via the trash button next to the row.
    // Accept the native confirm() prompt.
    page.on('dialog', (d) => void d.accept())
    const row = page.locator('li', { hasText: email })
    await row.getByRole('button', { name: 'Remover' }).click()
    await expect(row).toHaveCount(0, { timeout: 10_000 })
  })

  test('blocks super admin from removing their own account', async ({ page }) => {
    await page.goto('/admin/users')
    const ownRow = page.locator('li', { hasText: 'super@skip.dev' })
    await expect(ownRow).toBeVisible({ timeout: 15_000 })
    // The button is enabled but triggers a toast error guard
    await ownRow.getByRole('button', { name: 'Remover' }).click()
    await expect(page.getByText(/Não dá pra remover a própria conta/)).toBeVisible()
    // Row still there
    await expect(ownRow).toBeVisible()
  })
})
