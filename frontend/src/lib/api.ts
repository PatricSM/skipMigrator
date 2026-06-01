import { supabase } from './supabase'

const API_BASE = import.meta.env.VITE_API_URL as string

async function authHeaders(): Promise<HeadersInit> {
  const { data } = await supabase.auth.getSession()
  return {
    Authorization: `Bearer ${data.session?.access_token ?? ''}`,
  }
}

export interface Migration {
  ID: string
  UserID: string
  Status: 'queued' | 'running' | 'success' | 'failed'
  CreatedAt: string
  UpdatedAt: string
  SourceZipPath: string
  OutputZipPath?: string | null
  BuildLog?: string | null
  ErrorMessage?: string | null
  PixelPerfect: boolean
  Validate: boolean
  SupabaseStrategy: string
  SupabaseProjectRef?: string | null
}

export async function createMigration(file: File, opts: { pixelPerfect: boolean; validate: boolean }) {
  const fd = new FormData()
  fd.append('source', file)
  fd.append('pixel_perfect', String(opts.pixelPerfect))
  fd.append('validate', String(opts.validate))
  fd.append('supabase_strategy', 'extract')
  const headers = await authHeaders()
  const r = await fetch(`${API_BASE}/api/migrations`, { method: 'POST', headers, body: fd })
  if (!r.ok) throw new Error(await r.text())
  return (await r.json()) as { id: string; status: string }
}

export async function listMigrations(): Promise<Migration[]> {
  const headers = await authHeaders()
  const r = await fetch(`${API_BASE}/api/migrations`, { headers })
  if (!r.ok) throw new Error(await r.text())
  return (await r.json()) as Migration[]
}

export async function getMigration(id: string): Promise<Migration> {
  const headers = await authHeaders()
  const r = await fetch(`${API_BASE}/api/migrations/${id}`, { headers })
  if (!r.ok) throw new Error(await r.text())
  return (await r.json()) as Migration
}

export function downloadURL(id: string): string {
  return `${API_BASE}/api/migrations/${id}/download`
}

export interface Me {
  user_id: string
  email: string
  is_super_admin: boolean
}

export async function getMe(): Promise<Me> {
  const headers = await authHeaders()
  const r = await fetch(`${API_BASE}/api/me`, { headers })
  if (!r.ok) throw new Error(await r.text())
  return (await r.json()) as Me
}

export interface AdminUser {
  id: string
  email: string
  created_at: string
  last_sign_in_at?: string
}

export async function adminListUsers(): Promise<AdminUser[]> {
  const headers = await authHeaders()
  const r = await fetch(`${API_BASE}/api/admin/users`, { headers })
  if (!r.ok) throw new Error(await r.text())
  return (await r.json()) as AdminUser[]
}

export async function adminCreateUser(email: string, password?: string) {
  const headers = { ...(await authHeaders()), 'Content-Type': 'application/json' }
  const r = await fetch(`${API_BASE}/api/admin/users`, {
    method: 'POST',
    headers,
    body: JSON.stringify({ email, password }),
  })
  if (!r.ok) throw new Error(await r.text())
  return (await r.json()) as { user: AdminUser; generated_password: string }
}

export async function adminDeleteUser(id: string) {
  const headers = await authHeaders()
  const r = await fetch(`${API_BASE}/api/admin/users/${id}`, { method: 'DELETE', headers })
  if (!r.ok) throw new Error(await r.text())
}
