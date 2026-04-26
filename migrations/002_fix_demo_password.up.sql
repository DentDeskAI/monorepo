-- Align the seeded demo user with the documented login credentials.
UPDATE users
SET password_hash = crypt('demo1234', gen_salt('bf'))
WHERE email = 'admin@demo.kz';
