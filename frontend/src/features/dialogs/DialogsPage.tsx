import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Send, Bot, Loader2, MessageSquare } from 'lucide-react'
import { toast } from 'sonner'
import { get, post } from '@/lib/api'
import { timeAgo, initials, cn } from '@/lib/utils'
import type { MessageLog, Patient } from '@/types/domain'
import type { PaginatedResponse } from '@/types/api'

// ─── Conversation thread ──────────────────────────────────────────────────────

function MessageBubble({ msg }: { msg: MessageLog }) {
  const isOut = msg.direction === 'outbound'

  return (
    <div className={cn('flex', isOut ? 'justify-end' : 'justify-start')}>
      <div
        className={cn(
          'max-w-[72%] px-4 py-2.5 rounded-2xl text-sm leading-relaxed',
          isOut
            ? 'bg-brand-500 text-white rounded-br-sm'
            : 'bg-white border border-slate-200 text-slate-800 rounded-bl-sm shadow-sm',
        )}
      >
        <p>{msg.body}</p>
        <div className={cn('flex items-center gap-1.5 mt-1', isOut ? 'justify-end' : 'justify-start')}>
          <span className={cn('text-[10px]', isOut ? 'text-brand-100' : 'text-slate-400')}>
            {timeAgo(msg.created_at)}
          </span>
          {msg.llm_used && (
            <Bot className={cn('w-3 h-3', isOut ? 'text-brand-200' : 'text-slate-400')} />
          )}
        </div>
      </div>
    </div>
  )
}

// ─── Contact list item ────────────────────────────────────────────────────────

interface ContactItemProps {
  patient?: Patient
  lastMessage?: MessageLog
  isSelected: boolean
  onClick: () => void
}

function ContactItem({ patient, lastMessage, isSelected, onClick }: ContactItemProps) {
  const name = patient
    ? `${patient.first_name} ${patient.last_name}`
    : lastMessage?.from_phone ?? 'Неизвестный'

  return (
    <button
      onClick={onClick}
      className={cn(
        'w-full flex items-center gap-3 px-4 py-3 text-left transition-colors',
        isSelected ? 'bg-brand-50 border-r-2 border-brand-500' : 'hover:bg-slate-50',
      )}
    >
      <div className="w-9 h-9 rounded-full bg-slate-200 flex items-center justify-center
                      text-slate-600 text-sm font-semibold flex-shrink-0">
        {patient ? initials(patient.first_name, patient.last_name) : '?'}
      </div>
      <div className="flex-1 min-w-0">
        <p className="text-sm font-medium text-slate-800 truncate">{name}</p>
        {lastMessage && (
          <p className="text-xs text-slate-400 truncate mt-0.5">
            {lastMessage.direction === 'outbound' ? 'Вы: ' : ''}
            {lastMessage.body}
          </p>
        )}
      </div>
      {lastMessage && (
        <span className="text-[10px] text-slate-400 flex-shrink-0">
          {timeAgo(lastMessage.created_at)}
        </span>
      )}
    </button>
  )
}

// ─── Dialogs Page ─────────────────────────────────────────────────────────────

