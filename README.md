# Определение года события по упоминанию сановников — ЛР4 (авторизация и Swagger)

`Услуги` — `dignitarie_lists`: сановники (визири, консулы, воеводы) с датами пребывания в должности.
Поля по теме: `start_date` — дата вступления в должность и `end_date` — дата оставления должности.

Веб-сервис на Go (Gin + GORM), данные в PostgreSQL, изображения и видео — в MinIO, **авторизация по JWT**, чёрный список токенов (logout) — в **Redis**, документация — **Swagger**.

## Авторизация

- `POST /api/users/login` проверяет логин и пароль (bcrypt) и возвращает JWT (`access_token`, `token_type`, `expires_in`). Токен подписан HS256, живёт `JWT_EXPIRES_IN` (по умолчанию 1 час) и содержит `user_id` и уникальный идентификатор `jti`.
- Токен передаётся в заголовке `Authorization: Bearer <токен>`.
- `POST /api/users/logout` кладёт `jti` токена в чёрный список Redis (ключ `dignitarie_lists.jwt.blacklist.<jti>`) до конца срока жизни токена; после этого токен не принимается.
- Две роли: **гость** (без токена) и **создатель** (авторизованный пользователь). Создатель услуги подставляется из токена автоматически.
- Для гостя методы создателя отвечают `401`; чужую услугу удалить нельзя — `403`.

## HTTP-методы (10)

Все адреса начинаются с `/api`. Ответы — JSON, ошибки — `{"status": "error", "description": "..."}`.
Системные поля (id, статус, создатель, даты создания и формирования) клиент передать не может — они вычисляются на сервере.

### Домен «услуги» (`/api/dignitarie_lists`)

