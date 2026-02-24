# GML Auth — News Integration Design

**Date:** 2026-02-24

## Overview

Add `GET /api/news` endpoint to gml-auth that GML Launcher calls to display news in the launcher.
News is pulled from either Telegram or Discord — one source configured in `config.json`.

## Requirements

- `GET /api/news?limit=N&offset=N` → `[{id, title, description, createdAt}]`
- Default limit: 10
- Source: Telegram bot OR Discord bot — selected via `config.json`
- In-memory cache with background refresh (every 60s) — fast responses, rate-limit safe

## Architecture

```
GML Launcher → GET /api/news → NewsHandler → Cache (memory) → instant response
                                                  ↑
                                  background goroutine (60s) → Telegram / Discord API
```

## Config

File: `config.json` (next to `data/`)

```json
{
  "news": {
    "source": "telegram",
    "token": "BOT_TOKEN_HERE",
    "channel": "@mychannel",
    "refresh_seconds": 60
  }
}
```

For Discord:
```json
{
  "news": {
    "source": "discord",
    "token": "BOT_TOKEN_HERE",
    "channel": "123456789012345678",
    "refresh_seconds": 60
  }
}
```

## File Structure

```
config/
  config.go         — load/parse config.json
news/
  provider.go       — NewsProvider interface + NewsItem struct
  telegram.go       — Telegram Bot API client
  discord.go        — Discord API client
  cache.go          — in-memory cache + background refresh goroutine
handlers/
  news.go           — GET /api/news handler (limit/offset, serve from cache)
```

## Data Model

### NewsItem (internal + API response)

```go
type NewsItem struct {
    ID          int    `json:"id"`
    Title       string `json:"title"`
    Description string `json:"description"`
    CreatedAt   string `json:"createdAt"`
}
```

### NewsProvider interface

```go
type NewsProvider interface {
    Fetch(limit int) ([]NewsItem, error)
}
```

## Field Mapping

### Telegram → NewsItem
| Telegram field | NewsItem field | Notes |
|---|---|---|
| `message_id` | `id` | direct |
| first line of `text` | `title` | split on `\n`, max 100 chars |
| full `text` | `description` | full message |
| `date` (unix) | `createdAt` | ISO 8601 |

### Discord → NewsItem
| Discord field | NewsItem field | Notes |
|---|---|---|
| `id` (snowflake, truncated) | `id` | last 9 digits as int |
| first 80 chars of `content` | `title` | truncate at word boundary |
| `content` | `description` | full message |
| `timestamp` | `createdAt` | already ISO 8601 |

## API

### GET /api/news

Query params:
- `limit` — number of items (default: 10, max: 100)
- `offset` — skip N items (default: 0)

Response `200 OK`:
```json
[
  {
    "id": 42,
    "title": "Обновление сервера",
    "description": "Обновление сервера\nДобавлены новые плагины...",
    "createdAt": "2024-01-15T12:00:00Z"
  }
]
```

If news source not configured or cache empty: returns `[]`

## Error Handling

- Cache fetch fails → log error, return stale cache or `[]`
- Invalid config → server starts but logs warning, `/api/news` returns `[]`
- Rate limit hit → retry on next tick, serve stale cache

## Testing

- Unit tests for Telegram and Discord field mapping
- Unit test for cache (serve stale on error, refresh on tick)
- Integration test for `/api/news` handler (mock provider)
