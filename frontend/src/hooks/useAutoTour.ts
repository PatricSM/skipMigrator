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

    if (path === '/app') {
      if (!hasSeenTour() && getNextStage() === null) toRun = 'dashboard'
      else if (getNextStage() === 'finale') toRun = 'finale'
    } else if (path === '/app/new' && getNextStage() === null) {
      // arrives here only via the "Próximo →" of stage 1, which set via=dashboard-next
      if (sessionStorage.getItem('skipmigrator.tour.via') === 'dashboard-next') {
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
