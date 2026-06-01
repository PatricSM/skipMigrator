import '@testing-library/jest-dom/vitest'

// Stub Vite env vars consumed by lib/api + lib/supabase so imports don't blow up
// during tests. The actual values don't matter — no network in unit tests.
;(globalThis as unknown as { __vitest_env__?: Record<string, string> }).__vitest_env__ = {
  VITE_API_URL: 'http://test',
  VITE_SUPABASE_URL: 'http://test',
  VITE_SUPABASE_ANON_KEY: 'test',
}
import.meta.env.VITE_API_URL = 'http://test'
import.meta.env.VITE_SUPABASE_URL = 'http://test'
import.meta.env.VITE_SUPABASE_ANON_KEY = 'test'
