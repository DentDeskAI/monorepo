import { useLocation } from 'react-router-dom'
import { Bell } from 'lucide-react'
import { useAuth } from '@/hooks/useAuth'
import { initials } from '@/lib/utils'

const ROUTE_TITLES: Record<string, string> = {
  '/':          'Дашборд',
  '/dialogs':   'Диалоги',
  '/calendar':  'Календарь',
  '/patients':  'Пациенты',
  '/settings':  'Настройки',
}

export function TopBar() {
  const { pathname } = useLocation()
  const user = useAuth((s) => s.user)

  const title = ROUTE_TITLES[pathname] ?? 'DentDesk'

  return (
    <header
      className="fixed top-0 right-0 z-20 flex items-center justify-between
                 px-6 bg-white border-b border-slate-200 shadow-sm"
      style={{
        left: 'var(--sidebar-width)',
        height: 'var(--topbar-height)',
      }}
    >
      {/* ── Page title ── */}
      <h1 className="text-base font-semibold text-slate-800">{title}</h1>

      {/* ── Right controls ── */}
      <div className="flex items-center gap-3">
        {/* Notification bell — placeholder for future alerts */}
        <button
          className="relative w-8 h-8 flex items-center justify-center rounded-lg
                     text-slate-500 hover:bg-slate-100 transition-colors"
          aria-label="Уведомления"
        >
          <Bell className="w-4 h-4" />
          {/* Unread badge */}
          <span className="absolute top-1.5 right-1.5 w-1.5 h-1.5 rounded-full bg-brand-500" />
        </button>

        {/* User avatar */}
        <div className="flex items-center gap-2.5">
          <div
            className="w-8 h-8 rounded-full bg-brand-500 flex items-center justify-center
                       text-white text-xs font-semibold select-none"
          >
            {user ? initials(user.first_name, user.last_name) : '?'}
          </div>
          <div className="hidden sm:block">
            <p className="text-sm font-medium text-slate-700 leading-none">
              {user?.first_name} {user?.last_name}
            </p>
            <p className="text-xs text-slate-400 mt-0.5">{user?.role}</p>
          </div>
        </div>
      </div>
    </header>
  )
}
