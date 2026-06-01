import { useEffect, useRef } from 'react'
import { useLocation } from 'react-router-dom'
import {
  getNextStage,
  hasSeenTour,
  startAdminTour,
  startDashboardTour,
  startFinaleTour,
  startNewMigrationTour,
} from '@/lib/tour'
import { useMe } from './useMe'

type Stage = 'dashboard' | 'new-migration' | 'admin-users' | 'finale'

/**
 * Drives the multi-stage guided tour. Decides what to start based on:
 *   - current route
 *   - sessionStorage stage flag (set by the previous stage)
 *   - is_super_admin from /api/me
 *
 * Mount once at the App root.
 */
export function useAutoTour() {
  const me = useMe()
  const location = useLocation()
  const fired = useRef<string | null>(null)

  useEffect(() => {
    if (!me) return
    const path = location.pathname
    const via = (() => { try { return sessionStorage.getItem('skipmigrator.tour.via') } catch { return null } })()

    let toRun: Stage | null = null
    if (path === '/app') {
      if (!hasSeenTour() && getNextStage() === null) toRun = 'dashboard'
      else if (getNextStage() === 'finale') toRun = 'finale'
    } else if (path === '/app/new' && via === 'dashboard-next') {
      toRun = 'new-migration'
    } else if (path === '/admin/users' && getNextStage() === 'admin-users') {
      toRun = 'admin-users'
    }

    if (!toRun) return

    const key = `${path}:${toRun}`
    if (fired.current === key) return
    fired.current = key

    // Wait until the anchor element actually exists in the DOM before launching.
    // Better than a flat setTimeout — avoids races where the page is still
    // hydrating when we try to anchor the popover.
    const anchor = anchorFor(toRun)
    const launch = () => {
      if (toRun === 'new-migration') sessionStorage.removeItem('skipmigrator.tour.via')
      switch (toRun) {
        case 'dashboard':       startDashboardTour({ includeAdmin: me.is_super_admin }); break
        case 'new-migration':   startNewMigrationTour({ includeAdmin: me.is_super_admin }); break
        case 'admin-users':     startAdminTour(); break
        case 'finale':          startFinaleTour(); break
      }
    }
    if (!anchor) {
      // No specific anchor (welcome modal) — just give the page a tick to paint.
      const t = setTimeout(launch, 300)
      return () => clearTimeout(t)
    }
    waitForAnchor(anchor, 4000).then((found) => {
      if (found) launch()
    })
  }, [me, location.pathname])
}

function anchorFor(stage: Stage): string | null {
  switch (stage) {
    case 'dashboard':       return '[data-tour="dashboard-main"]'
    case 'new-migration':   return '[data-tour="upload-zone"]'
    case 'admin-users':     return '[data-tour="create-user-form"]'
    case 'finale':          return null
  }
}

async function waitForAnchor(selector: string, timeoutMs: number): Promise<boolean> {
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
