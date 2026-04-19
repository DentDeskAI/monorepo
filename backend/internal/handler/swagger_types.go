package handler

import "github.com/dentdesk/backend/internal/domain"

// ErrorResponse is a generic error payload.
type ErrorResponse struct {
	Error string `json:"error" example:"invalid request"`
}

// HealthResponse represents API health check status.
type HealthResponse struct {
	Status string `json:"status" example:"ok"`
}

// PatientListResponse is the paginated patient list payload.
type PatientListResponse struct {
	Data     []domain.Patient `json:"data"`
	Total    int64            `json:"total" example:"42"`
	Page     int              `json:"page" example:"1"`
	PageSize int              `json:"page_size" example:"20"`
}

// PatientDTO is a Swagger alias for domain.Patient.
type PatientDTO = domain.Patient

// AppointmentListResponse is the appointment list payload.
type AppointmentListResponse struct {
	Data []domain.Appointment `json:"data"`
}

// AppointmentDTO is a Swagger alias for domain.Appointment.
type AppointmentDTO = domain.Appointment

// StatusResponse is a generic success status payload.
type StatusResponse struct {
	Status string `json:"status" example:"sent"`
}
