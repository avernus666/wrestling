Wrestling — веб-приложение для систематизации техники борьбы, ОФП, домашних тренировок и личного прогресса.

Проект сохраняет существующий японский интерфейс Wrestling и добавляет полноценный серверный слой на Go + PostgreSQL. Backend отвечает за аккаунты, сессии, каталог, прогресс, тренировки, аналитику, навыки и community-функции.

Основные разделы
Стойка
Универсальная стойка, перемещения, контроль, проходы, броски, подсечки, подножки, зацепы и контратаки.

Партер
Позиции, переходы, свипы, удержания, удушающие и болевые. Опасные техники предназначены для постепенной отработки с тренером и подготовленным партнёром.

ОФП
Разминка, мобильность, сила, корпус, ноги, хват, резина, турник, брусья, скакалка, кондиционная работа и домашние тренировки.

Каталог
В текущем каталоге 369 элементов:

187 — стойка;
144 — партер;
38 — ОФП.
Каждый элемент содержит название, уровень сложности, группу, описание, урок и ссылку на обучающий поиск.

Backend
Backend построен на Go 1.23 и PostgreSQL.

Возможности
регистрация и вход;
криптографически случайные сессии;
bcrypt для паролей;
отзыв сессий;
профиль пользователя;
прогресс по технике;
избранное;
создание и завершение тренировочных сессий;
история тренировок;
аналитика минут, занятий и серий;
уровни навыков;
community-элементы;
комментарии;
жалобы;
каталог оборудования;
база безопасности;
тренировочные планы;
readiness checks;
audit events;
health/readiness/metrics endpoints.
API
Основные маршруты:

GET /api/health
GET /api/ready
GET /api/elements
GET /api/elements/{id}
GET /api/elements/category/{category}
GET /api/bars
GET /api/base
GET /api/safety
GET /api/stats
POST /api/auth/register
POST /api/auth/login
GET /api/auth/me
POST /api/auth/logout
GET /api/progress
PUT /api/progress
POST /api/workouts
POST /api/workouts/{id}
GET /api/history
GET /api/analytics
GET /api/skills
PUT /api/skills
community endpoints для пользовательского контента, комментариев и жалоб.
Полный контракт находится в openapi.yaml.

База данных
Миграции находятся в server-go/internal/db/migrations.

Миграции имеют checksum. Если уже применённый файл изменён, сервер останавливается вместо тихого изменения схемы.

Основные таблицы:

elements
bars
base_blocks
safety_items
users
sessions
user_progress
workout_sessions
workout_session_exercises
user_skills
community_elements
community_comments
community_reports
training_plans
training_plan_items
readiness_checks
audit_events
Безопасность
В проекте есть:

bcrypt cost 12;
токены с cryptographically secure random;
хранение только SHA-256 digest токена;
срок жизни и отзыв сессий;
лимит тела HTTP-запроса;
отдельный rate limit авторизации;
общий rate limit API;
request ID;
security headers;
CSP;
HSTS;
CORS allow-list;
SQL parameterization;
ownership checks;
transaction boundaries;
DB constraints;
migration checksum verification;
panic recovery;
graceful shutdown;
readiness/liveness;
community URL validation;
plain-text user generated content.
Подробности — SECURITY.md.

Домашние тренировки
Frontend содержит три этапа:

База — 30–40 минут;
Связки — 40–55 минут;
Раунды — 50–65 минут.
Есть короткая, стандартная и кондиционная сессии, readiness checklist, оборудование, недельный ритм и правила прогрессии.

Backend позволяет сохранять историю этих занятий и строить аналитику пользователя.

Запуск через Docker
Нужны Docker и Docker Compose.

docker compose up --build
После запуска:

приложение: http://localhost:8080
health: http://localhost:8080/api/health
readiness: http://localhost:8080/api/ready
OpenAPI: http://localhost:8080/openapi.yaml
Для production обязательно заменить локальные credentials PostgreSQL, включить TLS и задать точный CORS_ORIGINS.

Локальная разработка
Frontend:

npm install
npm start
Backend:

cd server-go
go mod download
go test ./...
go run ./cmd/server
Backend требует PostgreSQL и DATABASE_URL.

Структура
src/                     React-интерфейс
src/data/                каталог Wrestling и домашние тренировки
server-go/cmd/server     запуск API
server-go/internal       domain, services, handlers, repositories
server-go/internal/db    PostgreSQL и миграции
openapi.yaml             контракт API
Dockerfile               единый frontend + backend image
docker-compose.yml       локальный PostgreSQL + API
SECURITY.md              модель безопасности
ARCHITECTURE.md          архитектурные решения
OPERATIONS.md            эксплуатация
Проверка проекта
Backend имеет unit-тесты для middleware, CORS, обработчиков и сервисного слоя. Перед релизом рекомендуется выполнить:

cd server-go
go test ./...
go vet ./...
Для frontend:

npm install
npm test -- --runInBand
npm run build
Локальный React dev-server проксирует /api на http://localhost:8080, поэтому frontend и API можно запускать отдельно.

Принцип проекта
Wrestling — единый продукт без дополнительных приставок в названии. Интерфейс сохраняет японский визуальный стиль, а backend является источником истины для аккаунтов, пользовательского прогресса, тренировок и community-данных.

Quality and verification
The project includes unit/integration tests for the React application, Playwright smoke tests for desktop and mobile flows, and Go tests for HTTP routing and backend behavior. The catalog and training datasets also have integrity checks so duplicate technique names or missing training stages are caught early.

The application can be installed as a PWA. The service worker caches the application shell while API requests remain network-only so authenticated data is not silently cached as static content.

The visual design is intentionally preserved: this engineering pass does not replace the existing Wrestling visual system.

