import { Routes, Route, Navigate } from 'react-router-dom'
import { useEffect, useState } from 'react'
import type { Session } from '@supabase/supabase-js'
import { supabase } from './lib/supabase'
import Landing from './pages/Landing'
import Login from './pages/Login'
import Dashboard from './pages/Dashboard'
import NewMigration from './pages/NewMigration'
import MigrationDetail from './pages/MigrationDetail'
import AdminUsers from './pages/AdminUsers'
import { useAutoTour } from './hooks/useAutoTour'

export default function App() {
  const [session, setSession] = useState<Session | null | undefined>(undefined)
  useAutoTour()

  useEffect(() => {
    supabase.auth.getSession().then(({ data }) => setSession(data.session))
    const { data: sub } = supabase.auth.onAuthStateChange((_e, s) => setSession(s))
    return () => sub.subscription.unsubscribe()
  }, [])

  if (session === undefined) {
    return <div className="grid h-screen place-items-center text-muted-foreground">Carregando…</div>
  }

  return (
    <Routes>
      <Route path="/" element={<Landing authed={!!session} />} />
      <Route path="/login" element={session ? <Navigate to="/app" replace /> : <Login />} />
      <Route path="/app" element={session ? <Dashboard /> : <Navigate to="/login" replace />} />
      <Route path="/app/new" element={session ? <NewMigration /> : <Navigate to="/login" replace />} />
      <Route path="/app/m/:id" element={session ? <MigrationDetail /> : <Navigate to="/login" replace />} />
      <Route path="/admin/users" element={session ? <AdminUsers /> : <Navigate to="/login" replace />} />
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  )
}
