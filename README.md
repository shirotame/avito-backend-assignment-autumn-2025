# Тестовое задание Avito (Backend, 2025)
Репозиторий тестового задания для прохождения отбора на стажировку в Avito (Backend) (autumn 2025)

# Использованные технологии

- Go 1.25.3
- PostgreSQL 18
- [go-chi/chi](https://github.com/go-chi/chi) в качестве HTTP роутера
- [jackc/pgx](https://github.com/jackc/pgx) для работы с PostgreSQL
- [golang-migrate/migrate](https://github.com/golang-migrate/migrate) для миграций БД



# Запуск

Предварительно установите Docker и Docker Compose

Склонируйте репозиторий и перейдите в папку

```bash
git clone https://github.com/shirotame/avito-backend-assignment-autumn-2025
cd avito-backend-assignment-autumn-2025
```

Для запуска проекта используйте `docker-compose up --build`

```bash
docker-compose up --build
```

Проект запускается на порту `8080`. Документация OpenAPI доступна по адресу `localhost:8080/swagger/static/`.

Для запуска тестов необходимо запустить другой docker-compose файл

```bash
docker-compose --file "test.docker-compose.yml" up --build
```

При успешном прохождении тестов контейнер завершится с кодом *0*. `prservice-1 exited with code 0`.

# Дополнительные задания

- Был добавлен линтер `golangci-lint`. Его конфигруция находится в файле [`.golangci.yml`](https://github.com/shirotame/avito-backend-assignment-autumn-2025/blob/main/.golangci.yml)
- Были реализованы unit тесты слоя репозиториев
- Был создан простой эндпоинт статистики `/pullRequest/openByReviewers` (количество открытых PR по ревьюверу)
- Был написал e2e тест работы приложения

# Вопросы

- В задании не указан тот случай, если из всех свободных для замены ревьюверов есть только автор. Было
  сделано:
  - При случайном выборе из активных пользователей проверяется, если *ID* равен *ID* автора. В таком случае пытаемся 
  выбрать случайное, пока не закончится срез доступных пользователей. Если таких не нашлось, то тогда код 
  `NO_CANDIDATE` со статусом 409 Conflict.