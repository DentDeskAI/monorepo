# DentDesk — API reference & запуск

## Быстрый старт

```bash
cp .env.example .env
# по желанию — добавить ANTHROPIC_API_KEY и WHATSAPP_* ключи
docker compose up -d --build

# CRM:  http://localhost:3000
# API:  http://localhost:8082
# Demo: admin@demo.kz / demo1234
```

Без ключей `ANTHROPIC_API_KEY` и `WHATSAPP_TOKEN` система работает на **mock LLM** + не шлёт сообщения в WhatsApp наружу (но всё остальное — webhook, создание записей, CRM, SSE — работает полностью). Это удобно для разработки.

---

## Примеры API

### Login

```bash
curl -X POST http://localhost:8082/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@demo.kz","password":"demo1234"}'
```

Ответ:

```json
{
  "token": "eyJhbGciOi...",
  "user": {
    "id": "...", "email": "admin@demo.kz", "name": "Демо Админ",
    "role": "owner", "clinic_id": "11111111-..."
  }
}
```

### Список чатов

```bash
curl http://localhost:8082/api/chats \
  -H "Authorization: Bearer <TOKEN>"
```

### История сообщений

```bash
curl http://localhost:8082/api/chats/<conv_id>/messages \
  -H "Authorization: Bearer <TOKEN>"
```

### Отправить от имени оператора (бот замолкает)

```bash
curl -X POST http://localhost:8082/api/chats/<conv_id>/send \
  -H "Authorization: Bearer <TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{"body":"Добрый день! Это Марина из регистратуры."}'
```

### Вернуть диалог боту

```bash
curl -X POST http://localhost:8082/api/chats/<conv_id>/release \
  -H "Authorization: Bearer <TOKEN>"
```

### Календарь

```bash
curl "http://localhost:8082/api/calendar?from=2026-04-24T00:00:00Z&to=2026-05-01T00:00:00Z" \
  -H "Authorization: Bearer <TOKEN>"
```

### SSE (real-time)

```bash
curl -N "http://localhost:8082/api/events?token=<TOKEN>"
```

Отдаёт events типов `message` и `appointment`.

---

## WhatsApp webhook

### Verification (GET, Meta периодически проверяет)

Meta дергает:
```
GET /webhook/whatsapp?hub.mode=subscribe&hub.verify_token=<WHATSAPP_VERIFY_TOKEN>&hub.challenge=XYZ
```
Мы возвращаем `XYZ`, если токен совпадает.

### Inbound (POST, реальное сообщение пациента)

Пример тестового payload'а (можно отправить локально — бот ответит):

```bash
curl -X POST http://localhost:8082/webhook/whatsapp \
  -H "Content-Type: application/json" \
  -d '{
    "object": "whatsapp_business_account",
    "entry": [{
      "id": "BUSINESS_ACCOUNT_ID",
      "changes": [{
        "field": "messages",
        "value": {
          "messaging_product": "whatsapp",
          "metadata": {
            "display_phone_number": "77001234567",
            "phone_number_id": "PHONE_ID_OF_CLINIC"
          },
          "contacts": [{"profile":{"name":"Айдар"},"wa_id":"77019998877"}],
          "messages": [{
            "from": "77019998877",
            "id": "wamid.TESTID001",
            "timestamp": "1714800000",
            "type": "text",
            "text": {"body": "болит зуб, хочу записаться"}
          }]
        }
      }]
    }]
  }'
```

После этого в CRM появится новый чат, и Айгуль ответит.

---

## Как Айгуль думает (flow одной итерации)

1. Прилетел webhook → идемпотентность по `wa_message_id` в Redis.
2. Найдена клиника (по `phone_number_id`) → найден/создан пациент → найден/создан диалог.
3. Входящее сообщение сохранено в `messages` (UNIQUE по `wa_message_id`).
4. SSE event `message` → в CRM появилось.
5. Если диалог в `handoff` — бот молчит, человек разберётся.
6. Иначе: **LLM.Orchestrator.Handle**:
   - Если в `state.pending_slots` есть слоты, а пациент написал `"1"` / `"10:00"` / `"да"` — распознали выбор → `sched.CreateAppointment` → ответ "Записала вас на ...".
   - Иначе: `classify()` → JSON `{intent, service, doctor, when, language}`.
   - Если `intent ∈ {booking, urgent_pain}` → `sched.GetFreeSlots()` → кладём в state.
   - `generateReply()` с системным промптом Айгуль + история + метка `[СВОБОДНЫЕ СЛОТЫ: ...]` или `[НЕТ СЛОТОВ]`.
   - `ApplyGuardrails(reply)` — если LLM начал советовать лекарства, подменяем на безопасный шаблон.
7. Обновляем `conversations.context` (новое состояние диалога).
8. Сохраняем outbound сообщение, публикуем SSE.
9. `whatsapp.SendText()` реально отправляет пациенту.

---

## Переключение LLM / Scheduler

### LLM провайдер

В `.env`:

- `LLM_PROVIDER=anthropic` + `ANTHROPIC_API_KEY=...` (production default)
- `LLM_PROVIDER=groq` + `GROQ_API_KEY=...` (дешевле, быстрее)
- `LLM_PROVIDER=mock` (без ключей, детерминированные ответы — для e2e тестов)

### Scheduler

В `.env` `SCHEDULER_DEFAULT`:

- `local` — слоты из нашей PG (подходит клиникам без MacDent)
- `mock` — красивые фейковые слоты (подходит для демо клиентам)
- `macdent` — HTTP в MacDent API (заполните `macdent_base_url` и `macdent_api_key` на уровне клиники)

В проде `scheduler_type` в таблице `clinics` важнее ENV — каждая клиника может иметь свой.

---

## Deploy на Hetzner (коротко)

1. Hetzner CX22 (2 vCPU, 4 GB) + домен.
2. `docker compose up -d --build` на сервере.
3. Поставить **Caddy** или nginx перед портами 5173 (web) и 8080 (api) с TLS (`Caddyfile` — 3 строки):
   ```
   dentdesk.example.com {
     reverse_proxy web:80
   }
   ```
4. В Meta Developer Console → WhatsApp → настроить webhook:
   - URL: `https://dentdesk.example.com/webhook/whatsapp`
   - Verify token: значение `WHATSAPP_VERIFY_TOKEN` из `.env`
5. Subscribe на поле `messages`.

---

## Структура

```
dentdesk/
├── cmd/api/main.go          — HTTP сервер
├── cmd/worker/main.go       — напоминания и follow-up
├── internal/
│   ├── auth/                JWT + bcrypt
│   ├── patients/            repo
│   ├── conversations/       repo + идемпотентный insert
│   ├── appointments/        repo + due-очереди
│   ├── doctors/             repo
│   ├── scheduler/           интерфейс + 3 адаптера
│   ├── llm/                 Anthropic/Groq/Mock + persona + guardrails + orchestrator
│   ├── whatsapp/            Cloud API клиент + webhook parser
│   ├── notifications/       24h / 1h / follow-up
│   ├── realtime/            SSE hub
│   ├── http/                router + middleware + handlers
│   └── platform/            config, db, redis, logger, errors
├── migrations/              001_init.up.sql
├── web/                     React + Vite + Tailwind
├── docker/                  Dockerfile.api, .worker, .web, nginx.conf
├── docker-compose.yml
├── .env.example
├── Makefile
└── README.md
```
