# PG SecureLab Web UI

Веб-интерфейс для визуализации отчетов безопасности PostgreSQL.

## Структура

```
web/
├── api/          # Go API сервер
│   ├── server.go # HTTP handler
│   └── go.mod
└── frontend/     # Next.js React приложение
    ├── app/      # Next.js App Router
    ├── components/
    ├── types/
    └── package.json
```

## Быстрый старт

### 1. Запуск API сервера

```bash
cd web/api
go run server.go
# API будет доступно на http://localhost:8080
```

### 2. Запуск фронтенда

```bash
cd web/frontend
npm install
npm run dev
# UI будет доступен на http://localhost:3000
```

## API Endpoints

- `POST /api/analyze` - Анализ PostgreSQL по DSN
  ```json
  { "dsn": "postgres://user:pass@host:5432/dbname" }
  ```

- `POST /api/upload` - Загрузка готового JSON отчета
  ```
  multipart/form-data с полем "report"
  ```

- `GET /api/health` - Проверка здоровья сервиса

## Фронтенд возможности

- 📊 **Обзор** - общая информация о БД
- 👥 **Роли** - таблица и граф ролей/привилегий
- 🔒 **RLS** - статус Row Level Security
- ⚠️ **Нарушения** - список проблем безопасности
- 🎯 **What-If** - симулятор видимости данных

## Технологии

**Backend:**
- Go 1.21+
- pgx v5 - PostgreSQL драйвер
- rs/cors - CORS middleware

**Frontend:**
- Next.js 14 (App Router)
- TypeScript
- Tailwind CSS
- react-force-graph - визуализация графов
- Recharts - диаграммы
