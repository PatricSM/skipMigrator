import { useEffect, useState } from 'react'
import { supabase } from '@/lib/supabase'
import { getMe, type Me } from '@/lib/api'

/**
 * Fetches /api/me using the JWT directly from localStorage.
 *
 * We deliberately avoid `supabase.auth.getSession()` here because in some
 * conditions (notably right after a full-page navigation) the supabase-js
 * internal lock can stall the promise indefinitely. Reading the stored token
 * synchronously is reliable and the JWT is already what we'd hand to /api/me.
 *
 * Returns:
 *   null      — initial / still loading
 *   undefined — not authenticated
 *   Me        — authenticated user payload
 */
export function useMe(): Me | null | undefined {
  const [me, setMe] = useState<Me | null | undefined>(null)

  useEffect(() => {
    let cancelled = false

    const tokenFromStorage = (): string | null => {
      try {
        for (const k of Object.keys(localStorage)) {
          if (!k.startsWith('sb-') || !k.endsWith('-auth-token')) continue
          const raw = localStorage.getItem(k)
          if (!raw) continue
          const parsed = JSON.parse(raw)
          if (parsed?.access_token) return parsed.access_token as string
        }
      } catch {}
      return null
    }

    const load = async () => {
      const token = tokenFromStorage()
      if (!token) {
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
