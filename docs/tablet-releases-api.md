# NoisePlace API

Базовый URL локального API: `http://localhost:8080`.

## 1. Релизы

| Метод | Endpoint | Назначение | Успешный ответ |
|---|---|---|---|
| `POST` | `/api/v1/releases` | Создать релиз вместе со связями исполнителей | `201 Created` |
| `GET` | `/api/v1/releases` | Поиск релизов | `200 OK` |
| `GET` | `/api/v1/releases/{id}` | Получить релиз с исполнителями | `200 OK` |
| `PUT` | `/api/v1/releases/{id}` | Изменить данные релиза | `200 OK` |
| `DELETE` | `/api/v1/releases/{id}` | Удалить релиз | `204 No Content` |

### Параметры поиска релизов

Endpoint: `GET /api/v1/releases`

| Параметр | Обязательный | Значение по умолчанию | Назначение |
|---|---|---|---|
| `query` | Да | — | Поисковая строка |
| `limit` | Нет | `20` | Количество результатов, от `1` до `100` |
| `offset` | Нет | `0` | Смещение для пагинации, неотрицательное число |

Пример запроса:

`GET http://localhost:8080/api/v1/releases?query=test&limit=20&offset=0`

## 2. Исполнители релиза

| Метод | Endpoint | Назначение | Успешный ответ |
|---|---|---|---|
| `GET` | `/api/v1/releases/{id}/artists` | Получить исполнителей релиза | `200 OK` |
| `POST` | `/api/v1/releases/{id}/artists` | Добавить исполнителя к релизу | `201 Created` |
| `PUT` | `/api/v1/releases/{id}/artists/{artist_id}` | Изменить роль и позицию исполнителя | `200 OK` |
| `DELETE` | `/api/v1/releases/{id}/artists/{artist_id}` | Удалить связь исполнителя с релизом | `204 No Content` |

## 3. Роли исполнителей

Поддерживаемые роли:

| Роль | Назначение |
|---|---|
| `primary` | Основной исполнитель |
| `featured` | Приглашённый исполнитель |
| `producer` | Продюсер |

### Пример изменения роли и позиции

Метод: `PUT`

Endpoint: `/api/v1/releases/{id}/artists/{artist_id}`

Заголовок:

`Content-Type: application/json`

Тело запроса:

```json
{
  "role": "featured",
  "position": 1
}