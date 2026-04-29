-- migrations/002_seq_ids.up.sql
-- Add stable integer serial IDs to scheduling tables.
-- The /api/schedule/* endpoints use integer IDs (matching MacDent convention).
-- For local/mock clinics these serials serve as the integer identity.

ALTER TABLE doctors      ADD COLUMN IF NOT EXISTS seq_id SERIAL;
ALTER TABLE patients     ADD COLUMN IF NOT EXISTS seq_id SERIAL;
ALTER TABLE appointments ADD COLUMN IF NOT EXISTS seq_id SERIAL;

CREATE UNIQUE INDEX IF NOT EXISTS idx_doctors_seq_id      ON doctors(seq_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_patients_seq_id     ON patients(seq_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_appointments_seq_id ON appointments(seq_id);
