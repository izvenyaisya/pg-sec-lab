-- Инициализация "продакшн" базы для тестирования analyze

-- Создаем расширение для UUID
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Создаем разные схемы
CREATE SCHEMA IF NOT EXISTS app;
CREATE SCHEMA IF NOT EXISTS analytics;

-- Создаем роли
CREATE ROLE app_admin WITH LOGIN PASSWORD 'admin123' CREATEDB;
CREATE ROLE app_reader WITH LOGIN PASSWORD 'reader123';
CREATE ROLE app_writer NOLOGIN;

-- Назначаем роли
GRANT app_writer TO app_admin;

-- Создаем таблицы в разных схемах
CREATE TABLE public.users (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id uuid NOT NULL,
    username text NOT NULL,
    email text NOT NULL,
    password_hash text NOT NULL
);

CREATE TABLE public.sensitive_data (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid REFERENCES public.users(id),
    ssn text,
    credit_card text
);

CREATE TABLE app.logs (
    id serial PRIMARY KEY,
    user_id uuid,
    action text,
    created_at timestamp DEFAULT CURRENT_TIMESTAMP
);

-- Включаем RLS на некоторых таблицах
ALTER TABLE public.users ENABLE ROW LEVEL SECURITY;
CREATE POLICY user_isolation ON public.users
    FOR ALL
    USING (tenant_id = current_setting('app.tenant_id', true)::uuid);

-- Таблица БЕЗ RLS (для проверки findings)
CREATE TABLE public.unprotected_table (
    id serial PRIMARY KEY,
    data text
);

-- Выдаем привилегии
GRANT SELECT ON public.users TO app_reader;
GRANT SELECT, INSERT ON public.users TO app_writer;
GRANT ALL ON public.sensitive_data TO app_admin;
GRANT SELECT ON app.logs TO app_reader;

-- Вставляем тестовые данные
INSERT INTO public.users (id, tenant_id, username, email, password_hash) VALUES
    (gen_random_uuid(), '11111111-1111-1111-1111-111111111111', 'user1', 'user1@example.com', 'hash1'),
    (gen_random_uuid(), '22222222-2222-2222-2222-222222222222', 'user2', 'user2@example.com', 'hash2');

INSERT INTO public.unprotected_table (data) VALUES
    ('Unprotected data 1'),
    ('Unprotected data 2');

-- Настраиваем небезопасные параметры для тестирования findings
ALTER SYSTEM SET ssl = off;
SELECT pg_reload_conf();
