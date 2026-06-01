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

/**
 * Drives the multi-stage guided tour. Decides what to start based on:
 *   - current route
 *   - sessionStorage stage flag (set by the previous stage)
 *   - is_super_admin from /api/me
 *
 * Mount once at the App root. Idempotent within a session (per stage).
 */
export function useAutoTour() {
  const me = useMe()
  const location = useLocation()
  const fired = useRef<string | null>(null)

  useEffect(() => {
    if (!me) return // not loaded yet or signed out
    const path = location.pathname

    // Decide which stage (if any) belongs to this route
    let toRun: 'dashboard' | 'new-migration' | 'admin-users' | 'finale' | null = null

    const via = (() => { try { return sessionStorage.getItem('skipmigrator.tour.via') } catch { return null } })()

    if (path === '/app') {
      // First-time login → start at the beginning
      if (!hasSeenTour() && getNextStage() === null) toRun = 'dashboard'
      // Closing handoff from the new-migration stage when user is NOT admin
      else if (getNextStage() === 'finale') toRun = 'finale'
    } else if (path === '/app/new') {
      // We only auto-start the new-migration stage when we arrived via the
      // dashboard handoff. getNextStage() at this point already points to the
      // stage AFTER new-migration ("admin-users" or "finale"), so we look at
      // the `via` flag instead.
      if (via === 'dashboard-next') {
        toRun = 'new-migration'
        sessionStorage.removeItem('skipmigrator.tour.via')
      }
    } else if (path === '/admin/users' && getNextStage() === 'admin-users') {
      toRun = 'admin-users'
    }

    if (!toRun) return

    const key = `${path}:${toRun}`
    if (fired.current === key) return
    fired.current = key

    const t = setTimeout(() => {
      switch (toRun) {
        case 'dashboard':       startDashboardTour({ includeAdmin: me.is_super_admin }); break
        case 'new-migration':   startNewMigrationTour({ includeAdmin: me.is_super_admin }); break
        case 'admin-users':     startAdminTour(); break
        case 'finale':          startFinaleTour(); break
      }
    }, 500) // give the page a chance to paint anchor targets
    return () => clearTimeout(t)
  }, [me, location.pathname])
}
