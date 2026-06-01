import { driver, type Driver } from 'driver.js'
import 'driver.js/dist/driver.css'

/**
 * Guided onboarding tour using driver.js, split into 3+1 stages — one per
 * route. Stage handoffs happen via a sessionStorage flag + full-page nav
 * (window.location.href). useAutoTour reads the flag on each route mount and
 * starts the matching stage.
 *
 * Why full-page nav instead of react-router navigate(): driver.js is
 * DOM-imperative, fires from outside React, and our callbacks need to do a
 * route change. Calling navigate() from there is racy. window.location.href is
 * synchronous and reliable, at the cost of a reload (~50ms).
 *
 * Why use onDestroyStarted (not onNextClick): the last step of each stage
 * shows the "Concluir" button (default driver.js behavior for the final step),
 * which does NOT fire onNextClick. onDestroyStarted fires on every termination
 * (Concluir OR ×). We disambiguate by checking opts.state.activeIndex against
 * the configured "advance index" — if the user reached that step before
 * dismissing, we navigate to the next stage; otherwise we mark the tour done.
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
  try { sessionStorage.removeItem('skipmigrator.tour.via') } catch {}
}
export function resetTour() {
  try { localStorage.removeItem(COMPLETED_KEY) } catch {}
  try { sessionStorage.removeItem(STAGE_KEY) } catch {}
  try { sessionStorage.removeItem('skipmigrator.tour.via') } catch {}
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

interface StageHandoff {
  /** 0-based index of the step that, when reached, signals "advance to next stage on close". */
  advanceFromStep: number
  /** Where to navigate when the user clicks Concluir on the advance step. */
  nextPath: string
  /** What to set in sessionStorage so the next page knows which stage to start. */
  nextStage: TourStage | null
  /** Optional flag set in via so the next page knows the previous step closed forward. */
  viaFlag?: string
}

function makeStagedDriver(
  steps: Parameters<typeof driver>[0]['steps'],
  handoff: StageHandoff,
): Driver {
  return driver({
    showProgress: true,
    allowClose: true,
    overlayColor: 'rgba(0, 0, 0, 0.75)',
    popoverClass: 'skip-tour',
    nextBtnText: 'Próximo →',
    prevBtnText: '← Voltar',
    doneBtnText: 'Continuar →',
    progressText: '{{current}} de {{total}}',
    onDestroyStarted: (_el, _step, opts) => {
      const idx = opts.state.activeIndex ?? -1
      const reachedAdvance = idx >= handoff.advanceFromStep
      if (reachedAdvance) {
        if (handoff.viaFlag) sessionStorage.setItem('skipmigrator.tour.via', handoff.viaFlag)
        setNextStage(handoff.nextStage)
        // Defer to next tick so destroy() can finish cleanly before navigation.
        setTimeout(() => { window.location.href = handoff.nextPath }, 50)
      } else {
        // User dismissed before completing the stage — mark done so we don't re-prompt.
        markTourSeen()
      }
      opts.driver.destroy()
    },
    steps,
  })
}

// ── Stage 1: Dashboard (/app) ────────────────────────────────────────────
export function startDashboardTour(opts: { includeAdmin: boolean }): Driver {
  const steps: Parameters<typeof driver>[0]['steps'] = [
    {
      popover: {
        title: '👋 Bem-vindo ao Skip Migrator',
        description:
          'Em ~1 minuto te mostro como migrar um projeto Lovable para a stack Skip. Pode fechar (×) a qualquer momento.',
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
        title: '➕ Tudo começa aqui',
        description:
          'Click em <strong>Continuar →</strong> e te levo pra tela de upload pra ver as opções da migração.',
        side: 'bottom' as const,
        align: 'end' as const,
      },
    },
  ]
  const d = makeStagedDriver(steps, {
    advanceFromStep: 2,
    nextPath: '/app/new',
    nextStage: opts.includeAdmin ? 'admin-users' : 'finale',
    viaFlag: 'dashboard-next',
  })
  d.drive()
  return d
}

// ── Stage 2: New migration form (/app/new) ───────────────────────────────
export function startNewMigrationTour(opts: { includeAdmin: boolean }): Driver {
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
        description: 'Esse botão dispara a migração. Click <strong>Continuar →</strong> pra seguir o tour.',
        side: 'top' as const,
      },
    },
  ]
  const d = makeStagedDriver(steps, {
    advanceFromStep: 3,
    nextPath: opts.includeAdmin ? '/admin/users' : '/app',
    nextStage: opts.includeAdmin ? 'admin-users' : 'finale',
  })
  d.drive()
  return d
}

// ── Stage 3 (super admin only): /admin/users ─────────────────────────────
export function startAdminTour(): Driver {
  const steps: Parameters<typeof driver>[0]['steps'] = [
    {
      element: '[data-tour="create-user-form"]',
      popover: {
        title: '🛡️ Você é super admin',
        description:
          'Como super admin, você cria contas pra outros usuários aqui. A senha é opcional — se deixar vazio, geramos uma forte e mostramos <strong>uma vez</strong> pra você copiar e enviar pelo seu canal seguro.',
        side: 'bottom' as const,
      },
    },
  ]
  const d = makeStagedDriver(steps, {
    advanceFromStep: 0,
    nextPath: '/app',
    nextStage: 'finale',
  })
  d.drive()
  return d
}

// ── Finale on /app ───────────────────────────────────────────────────────
export function startFinaleTour(): Driver {
  const d = driver({
    showProgress: false,
    allowClose: true,
    overlayColor: 'rgba(0, 0, 0, 0.75)',
    popoverClass: 'skip-tour',
    doneBtnText: 'Boa migração!',
    onDestroyStarted: (_el, _step, opts) => {
      markTourSeen()
      opts.driver.destroy()
    },
    steps: [
      {
        popover: {
          title: '🎉 Pronto!',
          description:
            'Você pode rever este tour a qualquer momento clicando no ícone <strong>?</strong> no header.',
        },
      },
    ],
  })
  d.drive()
  return d
}

/**
 * Entry-point used by both auto-start (first login) and the "Refazer tour"
 * button. Resets stage + completion flag and starts at Stage 1.
 */
export function startTourFromBeginning(opts: { includeAdmin: boolean }): Driver {
  resetTour()
  return startDashboardTour(opts)
}
