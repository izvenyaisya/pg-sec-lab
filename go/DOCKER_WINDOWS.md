# Docker Testing на Windows

## Требования

- Docker Desktop для Windows
- PowerShell или CMD
- (Опционально) WSL2 для лучшей производительности

## Установка Docker Desktop

1. Скачать с https://www.docker.com/products/docker-desktop
2. Установить и запустить
3. Убедиться, что Docker работает:
```powershell
docker --version
docker-compose --version
```

## Быстрый старт

### PowerShell

```powershell
# Перейти в директорию проекта
cd d:\projects\study\course-4

# Запустить тесты
docker-compose up --build

# После завершения посмотреть результаты
cat test-results\summary.txt
cat output\generated.sql
```

### CMD

```cmd
cd d:\projects\study\course-4
docker-compose up --build
type test-results\summary.txt
```

## Использование без Makefile

Windows не поддерживает `make` по умолчанию, используйте команды напрямую:

### Запуск тестов
```powershell
docker-compose up --build --abort-on-container-exit pg-sec-lab
```

### Генерация SQL
```powershell
docker-compose run --rm pg-sec-lab generate --policy /app/policy.yaml --out /app/output/generated.sql
```

### Verify
```powershell
# Сначала запустить БД
docker-compose up -d postgres

# Затем verify
docker-compose run --rm pg-sec-lab verify `
  --policy /app/policy.yaml `
  --dsn "postgres://postgres:postgres@postgres:5432/testdb?sslmode=disable"
```

### Analyze
```powershell
docker-compose up -d postgres

docker-compose run --rm pg-sec-lab analyze `
  --dsn "postgres://postgres:postgres@postgres:5432/testdb?sslmode=disable" `
  --out /app/output/report.json
```

### Подключение к БД
```powershell
# Тестовая БД
docker-compose exec postgres psql -U postgres -d testdb

# Продакшн БД
docker-compose exec postgres-prod psql -U produser -d proddb
```

### Интерактивный shell
```powershell
docker-compose run --rm pg-sec-lab sh
```

## Просмотр результатов

### PowerShell с форматированием
```powershell
# Результаты тестов
Get-Content test-results\summary.txt

# SQL (первые 30 строк)
Get-Content output\generated.sql -Head 30

# JSON отчёт
Get-Content output\report.json | ConvertFrom-Json | ConvertTo-Json -Depth 10
```

### Логи
```powershell
# Все логи
docker-compose logs

# Только pg-sec-lab
docker-compose logs pg-sec-lab

# Следить за логами
docker-compose logs -f
```

## Управление

### Запуск
```powershell
# Все сервисы
docker-compose up -d

# Только БД
docker-compose up -d postgres postgres-prod
```

### Остановка
```powershell
# Остановить
docker-compose down

# Остановить и удалить volumes
docker-compose down -v
```

### Статус
```powershell
docker-compose ps
docker ps
```

## Очистка

```powershell
# Полная очистка
docker-compose down -v --rmi all
Remove-Item -Recurse -Force output, test-results -ErrorAction SilentlyContinue

# Очистка Docker системы
docker system prune -f
```

## Troubleshooting

### Docker Desktop не запускается

1. Проверьте WSL2:
```powershell
wsl --status
wsl --update
```

2. Перезапустите Docker Desktop

### Порты заняты

Измените порты в `docker-compose.yml`:
```yaml
ports:
  - "15432:5432"  # Вместо 5432
```

### Ошибки прав доступа

Запустите PowerShell от имени администратора:
```powershell
Start-Process powershell -Verb RunAs
```

### Медленная работа

1. Используйте WSL2 (рекомендуется)
2. Выделите больше ресурсов в Docker Desktop:
   - Settings → Resources
   - CPU: 4+ cores
   - Memory: 4+ GB

### Проблемы с сетью

```powershell
# Пересоздать сеть
docker network prune
docker-compose down
docker-compose up
```

## PowerShell скрипты

Создайте `run-tests.ps1`:
```powershell
#!/usr/bin/env pwsh

Write-Host "Запуск тестов pg-sec-lab..." -ForegroundColor Green

docker-compose up --build --abort-on-container-exit pg-sec-lab

if ($LASTEXITCODE -eq 0) {
    Write-Host "`n✅ Тесты завершены успешно!" -ForegroundColor Green
    Get-Content test-results\summary.txt
} else {
    Write-Host "`n❌ Тесты завершились с ошибками" -ForegroundColor Red
    docker-compose logs pg-sec-lab
}
```

Запуск:
```powershell
.\run-tests.ps1
```

Создайте `analyze.ps1`:
```powershell
#!/usr/bin/env pwsh

Write-Host "Запуск анализа конфигурации..." -ForegroundColor Green

docker-compose up -d postgres

Start-Sleep -Seconds 5

docker-compose run --rm pg-sec-lab analyze `
  --dsn "postgres://postgres:postgres@postgres:5432/testdb?sslmode=disable" `
  --out /app/output/report.json

if (Test-Path "output\report.json") {
    Write-Host "`n✅ Отчёт создан!" -ForegroundColor Green
    Get-Content output\report.json | ConvertFrom-Json | ConvertTo-Json -Depth 10
}

docker-compose down
```

## Альтернатива: WSL2

Для лучшей производительности используйте WSL2:

```powershell
# Установить WSL2
wsl --install

# Перейти в WSL
wsl

# В Linux shell
cd /mnt/d/projects/study/course-4
make test
```

## Batch файлы (CMD)

Создайте `run-tests.bat`:
```batch
@echo off
echo Запуск тестов pg-sec-lab...

docker-compose up --build --abort-on-container-exit pg-sec-lab

if %ERRORLEVEL% EQU 0 (
    echo.
    echo Тесты завершены успешно!
    type test-results\summary.txt
) else (
    echo.
    echo Тесты завершились с ошибками
    docker-compose logs pg-sec-lab
)

pause
```

## Полезные команды

```powershell
# Посмотреть образы
docker images | Select-String pg-sec-lab

# Посмотреть контейнеры
docker ps -a | Select-String pg-sec-lab

# Посмотреть volumes
docker volume ls | Select-String course-4

# Размер использованного места
docker system df

# Логи конкретного контейнера
docker logs pg-sec-lab-postgres

# Выполнить команду в контейнере
docker exec -it pg-sec-lab-postgres psql -U postgres
```

## Рекомендации для Windows

1. ✅ Используйте WSL2 для Docker Desktop
2. ✅ Выделите достаточно ресурсов (4+ GB RAM)
3. ✅ Храните проект на диске с SSD
4. ✅ Используйте PowerShell 7+ или Windows Terminal
5. ✅ Закройте антивирус для директории проекта (ускорит работу)

## Заключение

Docker отлично работает на Windows! Основные команды:

```powershell
# Запустить тесты
docker-compose up --build

# Посмотреть результаты  
cat test-results\summary.txt

# Очистить
docker-compose down -v
```

Для дополнительной информации см. [DOCKER_QUICKSTART.md](DOCKER_QUICKSTART.md)
