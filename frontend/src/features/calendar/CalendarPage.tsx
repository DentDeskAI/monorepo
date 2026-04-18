import { useState, useMemo } from 'react'
import { useQuery } from '@tanstack/react-query'
import { ChevronLeft, ChevronRight, CalendarDays, Loader2 } from 'lucide-react'
import {
  startOfWeek, endOfWeek, eachDayOfInterval,
  format, addWeeks, subWeeks, isSameDay, parseISO,
} from 'date-fns'
import { ru } from 'date-fns/locale'
import { get } from '@/lib/api'
import { cn, STATUS_LABELS } from '@/lib/utils'
import type { Appointment } from '@/types/domain'
import type { PaginatedResponse } from '@/types/api'

// ─── Hours to display on the grid ─────────────────────────────────────────────
const HOURS = Array.from({ length: 11 }, (_, i) => i + 8) // 08:00 – 18:00

function getMinuteOffset(iso: string) {
  const d = parseISO(iso)
  return d.getHours() * 60 + d.getMinutes()
}

function getDurationMins(starts: string, ends: string) {
  return Math.max(30, (parseISO(ends).getTime() - parseISO(starts).getTime()) / 60_000)
}

const SLOT_HEIGHT = 64 // px per hour

// ─── Appointment chip ─────────────────────────────────────────────────────────

function ApptChip({ appt }: { appt: Appointment }) {
  const top  = ((getMinuteOffset(appt.starts_at) - 8 * 60) / 60) * SLOT_HEIGHT
  const height = (getDurationMins(appt.starts_at, appt.ends_at) / 60) * SLOT_HEIGHT

  const color = appt.doctor?.color ?? '#14b8a6'

  return (
    <div
      className="absolute left-1 right-1 rounded-lg px-2 py-1 overflow-hidden
                 text-white text-xs cursor-pointer hover:brightness-110 transition-all
                 shadow-sm z-10"
      style={{ top, height: Math.max(28, height), background: color + 'e6' }}
      title={`${appt.patient?.first_name} ${appt.patient?.last_name} — ${appt.title}`}
    >
      <p className="font-semibold truncate">
        {appt.patient?.first_name} {appt.patient?.last_name}
      </p>
      {height > 44 && (
        <p className="opacity-80 truncate">{appt.title || STATUS_LABELS[appt.status]}</p>
      )}
    </div>
  )
}

// ─── Calendar Page ────────────────────────────────────────────────────────────

