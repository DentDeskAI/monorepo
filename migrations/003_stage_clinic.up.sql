-- migrations/003_stage_clinic.up.sql
-- Stage clinic "Стоматология Арман" — scheduler_type = 'local'.
-- Fully independent from MacDent. Login: stage@dentdesk.kz / stage1234

INSERT INTO clinics (id, name, timezone, scheduler_type, working_hours, slot_duration_min)
VALUES (
    '22222222-2222-2222-2222-222222222222',
    'Стоматология Арман',
    'Asia/Almaty',
    'local',
    '{"mon":["09:00","19:00"],"tue":["09:00","19:00"],"wed":["09:00","19:00"],"thu":["09:00","19:00"],"fri":["09:00","19:00"],"sat":["10:00","16:00"],"sun":null}'::jsonb,
    30
)
ON CONFLICT DO NOTHING;

-- Admin user — password: stage1234 (bcrypt cost 10)
INSERT INTO users (clinic_id, email, password_hash, role, name)
VALUES (
    '22222222-2222-2222-2222-222222222222',
    'stage@dentdesk.kz',
    '$2a$10$2p4zYyT7RWYMEYS04YH/nOLlnjFP//sR9rpKuWt7w3FLw90vv6G2O',
    'owner',
    'Арман Сейтқали'
)
ON CONFLICT DO NOTHING;

-- 4 Doctors
INSERT INTO doctors (id, clinic_id, name, specialty, active)
VALUES
    ('d1111111-0000-0000-0000-000000000001', '22222222-2222-2222-2222-222222222222', 'Айгерім Жақсыбекова', 'therapist',    TRUE),
    ('d1111111-0000-0000-0000-000000000002', '22222222-2222-2222-2222-222222222222', 'Бауыржан Оспанов',    'surgeon',      TRUE),
    ('d1111111-0000-0000-0000-000000000003', '22222222-2222-2222-2222-222222222222', 'Дина Сейткали',       'orthodontist', TRUE),
    ('d1111111-0000-0000-0000-000000000004', '22222222-2222-2222-2222-222222222222', 'Ерлан Қасымов',       'pediatric',    TRUE)
ON CONFLICT DO NOTHING;

-- 2 Chairs
INSERT INTO chairs (clinic_id, name, active)
VALUES
    ('22222222-2222-2222-2222-222222222222', 'Кабинет 1', TRUE),
    ('22222222-2222-2222-2222-222222222222', 'Кабинет 2', TRUE)
ON CONFLICT DO NOTHING;

-- 30 patients + 55 appointments
DO $$
DECLARE
    cid  UUID := '22222222-2222-2222-2222-222222222222';
    names  TEXT[] := ARRAY[
        'Айгерім Нұрланова',  'Бекзат Дюсупов',      'Гүлнар Ахметова',
        'Дәурен Сейткалиев',  'Еркежан Омарова',     'Жансая Қасенова',
        'Зарина Бектасова',   'Иман Жұмағалиев',     'Камила Мұхамедова',
        'Лейла Серікова',     'Мейіржан Дюсенов',    'Нұргүл Байжанова',
        'Олжас Оспанов',      'Перизат Алибекова',   'Рауан Жақыпова',
        'Сабина Тоқтарова',   'Тимур Нұрмаханов',   'Ұлбосын Кенжебаева',
        'Фатима Сейітова',    'Хасан Бегалиев',      'Шынар Дюсенова',
        'Ырысбек Жанабеков',  'Айдар Мәметов',       'Балжан Аманжолова',
        'Гани Рақымжанов',    'Дарига Сейткали',     'Еркін Байжанов',
        'Жібек Нұрланова',    'Зейнеп Омарбекова',   'Қанат Дауренов'
    ];
    phones TEXT[] := ARRAY[
        '+77011234501', '+77021234502', '+77031234503',
        '+77041234504', '+77051234505', '+77061234506',
        '+77071234507', '+77081234508', '+77091234509',
        '+77001234510', '+77011234511', '+77021234512',
        '+77031234513', '+77041234514', '+77051234515',
        '+77061234516', '+77071234517', '+77081234518',
        '+77091234519', '+77001234520', '+77011234521',
        '+77021234522', '+77031234523', '+77041234524',
        '+77051234525', '+77061234526', '+77071234527',
        '+77081234528', '+77091234529', '+77001234530'
    ];
    doc_ids UUID[] := ARRAY[
        'd1111111-0000-0000-0000-000000000001'::UUID,
        'd1111111-0000-0000-0000-000000000002'::UUID,
        'd1111111-0000-0000-0000-000000000003'::UUID,
        'd1111111-0000-0000-0000-000000000004'::UUID
    ];
    services TEXT[] := ARRAY[
        'Лечение кариеса',            'Удаление зуба',
        'Профессиональная чистка',    'Пломбирование',
        'Ортодонтическая консультация','Установка брекетов',
        'Детский осмотр',              'Реминерализация',
        'Имплантация (консультация)', 'Рентген зуба'
    ];
    pat_id UUID;
    i      INT;
