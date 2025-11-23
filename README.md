# Тестовое задание Avito (Backend, 2025)
Репозиторий тестового задания для прохождения отбора на стажировку в Avito (Backend) (autumn 2025)

# Запуск

Предварительно установите Docker и Docker Compose

Склонируйте репозиторий и перейдите в папку

```bash
git clone https://github.com/shirotame/avito-backend-assignment-autumn-2025
cd avito-backend-assignment-autumn-2025
```

Для запуска проекта используйте `docker-compose up --build`. Необходимо предварительно сделать `.env` файл с настройками
окружения (пример в [`example.env`](https://github.com/shirotame/avito-backend-assignment-autumn-2025/blob/main/example.env))

```bash
docker-compose up --build
```

Проект запускается на порту `8080`. Документация OpenAPI доступна по адресу `localhost:8080/swagger/static/`.

Для запуска тестов необходимо запустить другой docker-compose файл

```bash
docker-compose --file "test.docker-compose.yml" up --build
```

При успешном прохождении тестов контейнер завершится с кодом *0*. `prservice-1 exited with code 0`.


# Вопросы

- В задании не указано поведение в том случае, если команда создается с теми же участинками (одинаковые *id*). Было
сделано:
    - При создании пользователей, в случае SQL ошибки на **UNIQUE CONSTRAINT**, возвращается специальная ошибка
   `ErrUserAlreadyExists` и код `USER_EXISTS`, который пишется в ошибке со статусом 409 Conflict. 
- В задании также не указан тот случай, если из всех свободных для замены ревьюверов есть только автор. Было
  сделано:
  - При случайном выборе из активных пользователей проверяется, если *ID* равен *ID* автора. В таком случае пытаемся 
  выбрать случайное, пока не закончится срез доступных пользователей. Если таких не нашлось, то тогда код 
  `NO_CANDIDATE` со статусом 409 Conflict.