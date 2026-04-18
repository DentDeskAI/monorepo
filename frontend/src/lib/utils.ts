import { type ClassValue, clsx } from 'clsx'
import { twMerge } from 'tailwind-merge'
import { format, formatDistanceToNow, parseISO } from 'date-fns'
import { ru } from 'date-fns/locale'

// shadcn/ui standard cn() helper
export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

// ─── Date helpers ─────────────────────────────────────────────────────────────

export function formatDate(iso: string, fmt = 'dd MMM yyyy') {
  return format(parseISO(iso), fmt, { locale: ru })
}

export function formatTime(iso: string) {
  return format(parseISO(iso), 'HH:mm')
}

export function formatDatetime(iso: string) {
  return format(parseISO(iso), 'dd MMM yyyy, HH:mm', { locale: ru })
}

export function timeAgo(iso: string) {
  return formatDistanceToNow(parseISO(iso), { addSuffix: true, locale: ru })
}

// ─── String helpers ───────────────────────────────────────────────────────────

export function initials(firstName: string, lastName?: string) {
  const a = firstName.charAt(0).toUpperCase()
  const b = lastName ? lastName.charAt(0).toUpperCase() : ''
  return a + b
}

export function truncate(str: string, maxLen: number) {
  if (str.length <= maxLen) return str
  return str.slice(0, maxLen) + '…'
}

// ─── Appointment status helpers ───────────────────────────────────────────────

export const STATUS_LABELS: Record<string, string> = {
  scheduled:   'Запланировано',
  confirmed:   'Подтверждено',
  in_progress: 'В процессе',
  completed:   'Завершено',
  cancelled:   'Отменено',
  no_show:     'Не явился',
}

export const STATUS_COLORS: Record<string, string> = {
  scheduled:   'bg-blue-100 text-blue-700',
  confirmed:   'bg-brand-100 text-brand-700',
  in_progress: 'bg-amber-100 text-amber-700',
  completed:   'bg-green-100 text-green-700',
  cancelled:   'bg-red-100 text-red-700',
  no_show:     'bg-slate-100 text-slate-600',
}
