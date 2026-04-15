# Статусы
* `open` - "Открыта"
* `in_progress` - "В работе"
* `completed` - "Выполнено"
* `closed` - "Отменена"

---

# Приоритеты
* `low` - 'Низкий'
* `medium` - 'Средний'
* `high` - 'Высокий'
* `critical` - 'Критический'

---

# Endpoints
|   endpoints       |   что происходит  |
|     ---           |       ---         |
|**GET** `api/tasks`|



# TaskSite API

## Статусы задач
* `open` — Открыта
* `in_progress` — В работе
* `completed` — Выполнена
* `closed` — Отменена
* `has_solution` — Есть решение

---

## Сущности

### Task
| Поле | Тип | Описание |
|------|-----|----------|
| `id` | int | ID задачи |
| `username` | string | Имя исполнителя (опционально) |
| `group_name` | string | Название группы (опционально) |
| `name` | string | Название задачи |
| `description` | string | Описание |
| `author` | string | Автор задачи |
| `status` | string | Статус из списка выше |
| `solution_comment` | string | Комментарий к решению (опционально) |
| `created_at` | string | Дата создания (RFC3339) |
| `updated_at` | string | Дата обновления (опционально) |
| `completed_at` | string | Дата завершения (опционально) |

### TaskGroup
| Поле | Тип | Описание |
|------|-----|----------|
| `id` | int | ID группы |
| `name` | string | Название группы |
| `description` | string | Описание (опционально) |
| `created_at` | string | Дата создания (RFC3339) |

### LoginResponse
| Поле | Тип | Описание |
|------|-----|----------|
| `message` | string | Сообщение |
| `token` | string | Сессионный токен |
| `username` | string | Имя пользователя |
| `id` | int | ID пользователя |

---

## Endpoints

### Auth (без токена)
| Метод | Эндпоинт | Тело запроса | Ответ | Статусы |
|-------|----------|--------------|-------|---------|
| `POST` | `/api/register` | `{username, password}` | `LoginResponse` | 201, 400, 409 |
| `POST` | `/api/login` | `{username, password}` | `LoginResponse` | 200, 400, 401 |

### User (требуется `X-Session-Token`)
| Метод | Эндпоинт | Тело запроса | Ответ | Статусы |
|-------|----------|--------------|-------|---------|
| `POST` | `/api/me` | — | `LoginResponse` | 200, 401, 500 |
| `POST` | `/api/logout` | — | — | 200, 401 |

### Tasks (требуется `X-Session-Token`)
| Метод | Эндпоинт | Тело запроса | Ответ | Статусы |
|-------|----------|--------------|-------|---------|
| `GET` | `/api/tasks` | — | `Task[]` | 200, 401 |
| `GET` | `/api/tasks?status=open` | — | `Task[]` | 200, 401 |
| `POST` | `/api/tasks` | `{name, description, author}` | `Task` | 201, 400, 401, 500 |
| `GET` | `/api/tasks/{id}` | — | `Task` | 200, 400, 401, 500 |
| `PUT` | `/api/tasks/{id}` | `{name?, description?, author?, status?, solution_comment?}` | `Task` | 200, 400, 401, 403, 404, 500 |
| `DELETE` | `/api/tasks/{id}` | — | — | 204, 400, 401, 404, 500 |
| `POST` | `/api/tasks/{id}/claim` | — | `Task` | 200, 400, 401, 409, 500 |
| `POST` | `/api/tasks/{id}/complete` | — | `Task` | 200, 400, 401, 403, 404, 500 |
| `PUT` | `/api/tasks/{id}/group` | `{group_id}` | — | 204, 400, 401, 404, 500 |
| `GET` | `/api/tasks/ungrouped` | — | `Task[]` | 200, 401, 500 |
| `GET` | `/api/tasks/ungrouped?status=open` | — | `Task[]` | 200, 401, 500 |

### Groups (требуется `X-Session-Token`)
| Метод | Эндпоинт | Тело запроса | Ответ | Статусы |
|-------|----------|--------------|-------|---------|
| `POST` | `/api/groups` | `{name, description}` | `TaskGroup` | 201, 400, 401, 409, 500 |
| `GET` | `/api/groups` | — | `TaskGroup[]` | 200, 401, 500 |
| `GET` | `/api/groups/{id}/tasks` | — | `Task[]` | 200, 400, 401, 500 |
| `GET` | `/api/groups/{id}/tasks?status=open` | — | `Task[]` | 200, 400, 401, 500 |

### System
| Метод | Эндпоинт | Ответ | Статусы |
|-------|----------|-------|---------|
| `GET` | `/health` | `{status: "ok"}` | 200, 500 |
| `GET` | `/swagger/*` | Swagger UI | 200 |

---

## Заголовки
| Заголовок | Значение | Где используется |
|-----------|----------|-----------------|
| `X-Session-Token` | токен из `LoginResponse` | Все эндпоинты кроме `/register`, `/login`, `/health` |
| `Content-Type` | `application/json` | Все POST/PUT запросы с телом |

---

## Примеры запросов

### Регистрация
```
POST /api/register
Content-Type: application/json

{
  "username": "ivan",
  "password": "secret123"
}
```

### Вход
```
POST /api/login
Content-Type: application/json

{
  "username": "ivan",
  "password": "secret123"
}
```

### Создание задачи
```
POST /api/tasks
Content-Type: application/json
X-Session-Token: a1b2c3d4e5f6...

{
  "name": "Реализовать авторизацию",
  "description": "Добавить логин и регистрацию",
  "author": "ivan"
}
```

### Взять задачу в работу
```
POST /api/tasks/42/claim
X-Session-Token: a1b2c3d4e5f6...
```

### Обновить статус задачи
```
PUT /api/tasks/42
Content-Type: application/json
X-Session-Token: a1b2c3d4e5f6...

{
  "status": "completed",
  "solution_comment": "Готово, протестировано"
}
```

---

## Примечания
* Все даты в формате RFC3339: `2024-01-15T10:30:00Z`
* Токен сессии живёт 24 часа, продлевается при каждом запросе
* Задачу может редактировать/удалять только автор или назначенный исполнитель
* Завершить задачу может только тот, кому она назначена (`user_id`)
* При смене статуса на `open` сбрасывается `user_id`, на `completed` — заполняется `completed_at`
