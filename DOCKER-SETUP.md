# Быстрый запуск EasyCabinet через Docker

## Что будет запущено

- **MySQL 8.0** — база данных (порт 3306)
- **Backend (NestJS)** — API сервер (порт 4000)
- **Frontend (React)** — веб-интерфейс (порт 3000)

## Шаг 1: Проверка Docker

```bash
docker --version
docker-compose --version
```

Если Docker не установлен, скачайте [Docker Desktop](https://www.docker.com/products/docker-desktop/).

## Шаг 2: Запуск всех сервисов

```bash
# Запустить все контейнеры в фоновом режиме
docker-compose up -d

# Посмотреть логи
docker-compose logs -f

# Посмотреть статус
docker-compose ps
```

## Шаг 3: Проверка работы

После запуска откройте в браузере:

- **Frontend:** http://localhost:3000
- **Backend API:** http://localhost:4000
- **MySQL:** localhost:3306 (пользователь: cabinet, пароль: cabinetpass)

## Управление

```bash
# Остановить все сервисы
docker-compose down

# Остановить и удалить данные
docker-compose down -v

# Перезапустить сервис
docker-compose restart backend

# Посмотреть логи конкретного сервиса
docker-compose logs -f backend
docker-compose logs -f frontend
docker-compose logs -f mysql

# Пересобрать и запустить
docker-compose up -d --build
```

## Настройка (опционально)

Если нужно изменить настройки, отредактируйте `docker-compose.yml`:

### Изменить порты

```yaml
services:
  backend:
    ports:
      - "8080:4000"  # Внешний:Внутренний
  frontend:
    ports:
      - "8000:3000"
```

### Изменить пароли MySQL

```yaml
services:
  mysql:
    environment:
      MYSQL_ROOT_PASSWORD: новый_root_пароль
      MYSQL_PASSWORD: новый_пароль_пользователя
  backend:
    environment:
      DATABASE_URL: "mysql://cabinet:новый_пароль_пользователя@mysql:3306/easycabinet"
```

### Добавить Redis (для кеша)

```yaml
services:
  redis:
    image: redis:alpine
    container_name: easycabinet-redis
    ports:
      - "6379:6379"
  
  backend:
    environment:
      REDIS_URL: "redis://redis:6379"
    depends_on:
      - redis
```

## Troubleshooting

### Backend не запускается

```bash
# Проверить логи
docker-compose logs backend

# Проверить что MySQL готов
docker-compose logs mysql | grep "ready for connections"

# Перезапустить миграции вручную
docker-compose exec backend npx prisma migrate deploy
```

### Frontend показывает ошибки API

Проверьте что backend запущен:
```bash
curl http://localhost:4000
```

### Порты заняты

Если порты 3000, 4000 или 3306 уже используются, измените их в `docker-compose.yml`.

## Первый запуск

После успешного запуска:

1. Откройте http://localhost:3000
2. Зарегистрируйте первого пользователя
3. Проверьте что данные сохраняются в MySQL

## Остановка

```bash
# Остановить без удаления данных
docker-compose stop

# Остановить и удалить контейнеры (данные сохранятся)
docker-compose down

# Полная очистка (удалит все данные!)
docker-compose down -v
```

## Данные

Все данные хранятся в Docker volumes:
- `mysql_data` — база данных MySQL
- `backend_uploads` — загруженные файлы (скины/плащи)

Для бэкапа:
```bash
docker-compose exec mysql mysqldump -u cabinet -pcabinetpass easycabinet > backup.sql
```