export function DialogsPage() {
  const [selectedPatientId, setSelectedPatientId] = useState<string | null>(null)
  const [draft, setDraft] = useState('')
  const qc = useQueryClient()

  // Latest messages per contact (for sidebar)
  const { data: messages, isLoading } = useQuery({
    queryKey: ['messages', 'clinic'],
    queryFn: () => get<PaginatedResponse<MessageLog>>('/messages?page_size=50'),
    refetchInterval: 10_000, // poll every 10s for new messages
  })

  // Thread for selected patient
  const { data: thread } = useQuery({
    queryKey: ['messages', 'patient', selectedPatientId],
    queryFn: () =>
      get<PaginatedResponse<MessageLog>>(
        `/patients/${selectedPatientId}/messages?page_size=100`,
      ),
    enabled: !!selectedPatientId,
    refetchInterval: 5_000,
  })

  // Send message mutation
  const sendMutation = useMutation({
    mutationFn: () =>
      post('/messages/send', { patient_id: selectedPatientId, body: draft }),
    onSuccess: () => {
      setDraft('')
      qc.invalidateQueries({ queryKey: ['messages', 'patient', selectedPatientId] })
    },
    onError: () => toast.error('Ошибка при отправке сообщения'),
  })

  function handleSend(e: React.FormEvent) {
    e.preventDefault()
    if (!draft.trim() || !selectedPatientId) return
    sendMutation.mutate()
  }

  // Build deduplicated contact list from messages
  const contactMap = new Map<string | null, MessageLog>()
  for (const msg of messages?.data ?? []) {
    const key = msg.patient_id ?? msg.from_phone
    if (!contactMap.has(key)) contactMap.set(key, msg)
  }
  const contacts = Array.from(contactMap.values())

  const threadMessages = thread?.data ?? []

  return (
    <div className="max-w-6xl h-[calc(100vh-var(--topbar-height)-48px)]">
      <div className="flex h-full bg-white rounded-xl border border-slate-200 shadow-card overflow-hidden">

        {/* ── Contact list ── */}
        <div className="w-72 flex-shrink-0 border-r border-slate-100 flex flex-col">
          <div className="px-4 py-3 border-b border-slate-100">
            <h2 className="text-sm font-semibold text-slate-700">Диалоги</h2>
          </div>

          <div className="flex-1 overflow-y-auto divide-y divide-slate-50">
            {isLoading ? (
              <div className="flex items-center justify-center py-12">
                <Loader2 className="w-5 h-5 animate-spin text-slate-300" />
              </div>
            ) : contacts.length === 0 ? (
              <div className="py-12 text-center px-4">
                <MessageSquare className="w-7 h-7 text-slate-200 mx-auto mb-2" />
                <p className="text-sm text-slate-400">Нет сообщений</p>
              </div>
            ) : (
              contacts.map((msg) => (
                <ContactItem
                  key={msg.patient_id ?? msg.from_phone}
                  patient={msg.patient}
                  lastMessage={msg}
                  isSelected={selectedPatientId === msg.patient_id}
                  onClick={() => setSelectedPatientId(msg.patient_id ?? null)}
                />
              ))
            )}
          </div>
        </div>

        {/* ── Thread ── */}
        <div className="flex-1 flex flex-col min-w-0">
          {!selectedPatientId ? (
            <div className="flex-1 flex flex-col items-center justify-center text-center px-8">
              <MessageSquare className="w-12 h-12 text-slate-200 mb-3" />
              <p className="text-base font-medium text-slate-400">Выберите диалог</p>
              <p className="text-sm text-slate-300 mt-1">Входящие сообщения появятся автоматически</p>
            </div>
          ) : (
            <>
              {/* Messages */}
              <div className="flex-1 overflow-y-auto p-4 space-y-3">
                {threadMessages.map((msg) => (
                  <MessageBubble key={msg.id} msg={msg} />
                ))}
              </div>

              {/* Input */}
              <form
                onSubmit={handleSend}
                className="flex items-end gap-3 px-4 py-3 border-t border-slate-100"
              >
                <textarea
                  rows={1}
                  value={draft}
                  onChange={(e) => setDraft(e.target.value)}
                  onKeyDown={(e) => {
                    if (e.key === 'Enter' && !e.shiftKey) {
                      e.preventDefault()
                      handleSend(e)
                    }
                  }}
                  placeholder="Написать сообщение…"
                  className="flex-1 px-3 py-2.5 bg-slate-50 border border-slate-200 rounded-xl
                             text-sm resize-none focus:outline-none focus:ring-2
                             focus:ring-brand-400 transition"
                />
                <button
                  type="submit"
                  disabled={!draft.trim() || sendMutation.isPending}
                  className={cn(
                    'w-10 h-10 rounded-xl flex items-center justify-center flex-shrink-0 transition',
                    draft.trim()
                      ? 'bg-brand-500 hover:bg-brand-600 text-white'
                      : 'bg-slate-100 text-slate-300 cursor-not-allowed',
                  )}
                >
                  {sendMutation.isPending
                    ? <Loader2 className="w-4 h-4 animate-spin" />
                    : <Send className="w-4 h-4" />}
                </button>
              </form>
            </>
          )}
        </div>
      </div>
    </div>
  )
}
