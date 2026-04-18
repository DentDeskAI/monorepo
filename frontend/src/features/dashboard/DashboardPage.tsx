import { useQuery } from '@tanstack/react-query'
import {
  Users, CalendarCheck, MessageSquare,
  TrendingUp, Clock, CheckCircle2,
} from 'lucide-react'
import { get } from '@/lib/api'
import { formatDate, formatTime, STATUS_COLORS, STATUS_LABELS, cn } from '@/lib/utils'
import type { Appointment } from '@/types/domain'
import type { PaginatedResponse } from '@/types/api'

// ─── KPI Card ─────────────────────────────────────────────────────────────────

interface KpiCardProps {
  label: string
  value: string | number
  icon: React.ComponentType<{ className?: string }>
  trend?: string
  color: string
}

function KpiCard({ label, value, icon: Icon, trend, color }: KpiCardProps) {
  return (
    <div className="bg-white rounded-xl p-5 shadow-card border border-slate-100">
      <div className="flex items-start justify-between">
        <div>
          <p className="text-sm text-slate-500 font-medium">{label}</p>
          <p className="text-2xl font-bold text-slate-800 mt-1">{value}</p>
          {trend && (
            <div className="flex items-center gap-1 mt-2">
              <TrendingUp className="w-3.5 h-3.5 text-green-500" />
              <span className="text-xs text-green-600 font-medium">{trend}</span>
            </div>
          )}
        </div>
        <div className={cn('w-10 h-10 rounded-xl flex items-center justify-center flex-shrink-0', color)}>
          <Icon className="w-5 h-5" />
        </div>
      </div>
    </div>
  )
}

// ─── Recent Appointments ──────────────────────────────────────────────────────

function AppointmentRow({ appt }: { appt: Appointment }) {
  return (
    <div className="flex items-center gap-4 py-3 border-b border-slate-100 last:border-0">
      {/* Time */}
      <div className="text-center min-w-[44px]">
        <p className="text-xs font-semibold text-brand-600">{formatTime(appt.starts_at)}</p>
        <p className="text-[10px] text-slate-400">{formatDate(appt.starts_at, 'dd MMM')}</p>
      </div>

      {/* Doctor color dot */}
      <div
        className="w-2 h-2 rounded-full flex-shrink-0"
        style={{ background: appt.doctor?.color ?? '#94a3b8' }}
      />

      {/* Patient + title */}
      <div className="flex-1 min-w-0">
        <p className="text-sm font-medium text-slate-800 truncate">
          {appt.patient?.first_name} {appt.patient?.last_name}
        </p>
        <p className="text-xs text-slate-400 truncate">{appt.title || 'Без описания'}</p>
      </div>

      {/* Status badge */}
      <span className={cn('text-xs font-medium px-2 py-0.5 rounded-full flex-shrink-0',
        STATUS_COLORS[appt.status])}>
        {STATUS_LABELS[appt.status]}
      </span>
    </div>
  )
}

// ─── Dashboard Page ───────────────────────────────────────────────────────────

export function DashboardPage() {
  const today = new Date().toISOString().split('T')[0]

  const { data: todayAppts } = useQuery({
    queryKey: ['appointments', 'today'],
    queryFn: () =>
      get<PaginatedResponse<Appointment>>(
        `/appointments?date_from=${today}T00:00:00Z&date_to=${today}T23:59:59Z&page_size=10`,
      ),
  })

  const { data: patients } = useQuery({
    queryKey: ['patients', 'count'],
    queryFn: () => get<PaginatedResponse<unknown>>('/patients?page_size=1'),
  })

  const kpis: KpiCardProps[] = [
    {
      label: 'Пациентов всего',
      value: patients?.total ?? '—',
      icon: Users,
      trend: '+4 за неделю',
      color: 'bg-blue-50 text-blue-600',
    },
    {
      label: 'Приёмов сегодня',
      value: todayAppts?.total ?? '—',
      icon: CalendarCheck,
      color: 'bg-brand-50 text-brand-600',
    },
    {
      label: 'Новых сообщений',
      value: '12',
      icon: MessageSquare,
      trend: '3 без ответа',
      color: 'bg-amber-50 text-amber-600',
    },
    {
      label: 'Завершено сегодня',
      value: todayAppts?.data.filter((a) => a.status === 'completed').length ?? 0,
      icon: CheckCircle2,
      color: 'bg-green-50 text-green-600',
    },
  ]

  return (
    <div className="space-y-6 max-w-6xl">
      {/* KPI Grid */}
      <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
        {kpis.map((k) => (
          <KpiCard key={k.label} {...k} />
        ))}
      </div>

      {/* Today's appointments */}
      <div className="bg-white rounded-xl shadow-card border border-slate-100">
        <div className="flex items-center justify-between px-5 py-4 border-b border-slate-100">
          <div className="flex items-center gap-2">
            <Clock className="w-4 h-4 text-slate-500" />
            <h2 className="text-sm font-semibold text-slate-800">Приёмы сегодня</h2>
          </div>
          <span className="text-xs text-slate-400">
            {formatDate(new Date().toISOString())}
          </span>
        </div>

        <div className="px-5 divide-y divide-slate-50">
          {!todayAppts || todayAppts.data.length === 0 ? (
            <div className="py-10 text-center">
              <CalendarCheck className="w-8 h-8 text-slate-200 mx-auto mb-2" />
              <p className="text-sm text-slate-400">На сегодня приёмов нет</p>
            </div>
          ) : (
            todayAppts.data.map((appt) => (
              <AppointmentRow key={appt.id} appt={appt} />
            ))
          )}
        </div>
      </div>
    </div>
  )
}
