# Определение года события по упоминанию сановников — ЛР3 (веб-сервис)

`Услуги` — сановники (визири, консулы, воеводы) с годами пребывания в должности.
Поля по теме: `office_start` и `office_end` — год начала и конца пребывания в должности (до н.э. — отрицательные числа).

Веб-сервис на Go (Gin + GORM), данные в PostgreSQL, изображения и видео — в MinIO.
До ЛР4 текущий пользователь зафиксирован константой (id = 1) в функции-singleton `CurrentUser()` (`internal/app/handler/current_user.go`).

## HTTP-методы

Все адреса начинаются с `/api`. Ответы — JSON, ошибки — `{"status": "error", "description": "..."}`.
Системные поля (id, статус, создатель, даты) клиент передать не может — они вычисляются на сервере.

### Домен «сановники» (`/api/dignitaries`)

| Метод  | URL                                         | Тело запроса                                                                 | Ответ                                                                                  |
|--------|---------------------------------------------|------------------------------------------------------------------------------|----------------------------------------------------------------------------------------|
| GET    | `/api/dignitaries?office_year=1570` | —                                                                            | 200: массив опубликованных сановников; `is_creator` = 1, если создатель — текущий пользователь. Фильтр необязательный: `office_year` — год пребывания в должности (поля по теме `office_start` и `office_end`) |
| GET    | `/api/dignitaries/feed`                     | —                                                                            | 200: первый опубликованный сановник; `is_liked` = 1, если текущий пользователь поставил лайк |
| GET    | `/api/dignitaries/feed/:id`                 | —                                                                            | 200: сановник по id; 404, если не опубликован                                          |
| GET    | `/api/dignitaries/feed/:id?next=true`       | —                                                                            | 200: следующий после id опубликованный сановник (после последнего — первый)            |
| GET    | `/api/dignitaries/draft`                    | —                                                                            | 200: черновик текущего пользователя (id не указывается); 404, если черновика нет      |
| POST   | `/api/dignitaries`                          | `multipart/form-data`: `name`, `image` (файл `image/*`, до 10 МБ), `video` (файл `video/*`, до 50 МБ) | 201: созданный черновик. Файлы сохраняются в MinIO под сгенерированными латинскими именами, url — в БД. 409, если черновик уже есть |
| PUT    | `/api/dignitaries/draft/publish`            | `{"office": "...", "description": "...", "office_start": -70, "office_end": -70}` | 200: опубликованный сановник (статус «опубликован», заполнена дата формирования); 404, если черновика нет |
| DELETE | `/api/dignitaries/:id`                      | —                                                                            | 200: логическое удаление (статус «удален»); 403 — чужой сановник; 404 — не найден      |
| POST   | `/api/dignitaries/:id/like`                 | `{"liked": 1}` — поставить лайк, `{"liked": 0}` — отменить                    | 200: `{"dignitary_id", "is_liked", "likes_count"}`                                    |

Статус меняется только вперёд: создание даёт «черновик», публикация — «опубликован», удаление — «удален». Вернуть в черновик нельзя.

### Домен «пользователи» (`/api/users`)

| Метод | URL                   | Тело запроса                                                    | Ответ                                                         |
|-------|-----------------------|-----------------------------------------------------------------|---------------------------------------------------------------|
| POST  | `/api/users/register` | `{"login": "new_user", "full_name": "...", "password": "secret123"}` | 201: пользователь (без пароля); 409 — логин занят             |
| POST  | `/api/users/login`    | `{"login": "...", "password": "..."}`                            | 200: заглушка до ЛР4 — проверяет пароль, токен не выдаёт; 401 |
| POST  | `/api/users/logout`   | —                                                               | 200: заглушка до ЛР4                                          |

Коллекция запросов для Postman: [`postman_collection.json`](postman_collection.json).

## Таблицы БД

| Таблица           | Поля                                                                                                                                                       |
|-------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `users`           | `id` PK, `login` varchar(50) unique, `full_name` varchar(100), `password_hash` varchar(60) (bcrypt), `created_at`                                          |
| `dignitaries`     | `id` PK, `name` varchar(255), `office` varchar(255), `description` text, `status` varchar(20) (черновик/опубликован/удален), `image_url` varchar(500), `video_url` varchar(500), `office_start` int, `office_end` int, `created_at` — дата создания, `published_at` — дата формирования, `creator_id` FK → users |
| `dignitary_likes` | `id` PK, `user_id` FK → users, `dignitary_id` FK → dignitaries, unique(user_id, dignitary_id)                                                              |

Каскадного удаления нет (`ON DELETE RESTRICT`). Один черновик на пользователя — частичный уникальный индекс `idx_one_draft_per_creator`.

## Структура проекта

| Пакет                        | Назначение                                              |
|------------------------------|---------------------------------------------------------|
| `cmd/dignitaries`            | запуск веб-сервиса                                      |
| `cmd/migrate`                | миграции (AutoMigrate)                                  |
| `internal/app/ds`            | модели GORM                                             |
| `internal/app/serializer`    | структуры запросов и ответов JSON                       |
| `internal/app/repository`    | работа с PostgreSQL (GORM) и MinIO                      |
| `internal/app/handler`       | HTTP-обработчики, маршруты, singleton текущего пользователя |
| `internal/app/config`, `dsn` | конфигурация из `.env`                                  |

## Запуск

1. `cp .env.example .env` (Postgres снаружи на порту 5433)
2. Поднять Postgres, Adminer и MinIO:
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
   go run ./cmd/dignitaries
   ```
   API: http://localhost:8080/api/dignitaries
