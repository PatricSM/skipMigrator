import { type Page, expect } from '@playwright/test'

// `||` (not `??`) handles the case where GitHub Actions injects an empty
// string for an unset secret. `??` only falls back on null/undefined.
export const SUPER_ADMIN = {
  email: process.env.E2E_SUPER_EMAIL || 'super@skip.dev',
  password: process.env.E2E_SUPER_PASSWORD || 'GoSkip@123',
}

/**
 * Logs in via the /login form and waits until the dashboard is reachable.
 * Use this in beforeEach when a test needs an authenticated session.
 */
export async function loginAsSuperAdmin(page: Page) {
  await page.goto('/login')
  await page.getByPlaceholder('voce@email.com').fill(SUPER_ADMIN.email)
  await page.getByPlaceholder('Mínimo 8 caracteres').fill(SUPER_ADMIN.password)
  await page.getByRole('button', { name: 'Entrar' }).click()
  await expect(page).toHaveURL(/\/app$/, { timeout: 15_000 })
  // Force a clean reload so useMe() runs once with the session already in localStorage.
  // Without this, the post-redirect render races with the async session restore and
  // the header sometimes paints with is_super_admin still unknown.
  await page.reload()
  await expect(page.getByRole('heading', { name: 'Suas migrações' })).toBeVisible({ timeout: 15_000 })
}

/**
 * Clears the Supabase session from localStorage so the next navigation
 * starts in a signed-out state. Useful for tests that exercise the public
 * surface (landing, login).
 */
export async function clearSession(page: Page) {
  await page.addInitScript(() => {
    try {
      for (const k of Object.keys(localStorage)) {
        if (k.startsWith('sb-')) localStorage.removeItem(k)
      }
    } catch {}
  })
}
