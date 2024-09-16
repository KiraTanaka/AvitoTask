Для запуска проекта необходимо установить переменные окружения:
- `SERVER_ADDRESS` — адрес и порт, который будет слушать HTTP сервер при запуске. Пример: 0.0.0.0:8080.
- `POSTGRES_USERNAME` — имя пользователя для подключения к PostgreSQL.
- `POSTGRES_PASSWORD` — пароль для подключения к PostgreSQL.
- `POSTGRES_HOST` — хост для подключения к PostgreSQL (например, host.docker.internal).
- `POSTGRES_PORT` — порт для подключения к PostgreSQL (например, 5432).
- `POSTGRES_DATABASE` — имя базы данных PostgreSQL, которую будет использовать приложение.

Выполнить команды:
```
 docker build -t <имя образа> .
 docker run -d -p 8080:8080 --name <имя контейнера> --env-file <путь до файла с переменными окружения> <имя образа>
 ```

## Логика приложения

При развертывании приложения накатываются миграции в бд со следующими объектами:
- **Таблицы**:
  - tender
  - tender_version_hist
  - bid
  - bid_version_hist
  - bid_decision
- **Типы**:
  - service_type
  - tender_status
  - bid_author_type
  - bid_decision_type
  - bid_status
- **Триггерные функции**:
  - tender_version_hist_update_trigger_func
  - bid_version_hist_update_trigger_func

Данный проект реализует следующие эндпоинты:
- **GET**:
  - /
  - /api/ping
  - /api/tenders/
  - /api/tenders/my
  - /api/tenders/:tenderId/status
  - /api/bids/:id/list
  - /api/bids/my
  - /api/bids/:id/status
- **POST**:
  - /api/tenders/new
  - /api/bids/new
- **PUT**:
  - /api/tenders/:tenderId/status
  - /api/tenders/:tenderId/rollback/:version
  - /api/bids/:id/status
  - /api/bids/:id/rollback/:version
  - /api/bids/:id/submit_decision (Расширенный процесс согласования)
- **PATCH**:
  - /api/tenders/:tenderId/edit
  - /api/bids/:id/edit