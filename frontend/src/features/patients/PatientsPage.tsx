import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Search, UserPlus, Phone, Mail, ChevronRight, Loader2 } from 'lucide-react'
import { toast } from 'sonner'
import { get, post } from '@/lib/api'
import { formatDate, initials, cn } from '@/lib/utils'
import type { Patient } from '@/types/domain'
import type { PaginatedResponse, CreatePatientRequest } from '@/types/api'

// ─── Create patient modal ─────────────────────────────────────────────────────

interface CreateModalProps {
  open: boolean
  onClose: () => void
}

function CreatePatientModal({ open, onClose }: CreateModalProps) {
  const qc = useQueryClient()
  const [form, setForm] = useState<CreatePatientRequest>({
    first_name: '', last_name: '', phone: '', email: '',
  })

  const mutation = useMutation({
    mutationFn: (data: CreatePatientRequest) => post<Patient>('/patients', data),
    onSuccess: () => {
      toast.success('Пациент добавлен')
      qc.invalidateQueries({ queryKey: ['patients'] })
      onClose()
      setForm({ first_name: '', last_name: '', phone: '', email: '' })
    },
    onError: () => toast.error('Ошибка при создании пациента'),
  })

  function field(key: keyof CreatePatientRequest) {
    return {
      value: form[key] ?? '',
      onChange: (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) =>
        setForm((f) => ({ ...f, [key]: e.target.value })),
    }
  }

  if (!open) return null

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 backdrop-blur-sm">
      <div className="bg-white rounded-2xl shadow-2xl w-full max-w-md mx-4 p-6">
        <h2 className="text-base font-semibold text-slate-800 mb-5">Новый пациент</h2>

        <div className="space-y-4">
          <div className="grid grid-cols-2 gap-3">
            {(['first_name', 'last_name'] as const).map((k) => (
              <div key={k}>
                <label className="text-xs font-medium text-slate-500 block mb-1">
                  {k === 'first_name' ? 'Имя *' : 'Фамилия *'}
                </label>
                <input
                  {...field(k)}
                  required
                  className="w-full px-3 py-2 border border-slate-200 rounded-lg text-sm
                             focus:outline-none focus:ring-2 focus:ring-brand-400 transition"
                />
              </div>
            ))}
          </div>

          {[
            { key: 'phone' as const,  label: 'Телефон (WhatsApp) *', type: 'tel' },
            { key: 'email' as const,  label: 'Email',                type: 'email' },
          ].map(({ key, label, type }) => (
            <div key={key}>
              <label className="text-xs font-medium text-slate-500 block mb-1">{label}</label>
              <input
                type={type}
                {...field(key)}
                className="w-full px-3 py-2 border border-slate-200 rounded-lg text-sm
                           focus:outline-none focus:ring-2 focus:ring-brand-400 transition"
              />
            </div>
          ))}

          <div>
            <label className="text-xs font-medium text-slate-500 block mb-1">Заметки</label>
            <textarea
              rows={3}
              value={form.notes ?? ''}
              onChange={(e) => setForm((f) => ({ ...f, notes: e.target.value }))}
              className="w-full px-3 py-2 border border-slate-200 rounded-lg text-sm
                         focus:outline-none focus:ring-2 focus:ring-brand-400 transition resize-none"
            />
          </div>
        </div>

        <div className="flex gap-3 mt-6">
          <button
            onClick={onClose}
            className="flex-1 py-2 border border-slate-200 rounded-lg text-sm text-slate-600
                       hover:bg-slate-50 transition"
          >
            Отмена
          </button>
          <button
            onClick={() => mutation.mutate(form)}
            disabled={mutation.isPending || !form.first_name || !form.phone}
            className="flex-1 py-2 bg-brand-500 hover:bg-brand-600 disabled:opacity-60
                       text-white text-sm font-medium rounded-lg transition flex items-center justify-center gap-2"
          >
            {mutation.isPending && <Loader2 className="w-3.5 h-3.5 animate-spin" />}
            Создать
          </button>
        </div>
      </div>
    </div>
  )
}

// ─── Patient row ─────────────────────────────────────────────────────────────

