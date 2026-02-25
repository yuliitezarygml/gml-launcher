# GML — Minecraft Launcher Platform

**GML (GamerVII Minecraft Launcher)** — полноценная платформа для управления Minecraft-лаунчером. Включает серверную часть, веб-панель администратора, десктопный лаунчер, сервис авторизации и игровой сервер.

---

## Оглавление

- [Компоненты системы](#компоненты-системы)
- [Общая архитектура](#общая-архитектура)
- [Схема взаимодействия компонентов](#схема-взаимодействия-компонентов)
- [Поток авторизации](#поток-авторизации)
- [Поток запроса пользователя](#поток-запроса-пользователя)
- [Схема деплоя (Docker)](#схема-деплоя-docker)
- [Жизненный цикл лаунчера](#жизненный-цикл-лаунчера)
- [Агрегация новостей](#агрегация-новостей)
- [Технологический стек](#технологический-стек)
- [Установка и запуск](#установка-и-запуск)
- [API-эндпоинты](#api-эндпоинты)
- [Переменные окружения](#переменные-окружения)
- [Структура проекта](#структура-проекта)
- [Платформы](#платформы)
- [EasyCabinet — Личный кабинет игрока](#easycabinet--личный-кабинет-игрока)

---

## Компоненты системы

| Компонент | Язык / Фреймворк | Назначение |
|-----------|-----------------|------------|
| **Gml.Backend** | C# / ASP.NET Core 10 | Основной REST API + Reverse Proxy + Сервис скинов |
| **Gml.Web.Client** | TypeScript / Next.js 14 | Веб-панель администратора |
| **Gml.Launcher** | C# / Avalonia 11 | Кросс-платформенный десктопный лаунчер |
| **gml-auth** | Go 1.25 | Сервис авторизации + агрегация новостей |
| **ser** | Docker / Java 17 | Minecraft игровой сервер (Fabric 1.19.4) |
| **EasyCabinet** | TypeScript / NestJS + SolidJS | Личный кабинет игрока для AuroraLauncher |

---

## Общая архитектура

```mermaid
graph TB
    subgraph Clients["Клиенты"]
        BROWSER["Браузер (Веб-панель)"]
        LAUNCHER["Launcher (Avalonia Desktop)"]
        PLAYER["Игрок Minecraft"]
    end

    subgraph Docker["Docker: gml-network"]
        PROXY["gml-web-proxy\nYARP Reverse Proxy\n:8080 -> 5003"]

        subgraph Backend["Gml.Backend"]
            API["gml-web-api\nASP.NET Core\n:8082 -> 5000"]
            SKINS["gml-web-skins\nSkin Service\n:8085"]
        end

        FRONTEND["gml-web-frontend\nNext.js\n:8081 -> 3000"]
    end

    subgraph GameServer["Docker: ser"]
        MC["minecraft-server\nFabric 1.19.4 / Java 17\n:25565"]
        RCON["RCON\n:25575"]
    end

    subgraph External["Внешние сервисы"]
        AUTH["gml-auth\nGo Auth Server\n:5003"]
        MARKET["GML Market\ngml-market.recloud.tech"]
        DISCORD["Discord Bot"]
        TELEGRAM["Telegram Bot"]
    end

    subgraph Storage["Хранилище"]
        DB[("SQLite\nGmlBackendPanel")]
        FILES["Game Files\nProfiles and Mods"]
        TEXTURES["Skin Textures\n/app/Storage"]
        WORLD["World Data\n./ser/data"]
    end

    BROWSER -->|"HTTPS :5003"| PROXY
    LAUNCHER -->|"HTTP :5000"| API
    PLAYER -->|"Запускает игру"| LAUNCHER
    LAUNCHER -->|"Подключение :25565"| MC

    PROXY -->|"/api/*"| API
    PROXY -->|"/skins/*"| SKINS
    PROXY -->|"/"| FRONTEND

    FRONTEND -->|"API + SignalR"| API
    API -->|"Текстуры"| SKINS
    API -->|"Плагины/моды"| MARKET
    API -->|"Авторизация"| AUTH
    AUTH -->|"Новости"| DISCORD
    AUTH -->|"Новости"| TELEGRAM

    MC -->|"authlib-injector"| AUTH

    API --- DB
    API --- FILES
    SKINS --- TEXTURES
    MC --- WORLD
```

---

## Схема взаимодействия компонентов

```mermaid
graph LR
    subgraph "Gml.Launcher (Desktop)"
        LC["Gml.Client\nLibrary"]
        UI["Avalonia UI\nMVVM + ReactiveUI"]
        LSQ["SQLite\nLocal DB"]
        UI <--> LC
        UI <--> LSQ
    end

    subgraph "Gml.Backend"
        direction TB
        HANDLERS["Handlers\nBusiness Logic"]
        HUBS["SignalR Hubs\nReal-time"]
        REPOS["Repositories\nData Access"]
        INTEGRATIONS["Auth Integrations\nDataLife / Azuriom\nEasyCabinet / WebMC\nWordPress / NamelessMC"]
        HANDLERS --> REPOS
        HANDLERS --> INTEGRATIONS
    end

    subgraph "Gml.Web.Client (Next.js)"
        PAGES["Pages / App Router"]
        QUERY["TanStack Query\nData Fetching"]
        ZUSTAND["Zustand\nState"]
        SIGNALR_C["SignalR Client"]
        PAGES --> QUERY
        PAGES --> ZUSTAND
        PAGES --> SIGNALR_C
    end

    subgraph "gml-auth (Go)"
        AHANDLERS["Handlers\nAuth / Admin / News"]
        ASTORAGE["JSON Storage\nusers.json"]
        NEWS["News Aggregator\nTelegram + Discord"]
        AHANDLERS --> ASTORAGE
        AHANDLERS --> NEWS
    end

    subgraph "ser (Minecraft Server)"
        MCFAB["Fabric Server\n1.19.4"]
        AUTHINJ["authlib-injector\nCustom Auth"]
        MCFAB --> AUTHINJ
    end

    LC -->|"REST API"| HANDLERS
    QUERY -->|"REST API + JWT"| HANDLERS
    SIGNALR_C <-->|"WebSocket"| HUBS
    INTEGRATIONS -->|"POST /signin"| AHANDLERS
    AUTHINJ -->|"Verify player"| AHANDLERS
```

---

## Поток авторизации

```mermaid
sequenceDiagram
    actor Player as Игрок
    participant Launcher as Launcher
    participant API as Gml.Web.Api
    participant AuthService as gml-auth
    participant Storage as users.json
    participant MC as Minecraft Server

    Player->>Launcher: Вводит логин + пароль
    Launcher->>API: POST /api/v1/integrations/auth/signin
    API->>AuthService: POST /api/v1/integrations/auth/signin
    AuthService->>Storage: Поиск пользователя
    Storage-->>AuthService: Данные пользователя
    AuthService-->>API: UUID + IsSlim + статус
    API->>API: Генерация JWT токена
    API-->>Launcher: JWT Access Token + Refresh Token
    Launcher->>Launcher: Сохранение токена в SQLite
    Launcher-->>Player: Успешный вход, доступ к профилям

    Note over Launcher,API: Загрузка файлов игры
    Launcher->>API: GET /api/v1/profiles (Bearer JWT)
    API-->>Launcher: Список профилей + файлы
    Launcher->>Launcher: Скачивание и проверка файлов

    Note over Player,MC: Запуск игры
    Launcher->>MC: Подключение :25565
    MC->>AuthService: authlib-injector verify UUID
    AuthService-->>MC: Игрок авторизован
    MC-->>Player: Добро пожаловать на сервер!
```

---

## Поток запроса пользователя

```mermaid
sequenceDiagram
    actor Admin as Администратор
    participant Browser as Браузер
    participant Proxy as Reverse Proxy
    participant Frontend as Next.js
    participant API as ASP.NET Core API
    participant SignalR as SignalR Hub
    participant DB as SQLite

    Admin->>Browser: Открывает веб-панель
    Browser->>Proxy: GET / (порт 5003)
    Proxy->>Frontend: Маршрутизация на :8081
    Frontend-->>Browser: HTML + React App

    Admin->>Browser: Создаёт Minecraft-профиль
    Browser->>Proxy: POST /api/v1/profiles
    Proxy->>API: Маршрутизация на :8082
    API->>DB: Сохранение профиля
    DB-->>API: OK
    API-->>Proxy: 201 Created
    Proxy-->>Browser: Ответ

    API->>SignalR: Уведомление об обновлении
    SignalR-->>Browser: WebSocket Push (реал-тайм)
    Browser-->>Admin: Обновление UI без перезагрузки
```

---

## Схема деплоя (Docker)

```mermaid
graph TB
    subgraph Host["Хост-машина"]
        subgraph Ports["Открытые порты"]
            P5003["Port 5003\nWeb UI + API"]
            P5000["Port 5000\nAPI напрямую"]
            P25565["Port 25565\nMinecraft игра"]
            P25575["Port 25575\nRCON (admin)"]
        end

        subgraph GmlCompose["Gml.Backend/docker-compose.yml"]
            subgraph GmlNetwork["gml-network (bridge)"]
                PROXY_C["gml-web-proxy\n:8080\nYARP Gateway"]
                API_C["gml-web-api\n:8082 internal\nASP.NET Core"]
                FRONT_C["gml-web-frontend\n:8081 internal\nNext.js"]
                SKINS_C["gml-web-skins\n:8085 internal\nSkin Service"]
            end
        end

        subgraph SerCompose["ser/docker-compose.yml"]
            MC_C["minecraft-server\n:25565\nFabric 1.19.4"]
        end

        subgraph Volumes["Постоянные данные"]
            VOL1[("gml-system-data\nSQLite DB")]
            VOL2["./data/GmlBackend\nПрофили и файлы игры"]
            VOL3["./data/backups\nБэкапы БД"]
            VOL4["./data/TextureService\nСкины и текстуры"]
            VOL5["./ser/data\nМир и плагины Minecraft"]
        end
    end

    P5003 --> PROXY_C
    P5000 --> API_C
    P25565 --> MC_C
    P25575 --> MC_C

    PROXY_C -->|"/"| FRONT_C
    PROXY_C -->|"/api/"| API_C
    PROXY_C -->|"/skins/"| SKINS_C

    API_C --- VOL1
    API_C --- VOL2
    API_C --- VOL3
    SKINS_C --- VOL4
    MC_C --- VOL5
```

---

## Жизненный цикл лаунчера

```mermaid
stateDiagram-v2
    [*] --> Запуск: Пользователь открывает лаунчер

    Запуск --> ПроверкаСети: Инициализация Gml.Client
    ПроверкаСети --> ОкноВхода: Сервер недоступен
    ПроверкаСети --> ПроверкаТокена: Сервер доступен

    ПроверкаТокена --> ОкноВхода: Токен отсутствует или истёк
    ПроверкаТокена --> ЗагрузкаПрофилей: Токен действителен

    ОкноВхода --> Авторизация: Ввод логина/пароля
    Авторизация --> ОшибкаВхода: Неверные данные
    ОшибкаВхода --> ОкноВхода: Повторный ввод
    Авторизация --> ЗагрузкаПрофилей: Успех

    ЗагрузкаПрофилей --> ВыборПрофиля: Профили загружены
    ВыборПрофиля --> ЗагрузкаФайлов: Игрок выбирает профиль
    ЗагрузкаФайлов --> ПроверкаФайлов: Скачивание игровых файлов

    ПроверкаФайлов --> ЗапускМайнкрафт: Все файлы в порядке
    ПроверкаФайлов --> ЗагрузкаФайлов: Файлы повреждены/отсутствуют

    ЗапускМайнкрафт --> ИграЗапущена: Minecraft process started
    ИграЗапущена --> ВыборПрофиля: Игрок закрыл игру
    ВыборПрофиля --> [*]: Пользователь закрыл лаунчер
```

---

## Агрегация новостей

```mermaid
graph LR
    subgraph Sources["Источники новостей"]
        TG["Telegram Bot\ntoken + channel_id"]
        DS["Discord Bot\ntoken + channel_id"]
    end

    subgraph GmlAuth["gml-auth (Go)"]
        TGPROV["telegram.go\nTelegram Provider"]
        DSPROV["discord.go\nDiscord Provider"]
        MULTIPROV["multi.go\nMulti Provider"]
        CACHE["cache.go\nTTL Cache\n(refresh_seconds)"]
        NEWSHANDLER["news.go\nHTTP Handler"]
    end

    subgraph Endpoints["REST Endpoints"]
        EP1["GET /api/news\nВсе новости"]
        EP2["GET /api/news/telegram\nТолько Telegram"]
        EP3["GET /api/news/discord\nТолько Discord"]
    end

    TG --> TGPROV
    DS --> DSPROV
    TGPROV --> MULTIPROV
    DSPROV --> MULTIPROV
    MULTIPROV --> CACHE
    CACHE --> NEWSHANDLER
    NEWSHANDLER --> EP1
    NEWSHANDLER --> EP2
    NEWSHANDLER --> EP3

    EP1 -->|"JSON"| LAUNCHER_N["Launcher\nЛента новостей"]
    EP1 -->|"JSON"| WEB_N["Web Panel\nДашборд"]
```

---

## Технологический стек

```mermaid
mindmap
  root((GML Platform))
    Backend
      ASP.NET Core 10
      Entity Framework Core
      SQLite
      JWT + BouncyCastle
      SignalR WebSocket
      YARP Reverse Proxy
      Sentry Error Tracking
      AutoMapper
      FluentValidation
      Swagger/OpenAPI
    Web Frontend
      Next.js 14
      React 18
      TypeScript
      TanStack Query 5
      Zustand State
      Radix UI
      Tailwind CSS 3
      React Hook Form
      Zod Validation
      Recharts
      Axios
    Desktop Launcher
      Avalonia 11.3
      ReactiveUI
      MVVM Pattern
      SQLite Local DB
      Hardware.Info
      Sentry
      SixLabors.ImageSharp
    Auth Service
      Go 1.25
      net/http stdlib
      JSON File Storage
      UUID Generation
      Telegram Bot API
      Discord Bot API
      TTL Cache System
    Game Server
      Java 17
      Fabric Loader
      Minecraft 1.19.4
      authlib-injector
      RCON Protocol
      Docker Container
    EasyCabinet
      NestJS 11
      Fastify
      Prisma ORM
      MySQL 8.0
      JWT + Bcrypt
      SolidJS 1.9
      Vite 6
      TailwindCSS 4
      skinview3d
      Docker Compose
      S3 Compatible Storage
```

---

## Установка и запуск

### Требования

| Инструмент | Версия | Назначение |
|-----------|--------|------------|
| Docker | 24+ | Запуск Backend, Launcher API, Minecraft-сервера |
| Docker Compose | 2.x | Оркестрация контейнеров |
| .NET SDK | 8.0+ | Сборка Launcher |
| Go | 1.25+ | Запуск gml-auth |
| Node.js | 18+ | Разработка Web Client |

---

### 1. Клонировать репозиторий

```bash
git clone https://github.com/yuliitezarygml/gml-launcher.git
cd gml-launcher
```

---

### 2. Запустить Gml.Backend (API + Web + Skins)

```bash
# Скопировать и настроить переменные окружения
cp Gml.Backend/.env.example Gml.Backend/.env

# Открыть файл и задать SECURITY_KEY (любой длинный случайный ключ)
# nano Gml.Backend/.env

# Запустить все сервисы
cd Gml.Backend
docker-compose up -d

# Проверить статус
docker-compose ps
```

| Сервис | URL |
|--------|-----|
| Веб-панель | http://localhost:5003 |
| API | http://localhost:5000 |
| Swagger UI | http://localhost:5000/swagger |

---

### 3. Запустить gml-auth (авторизация + новости)

```bash
cd gml-auth

# Настроить Telegram и Discord боты (опционально, для новостей)
# Отредактировать config/config.json

go run main.go
# Сервер запустится на http://localhost:5003
```

> Или собрать бинарник:
> ```bash
> go build -o gml-auth main.go
> ./gml-auth
> ```

---

### 4. Запустить Minecraft-сервер (ser)

```bash
cd ser

# Запустить контейнер
docker-compose up -d

# Посмотреть логи запуска (первый старт может занять 2-3 минуты)
docker-compose logs -f

# Сервер готов когда появится: "Done! For help, type "help""
```

| Параметр | Значение |
|----------|----------|
| Адрес | localhost:25565 |
| Версия | Minecraft 1.19.4 |
| Тип | Fabric |
| Режим | Offline (авторизация через gml-auth) |
| Память | 4GB RAM |
| RCON порт | 25575 |

> **RCON-консоль** (удалённое управление сервером):
> ```bash
> # Подключиться к RCON (пароль указан в ser/data/server.properties)
> docker exec -it minecraft-server rcon-cli
> ```

---

### 5. Запустить Gml.Launcher (Desktop)

```bash
cd Gml.Launcher
dotnet run --project src/Gml.Launcher
```

> Или собрать для конкретной платформы:
> ```bash
> # Windows x64
> dotnet publish src/Gml.Launcher -r win-x64 -c Release
>
> # Linux x64
> dotnet publish src/Gml.Launcher -r linux-x64 -c Release
>
> # macOS ARM64 (Apple Silicon)
> dotnet publish src/Gml.Launcher -r osx-arm64 -c Release
> ```

---

### 6. Разработка Gml.Web.Client (опционально)

```bash
cd Gml.Web.Client

# Установить зависимости
npm install

# Скопировать и настроить окружение
cp .env.example .env.local
# Задать NEXT_PUBLIC_BACKEND_URL=http://localhost:5000/api/v1

# Запустить dev-сервер
npm run dev
# Открыть http://localhost:3000
```

---

### Порядок запуска (рекомендуемый)

```mermaid
graph LR
    A["1. gml-auth\n:5003"] --> B["2. Gml.Backend\n:5003 / :5000"]
    B --> C["3. ser (Minecraft)\n:25565"]
    B --> D["4. Gml.Launcher\nDesktop App"]
```

1. **gml-auth** — должен запуститься первым (Backend использует его для авторизации)
2. **Gml.Backend** — запускает API, прокси, веб-панель и сервис скинов
3. **ser** — Minecraft-сервер (зависит от gml-auth для проверки игроков)
4. **Gml.Launcher** — подключается к Backend API

---

## Поддерживаемые Auth-интеграции

| Система | Тип |
|---------|-----|
| DataLifeEngine | CMS |
| Azuriom | Игровая CMS |
| EasyCabinet | Minecraft Cabinet |
| UnicoreCMS | CMS |
| NamelessMC | Forum |
| CustomEndpoint | Произвольный endpoint |
| WebMC | Minecraft Web |
| WordPress | CMS |
| gml-auth | Встроенный Go-сервис |

---

## API-эндпоинты

### Gml.Web.Api (`/api/v1`)

| Метод | Путь | Описание |
|-------|------|----------|
| `POST` | `/integrations/auth/signin` | Вход игрока |
| `GET` | `/profiles` | Список профилей |
| `POST` | `/profiles` | Создание профиля |
| `GET` | `/profiles/{name}` | Профиль по имени |
| `DELETE` | `/profiles/{name}` | Удаление профиля |
| `GET` | `/players` | Список игроков |
| `GET` | `/players/{name}` | Игрок по имени |
| `GET` | `/files` | Файлы профиля |
| `GET` | `/news` | Лента новостей |
| `GET` | `/settings` | Настройки сервера |
| `POST` | `/settings` | Обновление настроек |
| `GET` | `/health` | Проверка работоспособности |

### gml-auth (`/`)

| Метод | Путь | Описание |
|-------|------|----------|
| `POST` | `/` | Авторизация (основной) |
| `POST` | `/api/v1/integrations/auth/signin` | Авторизация (версионный) |
| `POST` | `/api/v1/users/refresh` | Обновление токена |
| `GET` | `/admin/users` | Список пользователей |
| `POST` | `/admin/users` | Создание пользователя |
| `GET` | `/api/news` | Все новости |
| `GET` | `/api/news/telegram` | Новости из Telegram |
| `GET` | `/api/news/discord` | Новости из Discord |

---

## Переменные окружения

### Gml.Backend `.env`

```env
SECURITY_KEY=<секретный-ключ-JWT>
PROJECT_NAME=GmlBackendPanel
PROJECT_POLICYNAME=GmlServerPolicy
SERVICE_TEXTURE_ENDPOINT=http://gml-web-skins:8085
MARKET_ENDPOINT=https://gml-market.recloud.tech
SWAGGER_ENABLED=true
PORT_GML_FRONTEND=5003
```

### Gml.Web.Client `.env.local`

```env
NEXT_PUBLIC_BACKEND_URL=http://localhost:5000/api/v1
NEXT_PUBLIC_MARKETPLACE_URL=https://gml-market.recloud.tech
```

### gml-auth `config/config.json`

```json
{
  "news": {
    "refresh_seconds": 60,
    "telegram": {
      "token": "<telegram-bot-token>",
      "channel": "<channel-id>"
    },
    "discord": {
      "token": "<discord-bot-token>",
      "channel": "<channel-id>"
    }
  }
}
```

### ser `docker-compose.yml` (ключевые параметры)

```yaml
TYPE: FABRIC
VERSION: 1.19.4
EULA: "TRUE"
ONLINE_MODE: "FALSE"      # Авторизация через authlib-injector
MEMORY: 4G
USE_AIKAR_FLAGS: "true"   # Оптимизированные JVM флаги
```

---

## Структура проекта

```
gml/
├── Gml.Backend/                       # Backend инфраструктура
│   ├── src/
│   │   ├── Gml.Web.Api/              # Основной REST API (C# .NET 10)
│   │   │   └── src/Gml.Web.Api/
│   │   │       ├── Program.cs        # Точка входа
│   │   │       ├── Core/
│   │   │       │   ├── Handlers/     # Бизнес-логика
│   │   │       │   ├── Hubs/         # SignalR хабы
│   │   │       │   ├── Integrations/ # Auth-интеграции
│   │   │       │   └── Repositories/ # Доступ к данным
│   │   │       └── Data/             # БД контекст + миграции
│   │   ├── Gml.Web.Proxy/            # Reverse Proxy (YARP)
│   │   ├── Gml.Web.Client/           # Веб-фронтенд (Next.js)
│   │   └── Gml.Web.Skin.Service/     # Сервис скинов
│   ├── docker-compose.yml
│   └── .env.example
│
├── Gml.Launcher/                      # Десктопный лаунчер (Avalonia)
│   └── src/
│       ├── Gml.Launcher/             # Основной проект UI
│       │   ├── Program.cs            # Точка входа
│       │   ├── ViewModels/           # MVVM ViewModels
│       │   ├── Views/                # XAML интерфейс
│       │   └── Core/Services/        # Сервисы лаунчера
│       └── Gml.Client/               # Клиентская библиотека API
│
├── Gml.Web.Client/                    # Standalone веб-клиент (Next.js)
│   └── src/
│       ├── app/                      # Next.js App Router
│       │   ├── auth/                 # Страницы входа/регистрации
│       │   └── dashboard/            # Панель управления
│       ├── components/               # React компоненты
│       └── shared/                   # Утилиты и хуки
│
├── gml-auth/                          # Auth сервис (Go)
│   ├── main.go                       # Точка входа, HTTP роутер
│   ├── handlers/
│   │   ├── auth.go                   # SignIn + Refresh
│   │   ├── admin.go                  # Управление пользователями
│   │   └── news.go                   # Новостной API
│   ├── models/user.go                # Модель пользователя
│   ├── storage/storage.go            # JSON-хранилище
│   ├── news/
│   │   ├── telegram.go               # Telegram провайдер
│   │   ├── discord.go                # Discord провайдер
│   │   ├── multi.go                  # Мульти-провайдер
│   │   └── cache.go                  # TTL кэш
│   └── config/config.go              # Конфигурация
│
├── ser/                               # Minecraft игровой сервер
│   ├── docker-compose.yml            # Fabric 1.19.4 / Java 17
│   └── data/
│       ├── server.properties         # Настройки сервера
│       └── libraries/
│           └── authlib-injector.jar  # Кастомная авторизация
│
└── EasyCabinet/                       # Личный кабинет игрока
    ├── packages/
    │   ├── backend/                  # NestJS Backend API
    │   │   ├── src/
    │   │   │   ├── auth/            # Авторизация (JWT + Cookie)
    │   │   │   ├── users/           # Управление пользователями
    │   │   │   ├── aurora/          # Aurora Launcher API
    │   │   │   ├── files/           # Загрузка скинов/плащей
    │   │   │   └── emails/          # Email-уведомления
    │   │   ├── prisma/
    │   │   │   └── schema.prisma    # Схема MySQL БД
    │   │   └── uploads/             # Локальное хранилище файлов
    │   │
    │   └── frontend/                 # SolidJS Frontend
    │       ├── src/
    │       │   ├── pages/           # Страницы (index, login, profile)
    │       │   ├── components/      # UI-компоненты
    │       │   ├── api/             # API-клиенты
    │       │   └── services/        # Axios, Notiflix
    │       └── public/              # Статические файлы
    │
    ├── docker-compose.yml            # MySQL + Backend + Frontend
    └── pnpm-workspace.yaml           # Monorepo конфигурация
```

---

## Платформы для Gml.Launcher

| ОС | Архитектуры |
|----|------------|
| Windows | x86, x64, ARM64 |
| Linux | x64, musl-x64, ARM64, ARM musl |
| macOS | x64, ARM64 (Apple Silicon) |

---

## EasyCabinet — Личный кабинет игрока

**EasyCabinet** — это веб-приложение для управления профилем игрока Minecraft, совместимое с [AuroraLauncher](https://github.com/AuroraTeam/AuroraLauncher). Позволяет игрокам регистрироваться, входить в систему, загружать скины и плащи.

### Архитектура EasyCabinet

```mermaid
graph TB
    subgraph Client["Клиент"]
        BROWSER["Браузер\nSolidJS App"]
    end

    subgraph Docker["Docker Compose"]
        FRONTEND["easycabinet-frontend\nNginx + SolidJS\n:3000"]
        BACKEND["easycabinet-backend\nNestJS API\n:4000"]
        MYSQL["easycabinet-mysql\nMySQL 8.0\n:3306"]
    end

    subgraph Storage["Хранилище"]
        DB[("MySQL Database\nusers, sessions")]
        UPLOADS["Uploads Volume\nskins, capes"]
    end

    BROWSER -->|"HTTP :3000"| FRONTEND
    FRONTEND -->|"API :4000"| BACKEND
    BACKEND --> MYSQL
    BACKEND --- UPLOADS
    MYSQL --- DB
```

### Технологический стек

**Backend (NestJS):**
- NestJS 11 + Fastify
- Prisma ORM + MySQL
- JWT авторизация + Cookie-based refresh tokens
- Bcrypt для хеширования паролей
- Загрузка файлов (скины/плащи)
- Email-уведомления (Nodemailer)
- Опциональная поддержка Redis для кэша
- Опциональная поддержка S3 для хранения файлов

**Frontend (SolidJS):**
- SolidJS 1.9 (реактивный фреймворк)
- Vite 6 (сборщик)
- TailwindCSS 4 (стилизация)
- Axios (HTTP-клиент)
- SolidJS Router (маршрутизация)
- skinview3d (3D-превью скинов)
- Notiflix (уведомления)

### Установка EasyCabinet

#### Требования

| Инструмент | Версия | Назначение |
|-----------|--------|------------|
| Docker | 24+ | Запуск всех сервисов |
| Docker Compose | 2.x | Оркестрация контейнеров |
| pnpm | 9+ | Менеджер пакетов (опционально для разработки) |

#### Быстрый старт с Docker

```bash
cd EasyCabinet

# Запустить все сервисы (MySQL, Backend, Frontend)
docker-compose up -d

# Проверить статус
docker-compose ps

# Посмотреть логи
docker-compose logs -f
```

| Сервис | URL | Описание |
|--------|-----|----------|
| Frontend | http://localhost:3000 | Веб-интерфейс |
| Backend API | http://localhost:4000 | REST API |
| MySQL | localhost:3306 | База данных |

#### Установка для разработки

```bash
cd EasyCabinet

# Установить зависимости (требуется pnpm)
pnpm install

# Настроить Backend
cd packages/backend
cp .env.example .env
# Отредактировать .env (см. раздел "Настройка Backend")

# Создать базу данных MySQL
mysql -u root -p
CREATE DATABASE easycabinet;
exit;

# Применить миграции Prisma
pnpx prisma migrate deploy

# Запустить Backend в dev-режиме
pnpm run start:dev
# Backend запустится на http://localhost:4000

# В новом терминале запустить Frontend
cd packages/frontend
pnpm run dev
# Frontend запустится на http://localhost:5173
```

### Настройка Backend

Создайте файл `packages/backend/.env` на основе `.env.example`:

```env
# Сервер
HOST="0.0.0.0"
PORT=4000
FRONTEND_URL="http://localhost:3000"
BACKEND_URL="http://localhost:4000"

# Проект
PROJECT_NAME="EasyCabinet"

# JWT токены
JWT_SECRET="измените-на-случайную-строку"
JWT_EXPIRES_IN="30m"

# Cookie (для refresh token)
COOKIE_SECURE=false  # true для HTTPS
COOKIE_SECRET="измените-на-случайную-строку"
COOKIE_DOMAIN="localhost"
COOKIE_EXPIRES_IN=2592000  # 30 дней в секундах

# База данных MySQL
DATABASE_URL="mysql://cabinet:cabinetpass@localhost:3306/easycabinet"

# Email (опционально)
USE_SENDMAIL=false
EMAIL_FROM="no-reply@example.com"
SMTP_HOST="localhost"
SMTP_PORT=587
SMTP_SECURE=false
SMTP_USER="user"
SMTP_PASS="pass"

# Redis (опционально, для кэша)
REDIS_URL=""

# S3 хранилище (опционально)
STORAGE_FORMAT="local"  # local или s3
S3_ACCESS_KEY_ID="minioadmin"
S3_SECRET_ACCESS_KEY="minioadmin"
S3_ENDPOINT="http://127.0.0.1:9000"
S3_REGION="not-used"
S3_BUCKET="easy-cabinet"
S3_PUBLIC_URL="http://127.0.0.1:9000"
```

### Настройка Frontend

Создайте файл `packages/frontend/.env`:

```env
VITE_API_URL="http://localhost:4000"
```

### API Эндпоинты EasyCabinet

#### Авторизация (`/auth`)

| Метод | Путь | Описание | Тело запроса |
|-------|------|----------|--------------|
| `POST` | `/auth/register` | Регистрация | `{ email, login, password }` |
| `POST` | `/auth/login` | Вход | `{ login, password }` |
| `POST` | `/auth/logout` | Выход | - |
| `POST` | `/auth/refresh` | Обновление токена | - (cookie) |
| `POST` | `/auth/reset-password` | Сброс пароля | `{ email }` |
| `POST` | `/auth/change-password` | Смена пароля | `{ resetToken, password }` |

#### Пользователи (`/users`)

| Метод | Путь | Описание | Авторизация |
|-------|------|----------|-------------|
| `GET` | `/users` | Получить профиль | JWT Bearer |
| `PUT` | `/users` | Обновить профиль (скин/плащ) | JWT Bearer |

#### Aurora Launcher API (`/aurora`)

Эндпоинты для интеграции с AuroraLauncher:

| Метод | Путь | Описание |
|-------|------|----------|
| `POST` | `/aurora/auth` | Авторизация игрока |
| `POST` | `/aurora/join` | Присоединение к серверу |
| `POST` | `/aurora/hasJoined` | Проверка сессии |
| `POST` | `/aurora/profile` | Получить профиль игрока |
| `POST` | `/aurora/profiles` | Получить несколько профилей |

### Структура базы данных

```sql
CREATE TABLE User (
  id INT PRIMARY KEY AUTO_INCREMENT,
  uuid VARCHAR(36) UNIQUE DEFAULT (UUID()),
  email VARCHAR(255) UNIQUE NOT NULL,
  login VARCHAR(255) UNIQUE NOT NULL,
  password VARCHAR(60) NOT NULL,  -- bcrypt hash
  resetToken VARCHAR(32) UNIQUE,
  accessToken VARCHAR(255),       -- для Aurora API
  serverId VARCHAR(255),           -- для Aurora API
  skinHash VARCHAR(255),           -- хеш файла скина
  capeHash VARCHAR(255),           -- хеш файла плаща
  isAlex TINYINT                   -- тип модели (Alex/Steve)
);
```

### Привязка к AuroraLauncher

Для интеграции EasyCabinet с AuroraLauncher измените конфигурацию LaunchServer:

```hjson
auth:
{
    type: json
    authUrl: http://ДОМЕН_BACKEND/aurora/auth
    joinUrl: http://ДОМЕН_BACKEND/aurora/join
    hasJoinedUrl: http://ДОМЕН_BACKEND/aurora/hasJoined
    profileUrl: http://ДОМЕН_BACKEND/aurora/profile
    profilesUrl: http://ДОМЕН_BACKEND/aurora/profiles
}
injector:
{
    skinDomains: [
        "ДОМЕН_BACKEND"
    ]
}
```

Замените `ДОМЕН_BACKEND` на адрес вашего Backend-сервера (например, `api.example.com` или `localhost:4000`).

### Структура проекта EasyCabinet

```
EasyCabinet/
├── packages/
│   ├── backend/                    # NestJS Backend API
│   │   ├── src/
│   │   │   ├── auth/              # Модуль авторизации
│   │   │   │   ├── auth.controller.ts
│   │   │   │   ├── auth.service.ts
│   │   │   │   ├── jwt.strategy.ts
│   │   │   │   └── dto/           # DTO для валидации
│   │   │   ├── users/             # Модуль пользователей
│   │   │   │   ├── users.controller.ts
│   │   │   │   ├── users.service.ts
│   │   │   │   └── prisma.service.ts
│   │   │   ├── aurora/            # Aurora Launcher API
│   │   │   │   ├── aurora.controller.ts
│   │   │   │   └── aurora.service.ts
│   │   │   ├── files/             # Сервис файлов (скины/плащи)
│   │   │   ├── emails/            # Email-уведомления
│   │   │   └── main.ts            # Точка входа
│   │   ├── prisma/
│   │   │   └── schema.prisma      # Схема БД
│   │   ├── uploads/               # Загруженные файлы
│   │   ├── .env.example
│   │   ├── Dockerfile
│   │   └── package.json
│   │
│   └── frontend/                   # SolidJS Frontend
│       ├── src/
│       │   ├── pages/             # Страницы приложения
│       │   │   ├── index.jsx      # Главная страница
│       │   │   ├── login.jsx      # Вход
│       │   │   ├── register.jsx   # Регистрация
│       │   │   ├── profile.jsx    # Профиль игрока
│       │   │   ├── forgot-password.jsx
│       │   │   └── change-password.jsx
│       │   ├── components/        # React-компоненты
│       │   │   └── Header.jsx
│       │   ├── api/               # API-клиенты
│       │   │   ├── auth.js
│       │   │   └── user.js
│       │   ├── services/          # Сервисы (axios, notiflix)
│       │   ├── layouts/           # Layouts
│       │   ├── App.jsx
│       │   └── index.css
│       ├── public/
│       ├── .env
│       ├── Dockerfile
│       ├── nginx.conf
│       └── package.json
│
├── docker-compose.yml              # Оркестрация всех сервисов
├── pnpm-workspace.yaml             # Monorepo конфигурация
└── package.json
```

### Возможности EasyCabinet

- ✅ Регистрация новых игроков
- ✅ Авторизация с JWT + Refresh Tokens
- ✅ Профиль игрока с 3D-превью скина
- ✅ Загрузка скинов (Steve/Alex модели)
- ✅ Загрузка плащей (capes)
- ✅ Сброс пароля через email
- ✅ Интеграция с AuroraLauncher
- ✅ Docker-деплой из коробки
- ✅ Поддержка S3-хранилищ (MinIO, AWS S3)
- ✅ Опциональный Redis для кэша

### Управление EasyCabinet

```bash
# Запустить все сервисы
docker-compose up -d

# Остановить все сервисы
docker-compose down

# Пересобрать контейнеры после изменений
docker-compose up --build -d

# Посмотреть логи
docker-compose logs -f backend
docker-compose logs -f frontend

# Выполнить миграции БД
docker-compose exec backend npx prisma migrate deploy

# Создать нового пользователя вручную (через MySQL)
docker-compose exec mysql mysql -u cabinet -pcabinetpass easycabinet
```

### Поток работы игрока

```mermaid
sequenceDiagram
    actor Player as Игрок
    participant Browser as Браузер
    participant Frontend as SolidJS Frontend
    participant Backend as NestJS Backend
    participant MySQL as MySQL DB
    participant Launcher as AuroraLauncher

    Player->>Browser: Открывает http://localhost:3000
    Browser->>Frontend: GET /
    Frontend-->>Browser: Главная страница

    Player->>Browser: Нажимает "Создать аккаунт"
    Browser->>Frontend: Переход на /register
    Player->>Browser: Вводит email, login, password
    Browser->>Backend: POST /auth/register
    Backend->>MySQL: INSERT User
    MySQL-->>Backend: OK
    Backend-->>Browser: 201 Created
    Browser-->>Player: "Регистрация успешна"

    Player->>Browser: Переход на /login
    Player->>Browser: Вводит login + password
    Browser->>Backend: POST /auth/login
    Backend->>MySQL: SELECT User WHERE login=?
    MySQL-->>Backend: User data
    Backend->>Backend: Проверка bcrypt hash
    Backend-->>Browser: JWT accessToken + refreshToken (cookie)
    Browser-->>Player: Перенаправление на /profile

    Player->>Browser: Загружает скин
    Browser->>Backend: PUT /users (multipart/form-data)
    Backend->>Backend: Обработка изображения (sharp)
    Backend->>MySQL: UPDATE User SET skinHash=?
    Backend-->>Browser: 200 OK
    Browser-->>Player: "Скин обновлён"

    Note over Player,Launcher: Запуск игры через AuroraLauncher
    Launcher->>Backend: POST /aurora/auth {login, password}
    Backend->>MySQL: Проверка пользователя
    Backend-->>Launcher: UUID + accessToken
    Launcher->>Launcher: Сохранение токена
    Launcher->>Backend: POST /aurora/profile {uuid, unsigned=false}
    Backend-->>Launcher: Профиль + URL скина/плаща
    Launcher-->>Player: Запуск Minecraft с кастомным скином
```

### Переменные окружения EasyCabinet

#### Backend `.env`

```env
# Основные настройки
HOST="0.0.0.0"
PORT=4000
FRONTEND_URL="http://localhost:3000"
BACKEND_URL="http://localhost:4000"
PROJECT_NAME="EasyCabinet"

# JWT авторизация
JWT_SECRET="ваш-секретный-ключ-минимум-32-символа"
JWT_EXPIRES_IN="30m"

# Cookie (refresh token)
COOKIE_SECURE=false              # true для production с HTTPS
COOKIE_SECRET="cookie-секрет"
COOKIE_DOMAIN="localhost"        # ваш домен для production
COOKIE_EXPIRES_IN=2592000        # 30 дней

# База данных
DATABASE_URL="mysql://cabinet:cabinetpass@mysql:3306/easycabinet"

# Email (опционально)
USE_SENDMAIL=false
EMAIL_FROM="no-reply@example.com"
SMTP_HOST="localhost"
SMTP_PORT=587
SMTP_SECURE=false
SMTP_USER="user"
SMTP_PASS="pass"

# Redis (опционально)
REDIS_URL=""

# S3 хранилище (опционально)
STORAGE_FORMAT="local"           # local или s3
S3_ACCESS_KEY_ID="minioadmin"
S3_SECRET_ACCESS_KEY="minioadmin"
S3_ENDPOINT="http://127.0.0.1:9000"
S3_REGION="not-used"
S3_BUCKET="easy-cabinet"
S3_PUBLIC_URL="http://127.0.0.1:9000"
```

#### Frontend `.env`

```env
VITE_API_URL="http://localhost:4000"
```

### Страницы EasyCabinet

| Путь | Описание | Авторизация |
|------|----------|-------------|
| `/` | Главная страница с кнопками скачивания лаунчера | Нет |
| `/register` | Регистрация нового игрока | Нет |
| `/login` | Вход в систему | Нет |
| `/profile` | Профиль игрока + загрузка скинов/плащей | Да |
| `/forgot-password` | Запрос сброса пароля | Нет |
| `/change-password` | Смена пароля по токену | Нет |

### Особенности реализации

**Безопасность:**
- Пароли хешируются с помощью bcrypt (10 раундов)
- JWT токены для API-запросов (короткий срок жизни)
- HTTP-only signed cookies для refresh tokens (долгий срок жизни)
- CORS настроен только для FRONTEND_URL
- Валидация всех входных данных через class-validator

**Хранение файлов:**
- Локальное хранилище: `uploads/skins/` и `uploads/capes/`
- S3-совместимое хранилище (MinIO, AWS S3, DigitalOcean Spaces)
- Автоматическая обработка изображений через sharp
- Генерация уникальных хешей для файлов

**Кэширование:**
- Встроенный file-based кэш (cache-manager-fs-hash)
- Опциональная поддержка Redis для production

### Troubleshooting

**Проблема:** Backend не может подключиться к MySQL

```bash
# Проверить статус MySQL
docker-compose ps mysql

# Посмотреть логи MySQL
docker-compose logs mysql

# Убедиться что healthcheck прошёл
docker-compose exec mysql mysqladmin ping -h localhost
```

**Проблема:** Frontend не может подключиться к Backend

```bash
# Проверить что Backend запущен
curl http://localhost:4000

# Проверить переменную VITE_API_URL в frontend/.env
cat packages/frontend/.env
```

**Проблема:** Ошибка при загрузке скина

```bash
# Проверить права на папку uploads
docker-compose exec backend ls -la /app/uploads

# Проверить логи Backend
docker-compose logs backend | grep -i error
```

### Production деплой

Для production-окружения:

1. Измените `COOKIE_SECURE=true` в backend/.env
2. Настройте HTTPS через Nginx или Traefik
3. Используйте сильные случайные значения для `JWT_SECRET` и `COOKIE_SECRET`
4. Настройте SMTP для email-уведомлений
5. Рассмотрите использование Redis для кэша
6. Настройте S3-хранилище для масштабируемости
7. Настройте регулярные бэкапы MySQL

```bash
# Пример backup MySQL
docker-compose exec mysql mysqldump -u cabinet -pcabinetpass easycabinet > backup.sql

# Восстановление
docker-compose exec -T mysql mysql -u cabinet -pcabinetpass easycabinet < backup.sql
```