export function CalendarPage() {
  const [currentWeek, setCurrentWeek] = useState(new Date())

  const weekStart = startOfWeek(currentWeek, { weekStartsOn: 1 }) // Monday
  const weekEnd   = endOfWeek(currentWeek, { weekStartsOn: 1 })
  const days      = eachDayOfInterval({ start: weekStart, end: weekEnd })

  // Load appointments for this week
  const dateFrom = format(weekStart, "yyyy-MM-dd'T'HH:mm:ss'Z'")
  const dateTo   = format(weekEnd,   "yyyy-MM-dd'T'23:59:59'Z'")

  const { data, isLoading } = useQuery({
    queryKey: ['appointments', 'week', dateFrom],
    queryFn: () =>
      get<PaginatedResponse<Appointment>>(
        `/appointments?date_from=${dateFrom}&date_to=${dateTo}&page_size=200`,
      ),
  })

  // Group by day
  const byDay = useMemo(() => {
    const map = new Map<string, Appointment[]>()
    for (const d of days) map.set(format(d, 'yyyy-MM-dd'), [])
    for (const appt of data?.data ?? []) {
      const key = format(parseISO(appt.starts_at), 'yyyy-MM-dd')
      if (map.has(key)) map.get(key)!.push(appt)
    }
    return map
  }, [data, days])

  const weekLabel = `${format(weekStart, 'd MMM', { locale: ru })} – ${format(weekEnd, 'd MMM yyyy', { locale: ru })}`

  return (
    <div className="max-w-7xl space-y-4">
      {/* Controls */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <CalendarDays className="w-5 h-5 text-slate-500" />
          <h2 className="text-base font-semibold text-slate-800 capitalize">{weekLabel}</h2>
        </div>

        <div className="flex items-center gap-2">
          <button
            onClick={() => setCurrentWeek((w) => subWeeks(w, 1))}
            className="w-8 h-8 rounded-lg border border-slate-200 flex items-center justify-center
                       hover:bg-slate-50 transition text-slate-600"
          >
            <ChevronLeft className="w-4 h-4" />
          </button>
          <button
            onClick={() => setCurrentWeek(new Date())}
            className="px-3 h-8 rounded-lg border border-slate-200 text-xs font-medium
                       hover:bg-slate-50 transition text-slate-600"
          >
            Сегодня
          </button>
          <button
            onClick={() => setCurrentWeek((w) => addWeeks(w, 1))}
            className="w-8 h-8 rounded-lg border border-slate-200 flex items-center justify-center
                       hover:bg-slate-50 transition text-slate-600"
          >
            <ChevronRight className="w-4 h-4" />
          </button>
        </div>
      </div>

      {/* Grid */}
      <div className="bg-white rounded-xl border border-slate-200 shadow-card overflow-hidden">
        {isLoading ? (
          <div className="flex items-center justify-center py-24 gap-2 text-slate-400">
            <Loader2 className="w-5 h-5 animate-spin" />
            <span className="text-sm">Загрузка расписания…</span>
          </div>
        ) : (
          <div className="flex overflow-x-auto">
            {/* Time gutter */}
            <div className="flex-shrink-0 w-14 border-r border-slate-100">
              {/* Header spacer */}
              <div className="h-12 border-b border-slate-100" />
              {HOURS.map((h) => (
                <div
                  key={h}
                  className="flex items-start justify-end pr-2 text-[10px] text-slate-400"
                  style={{ height: SLOT_HEIGHT }}
                >
                  <span className="-mt-2">{String(h).padStart(2, '0')}:00</span>
                </div>
              ))}
            </div>

            {/* Day columns */}
            {days.map((day) => {
              const dayKey    = format(day, 'yyyy-MM-dd')
              const isToday   = isSameDay(day, new Date())
              const dayAppts  = byDay.get(dayKey) ?? []

              return (
                <div key={dayKey} className="flex-1 min-w-[100px] border-r border-slate-100 last:border-0">
                  {/* Day header */}
                  <div className={cn(
                    'h-12 flex flex-col items-center justify-center border-b border-slate-100',
                    isToday && 'bg-brand-50',
                  )}>
                    <span className="text-[10px] text-slate-400 uppercase font-medium">
                      {format(day, 'EEE', { locale: ru })}
                    </span>
                    <span className={cn(
                      'text-sm font-semibold mt-0.5',
                      isToday ? 'text-brand-600' : 'text-slate-700',
                    )}>
                      {format(day, 'd')}
                    </span>
                  </div>

                  {/* Hour cells */}
                  <div
                    className="relative"
                    style={{ height: SLOT_HEIGHT * HOURS.length }}
                  >
                    {/* Hour grid lines */}
                    {HOURS.map((h, idx) => (
                      <div
                        key={h}
                        className="absolute left-0 right-0 border-b border-slate-100"
                        style={{ top: idx * SLOT_HEIGHT, height: SLOT_HEIGHT }}
                      />
                    ))}

                    {/* Today highlight */}
                    {isToday && (
                      <div className="absolute inset-0 bg-brand-500/[0.03] pointer-events-none" />
                    )}

                    {/* Appointment chips */}
                    {dayAppts.map((appt) => (
                      <ApptChip key={appt.id} appt={appt} />
                    ))}
                  </div>
                </div>
              )
            })}
          </div>
        )}
      </div>
    </div>
  )
}