| № | Метод  | URL                                    | Доступ   | Параметры / тело запроса                                                           | Ответ                                                                                                  |
|---|--------|----------------------------------------|----------|-------------------------------------------------------------------------------------|--------------------------------------------------------------------------------------------------------|
| 1 | GET    | `/api/dignitarie_lists`                | гость    | необязательный `date` — календарная дата ГГГГ-ММ-ДД                                 | 200: массив опубликованных услуг: `id`, `name`, `image_url`, `start_date`, `end_date`, `likes_count`, `is_creator` (0/1; для гостя 0) |
| 2 | GET    | `/api/dignitarie_lists/feed`           | гость    | необязательные `id` и `next`: без `id` — первая услуга, `id` — по ID, `id`+`next=true` — следующая | 200: `id`, `name`, `description`, `image_url`, `video_url`, `start_date`, `end_date`, `likes_count`, `is_liked` (0/1), `is_creator` (0/1); 404 |
| 3 | GET    | `/api/dignitarie_lists/draft`          | создатель| — (id не указывается)                                                               | 200: черновик пользователя (не более одного): `id`, `name`, `description`, `image_url`, `video_url`, `start_date`, `end_date`, `status`; 401; 404 |
| 4 | POST   | `/api/dignitarie_lists`                | создатель| `multipart/form-data`: `name`, файлы `image` (image/*, до 10 МБ), `video` (video/*, до 50 МБ) | 201: черновик; файлы в MinIO под латинскими именами, в БД их url; создатель из токена; 401; 409 — черновик уже есть |
| 5 | PUT    | `/api/dignitarie_lists/draft/publish`  | создатель| JSON: `description`, `start_date`, `end_date` (ГГГГ-ММ-ДД)                          | 200: опубликованная услуга (статус «опубликован», дата формирования); 401 для гостя; 404            |
| 6 | DELETE | `/api/dignitarie_lists/:id`            | создатель| —                                                                                   | 200: логическое удаление своей услуги; 401; 403 — чужая; 404                                          |
| 7 | POST   | `/api/dignitarie_lists/:id/like`       | создатель| JSON: `liked` — 1 поставить лайк, 0 отменить                                        | 200: `dignitarie_list_id`, `is_liked`, `likes_count`; 401                                             |

### Домен «пользователи» (`/api/users`)

| № | Метод | URL                   | Доступ    | Тело запроса                                              | Ответ                                                                  |
|---|-------|-----------------------|-----------|-----------------------------------------------------------|------------------------------------------------------------------------|
| 8 | POST  | `/api/users/register` | гость     | JSON: `login` (3–50 символов), `password` (от 6 символов) | 201: `id`, `login`; 409 — логин занят                                  |
| 9 | POST  | `/api/users/login`    | гость     | JSON: `login`, `password`                                 | 200: `access_token`, `token_type`, `expires_in`; 401 — неверные данные |
| 10| POST  | `/api/users/logout`   | создатель | — (токен в заголовке)                                     | 200: токен добавлен в чёрный список Redis; 401                          |

Лента работает и без параметров, потому что при переходе с панели вкладок идентификатор не указывается: возвращается первая опубликованная услуга. «Далее» — тот же метод ленты с параметрами `id` и `next=true`.

## Swagger

Документация генерируется утилитой `swag` из комментариев к методам (`docs/swagger.json`, `docs/swagger.yaml`). Страница: http://localhost:8080/swagger/index.html.
Без авторизации доступны только методы чтения (список, лента) и вход/регистрация. Чтобы вызывать методы создателя: выполнить `POST /users/login`, нажать **Authorize** и вставить `Bearer <access_token>`.

Перегенерация документации после изменения комментариев:
```bash
go run github.com/swaggo/swag/cmd/swag init -g cmd/dignitarie_lists/main.go
```

## Таблицы БД

### `users` — пользователи

| Поле            | Тип          | Ограничения       | Описание          |
|-----------------|--------------|-------------------|-------------------|
| `id`            | bigint       | PK                | идентификатор     |
| `login`         | varchar(50)  | NOT NULL, UNIQUE  | логин             |
| `password_hash` | varchar(60)  | NOT NULL          | bcrypt-хеш пароля |

### `dignitarie_lists` — услуги

| Поле           | Тип          | Ограничения                                  | Описание                                  |
|----------------|--------------|----------------------------------------------|-------------------------------------------|
| `id`           | bigint       | PK                                           | идентификатор                             |
| `name`         | varchar(255) | NOT NULL                                     | наименование                              |
| `description`  | text         | необязательное                               | краткое описание                          |
| `status`       | varchar(20)  | NOT NULL                                     | черновик / опубликован / удален           |
| `image_url`    | varchar(500) | NOT NULL                                     | url изображения в MinIO                   |
| `video_url`    | varchar(500) | NOT NULL                                     | url видео в MinIO                         |
| `start_date`   | date         | необязательное                               | поле по теме: дата вступления в должность |
| `end_date`     | date         | необязательное                               | поле по теме: дата оставления должности   |
| `created_at`   | timestamptz  | NOT NULL                                     | дата создания                             |
| `published_at` | timestamptz  | необязательное                               | дата формирования (публикации)            |
| `creator_id`   | bigint       | NOT NULL, FK → users(id), ON DELETE RESTRICT | создатель (из токена)                     |

Один черновик на пользователя — частичный уникальный индекс `idx_one_draft_per_creator`.

### `dignitarie_list_likes` — лайки (многие-ко-многим пользователь–услуга)

| Поле                 | Тип    | Ограничения                                             | Описание      |
|----------------------|--------|---------------------------------------------------------|---------------|
| `id`                 | bigint | PK                                                      | идентификатор |
| `user_id`            | bigint | NOT NULL, FK → users(id), ON DELETE RESTRICT            | кто поставил  |
| `dignitarie_list_id` | bigint | NOT NULL, FK → dignitarie_lists(id), ON DELETE RESTRICT | какой услуге  |

Пара (`user_id`, `dignitarie_list_id`) уникальна. Каскадного удаления нет.

## Структура проекта

| Пакет                        | Назначение                                                    |
|------------------------------|---------------------------------------------------------------|
| `cmd/dignitarie_lists`       | запуск веб-сервиса, общее описание Swagger                    |
| `cmd/migrate`                | миграции (AutoMigrate)                                        |
| `docs`                       | сгенерированная Swagger-документация                          |
| `internal/app/ds`            | модели GORM и содержимое JWT                                  |
| `internal/app/serializer`    | структуры запросов и ответов JSON                             |
| `internal/app/repository`    | работа с PostgreSQL (GORM) и MinIO                            |
| `internal/app/redis`         | чёрный список токенов в Redis                                 |
| `internal/app/handler`       | HTTP-обработчики, маршруты, middleware авторизации            |
| `internal/app/config`, `dsn` | конфигурация из `.env`                                        |

## Запуск

1. `cp .env.example .env` (Postgres снаружи на порту 5433, Redis — 6379)
2. Поднять Postgres, Adminer, MinIO и Redis:
   ```bash
   docker compose up -d
   ```
3. Создать таблицы:
   ```bash
   go run ./cmd/migrate
   ```
4. Заполнить данными: Adminer http://localhost:8081 (PostgreSQL, сервер `db`, пользователь/пароль/БД из `.env`) → «SQL-запрос» → содержимое `seed.sql`. Пароль всех тестовых пользователей — `password123`.
5. Запустить веб-сервис:
   ```bash
   go run ./cmd/dignitarie_lists
   ```
   API: http://localhost:8080/api/dignitarie_lists, Swagger: http://localhost:8080/swagger/index.html
