#!/bin/sh

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Устанавливаем пароль для psql
export PGPASSWORD=postgres

# Функция для логирования
log() {
    echo "${BLUE}[$(date +'%Y-%m-%d %H:%M:%S')]${NC} $1"
}

success() {
    echo "${GREEN}✅ $1${NC}"
}

error() {
    echo "${RED}❌ $1${NC}"
}

warning() {
    echo "${YELLOW}⚠️  $1${NC}"
}

# Директории для результатов
RESULTS_DIR="/app/test-results"
OUTPUT_DIR="/app/output"
mkdir -p "$RESULTS_DIR" "$OUTPUT_DIR"

# Счетчики тестов
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# Функция для запуска теста
run_test() {
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    TEST_NAME=$1
    log "Запуск теста: $TEST_NAME"
    
    if eval "$2"; then
        success "$TEST_NAME"
        PASSED_TESTS=$((PASSED_TESTS + 1))
        return 0
    else
        error "$TEST_NAME"
        FAILED_TESTS=$((FAILED_TESTS + 1))
        return 1
    fi
}

# Ждем готовности PostgreSQL
log "Ожидание готовности PostgreSQL..."
sleep 5

# Проверяем доступность базы данных
if ! pg_isready -h postgres -U postgres > /dev/null 2>&1; then
    error "PostgreSQL недоступен!"
    exit 1
fi

success "PostgreSQL готов к работе"

echo ""
echo "======================================"
echo "   ТЕСТИРОВАНИЕ pg-sec-lab"
echo "======================================"
echo ""

# ===========================================
# ТЕСТ 1: Проверка справки
# ===========================================
run_test "Тест 1: Справка (--help)" \
    "/app/pg-sec-lab --help | grep -q 'Available Commands'"

# ===========================================
# ТЕСТ 2: Генерация SQL (stdout)
# ===========================================
run_test "Тест 2: Генерация SQL в stdout" \
    "/app/pg-sec-lab generate --policy /app/policy.yaml | grep -q 'CREATE ROLE'"

# ===========================================
# ТЕСТ 3: Генерация SQL (файл)
# ===========================================
run_test "Тест 3: Генерация SQL в файл" \
    "/app/pg-sec-lab generate --policy /app/policy.yaml --out $OUTPUT_DIR/generated.sql && test -f $OUTPUT_DIR/generated.sql"

# ===========================================
# ТЕСТ 4: Проверка содержимого SQL
# ===========================================
log "Тест 4: Проверка содержимого сгенерированного SQL"
if [ -f "$OUTPUT_DIR/generated.sql" ]; then
    TOTAL_TESTS=$((TOTAL_TESTS + 4))
    
    if grep -q "CREATE ROLE" "$OUTPUT_DIR/generated.sql"; then
        success "  - Содержит CREATE ROLE"
        PASSED_TESTS=$((PASSED_TESTS + 1))
    else
        error "  - Не содержит CREATE ROLE"
        FAILED_TESTS=$((FAILED_TESTS + 1))
    fi
    
    if grep -q "ENABLE ROW LEVEL SECURITY" "$OUTPUT_DIR/generated.sql"; then
        success "  - Содержит ENABLE ROW LEVEL SECURITY"
        PASSED_TESTS=$((PASSED_TESTS + 1))
    else
        error "  - Не содержит ENABLE ROW LEVEL SECURITY"
        FAILED_TESTS=$((FAILED_TESTS + 1))
    fi
    
    if grep -q "CREATE POLICY" "$OUTPUT_DIR/generated.sql"; then
        success "  - Содержит CREATE POLICY"
        PASSED_TESTS=$((PASSED_TESTS + 1))
    else
        error "  - Не содержит CREATE POLICY"
        FAILED_TESTS=$((FAILED_TESTS + 1))
    fi
    
    if grep -q "GRANT" "$OUTPUT_DIR/generated.sql"; then
        success "  - Содержит GRANT"
        PASSED_TESTS=$((PASSED_TESTS + 1))
    else
        error "  - Не содержит GRANT"
        FAILED_TESTS=$((FAILED_TESTS + 1))
    fi
fi