function PatientRow({ patient }: { patient: Patient }) {
  return (
    <div
      className="flex items-center gap-4 p-4 hover:bg-slate-50 cursor-pointer
                 border-b border-slate-100 last:border-0 transition-colors group"
    >
      {/* Avatar */}
      <div className="w-9 h-9 rounded-full bg-brand-100 text-brand-700 flex items-center
                      justify-center text-sm font-semibold flex-shrink-0">
        {initials(patient.first_name, patient.last_name)}
      </div>

      {/* Name + meta */}
      <div className="flex-1 min-w-0">
        <p className="text-sm font-medium text-slate-800">
          {patient.first_name} {patient.last_name}
        </p>
        <div className="flex items-center gap-3 mt-0.5">
          <span className="flex items-center gap-1 text-xs text-slate-400">
            <Phone className="w-3 h-3" /> {patient.phone}
          </span>
          {patient.email && (
            <span className="flex items-center gap-1 text-xs text-slate-400 truncate">
              <Mail className="w-3 h-3" /> {patient.email}
            </span>
          )}
        </div>
      </div>

      {/* Added date */}
      <p className="text-xs text-slate-400 flex-shrink-0 hidden sm:block">
        {formatDate(patient.created_at)}
      </p>

      <ChevronRight className="w-4 h-4 text-slate-300 group-hover:text-slate-400 flex-shrink-0" />
    </div>
  )
}

// ─── Patients Page ────────────────────────────────────────────────────────────

export function PatientsPage() {
  const [search, setSearch]       = useState('')
  const [page, setPage]           = useState(1)
  const [showCreate, setShowCreate] = useState(false)

  const PAGE_SIZE = 20

  const { data, isLoading } = useQuery({
    queryKey: ['patients', page, PAGE_SIZE],
    queryFn: () =>
      get<PaginatedResponse<Patient>>(`/patients?page=${page}&page_size=${PAGE_SIZE}`),
  })

  // Client-side search filter (for MVP — move to server-side for large datasets)
  const filtered = (data?.data ?? []).filter((p) => {
    const q = search.toLowerCase()
    return (
      p.first_name.toLowerCase().includes(q) ||
      p.last_name.toLowerCase().includes(q) ||
      p.phone.includes(q) ||
      (p.email?.toLowerCase().includes(q) ?? false)
    )
  })

  const totalPages = Math.ceil((data?.total ?? 0) / PAGE_SIZE)

  return (
    <>
      <CreatePatientModal open={showCreate} onClose={() => setShowCreate(false)} />

      <div className="max-w-5xl space-y-4">
        {/* Header */}
        <div className="flex items-center justify-between">
          <div>
            <h2 className="text-lg font-semibold text-slate-800">Пациенты</h2>
            <p className="text-sm text-slate-400 mt-0.5">
              {data?.total ?? 0} записей
            </p>
          </div>
          <button
            onClick={() => setShowCreate(true)}
            className="flex items-center gap-2 px-4 py-2 bg-brand-500 hover:bg-brand-600
                       text-white text-sm font-medium rounded-lg transition"
          >
            <UserPlus className="w-4 h-4" />
            Добавить
          </button>
        </div>

        {/* Search */}
        <div className="relative">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400" />
          <input
            type="search"
            placeholder="Поиск по имени, телефону или email..."
            value={search}
            onChange={(e) => { setSearch(e.target.value); setPage(1) }}
            className="w-full pl-9 pr-4 py-2.5 bg-white border border-slate-200 rounded-xl
                       text-sm focus:outline-none focus:ring-2 focus:ring-brand-400 transition"
          />
        </div>

        {/* Table */}
        <div className="bg-white rounded-xl shadow-card border border-slate-100 overflow-hidden">
          {isLoading ? (
            <div className="flex items-center justify-center py-16 gap-2 text-slate-400">
              <Loader2 className="w-5 h-5 animate-spin" />
              <span className="text-sm">Загрузка...</span>
            </div>
          ) : filtered.length === 0 ? (
            <div className="py-16 text-center">
              <p className="text-sm text-slate-400">Пациенты не найдены</p>
            </div>
          ) : (
            filtered.map((p) => <PatientRow key={p.id} patient={p} />)
          )}
        </div>

        {/* Pagination */}
        {totalPages > 1 && (
          <div className="flex items-center justify-between">
            <p className="text-xs text-slate-400">
              Страница {page} из {totalPages}
            </p>
            <div className="flex gap-2">
              <button
                onClick={() => setPage((p) => Math.max(1, p - 1))}
                disabled={page === 1}
                className={cn(
                  'px-3 py-1.5 text-xs rounded-lg border transition',
                  page === 1
                    ? 'border-slate-200 text-slate-300 cursor-not-allowed'
                    : 'border-slate-200 text-slate-600 hover:bg-slate-50',
                )}
              >
                ← Назад
              </button>
              <button
                onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
                disabled={page === totalPages}
                className={cn(
                  'px-3 py-1.5 text-xs rounded-lg border transition',
                  page === totalPages
                    ? 'border-slate-200 text-slate-300 cursor-not-allowed'
                    : 'border-slate-200 text-slate-600 hover:bg-slate-50',
                )}
              >
                Вперёд →
              </button>
            </div>
          </div>
        )}
      </div>
    </>
  )
}
