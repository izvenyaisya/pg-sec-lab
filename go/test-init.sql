-- Инициализация тестовой базы данных

-- Создаем расширение для UUID (если не установлено)
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Создаем тестовые таблицы
CREATE TABLE IF NOT EXISTS customers (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id uuid NOT NULL,
    email text NOT NULL,
    created_at timestamp DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS orders (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id uuid NOT NULL,
    customer_id uuid REFERENCES customers(id),
    amount numeric NOT NULL,
    created_at timestamp DEFAULT CURRENT_TIMESTAMP
);

-- Вставляем тестовые данные
INSERT INTO customers (id, tenant_id, email) VALUES
    ('00000000-0000-0000-0000-000000000001', '11111111-1111-1111-1111-111111111111', 'alice@tenant1.com'),
    ('00000000-0000-0000-0000-000000000002', '11111111-1111-1111-1111-111111111111', 'bob@tenant1.com'),
    ('00000000-0000-0000-0000-000000000003', '22222222-2222-2222-2222-222222222222', 'charlie@tenant2.com');

INSERT INTO orders (id, tenant_id, customer_id, amount) VALUES
    ('00000000-0000-0000-0001-000000000001', '11111111-1111-1111-1111-111111111111', '00000000-0000-0000-0000-000000000001', 100.50),
    ('00000000-0000-0000-0001-000000000002', '11111111-1111-1111-1111-111111111111', '00000000-0000-0000-0000-000000000002', 250.75),
    ('00000000-0000-0000-0001-000000000003', '22222222-2222-2222-2222-222222222222', '00000000-0000-0000-0000-000000000003', 500.00);

-- Создаем тестового пользователя для проверки
-- Пользователь postgres уже существует и имеет доступ без пароля в Docker
