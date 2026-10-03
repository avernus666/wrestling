# Архитектура Wrestling

## Общая схема

`React UI -> HTTP API -> service layer -> repository ports -> PostgreSQL`

Frontend остаётся визуально прежним: японская дневная/ночная тема, каталог, тренировки и прогресс. Backend добавлен отдельным слоем и не требует переписывать дизайн.

## Backend

- `server-go/cmd/server` — composition root.
- `internal/config` — конфигурация окружения.
- `internal/httpserver` — router, middleware, rate limiting, recovery.
- `internal/handlers` — HTTP boundary и валидация входных данных.
- `internal/services` — бизнес-логика.
- `internal/ports` — интерфейсы repository/service boundary.
- `internal/adapters/postgres` — PostgreSQL implementation.
- `internal/models` — API/domain models.
- `internal/db` — pool и checksum-миграции.
- `internal/observability` — metrics.

## Данные

Каталог Wrestling содержит 369 элементов:
- 187 стойка;
- 144 партер;
- 38 ОФП.

PostgreSQL хранит каталог, пользователей, сессии, прогресс, избранное, тренировки, навыки, community-контент, планы, readiness checks и audit events.

## SOLID

Dependency inversion реализован через `ports`: сервисы не знают о конкретном PostgreSQL repository. Composition root связывает реализации. HTTP, бизнес-логика и persistence разделены. Классы в стиле классического OOP не используются там, где функциональная композиция Go/React проще.

## Надёжность

- транзакционное создание и завершение workout session;
- foreign keys и CHECK constraints;
- индексы для основных фильтров;
- migration checksum verification;
- graceful shutdown;
- readiness/liveness endpoints;
- ограничение запросов;
- recovery от panic.
