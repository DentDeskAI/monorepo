import { NavLink, useNavigate } from 'react-router-dom'
import {
  LayoutDashboard,
  MessageSquare,
  CalendarDays,
  Users,
  Settings,
  LogOut,
  Stethoscope,
} from 'lucide-react'
import { useAuth } from '@/hooks/useAuth'
import { cn } from '@/lib/utils'

interface NavItem {
  label: string
  to: string
  icon: React.ComponentType<{ className?: string }>
}

const NAV_ITEMS: NavItem[] = [
  { label: 'Дашборд',   to: '/',         icon: LayoutDashboard },
  { label: 'Диалоги',   to: '/dialogs',  icon: MessageSquare   },
  { label: 'Календарь', to: '/calendar', icon: CalendarDays    },
  { label: 'Пациенты',  to: '/patients', icon: Users           },
]

export function Sidebar() {
  const { user, logout } = useAuth()
  const navigate = useNavigate()

  function handleLogout() {
    logout()
    navigate('/login')
  }

  return (
    <aside
      className="fixed inset-y-0 left-0 flex flex-col z-30"
      style={{ width: 'var(--sidebar-width)', background: 'var(--sidebar-bg)' }}
    >
      {/* ── Logo ── */}
      <div className="flex items-center gap-2.5 px-4 py-5 border-b border-white/10">
        <div className="w-8 h-8 rounded-lg bg-brand-500 flex items-center justify-center flex-shrink-0">
          <Stethoscope className="w-4 h-4 text-white" />
        </div>
        <div>
          <p className="text-white font-semibold text-sm leading-none">DentDesk</p>
          <p className="text-slate-400 text-xs mt-0.5 truncate max-w-[140px]">
            {user?.first_name} {user?.last_name}
          </p>
        </div>
      </div>

      {/* ── Navigation ── */}
      <nav className="flex-1 px-3 py-4 space-y-0.5 overflow-y-auto">
        {NAV_ITEMS.map((item) => (
          <NavLink
            key={item.to}
            to={item.to}
            end={item.to === '/'}
            className={({ isActive }) =>
              cn('sidebar-link', isActive && 'active')
            }
          >
            <item.icon className="w-4 h-4 flex-shrink-0" />
            {item.label}
          </NavLink>
        ))}
      </nav>

      {/* ── Bottom actions ── */}
      <div className="px-3 py-4 border-t border-white/10 space-y-0.5">
        <NavLink
          to="/settings"
          className={({ isActive }) => cn('sidebar-link', isActive && 'active')}
        >
          <Settings className="w-4 h-4 flex-shrink-0" />
          Настройки
        </NavLink>

        <button
          onClick={handleLogout}
          className="sidebar-link w-full text-left hover:text-red-400"
        >
          <LogOut className="w-4 h-4 flex-shrink-0" />
          Выйти
        </button>
      </div>
    </aside>
  )
}
