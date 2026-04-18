import { Outlet } from 'react-router-dom'
import { Sidebar } from './Sidebar'
import { TopBar } from './TopBar'

/**
 * AppShell is the persistent admin layout.
 * Sidebar (fixed left) + TopBar (fixed top) + scrollable main content.
 */
export function AppShell() {
  return (
    <div className="flex h-screen overflow-hidden bg-surface-subtle">
      {/* ── Sidebar ── */}
      <Sidebar />

      {/* ── Main area ── */}
      <div
        className="flex flex-col flex-1 overflow-hidden"
        style={{ marginLeft: 'var(--sidebar-width)' }}
      >
        <TopBar />

        <main
          className="flex-1 overflow-y-auto p-6"
          style={{ marginTop: 'var(--topbar-height)' }}
        >
          <Outlet />
        </main>
      </div>
    </div>
  )
}
