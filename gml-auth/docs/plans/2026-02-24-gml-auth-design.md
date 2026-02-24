# GML Auth — Design Document
**Date:** 2026-02-24

## Overview

Go HTTP server providing custom authentication backend for GML Launcher.
Implements the GML Launcher custom auth protocol with JSON file as database.

## Requirements

- Handle POST `/api/v1/integrations/auth/signin` for GML Launcher
- Login + Password auth (no 2FA)
- Plain text passwords (easy manual editing)
- User blocking with reason (403)
- Admin REST API for user management
- Zero external dependencies (stdlib only)

## Architecture

Single Go HTTP server, one port (5003), two route groups:
- Public: `/api/v1/integrations/auth/signin`
- Admin: `/admin/users`

## Data Model

File: `data/users.json`

```json
{
  "users": [
    {
      "uuid": "c07a9841-2275-4ba0-8f1c-2e1599a1f22f",
      "login": "GamerVII",
      "password": "mypassword",
      "is_slim": false,
      "blocked": false,
      "block_reason": ""
    }
  ]
}
```

## Endpoints

### Auth (for GML Launcher)

**POST** `/api/v1/integrations/auth/signin`

Request:
```json
{ "Login": "user", "Password": "pass", "Totp": "" }
```

Responses:
- `200` — success with user data
- `401` — wrong password
- `404` — user not found
- `403` — user blocked

### Admin API

| Method | Path | Action |
|--------|------|--------|
| GET | `/admin/users` | list all users |
| POST | `/admin/users` | create user |
| DELETE | `/admin/users/{login}` | delete user |
| PATCH | `/admin/users/{login}/block` | block user |
| PATCH | `/admin/users/{login}/unblock` | unblock user |

## Project Structure

```
gml-auth/
├── main.go
├── handlers/
│   ├── auth.go
│   └── admin.go
├── storage/
│   └── storage.go
├── models/
│   └── user.go
└── data/
    └── users.json
```
