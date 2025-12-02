# Архитектура pg-sec-lab

## Обзор

`pg-sec-lab` реализует подход Policy-as-Code для управления безопасностью PostgreSQL. Инструмент преобразует декларативные политики в SQL-скрипты и обеспечивает их проверку.

## Архитектурные принципы

### 1. Separation of Concerns

Проект разделён на чёткие модули:

- **cmd/** - CLI интерфейс (cobra commands)
- **internal/policy/** - Модель данных и загрузка политик
- **internal/generator/** - Генерация SQL
- **internal/verifier/** - Проверка политик
- **internal/configcheck/** - Анализ конфигурации

### 2. Declarative over Imperative

Вместо написания SQL вручную, пользователь описывает желаемое состояние в YAML:

```yaml
# Декларативно
roles:
  analyst:
    login: false
    privileges:
      - object: "public.orders"
        actions: ["SELECT"]
```

Инструмент сам генерирует необходимый SQL.

### 3. Validation First

Политики валидируются до генерации SQL:
- RLS политики должны иметь select_policy
- Маски должны иметь exposed_as
- Объекты должны существовать

## Компоненты

### Policy Loader (internal/policy)

**Назначение**: Загрузка и валидация YAML-политик

**Структура данных**:
```go
type Policy struct {
    Metadata  Metadata
    Tenants   TenantConfig
    Roles     map[string]Role
    Tables    map[string]TablePolicy
}
```

**Валидация**:
- Проверка обязательных полей
- Валидация связей между сущностями
- Проверка корректности SQL-выражений (basic)

### SQL Generator (internal/generator)

**Назначение**: Преобразование Policy в SQL-скрипты

**Процесс генерации**:
1. Создание ролей (`CREATE ROLE`)
2. Установка атрибутов ролей (LOGIN, CREATEDB)
3. Включение RLS (`ALTER TABLE ... ENABLE ROW LEVEL SECURITY`)
4. Создание политик (`CREATE POLICY`)
5. Создание представлений для маскировки (`CREATE VIEW`)
6. Выдача привилегий (`GRANT`)

**Безопасность**:
- Все идентификаторы экранируются через `pqQuoteIdent()`
- Использование FORCE ROW LEVEL SECURITY
- Минимальные привилегии для ролей

### Verifier (internal/verifier)

**Назначение**: Проверка корректности политик на тестовой БД

**Алгоритм**:
1. Создание временной схемы (`pg_sec_lab_test_*`)
2. Создание тестовых таблиц
3. Вставка тестовых данных с разными tenant_id
4. Применение сгенерированных политик
5. Проверка изоляции данных между тенантами
6. Очистка (DROP SCHEMA CASCADE)

**Проверки**:
- ✅ RLS работает (каждый тенант видит только свои данные)
- ✅ Роли имеют правильные привилегии
- ✅ Маскировка применяется корректно

### Config Checker (internal/configcheck)

**Назначение**: Анализ существующей конфигурации PostgreSQL

**Собираемая информация**:
- Версия PostgreSQL
- Важные параметры безопасности (SSL, password_encryption, logging)
- Список ролей с атрибутами
- Список таблиц с информацией о RLS
- Привилегии ролей

**Findings** (обнаруженные проблемы):
- `NO_RLS` - таблица без RLS
- `SSL_DISABLED` - SSL отключён
- `SUPERUSER_LOGIN` - superuser с возможностью входа
- `BYPASS_RLS` - роль может обходить RLS

## Потоки данных

### Генерация SQL

```
policy.yaml
    ↓
[policy.Load()]
    ↓
Policy struct
    ↓
[generator.GenerateSQL()]
    ↓
SQL string
    ↓
output.sql / stdout
```

### Проверка политик

```
policy.yaml + DSN
    ↓
[policy.Load()]
    ↓
Policy struct
    ↓
[verifier.Verify()]
    ├─ Connect to DB
    ├─ Create test schema
    ├─ Create test tables
    ├─ Insert test data
    ├─ Apply policies
    ├─ Run tests
    └─ Cleanup
    ↓
Success/Failure
```

### Анализ конфигурации

```
DSN
    ↓
[configcheck.Analyze()]
    ├─ Query pg_settings
    ├─ Query pg_roles
    ├─ Query pg_class
    └─ Query information_schema
    ↓
Report struct
    ↓
[json.Marshal()]
    ↓
report.json
```

## Расширяемость

### Добавление новых типов политик

1. Расширить `policy.Policy` структуру
2. Добавить валидацию в `policy.validate()`
3. Реализовать генератор SQL в `generator/`
4. Добавить проверки в `verifier/`

Пример (audit policies):
```go
// internal/policy/model.go
type AuditPolicy struct {
    Enabled bool   `yaml:"enabled"`
    Table   string `yaml:"table"`
}

type Policy struct {
    // ... existing fields
    Audit AuditPolicy `yaml:"audit"`
}

// internal/generator/generator.go
func generateAudit(p *policy.Policy) string {
    if !p.Audit.Enabled {
        return ""
    }
    // Generate audit trigger SQL
}
```

### Добавление новых команд

```go
// cmd/export.go
package cmd

var exportCmd = &cobra.Command{
    Use:   "export",
    Short: "Export current database policies to YAML",
    RunE:  runExport,
}

func init() {
    rootCmd.AddCommand(exportCmd)
}

func runExport(cmd *cobra.Command, args []string) error {
    // Implementation
}
```

## Безопасность

### SQL Injection Prevention

- Все пользовательские идентификаторы экранируются
- Параметризованные запросы в verifier и analyzer
- Валидация входных данных

### Privilege Escalation Prevention

- Роли создаются с минимальными привилегиями
- NOLOGIN по умолчанию для служебных ролей
- FORCE ROW LEVEL SECURITY предотвращает обход

### Secrets Management

- DSN передаётся через CLI флаги
- Рекомендуется использование переменных окружения
- Никакого хранения учётных данных в коде/конфигах

## Performance

### Оптимизации

1. **Batch operations**: Группировка SQL-операций
2. **Connection pooling**: Использование pgx connection pool для analyze
3. **Minimal queries**: Только необходимые запросы к pg_catalog

### Ограничения

- Verifier создаёт временную схему (требует места)
- Analyzer делает несколько SELECT запросов (может быть медленным на больших БД)

## Testing Strategy

### Unit Tests

```go
// internal/generator/generator_test.go
func TestGenerateSQL(t *testing.T) {
    p := &policy.Policy{
        // ... test data
    }
    sql, err := GenerateSQL(p)
    assert.NoError(t, err)
    assert.Contains(t, sql, "CREATE ROLE")
}
```

### Integration Tests

```go
// internal/verifier/verifier_test.go
func TestVerify(t *testing.T) {
    // Requires running PostgreSQL
    conn := setupTestDB(t)
    defer conn.Close()
    
    p := loadTestPolicy(t)
    err := Verify(context.Background(), p, conn)
    assert.NoError(t, err)
}
```

### End-to-End Tests

```bash
# test.sh
./pg-sec-lab generate --policy test_policy.yaml --out test_out.sql
./pg-sec-lab verify --policy test_policy.yaml --dsn "$TEST_DSN"
./pg-sec-lab analyze --dsn "$TEST_DSN" --out test_report.json
```

## Roadmap

### Planned Features

- [ ] Export existing DB policies to YAML
- [ ] Diff between policy and actual state
- [ ] Support for INSERT/UPDATE/DELETE RLS policies
- [ ] Column-level permissions
- [ ] Dynamic data masking functions
- [ ] Policy versioning and migrations
- [ ] Web UI for policy management
- [ ] Terraform provider
- [ ] Kubernetes operator

### Future Improvements

- [ ] Better error messages with suggestions
- [ ] Policy templates library
- [ ] CI/CD integration examples
- [ ] Performance benchmarks
- [ ] Multi-database support (MySQL, Oracle)

## Contributing

При добавлении новых функций следуйте этим принципам:

1. **Keep it simple**: Не усложняйте без необходимости
2. **Test everything**: Покрытие тестами > 80%
3. **Document changes**: Обновляйте README и EXAMPLES
4. **Security first**: Всегда думайте о безопасности
5. **Backward compatible**: Не ломайте существующие API

## Resources

- [PostgreSQL RLS Documentation](https://www.postgresql.org/docs/current/ddl-rowsecurity.html)
- [PostgreSQL Security Best Practices](https://www.postgresql.org/docs/current/security.html)
- [pgx Driver](https://github.com/jackc/pgx)
- [Cobra CLI Framework](https://github.com/spf13/cobra)
