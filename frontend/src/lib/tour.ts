import { driver, type Driver } from 'driver.js'
import 'driver.js/dist/driver.css'

/**
 * Guided onboarding tour using driver.js.
 *
 * Steps span 3 routes (/app, /app/new, /admin/users). Inter-route transitions
 * happen via the onNextClick callback: we navigate, wait for the new element
 * to mount, then advance the driver instance manually.
 *
 * Persistence: localStorage flag — show once, dismissible. The user can replay
 * via the "Refazer tour" button rendered in AppHeader.
 */

const STORAGE_KEY = 'skipmigrator.tour.completed'

export function hasSeenTour(): boolean {
  try {
    return localStorage.getItem(STORAGE_KEY) === '1'
  } catch {
    return false
  }
}

export function markTourSeen() {
  try {
    localStorage.setItem(STORAGE_KEY, '1')
  } catch {}
}

export function resetTour() {
  try {
    localStorage.removeItem(STORAGE_KEY)
  } catch {}
}

/**
 * Wait for a selector to appear in the DOM (post-navigation mount).
 * Returns true if the element shows up within `timeoutMs`, false otherwise.
 */
async function waitForElement(selector: string, timeoutMs = 4000): Promise<boolean> {
  const start = Date.now()
  return new Promise((resolve) => {
    const tick = () => {
      if (document.querySelector(selector)) return resolve(true)
      if (Date.now() - start > timeoutMs) return resolve(false)
      requestAnimationFrame(tick)
    }
    tick()
  })
}

interface StartTourOpts {
  navigate: (path: string) => void
  /** When false, super-admin-only steps (Admin link, /admin/users) are omitted. */
  includeAdmin: boolean
}

export function startTour({ navigate, includeAdmin }: StartTourOpts): Driver {
  const baseSteps = [
    {
      popover: {
        title: '👋 Bem-vindo ao Skip Migrator',
        description:
          'Em ~1 minuto te mostro como migrar um projeto Lovable para a stack Skip. Pode pular a qualquer momento.',
      },
    },
    {
      element: '[data-tour="new-migration-btn"]',
      popover: {
        title: '1️⃣ Nova migração',
        description: 'Tudo começa por aqui. Vou levar você para a tela de upload.',
        side: 'bottom' as const,
        align: 'end' as const,
      },
      onNextClick: async (_el: Element | undefined, _step: unknown, opts: { driver: Driver }) => {
        navigate('/app/new')
        if (await waitForElement('[data-tour="upload-zone"]')) {
          opts.driver.moveNext()
        }
      },
    },
    {
      element: '[data-tour="upload-zone"]',
      popover: {
        title: '2️⃣ Suba o ZIP',
        description:
          'Arraste o ZIP do seu projeto Lovable aqui (até 100 MB). Garanta que <code>node_modules</code> e <code>dist</code> não estejam dentro.',
        side: 'bottom' as const,
        align: 'center' as const,
      },
    },
    {
      element: '[data-tour="validate-option"]',
      popover: {
        title: '3️⃣ Validar build (recomendado)',
        description:
          'Roda <code>pnpm install + tsc + pnpm build</code> antes de devolver. Adiciona ~90s mas garante que o ZIP de saída funciona.',
        side: 'bottom' as const,
        align: 'start' as const,
      },
    },
    {
      element: '[data-tour="pixel-perfect-option"]',
      popover: {
        title: '4️⃣ Pixel-perfect (opcional)',
        description:
          'Sobrescreve <code>components/ui/*</code> com as versões do seu projeto. Use se quiser réplica visual exata.',
        side: 'bottom' as const,
        align: 'start' as const,
      },
    },
    {
      element: '[data-tour="submit-btn"]',
      popover: {
        title: '5️⃣ Iniciar',
        description: 'Click aqui e você é levado para o card de status da migração.',
        side: 'top' as const,
      },
      onNextClick: async (_el: Element | undefined, _step: unknown, opts: { driver: Driver }) => {
        navigate('/app')
        if (await waitForElement('[data-tour="dashboard-main"]')) {
          opts.driver.moveNext()
        }
      },
    },
    {
      element: '[data-tour="dashboard-main"]',
      popover: {
        title: '6️⃣ Seu histórico',
        description:
          'Migrações ficam aqui em tempo real. Quando concluídas, aparece o botão para baixar o ZIP.',
        side: 'bottom' as const,
      },
    },
  ]

  const adminSteps = [
    {
      element: '[data-tour="admin-link"]',
      popover: {
        title: '🛡️ Você é super admin',
        description: 'Como super admin, você cria contas para outros usuários aqui.',
        side: 'bottom' as const,
      },
      onNextClick: async (_el: Element | undefined, _step: unknown, opts: { driver: Driver }) => {
        navigate('/admin/users')
        if (await waitForElement('[data-tour="create-user-form"]')) {
          opts.driver.moveNext()
        }
      },
    },
    {
      element: '[data-tour="create-user-form"]',
      popover: {
        title: '➕ Criar conta',
        description:
          'Preencha o e-mail. A senha é opcional — se deixar vazio, geramos uma forte e mostramos <strong>uma vez</strong> para você copiar e enviar pelo seu canal seguro.',
        side: 'bottom' as const,
      },
      onNextClick: async (_el: Element | undefined, _step: unknown, opts: { driver: Driver }) => {
        navigate('/app')
        if (await waitForElement('[data-tour="dashboard-main"]')) {
          opts.driver.moveNext()
        }
      },
    },
  ]

  const finalStep = [
    {
      popover: {
        title: '🎉 Pronto!',
        description:
          'Você pode rever este tour a qualquer momento pelo ícone <strong>?</strong> no header. Boa migração!',
      },
    },
  ]

  const steps = includeAdmin
    ? [...baseSteps, ...adminSteps, ...finalStep]
    : [...baseSteps, ...finalStep]

  const d = driver({
    showProgress: true,
    allowClose: true,
    overlayColor: 'rgba(0, 0, 0, 0.75)',
    popoverClass: 'skip-tour',
    nextBtnText: 'Próximo →',
    prevBtnText: '← Voltar',
    doneBtnText: 'Concluir',
    progressText: '{{current}} de {{total}}',
    onDestroyed: () => markTourSeen(),
    steps,
  })

  d.drive()
  return d
}
