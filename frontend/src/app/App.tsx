import { useEffect } from 'react'
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import { useAuth } from '@/hooks/useAuth'

// Layouts
import { AppShell } from '@/components/layout/AppShell'

// Pages
import { DashboardPage } from '@/features/dashboard/DashboardPage'
import { DialogsPage }   from '@/features/dialogs/DialogsPage'
import { CalendarPage }  from '@/features/calendar/CalendarPage'
import { PatientsPage }  from '@/features/patients/PatientsPage'
import { LoginPage }     from '@/features/auth/LoginPage'
import { LandingPage }   from '@/features/landing/LandingPage'

// ─── Auth guard ───────────────────────────────────────────────────────────────

function RequireAuth({ children }: { children: React.ReactNode }) {
  const user = useAuth((s) => s.user)
  if (!user) return <Navigate to="/login" replace />
  return <>{children}</>
}

function GuestOnly({ children }: { children: React.ReactNode }) {
  const user = useAuth((s) => s.user)
  if (user) return <Navigate to="/app" replace />
  return <>{children}</>
}

// ─── App ──────────────────────────────────────────────────────────────────────

export default function App() {
  const init = useAuth((s) => s.init)
  const user = useAuth((s) => s.user)

  // Rehydrate auth state from localStorage on first render
  useEffect(() => { init() }, [init])

  return (
    <BrowserRouter>
      <Routes>
        {/* Public routes */}
        <Route path="/" element={<LandingPage />} />
        <Route
          path="/login"
          element={
            <GuestOnly>
              <LoginPage />
            </GuestOnly>
          }
        />

        {/* Protected routes — wrapped in AppShell (sidebar + topbar) */}
        <Route
          path="/app"
          element={
            <RequireAuth>
              <AppShell />
            </RequireAuth>
          }
        >
          <Route index element={<DashboardPage />} />
          <Route path="dialogs"     element={<DialogsPage />} />
          <Route path="calendar"    element={<CalendarPage />} />
          <Route path="patients"    element={<PatientsPage />} />
          <Route path="patients/:id" element={<PatientsPage />} />
          <Route path="*" element={<Navigate to="/app" replace />} />
        </Route>

        {/* Catch-all */}
        <Route path="*" element={<Navigate to={user ? '/app' : '/'} replace />} />
      </Routes>
    </BrowserRouter>
  )
}
