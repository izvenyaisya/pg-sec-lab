# Docker Testing Guide для pg-sec-lab

## Описание

Docker-окружение для полного тестирования всех функций pg-sec-lab:
- Автоматическая сборка приложения
- Развертывание PostgreSQL для тестов
- Запуск всех команд (generate, verify, analyze)
- Проверка результатов

## Компоненты

### 1. Dockerfile
Многоэтапная сборка:
- **Stage 1 (builder)**: Компиляция Go-приложения
- **Stage 2 (final)**: Минималистичный образ с бинарником

### 2. docker-compose.yml
Три сервиса:
- **postgres**: Тестовая база данных (port 5432)
- **postgres-prod**: "Продакшн" база для analyze (port 5433)
- **pg-sec-lab**: Приложение с автотестами

### 3. docker-test.sh
Скрипт автоматического тестирования (11+ тестов):
- Проверка справки
- Генерация SQL
- Проверка содержимого SQL
- Создание таблиц
- Команда verify
- Команда analyze
- Проверка JSON-отчёта
- Применение политик
- Проверка ролей и RLS

### 4. test-init.sql
Инициализация тестовой БД:
- Создание таблиц customers, orders
- Вставка тестовых данных с разными tenant_id

### 5. test-prod-init.sql
Инициализация "продакшн" БД:
- Множество схем и таблиц
- Разные роли с привилегиями
- Таблицы с RLS и без (для проверки findings)
- Небезопасные настройки для тестирования

## Использование

### Быстрый старт

```bash
# Запуск всех тестов
docker-compose up --build

# Или только сборка и тесты pg-sec-lab
docker-compose up --build pg-sec-lab
```

### Пошаговое использование

#### 1. Сборка образа
```bash
docker build -t pg-sec-lab:latest .
```

#### 2. Запуск PostgreSQL
```bash
docker-compose up -d postgres postgres-prod
```

#### 3. Проверка готовности БД
```bash
docker-compose exec postgres pg_isready
```

#### 4. Запуск тестов
```bash
docker-compose up pg-sec-lab
```

#### 5. Просмотр результатов
```bash
# Логи тестов
docker-compose logs pg-sec-lab

# Сгенерированные файлы
ls -la output/

# Итоговый отчёт
cat test-results/summary.txt
```

### Интерактивный режим

```bash
# Запустить контейнер в интерактивном режиме
docker-compose run --rm pg-sec-lab sh

# Внутри контейнера можно запускать команды:
/app/pg-sec-lab --help
/app/pg-sec-lab generate --policy /app/policy.yaml
/app/pg-sec-lab analyze --dsn "postgres://postgres:postgres@postgres:5432/testdb"
```

### Отдельные команды

```bash
# Только генерация
docker-compose run --rm pg-sec-lab generate --policy /app/policy.yaml

# Только verify
docker-compose run --rm pg-sec-lab verify \
  --policy /app/policy.yaml \
  --dsn "postgres://postgres:postgres@postgres:5432/testdb?sslmode=disable"

# Только analyze
docker-compose run --rm pg-sec-lab analyze \
  --dsn "postgres://postgres:postgres@postgres:5432/testdb?sslmode=disable" \
  --out /app/output/report.json
```

## Структура результатов

После запуска тестов создаются директории:

```
output/
├── generated.sql       # Сгенерированный SQL
└── report.json        # JSON-отчёт анализа

test-results/
├── summary.txt        # Итоговый отчёт
├── verify-output.log  # Логи команды verify
├── analyze-output.log # Логи команды analyze
└── apply-sql.log      # Логи применения SQL
```

## Тесты

### Список тестов

1. ✅ Тест справки (--help)
2. ✅ Генерация SQL в stdout
3. ✅ Генерация SQL в файл
4. ✅ Проверка содержимого SQL (CREATE ROLE, RLS, POLICY, GRANT)
5. ✅ Создание тестовых таблиц
6. ✅ Команда verify
7. ✅ Команда analyze
8. ✅ Проверка JSON-отчёта (instance, roles, tables, findings)
9. ✅ Применение SQL
10. ✅ Проверка созданных ролей
11. ✅ Проверка включения RLS

### Ожидаемые результаты

При успешном прохождении всех тестов:
```
====================================
   ИТОГОВЫЕ РЕЗУЛЬТАТЫ
====================================
Всего тестов:    15+
Успешных:        15+
Провалено:       0
Процент успеха:  100.00%
====================================
```

## Troubleshooting

### PostgreSQL не запускается

```bash
# Проверить логи
docker-compose logs postgres

# Пересоздать контейнер
docker-compose down -v
docker-compose up -d postgres
```

### Тесты падают

```bash
# Посмотреть детальные логи
docker-compose logs pg-sec-lab

# Зайти в контейнер и запустить вручную
docker-compose run --rm pg-sec-lab sh
/app/docker-test.sh
```

### Ошибки подключения к БД

```bash
# Проверить сеть
docker network ls
docker network inspect course-4_pg-sec-lab-network

# Проверить доступность с другого контейнера
docker-compose run --rm pg-sec-lab pg_isready -h postgres -U postgres
```

### Очистка

```bash
# Остановить все контейнеры
docker-compose down

# Удалить volumes (БД)
docker-compose down -v

# Полная очистка
docker-compose down -v --rmi all
rm -rf output/ test-results/
```

## CI/CD интеграция

### GitHub Actions

```yaml
name: Docker Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      
      - name: Run Docker tests
        run: |
          docker-compose up --build --abort-on-container-exit
          docker-compose down -v
      
      - name: Upload test results
        uses: actions/upload-artifact@v2
        with:
          name: test-results
          path: |
            output/
            test-results/
```

### GitLab CI

```yaml
test:
  image: docker:latest
  services:
    - docker:dind
  script:
    - docker-compose up --build --abort-on-container-exit
  artifacts:
    paths:
      - output/
      - test-results/
```

## Дополнительные возможности

### Тестирование с разными версиями PostgreSQL

Измените в `docker-compose.yml`:
```yaml
postgres:
  image: postgres:15-alpine  # или 14, 13
```

### Добавление своих тестов

Отредактируйте `docker-test.sh`:
```bash
# Добавить новый тест
run_test "Тест 12: Моя проверка" \
    "команда для проверки"
```

### Профилирование производительности

```bash
# Добавить time к командам в docker-test.sh
time /app/pg-sec-lab generate --policy /app/policy.yaml
```

## Лучшие практики

1. **Всегда запускайте тесты в Docker** перед коммитом
2. **Проверяйте логи** при падении тестов
3. **Очищайте volumes** между запусками для чистоты тестов
4. **Используйте --build** чтобы пересобрать после изменений кода
5. **Сохраняйте test-results/** в версионном контроле (для CI)

## Производительность

Время выполнения на типичной машине:
- Сборка: ~30 секунд
- Запуск PostgreSQL: ~5 секунд
- Выполнение тестов: ~10-15 секунд
- **Итого**: ~1 минута

## Заключение

Docker-окружение предоставляет:
✅ Изолированное тестирование  
✅ Воспроизводимые результаты  
✅ Автоматизированные проверки  
✅ Готовность для CI/CD  
✅ Простоту запуска (`docker-compose up`)  