BEGIN
    -- Insert 30 patients
    FOR i IN 1..30 LOOP
        INSERT INTO patients (clinic_id, phone, name, language)
        VALUES (cid, phones[i], names[i], 'ru')
        ON CONFLICT DO NOTHING;
    END LOOP;

    -- 30 completed past appointments (1–30 days ago, one per patient)
    FOR i IN 1..30 LOOP
        SELECT id INTO pat_id FROM patients
        WHERE clinic_id = cid AND phone = phones[i];

        INSERT INTO appointments (clinic_id, patient_id, doctor_id, starts_at, ends_at, service, status, source)
        VALUES (
            cid,
            pat_id,
            doc_ids[((i - 1) % 4) + 1],
            NOW() - (INTERVAL '1 day' * i) + INTERVAL '10 hours',
            NOW() - (INTERVAL '1 day' * i) + INTERVAL '10 hours 30 minutes',
            services[((i - 1) % 10) + 1],
            'completed',
            'operator'
        );
    END LOOP;

    -- 5 cancelled appointments (in the past)
    FOR i IN 1..5 LOOP
        SELECT id INTO pat_id FROM patients
        WHERE clinic_id = cid AND phone = phones[i];

        INSERT INTO appointments (clinic_id, patient_id, doctor_id, starts_at, ends_at, service, status, source)
        VALUES (
            cid,
            pat_id,
            doc_ids[((i - 1) % 4) + 1],
            NOW() - (INTERVAL '1 day' * (i + 5)) + INTERVAL '14 hours',
            NOW() - (INTERVAL '1 day' * (i + 5)) + INTERVAL '14 hours 30 minutes',
            services[((i - 1) % 10) + 1],
            'cancelled',
            'bot'
        );
    END LOOP;

    -- 20 future appointments (spread over next 20 days, 2 per day-slot)
    FOR i IN 1..20 LOOP
        SELECT id INTO pat_id FROM patients
        WHERE clinic_id = cid AND phone = phones[((i - 1) % 30) + 1];

        INSERT INTO appointments (clinic_id, patient_id, doctor_id, starts_at, ends_at, service, status, source)
        VALUES (
            cid,
            pat_id,
            doc_ids[((i - 1) % 4) + 1],
            NOW() + (INTERVAL '1 day' * i) + INTERVAL '9 hours'  + (INTERVAL '30 minutes' * ((i % 4) * 2)),
            NOW() + (INTERVAL '1 day' * i) + INTERVAL '9 hours 30 minutes' + (INTERVAL '30 minutes' * ((i % 4) * 2)),
            services[((i - 1) % 10) + 1],
            CASE WHEN i % 3 = 0 THEN 'confirmed' ELSE 'scheduled' END,
            'operator'
        );
    END LOOP;
END $$;
