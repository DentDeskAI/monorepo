-- migrations/001_init.up.sql
-- DentDesk MVP: единая миграция, идемпотентная.

CREATE EXTENSION IF NOT EXISTS "pgcrypto";
CREATE EXTENSION IF NOT EXISTS "citext";

-- ===== Клиники (tenant) =====
CREATE TABLE IF NOT EXISTS clinics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    timezone TEXT NOT NULL DEFAULT 'Asia/Almaty',
    whatsapp_phone_id TEXT UNIQUE,
    scheduler_type TEXT NOT NULL DEFAULT 'local' CHECK (scheduler_type IN ('local','mock','macdent')),
    macdent_base_url TEXT,
    macdent_api_key TEXT,
    working_hours JSONB NOT NULL DEFAULT '{"mon":["09:00","20:00"],"tue":["09:00","20:00"],"wed":["09:00","20:00"],"thu":["09:00","20:00"],"fri":["09:00","20:00"],"sat":["10:00","18:00"],"sun":null}'::jsonb,
    slot_duration_min INT NOT NULL DEFAULT 30,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ===== Сотрудники CRM =====
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    clinic_id UUID NOT NULL REFERENCES clinics(id) ON DELETE CASCADE,
    email CITEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    role TEXT NOT NULL CHECK (role IN ('owner','admin','operator')),
    name TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ===== Врачи =====
CREATE TABLE IF NOT EXISTS doctors (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    clinic_id UUID NOT NULL REFERENCES clinics(id) ON DELETE CASCADE,
    external_id TEXT,
    name TEXT NOT NULL,
    specialty TEXT,
    active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(clinic_id, external_id)
);

-- ===== Кресла =====
CREATE TABLE IF NOT EXISTS chairs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    clinic_id UUID NOT NULL REFERENCES clinics(id) ON DELETE CASCADE,
    external_id TEXT,
    name TEXT NOT NULL,
    active BOOLEAN NOT NULL DEFAULT TRUE
);

-- ===== Пациенты =====
CREATE TABLE IF NOT EXISTS patients (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    clinic_id UUID NOT NULL REFERENCES clinics(id) ON DELETE CASCADE,
    phone TEXT NOT NULL,
    name TEXT,
    external_id TEXT,
    language TEXT NOT NULL DEFAULT 'ru' CHECK (language IN ('ru','kk')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(clinic_id, phone)
);
CREATE INDEX IF NOT EXISTS idx_patients_phone ON patients(clinic_id, phone);

-- ===== Диалоги =====
CREATE TABLE IF NOT EXISTS conversations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    clinic_id UUID NOT NULL REFERENCES clinics(id) ON DELETE CASCADE,
    patient_id UUID NOT NULL REFERENCES patients(id) ON DELETE CASCADE,
    status TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active','handoff','closed')),
    context JSONB NOT NULL DEFAULT '{}'::jsonb,
    last_message_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(clinic_id, patient_id)
);
CREATE INDEX IF NOT EXISTS idx_conv_clinic_lastmsg ON conversations(clinic_id, last_message_at DESC);

-- ===== Сообщения =====
CREATE TABLE IF NOT EXISTS messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    conversation_id UUID NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    wa_message_id TEXT UNIQUE,
    direction TEXT NOT NULL CHECK (direction IN ('inbound','outbound')),
    sender TEXT NOT NULL CHECK (sender IN ('patient','bot','operator')),
    body TEXT NOT NULL,
    meta JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_messages_conv ON messages(conversation_id, created_at);

-- ===== Записи =====
CREATE TABLE IF NOT EXISTS appointments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    clinic_id UUID NOT NULL REFERENCES clinics(id) ON DELETE CASCADE,
    patient_id UUID NOT NULL REFERENCES patients(id) ON DELETE CASCADE,
    doctor_id UUID REFERENCES doctors(id) ON DELETE SET NULL,
    chair_id UUID REFERENCES chairs(id) ON DELETE SET NULL,
    external_id TEXT,
    starts_at TIMESTAMPTZ NOT NULL,
    ends_at TIMESTAMPTZ NOT NULL,
    service TEXT,
    status TEXT NOT NULL DEFAULT 'scheduled' CHECK (status IN ('scheduled','confirmed','cancelled','completed','no_show')),
    source TEXT NOT NULL DEFAULT 'bot' CHECK (source IN ('bot','operator','import')),
    reminder_24h_sent_at TIMESTAMPTZ,
    reminder_2h_sent_at TIMESTAMPTZ,
    followup_sent_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (ends_at > starts_at)
);
CREATE INDEX IF NOT EXISTS idx_appt_clinic_starts ON appointments(clinic_id, starts_at);
CREATE INDEX IF NOT EXISTS idx_appt_doctor_starts ON appointments(doctor_id, starts_at);
CREATE INDEX IF NOT EXISTS idx_appt_patient ON appointments(patient_id);
-- Предотвращаем пересечение записей на одного врача (в пределах tenant'а)
CREATE INDEX IF NOT EXISTS idx_appt_doctor_active
    ON appointments(doctor_id, starts_at, ends_at)
    WHERE status IN ('scheduled','confirmed');

-- ===== Seed: демо-клиника для локальной разработки =====
-- Пароль demo1234, bcrypt cost 10 — сгенерировано заранее.
INSERT INTO clinics (id, name, timezone, scheduler_type)
VALUES ('11111111-1111-1111-1111-111111111111', 'Demo Dental Almaty', 'Asia/Almaty', 'mock')
ON CONFLICT DO NOTHING;

INSERT INTO users (clinic_id, email, password_hash, role, name)
VALUES (
    '11111111-1111-1111-1111-111111111111',
    'admin@demo.kz',
    '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy',
    'owner',
    'Демо Админ'
)
ON CONFLICT DO NOTHING;

INSERT INTO doctors (clinic_id, name, specialty)
VALUES
  ('11111111-1111-1111-1111-111111111111', 'Айгерим Касымова', 'therapist'),
  ('11111111-1111-1111-1111-111111111111', 'Бауыржан Ахметов', 'surgeon'),
  ('11111111-1111-1111-1111-111111111111', 'Динара Серикова', 'orthodontist')
ON CONFLICT DO NOTHING;

INSERT INTO chairs (clinic_id, name)
VALUES
  ('11111111-1111-1111-1111-111111111111', 'Кабинет 1'),
  ('11111111-1111-1111-1111-111111111111', 'Кабинет 2')
ON CONFLICT DO NOTHING;
