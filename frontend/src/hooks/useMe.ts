import { useEffect, useState } from 'react'
import { supabase } from '@/lib/supabase'
import { getMe, type Me } from '@/lib/api'

/**
 * Fetches /api/me whenever the auth session changes. Returns null while
 * unknown and `undefined` for not authenticated.
 */
export function useMe(): Me | null | undefined {
  const [me, setMe] = useState<Me | null | undefined>(null)
  console.log('[useMe] render', me)

  useEffect(() => {
    console.log('[useMe] effect mount')
    let cancelled = false
    const load = async () => {
      console.log('[useMe] load called')
      const { data } = await supabase.auth.getSession()
      console.log('[useMe] session?', !!data.session)
      if (!data.session) {
        if (!cancelled) setMe(undefined)
        return
      }
      try {
        const m = await getMe()
        console.log('[useMe] got', m)
        if (!cancelled) setMe(m)
      } catch (e) {
        console.log('[useMe] err', e)
        if (!cancelled) setMe(undefined)
      }
    }
    load()
    const { data: sub } = supabase.auth.onAuthStateChange(() => load())
    return () => { cancelled = true; sub.subscription.unsubscribe() }
  }, [])

  return me
}
