// ─── Generic API response wrappers ───────────────────────────────────────────

export interface PaginatedResponse<T> {
  data: T[]
  total: number
  page: number
  page_size: number
}

export interface ApiError {
  error: string
}

// ─── Auth ─────────────────────────────────────────────────────────────────────

export interface AuthResponse {
  token: string
  expires_at: string
  user: {
    id: string
    clinic_id: string
    email: string
    first_name: string
    last_name: string
    role: string
  }
}

export interface LoginRequest {
  email: string
  password: string
}

export interface RegisterRequest {
  clinic_name: string
  email: string
  password: string
  first_name: string
  last_name: string
}

// ─── Appointments ─────────────────────────────────────────────────────────────

export interface CreateAppointmentRequest {
  patient_id: string
  doctor_id: string
  starts_at: string
  ends_at: string
  title?: string
  notes?: string
}

// ─── Patients ─────────────────────────────────────────────────────────────────

export interface CreatePatientRequest {
  first_name: string
  last_name: string
  phone: string
  email?: string
  date_of_birth?: string
  gender?: string
  notes?: string
}
