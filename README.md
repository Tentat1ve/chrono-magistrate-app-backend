# Определение года события по упоминанию сановников — ЛР2

`Услуги` — сановники (визири, консулы, воеводы) с годами пребывания в должности.
Поля по теме: `office_start` и `office_end` — год начала и конца пребывания в должности (до н.э. — отрицательные числа).

## База данных (PostgreSQL, GORM)

| Таблица           | Поля                                                                                                                                                       |
|-------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `users`           | `id` PK, `login` varchar(50) unique, `full_name` varchar(100), `created_at`                                                                                 |
| `dignitaries`     | `id` PK, `name` varchar(255), `office` varchar(255), `description` text, `status` varchar(20) (черновик/опубликован/удален), `image_url` varchar(500), `video_url` varchar(500), `office_start` int, `office_end` int, `created_at` — дата создания, `published_at` — дата формирования, `creator_id` FK → users |
| `dignitary_likes` | `id` PK, `user_id` FK → users, `dignitary_id` FK → dignitaries, unique(user_id, dignitary_id)                                                              |

Каскадного удаления нет (`ON DELETE RESTRICT`). Один черновик на пользователя — частичный уникальный индекс `idx_one_draft_per_creator`.

## HTTP-методы

| Метод | URL                                                         | Что делает                                     | Как          |
|-------|-------------------------------------------------------------|------------------------------------------------|--------------|
| GET   | `/dignitaries?office_year=1570`                             | плитка, фильтр «был в должности в году»        | ORM          |
| GET   | `/dignitary_feed`, `/dignitary_feed/:id[?next=true]`        | лента, из БД одна строка                       | ORM          |
| GET   | `/dignitary_draft`                                          | черновик текущего пользователя                 | ORM          |
| POST  | `/dignitary_draft`                                          | «Далее» — создание черновика                   | ORM          |
| POST  | `/dignitary_draft/publish`                                  | «Опубликовать» — смена статуса                 | ORM          |
| POST  | `/dignitaries/:id/delete`                                   | логическое удаление                            | SQL `UPDATE` (курсор) |

Файлы при создании на сервер не передаются — форма отправляет только имена файлов, в БД записываются неверные url.
Если url пустой или файл недоступен, показываются фото и видео по умолчанию из `static/img/`.

## Запуск

1. `cp .env.example .env` (Postgres снаружи на порту 5433)
2. Поднять Postgres, Adminer и Minio:
   ```bash
   docker compose up -d
   ```
3. Создать таблицы:
   ```bash
   go run ./cmd/migrate
   ```
4. Заполнить данными: Adminer http://localhost:8081 (PostgreSQL, сервер `db`, пользователь/пароль/БД из `.env`) → «SQL-запрос» → содержимое `seed.sql`.
5. Запустить сервер:
   ```bash
   go run ./cmd/dignitaries
   ```
   Открыть http://localhost:8080/dignitaries
