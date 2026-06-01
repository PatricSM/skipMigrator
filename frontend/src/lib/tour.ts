import { driver, type Driver } from 'driver.js'
import 'driver.js/dist/driver.css'

/**
 * Guided onboarding tour using driver.js, split into 3 stages — one per route.
 *
 * Transitions between stages happen via a sessionStorage flag + full-page nav
 * (window.location.href). useAutoTour reads the flag on each route mount and
 * starts the matching stage. This avoids the driver.js × React Router
 * impedance mismatch (driver runs DOM-imperatively, navigate() needs the React
 * context to be live).
 */

const COMPLETED_KEY = 'skipmigrator.tour.completed'
const STAGE_KEY = 'skipmigrator.tour.stage'

export type TourStage = 'new-migration' | 'admin-users' | 'finale'

export function hasSeenTour(): boolean {
  try { return localStorage.getItem(COMPLETED_KEY) === '1' } catch { return false }
}
export function markTourSeen() {
  try { localStorage.setItem(COMPLETED_KEY, '1') } catch {}
  try { sessionStorage.removeItem(STAGE_KEY) } catch {}
}
export function resetTour() {
  try { localStorage.removeItem(COMPLETED_KEY) } catch {}
  try { sessionStorage.removeItem(STAGE_KEY) } catch {}
}
export function getNextStage(): TourStage | null {
  try {
    const v = sessionStorage.getItem(STAGE_KEY)
    return v === 'new-migration' || v === 'admin-users' || v === 'finale' ? v : null
  } catch { return null }
}
function setNextStage(s: TourStage | null) {
  try {
    if (s === null) sessionStorage.removeItem(STAGE_KEY)
    else sessionStorage.setItem(STAGE_KEY, s)
  } catch {}
}

// Common driver factory: applies our dark+indigo theme to every stage.
function makeDriver(steps: Parameters<typeof driver>[0]['steps'], onDestroyed?: () => void): Driver {
  return driver({
    showProgress: true,
    allowClose: true,
    overlayColor: 'rgba(0, 0, 0, 0.75)',
    popoverClass: 'skip-tour',
    nextBtnText: 'Próximo →',
    prevBtnText: '← Voltar',
    doneBtnText: 'Concluir',
    progressText: '{{current}} de {{total}}',
    onDestroyed,
    steps,
  })
}

// ──────────────────────────────────────────────────────────────────────────
// Stage 1: Dashboard (/app)
// Welcome → nova-migração button. Last step navigates to /app/new.
// ──────────────────────────────────────────────────────────────────────────
export function startDashboardTour(opts: { includeAdmin: boolean }): Driver {
  let advancedToNext = false

  const steps: Parameters<typeof driver>[0]['steps'] = [
    {
      popover: {
        title: '👋 Bem-vindo ao Skip Migrator',
        description:
          'Em ~1 minuto te mostro como migrar um projeto Lovable para a stack Skip. Pode pular a qualquer momento clicando no X.',
      },
    },
    {
      element: '[data-tour="dashboard-main"]',
      popover: {
        title: 'Seu histórico',
        description: 'Suas migrações aparecem aqui em tempo real. Quando concluídas, aparece o botão para baixar o ZIP.',
        side: 'bottom' as const,
      },
    },
    {
      element: '[data-tour="new-migration-btn"]',
      popover: {
        title: '➕ Nova migração',
        description: 'Tudo começa por aqui. Click em <strong>Próximo →</strong> e te levo pra tela de upload.',
        side: 'bottom' as const,
        align: 'end' as const,
      },
      onNextClick: (_el, _step, popoverOpts) => {
        advancedToNext = true
        setNextStage(opts.includeAdmin ? 'admin-users' : 'finale')
        // chain a /app/new visit first — that page has its own stage tour
        sessionStorage.setItem('skipmigrator.tour.via', 'dashboard-next')
        popoverOpts.driver.destroy()
        setTimeout(() => { window.location.href = '/app/new' }, 50)
      },
    },
  ]
  return makeDriver(steps, () => {
    if (!advancedToNext) markTourSeen() // user closed mid-stage
  }).drive() as unknown as Driver
}

