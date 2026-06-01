import { test, expect } from '@playwright/test'
import { SUPER_ADMIN, clearSession } from './helpers'

test.describe('Auth', () => {
  test.beforeEach(async ({ page }) => {
    await clearSession(page)
  })

  test('login form has the header and no public signup', async ({ page }) => {
    await page.goto('/login')
    // Header (logo) present even on login
    await expect(page.locator('header').getByRole('link', { name: /Skip Migrator/ })).toBeVisible()

    // Form
    await expect(page.getByRole('heading', { name: 'Entrar' })).toBeVisible()
    await expect(page.getByPlaceholder('voce@email.com')).toBeVisible()
    await expect(page.getByPlaceholder('Mínimo 8 caracteres')).toBeVisible()

    // No public signup
    await expect(page.getByRole('button', { name: /Cadastre-se/ })).toHaveCount(0)
    await expect(page.getByText(/Acesso por convite/)).toBeVisible()
  })

  test('rejects invalid credentials with a toast', async ({ page }) => {
    await page.goto('/login')
    await page.getByPlaceholder('voce@email.com').fill('bad@example.com')
    await page.getByPlaceholder('Mínimo 8 caracteres').fill('wrongpassword')
    await page.getByRole('button', { name: 'Entrar' }).click()
    await expect(page.getByText(/Invalid login credentials|Falha na autenticação/i)).toBeVisible({ timeout: 10_000 })
    await expect(page).toHaveURL(/\/login$/)
  })

  test('successful login lands on /app dashboard', async ({ page }) => {
    await page.goto('/login')
    await page.getByPlaceholder('voce@email.com').fill(SUPER_ADMIN.email)
    await page.getByPlaceholder('Mínimo 8 caracteres').fill(SUPER_ADMIN.password)
    await page.getByRole('button', { name: 'Entrar' }).click()
    await expect(page).toHaveURL(/\/app$/, { timeout: 15_000 })
    await expect(page.getByRole('heading', { name: 'Suas migrações' })).toBeVisible()
  })
})
