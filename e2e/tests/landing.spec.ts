import { test, expect } from '@playwright/test'
import { clearSession } from './helpers'

test.describe('Landing', () => {
  test.beforeEach(async ({ page }) => {
    await clearSession(page)
  })

  test('renders hero + CTAs in pt-BR', async ({ page }) => {
    await page.goto('/')
    await expect(page).toHaveTitle(/Skip Migrator/)
    await expect(page.getByRole('heading', { level: 1 })).toContainText('Lovable')
    await expect(page.getByRole('heading', { level: 1 })).toContainText('Skip')
    await expect(page.getByRole('link', { name: /Começar migração/ })).toBeVisible()
    await expect(page.getByRole('link', { name: 'Ver no GitHub' })).toBeVisible()
  })

  test('header brand link points to /', async ({ page }) => {
    await page.goto('/')
    const brand = page.getByRole('link', { name: /Skip Migrator/ }).first()
    await expect(brand).toHaveAttribute('href', '/')
  })

  test('shows Entrar CTA when signed out', async ({ page }) => {
    await page.goto('/')
    await expect(page.getByRole('link', { name: 'Entrar' })).toBeVisible()
  })

  test('how-it-works lists the 4 steps', async ({ page }) => {
    await page.goto('/#como-funciona')
    await expect(page.getByText('Envie o ZIP do projeto Lovable')).toBeVisible()
    await expect(page.getByText('Escolha as opções')).toBeVisible()
    await expect(page.getByText('Nós migramos + validamos')).toBeVisible()
    await expect(page.getByText('Baixe o resultado')).toBeVisible()
  })
})
