package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/dentdesk/backend/internal/domain"
)

// ─── User Repository ──────────────────────────────────────────────────────────

type userRepo struct{ db *gorm.DB }

func NewUserRepository(db *gorm.DB) *userRepo {
	return &userRepo{db: db}
}

func (r *userRepo) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	var u domain.User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&u).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &u, err
}

func (r *userRepo) Create(ctx context.Context, user *domain.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

// ─── Patient Repository ───────────────────────────────────────────────────────

type patientRepo struct{ db *gorm.DB }

func NewPatientRepository(db *gorm.DB) PatientRepository {
	return &patientRepo{db: db}
}

func (r *patientRepo) FindByID(ctx context.Context, clinicID, patientID uuid.UUID) (*domain.Patient, error) {
	var p domain.Patient
	err := r.db.WithContext(ctx).
		Where("clinic_id = ? AND id = ?", clinicID, patientID).
		First(&p).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &p, err
}

func (r *patientRepo) FindByPhone(ctx context.Context, clinicID uuid.UUID, phone string) (*domain.Patient, error) {
	var p domain.Patient
	err := r.db.WithContext(ctx).
		Where("clinic_id = ? AND phone = ?", clinicID, phone).
		First(&p).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &p, err
}

func (r *patientRepo) List(ctx context.Context, clinicID uuid.UUID, page, pageSize int) ([]domain.Patient, int64, error) {
	var patients []domain.Patient
	var total int64

	offset := (page - 1) * pageSize
	q := r.db.WithContext(ctx).Where("clinic_id = ?", clinicID)

	if err := q.Model(&domain.Patient{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&patients).Error; err != nil {
		return nil, 0, err
	}

	return patients, total, nil
}

func (r *patientRepo) Create(ctx context.Context, patient *domain.Patient) error {
	return r.db.WithContext(ctx).Create(patient).Error
}

func (r *patientRepo) Update(ctx context.Context, patient *domain.Patient) error {
	return r.db.WithContext(ctx).
		Where("clinic_id = ?", patient.ClinicID).
		Save(patient).Error
}

func (r *patientRepo) Delete(ctx context.Context, clinicID, patientID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("clinic_id = ? AND id = ?", clinicID, patientID).
		Delete(&domain.Patient{}).Error
}

// ─── Appointment Repository ───────────────────────────────────────────────────

type appointmentRepo struct{ db *gorm.DB }

func NewAppointmentRepository(db *gorm.DB) AppointmentRepository {
	return &appointmentRepo{db: db}
}

func (r *appointmentRepo) FindByID(ctx context.Context, clinicID, id uuid.UUID) (*domain.Appointment, error) {
	var a domain.Appointment
	err := r.db.WithContext(ctx).
		Preload("Patient").
		Preload("Doctor").
		Where("clinic_id = ? AND id = ?", clinicID, id).
		First(&a).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &a, err
}

func (r *appointmentRepo) ListByClinic(ctx context.Context, clinicID uuid.UUID, f AppointmentFilter) ([]domain.Appointment, error) {
	var appts []domain.Appointment

	q := r.db.WithContext(ctx).
		Preload("Patient").
		Preload("Doctor").
		Where("clinic_id = ?", clinicID)

	if f.DoctorID != nil {
		q = q.Where("doctor_id = ?", *f.DoctorID)
	}
	if f.PatientID != nil {
		q = q.Where("patient_id = ?", *f.PatientID)
	}
	if f.DateFrom != nil {
		q = q.Where("starts_at >= ?", *f.DateFrom)
	}
	if f.DateTo != nil {
		q = q.Where("starts_at <= ?", *f.DateTo)
	}
	if f.Status != nil {
		q = q.Where("status = ?", *f.Status)
	}

	offset := 0
	if f.Page > 1 {
		offset = (f.Page - 1) * f.PageSize
	}
	if f.PageSize == 0 {
		f.PageSize = 50
	}

	err := q.Order("starts_at ASC").Offset(offset).Limit(f.PageSize).Find(&appts).Error
	return appts, err
}

func (r *appointmentRepo) Create(ctx context.Context, appt *domain.Appointment) error {
	return r.db.WithContext(ctx).Create(appt).Error
}

func (r *appointmentRepo) Update(ctx context.Context, appt *domain.Appointment) error {
	return r.db.WithContext(ctx).
		Where("clinic_id = ?", appt.ClinicID).
		Save(appt).Error
}

func (r *appointmentRepo) Delete(ctx context.Context, clinicID, id uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("clinic_id = ? AND id = ?", clinicID, id).
		Delete(&domain.Appointment{}).Error
}

// ─── MessageLog Repository ────────────────────────────────────────────────────

type messageLogRepo struct{ db *gorm.DB }

func NewMessageLogRepository(db *gorm.DB) MessageLogRepository {
	return &messageLogRepo{db: db}
}

func (r *messageLogRepo) Create(ctx context.Context, msg *domain.MessageLog) error {
	return r.db.WithContext(ctx).Create(msg).Error
}

func (r *messageLogRepo) ListByPatient(ctx context.Context, clinicID, patientID uuid.UUID, page, pageSize int) ([]domain.MessageLog, int64, error) {
	var msgs []domain.MessageLog
	var total int64
	offset := (page - 1) * pageSize

	q := r.db.WithContext(ctx).Where("clinic_id = ? AND patient_id = ?", clinicID, patientID)
	if err := q.Model(&domain.MessageLog{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Order("created_at ASC").Offset(offset).Limit(pageSize).Find(&msgs).Error
	return msgs, total, err
}

func (r *messageLogRepo) ListByClinic(ctx context.Context, clinicID uuid.UUID, page, pageSize int) ([]domain.MessageLog, int64, error) {
	var msgs []domain.MessageLog
	var total int64
	offset := (page - 1) * pageSize

	q := r.db.WithContext(ctx).Where("clinic_id = ?", clinicID)
	if err := q.Model(&domain.MessageLog{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&msgs).Error
	return msgs, total, err
}

func (r *messageLogRepo) UpdateStatus(ctx context.Context, wamid string, status domain.MessageStatus) error {
	return r.db.WithContext(ctx).
		Model(&domain.MessageLog{}).
		Where("whats_app_message_id = ?", wamid).
		Update("status", status).Error
}

// ─── Clinic Repository ────────────────────────────────────────────────────────

type clinicRepo struct{ db *gorm.DB }

func NewClinicRepository(db *gorm.DB) ClinicRepository {
	return &clinicRepo{db: db}
}

func (r *clinicRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.Clinic, error) {
	var c domain.Clinic
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&c).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &c, err
}

func (r *clinicRepo) FindByPhoneNumberID(ctx context.Context, phoneNumberID string) (*domain.Clinic, error) {
	// TODO: Store phone_number_id on Clinic model when supporting multiple clinics per deployment.
	// For MVP single-tenant-per-deployment, return the first active clinic.
	var c domain.Clinic
	err := r.db.WithContext(ctx).Where("is_active = true").First(&c).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("no active clinic found for phone_number_id %s", phoneNumberID)
	}
	return &c, err
}

func (r *clinicRepo) Create(ctx context.Context, clinic *domain.Clinic) error {
	return r.db.WithContext(ctx).Create(clinic).Error
}

// ─── Sentinel errors ──────────────────────────────────────────────────────────

var ErrNotFound = errors.New("record not found")
