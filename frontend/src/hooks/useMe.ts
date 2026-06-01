import { useEffect, useState } from 'react'
import { supabase } from '@/lib/supabase'
import { getMe, type Me } from '@/lib/api'

/**
 * Fetches /api/me whenever the auth session changes. Returns null while
 * unknown and `undefined` for not authenticated.
 */
export function useMe(): Me | null | undefined {
  const [me, setMe] = useState<Me | null | undefined>(null)

  useEffect(() => {
    let cancelled = false
    const load = async () => {
      const { data } = await supabase.auth.getSession()
      if (!data.session) {
        if (!cancelled) setMe(undefined)
        return
      }
      try {
        const m = await getMe()
        if (!cancelled) setMe(m)
      } catch {
        if (!cancelled) setMe(undefined)
      }
    }
    load()
    const { data: sub } = supabase.auth.onAuthStateChange(() => load())
    return () => { cancelled = true; sub.subscription.unsubscribe() }
  }, [])

  return me
}