// ──────────────────────────────────────────────────────────────────────────
// Stage 2: New migration form (/app/new)
// Walks through the 3 form sections, ends by going to /app (or /admin/users).
// ──────────────────────────────────────────────────────────────────────────
export function startNewMigrationTour(opts: { includeAdmin: boolean }): Driver {
  let advancedToNext = false
  const steps: Parameters<typeof driver>[0]['steps'] = [
    {
      element: '[data-tour="upload-zone"]',
      popover: {
        title: '1️⃣ Suba o ZIP',
        description:
          'Arraste o ZIP do seu projeto Lovable aqui (até 100 MB). Garanta que <code>node_modules</code> e <code>dist</code> não estejam dentro.',
        side: 'bottom' as const,
      },
    },
    {
      element: '[data-tour="validate-option"]',
      popover: {
        title: '2️⃣ Validar build (recomendado)',
        description:
          'Roda <code>pnpm install + tsc + pnpm build</code> antes de devolver. Adiciona ~90s mas garante que o ZIP de saída funciona.',
        side: 'bottom' as const,
      },
    },
    {
      element: '[data-tour="pixel-perfect-option"]',
      popover: {
        title: '3️⃣ Pixel-perfect (opcional)',
        description:
          'Sobrescreve <code>components/ui/*</code> com as versões do seu projeto. Liga só se quiser réplica visual exata.',
        side: 'bottom' as const,
      },
    },
    {
      element: '[data-tour="submit-btn"]',
      popover: {
        title: '4️⃣ Iniciar',
        description: 'Click aqui (ou no botão <strong>Próximo →</strong>) pra terminar o tour e voltar ao painel.',
        side: 'top' as const,
      },
      onNextClick: (_el, _step, popoverOpts) => {
        advancedToNext = true
        setNextStage(opts.includeAdmin ? 'admin-users' : 'finale')
        popoverOpts.driver.destroy()
        const target = opts.includeAdmin ? '/admin/users' : '/app'
        setTimeout(() => { window.location.href = target }, 50)
      },
    },
  ]
  return makeDriver(steps, () => {
    if (!advancedToNext) markTourSeen()
  }).drive() as unknown as Driver
}

// ──────────────────────────────────────────────────────────────────────────
// Stage 3 (super admin only): Admin → /admin/users
// ──────────────────────────────────────────────────────────────────────────
export function startAdminTour(): Driver {
  let advancedToNext = false
  const steps: Parameters<typeof driver>[0]['steps'] = [
    {
      element: '[data-tour="create-user-form"]',
      popover: {
        title: '🛡️ Você é super admin',
        description:
          'Como super admin, você cria contas pra outros usuários aqui. A senha é opcional — se deixar vazio, geramos uma forte e mostramos <strong>uma vez</strong> pra você copiar e enviar pelo seu canal seguro.',
        side: 'bottom' as const,
      },
      onNextClick: (_el, _step, popoverOpts) => {
        advancedToNext = true
        setNextStage('finale')
        popoverOpts.driver.destroy()
        setTimeout(() => { window.location.href = '/app' }, 50)
      },
    },
  ]
  return makeDriver(steps, () => {
    if (!advancedToNext) markTourSeen()
  }).drive() as unknown as Driver
}

// ──────────────────────────────────────────────────────────────────────────
// Final farewell back on /app
// ──────────────────────────────────────────────────────────────────────────
export function startFinaleTour(): Driver {
  const steps: Parameters<typeof driver>[0]['steps'] = [
    {
      popover: {
        title: '🎉 Pronto!',
        description:
          'Você pode rever este tour a qualquer momento clicando no ícone <strong>?</strong> no header. Boa migração!',
      },
    },
  ]
  return makeDriver(steps, () => markTourSeen()).drive() as unknown as Driver
}

/**
 * Entry-point used by both auto-start (first login) and the "Refazer tour"
 * button. Always resets stage + completion flag and starts at Stage 1.
 */
export function startTourFromBeginning(opts: { includeAdmin: boolean }): Driver {
  resetTour()
  return startDashboardTour(opts)
}
