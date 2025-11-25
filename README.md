
# Review Service (Avito Internship)

Сервис для автоматического назначения ревьюеров на Pull Requests.

## Краткое описание

- Создаёт команды и управляет их составом.
- Назначает ревьюеров из команды автора PR (до 2 человек).
- Позволяет менять статус PR (открыт/объединён).
- Поддерживает активацию/деактивацию пользователей.
- Обеспечивает переназначение ревьюеров.

## Quick Start


1. Запустите проект:
```bash
docker-compose up -d && goose -dir migrations postgres "user=review_user password=review_password dbname=review_service host=localhost port=5432 sslmode=disable" up && go run cmd/main.go
```
2. Протестируйсте
```bash
curl http://localhost:8080/health
sslmode=disable" up && go run cmd/main.go
```
 Ответ: {"status": "ok"}

## Основные эндпоинты


### Команды
- `POST /team/add` — создать команду  
- `GET /team/get?team_name=X` — получить команду  

### Пользователи
- `POST /users/setIsActive` — активировать/деактивировать пользователя  
- `GET /users/getReview?user_id=X` — получить PR для ревью  

### Pull Requests
- `POST /pullRequest/create` — создать PR  
- `POST /pullRequest/merge` — объединить PR  
- `POST /pullRequest/reassign` — переназначить ревьюера  

## Пример создания PR

```bash
curl -X POST http://localhost:8080/pullRequest/create \
  -H "Content-Type: application/json" \
  -d '{
    "pull_request_id": "pr-1",
    "pull_request_name": "Fix bug",
    "author_id": "user-1"
  }'
```