# Определение года события по упоминанию сановников — ЛР1

`Услуги` — сановники (визири, консулы, воеводы) с годами пребывания в должности.
Поля по теме: `OfficeStart` и `OfficeEnd` — год начала и конца пребывания в должности (до н.э. — отрицательные числа).

## Страницы (3 GET-запроса)

| Страница    | URL                                                     | Обработчик          |
|-------------|---------------------------------------------------------|---------------------|
| Лента       | `/dignitary_feed`, `/dignitary_feed/:id`, `/dignitary_feed/:id?next=true` | `GetDignitaryFeed`  |
| Добавление  | `/dignitary_draft`                                      | `GetDignitaryDraft` |
| Плитка      | `/dignitaries?office_year=1570`                         | `GetDignitaries`    |

Фильтр плитки: сановники, бывшие в должности в указанном году (`OfficeStart <= office_year <= OfficeEnd`).

Данные — коллекция в `internal/app/repository/repository.go` (без БД): статусы `черновик`/`опубликован`/`удален`, лайки — массив ID пользователей, количество считается в обработчике.

## Запуск

1. Положить в папку `media/` изображения и видео с ключами из коллекции:
   `ibrahim_pasha`, `sokollu_mehmed`, `cicero`, `caesar`, `shein`, `kara_mustafa`, `rustem_pasha`, `pompey` — для каждого `.jpg` и `.mp4`.
2. Поднять Minio (создаст публичный бакет `dignitaries` и загрузит файлы из `media/`):
   ```bash
   docker compose up -d
   ```
   Консоль Minio: http://localhost:9001 (minioadmin / minioadminpassword).
3. Запустить сервер:
   ```bash
   go run ./cmd/dignitaries
   ```
   Открыть http://localhost:8080/dignitaries
