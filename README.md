# DentDesk MVP

WhatsApp-бот "Айгуль" + mini-CRM для стоматологий (Алматы).

## Стек

- **Backend**: Go 1.22, Gin, sqlx, PostgreSQL 16, Redis 7
- **Frontend**: React 18, Vite, Tailwind, shadcn/ui
- **LLM**: Anthropic Claude (с fallback на Groq/Llama)
- **Integrations**: WhatsApp Cloud API, MacDent API (через adapter)
- **Infra**: Docker Compose

## Быстрый старт

```bash
cp .env.example .env
# заполнить WHATSAPP_*, ANTHROPIC_API_KEY
docker-compose up --build
# миграции применятся автоматически из /migrations
# CRM: http://localhost:5173 (admin@demo.kz / demo1234)
# API: http://localhost:8080
```

## Структура

См. `docs/ARCHITECTURE.md` или блок "Структура проекта" ниже.

```
cmd/              — точки входа (api + worker)
internal/         — бизнес-логика, разбитая по доменам
migrations/       — SQL миграции (idempotent)
web/              — React фронт
docker/           — Dockerfile'ы
```

## Ключевые решения

1. **Модульный монолит** на Go — быстрее для MVP, чистые швы под распил.
2. **Scheduler интерфейс** с 3 имплементациями (`local`, `mock`, `macdent`) — можно продавать клиникам без MacDent.
3. **LLM orchestrator** — отдельный слой с guardrails, system prompt, intent/entity extraction. Сменить Claude на Groq = одна строка в `cmd/api/main.go`.
4. **SSE вместо WebSocket** для realtime-чатов в CRM — проще, достаточно для push'а операторам.
5. **Идемпотентность webhook'а** — дедуп по `wa_message_id` через Postgres UNIQUE + Redis SETNX.
6. **Worker отдельный бинарь** — напоминания за 24ч и за 2ч, follow-up через 3 дня после визита.

## План запуска MVP (1–2 дня)

**День 1 (бэкенд)**
- [ ] `docker-compose up` → PG + Redis работают
- [ ] Миграции применены, seed-данные есть
- [ ] `/webhook/whatsapp` принимает сообщения (можно через ngrok + Meta test number)
- [ ] LLM отвечает с персоной Айгуль, детектит intent `booking`
- [ ] MockAdapter выдаёт слоты, Appointment создаётся

**День 2 (фронт + прод)**
- [ ] CRM логин, список чатов, Chat UI, календарь, список пациентов
- [ ] SSE работает — новое сообщение → появляется в UI оператора
- [ ] Worker шлёт напоминание за 24ч (тест через `ALTER` → старая дата)
- [ ] Deploy на Hetzner: `docker-compose.yml` + Caddy для TLS + домен
- [ ] Зарегистрировать WhatsApp Business API → вебхук на домен

## Что НЕ входит в MVP

Финансы, оплата, мед.карта, сложная аналитика, мульти-язык UI (только RU+KK в боте).
