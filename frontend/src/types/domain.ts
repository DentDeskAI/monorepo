// ─── Core domain types — mirror backend domain models ────────────────────────

export type UUID = string

export type UserRole = 'admin' | 'receptionist' | 'doctor'

export interface Clinic {
  id: UUID
  name: string
  slug: string
  phone: string
  email: string
  address: string
  logo_url: string
  is_active: boolean
  created_at: string
}

export interface User {
  id: UUID
  clinic_id: UUID
  email: string
  first_name: string
  last_name: string
  role: UserRole
  is_active: boolean
}

export interface Doctor {
  id: UUID
  clinic_id: UUID
  first_name: string
  last_name: string
  speciality: string
  avatar_url: string
  color: string // hex color for calendar display
}

export interface Patient {
  id: UUID
  clinic_id: UUID
  first_name: string
  last_name: string
  phone: string
  email: string
  date_of_birth?: string // ISO-8601
  gender: string
  notes: string
  created_at: string
}

export type AppointmentStatus =
  | 'scheduled'
  | 'confirmed'
  | 'in_progress'
  | 'completed'
  | 'cancelled'
  | 'no_show'

export interface Appointment {
  id: UUID
  clinic_id: UUID
  patient_id: UUID
  patient?: Patient
  doctor_id: UUID
  doctor?: Doctor
  starts_at: string // RFC3339
  ends_at: string
  status: AppointmentStatus
  title: string
  notes: string
  reminder_sent_at?: string
}

export type MessageDirection = 'inbound' | 'outbound'
export type MessageStatus = 'pending' | 'sent' | 'delivered' | 'read' | 'failed'

export interface MessageLog {
  id: UUID
  clinic_id: UUID
  patient_id?: UUID
  patient?: Patient
  wa_message_id: string
  from_phone: string
  to_phone: string
  direction: MessageDirection
  status: MessageStatus
  message_type: string
  body: string
  media_url?: string
  llm_used: boolean
  created_at: string
}