# ===========================================
# ТЕСТ 5: Создание тестовых таблиц
# ===========================================
log "Тест 5: Создание тестовых таблиц в PostgreSQL"
PSQL="psql -h postgres -U postgres -d testdb -t -A"

$PSQL <<EOF
CREATE TABLE IF NOT EXISTS customers (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id uuid NOT NULL,
    email text NOT NULL
);

CREATE TABLE IF NOT EXISTS orders (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id uuid NOT NULL,
    amount numeric NOT NULL
);

INSERT INTO customers (id, tenant_id, email) 
VALUES 
    (gen_random_uuid(), '11111111-1111-1111-1111-111111111111'::uuid, 'user1@tenant1.com'),
    (gen_random_uuid(), '22222222-2222-2222-2222-222222222222'::uuid, 'user2@tenant2.com');

INSERT INTO orders (id, tenant_id, amount)
VALUES
    (gen_random_uuid(), '11111111-1111-1111-1111-111111111111'::uuid, 100.00),
    (gen_random_uuid(), '22222222-2222-2222-2222-222222222222'::uuid, 200.00);
EOF

if [ $? -eq 0 ]; then
    success "Тестовые таблицы созданы"
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    PASSED_TESTS=$((PASSED_TESTS + 1))
else
    error "Ошибка создания таблиц"
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    FAILED_TESTS=$((FAILED_TESTS + 1))
fi

# ===========================================
# ТЕСТ 6: Команда verify
# ===========================================
log "Тест 6: Проверка политик (verify)"
if /app/pg-sec-lab verify \
    --policy /app/policy.yaml \
    --dsn "postgres://postgres@postgres:5432/testdb?sslmode=disable" \
    > "$RESULTS_DIR/verify-output.log" 2>&1; then
    success "Команда verify выполнена успешно"
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    PASSED_TESTS=$((PASSED_TESTS + 1))
else
    warning "Команда verify завершилась с ошибкой (может быть ожидаемо)"
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    FAILED_TESTS=$((FAILED_TESTS + 1))
    cat "$RESULTS_DIR/verify-output.log"
fi

# ===========================================
# ТЕСТ 7: Команда analyze  
# ===========================================
log "Тест 7: Анализ конфигурации (analyze)"

# Принудительно закрываем все неиспользуемые соединения к БД
psql -h postgres -U postgres -d testdb -c "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = 'testdb' AND pid <> pg_backend_pid() AND state = 'idle';" > /dev/null 2>&1 || true

# Небольшая задержка для гарантии освобождения соединений
sleep 1

if /app/pg-sec-lab analyze \
    --dsn "postgres://postgres@postgres:5432/testdb?sslmode=disable" \
    --out "$OUTPUT_DIR/report.json" \
    > "$RESULTS_DIR/analyze-output.log" 2>&1; then
    success "Команда analyze выполнена успешно"
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    PASSED_TESTS=$((PASSED_TESTS + 1))
else
    error "Команда analyze завершилась с ошибкой"
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    FAILED_TESTS=$((FAILED_TESTS + 1))
    cat "$RESULTS_DIR/analyze-output.log"
fi

# ===========================================
# ТЕСТ 8: Проверка JSON-отчёта
# ===========================================
log "Тест 8: Проверка содержимого JSON-отчёта"
if [ -f "$OUTPUT_DIR/report.json" ]; then
    TOTAL_TESTS=$((TOTAL_TESTS + 4))
    
    if grep -q '"instance"' "$OUTPUT_DIR/report.json"; then
        success "  - JSON содержит instance"
        PASSED_TESTS=$((PASSED_TESTS + 1))
    else
        error "  - JSON не содержит instance"
        FAILED_TESTS=$((FAILED_TESTS + 1))
    fi
    
    if grep -q '"roles"' "$OUTPUT_DIR/report.json"; then
        success "  - JSON содержит roles"
        PASSED_TESTS=$((PASSED_TESTS + 1))
    else
        error "  - JSON не содержит roles"
        FAILED_TESTS=$((FAILED_TESTS + 1))
    fi
    
    if grep -q '"tables"' "$OUTPUT_DIR/report.json"; then
        success "  - JSON содержит tables"
        PASSED_TESTS=$((PASSED_TESTS + 1))
    else
        error "  - JSON не содержит tables"
        FAILED_TESTS=$((FAILED_TESTS + 1))
    fi
    
    if grep -q '"findings"' "$OUTPUT_DIR/report.json"; then
        success "  - JSON содержит findings"
        PASSED_TESTS=$((PASSED_TESTS + 1))
    else
        error "  - JSON не содержит findings"
        FAILED_TESTS=$((FAILED_TESTS + 1))
    fi
    
    # Красиво выводим JSON
    log "Содержимое report.json:"
    cat "$OUTPUT_DIR/report.json"
