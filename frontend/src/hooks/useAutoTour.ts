import { useEffect, useRef } from 'react'
import { useNavigate, useLocation } from 'react-router-dom'
import { hasSeenTour, startTour } from '@/lib/tour'
import { useMe } from './useMe'

/**
 * Fires the guided tour automatically the first time a logged-in user lands
 * on /app. Skips if:
 *   - user isn't loaded yet
 *   - user already dismissed the tour
 *   - we're not on the dashboard
 *
 * Mount this once at the App root.
 */
export function useAutoTour() {
  const me = useMe()
  const navigate = useNavigate()
  const location = useLocation()
  const fired = useRef(false)

  useEffect(() => {
    if (fired.current) return
    if (!me || me === undefined) return
    if (location.pathname !== '/app') return
    if (hasSeenTour()) return
    fired.current = true
    // Slight delay so the dashboard has time to paint anchor targets.
    const t = setTimeout(() => {
      startTour({ navigate, includeAdmin: me.is_super_admin })
    }, 400)
    return () => clearTimeout(t)
  }, [me, location.pathname, navigate])
}