fi

# ===========================================
# ТЕСТ 9: Применение сгенерированных политик
# ===========================================
log "Тест 9: Применение сгенерированного SQL"
if [ -f "$OUTPUT_DIR/generated.sql" ]; then
    if $PSQL -f "$OUTPUT_DIR/generated.sql" > "$RESULTS_DIR/apply-sql.log" 2>&1; then
        success "SQL политики применены успешно"
        TOTAL_TESTS=$((TOTAL_TESTS + 1))
        PASSED_TESTS=$((PASSED_TESTS + 1))
    else
        warning "Применение SQL завершилось с ошибками (возможно роли уже существуют)"
        TOTAL_TESTS=$((TOTAL_TESTS + 1))
        FAILED_TESTS=$((FAILED_TESTS + 1))
        cat "$RESULTS_DIR/apply-sql.log"
    fi
fi

# ===========================================
# ТЕСТ 10: Проверка созданных ролей
# ===========================================
log "Тест 10: Проверка созданных ролей"
ROLE_COUNT=$($PSQL -c "SELECT COUNT(*) FROM pg_roles WHERE rolname IN ('analyst', 'support', 'reporting_app')")

if [ "$ROLE_COUNT" -ge 1 ]; then
    success "Роли созданы (найдено: $ROLE_COUNT)"
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    PASSED_TESTS=$((PASSED_TESTS + 1))
else
    warning "Роли не найдены"
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    FAILED_TESTS=$((FAILED_TESTS + 1))
fi

# ===========================================
# ТЕСТ 11: Проверка RLS на таблицах
# ===========================================
log "Тест 11: Проверка включения RLS"
RLS_COUNT=$($PSQL -c "SELECT COUNT(*) FROM pg_class c JOIN pg_namespace n ON c.relnamespace = n.oid WHERE c.relrowsecurity = true AND n.nspname = 'public'")

if [ "$RLS_COUNT" -ge 1 ]; then
    success "RLS включен на таблицах (найдено: $RLS_COUNT)"
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    PASSED_TESTS=$((PASSED_TESTS + 1))
else
    warning "RLS не включен"
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    FAILED_TESTS=$((FAILED_TESTS + 1))
fi

# ===========================================
# Сохранение результатов
# ===========================================
cat > "$RESULTS_DIR/summary.txt" <<EOF
====================================
   ИТОГОВЫЕ РЕЗУЛЬТАТЫ ТЕСТИРОВАНИЯ
====================================

Дата: $(date)

Всего тестов:    $TOTAL_TESTS
Успешных:        $PASSED_TESTS
Провалено:       $FAILED_TESTS

Процент успеха:  $(awk "BEGIN {printf \"%.2f\", ($PASSED_TESTS/$TOTAL_TESTS)*100}")%

Сгенерированные файлы:
- $OUTPUT_DIR/generated.sql
- $OUTPUT_DIR/report.json

Логи:
- $RESULTS_DIR/verify-output.log
- $RESULTS_DIR/analyze-output.log
- $RESULTS_DIR/apply-sql.log
EOF

echo ""
echo "======================================"
echo "   ИТОГОВЫЕ РЕЗУЛЬТАТЫ"
echo "======================================"
cat "$RESULTS_DIR/summary.txt"
echo "======================================"
echo ""

# Выход с соответствующим кодом
if [ $FAILED_TESTS -eq 0 ]; then
    success "ВСЕ ТЕСТЫ ПРОЙДЕНЫ!"
    exit 0
else
    error "НЕКОТОРЫЕ ТЕСТЫ ПРОВАЛЕНЫ"
    exit 1
fi
