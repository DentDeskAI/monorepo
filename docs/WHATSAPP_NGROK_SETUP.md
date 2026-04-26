# WhatsApp + ngrok Setup

This project already supports:

- inbound WhatsApp webhooks on `GET/POST /webhook/whatsapp`
- outbound operator replies from CRM on `POST /api/chats/:id/send`
- automatic reminders from the worker

## 1. Start the stack

From the repo root:

```powershell
docker compose up -d --build
```

Local URLs:

- CRM: `http://localhost:3000`
- API: `http://localhost:8082`

## 2. Expose the CRM with ngrok

Tunnel the web container, not the API container. The nginx layer already proxies `/api/*` and `/webhook/*` to the backend.

```powershell
ngrok http 3000
```

Copy the public HTTPS URL from ngrok, for example:

```text
https://abc123.ngrok-free.app
```

If you also want the MacDent helper endpoint to generate the public URL correctly, set:

```env
PUBLIC_API_BASE_URL=https://abc123.ngrok-free.app
```

## 3. Configure WhatsApp webhook in Meta

In Meta Developer Dashboard > WhatsApp > Configuration:

- Callback URL: `https://abc123.ngrok-free.app/webhook/whatsapp`
- Verify token: your `WHATSAPP_VERIFY_TOKEN` from `.env`

Then subscribe the app to the `messages` field.

## 4. Connect the number to your clinic

The inbound webhook resolves the clinic by `whatsapp_phone_id`, so the clinic record must contain the same Meta phone number id as `WHATSAPP_PHONE_NUMBER_ID`.

You can now set it through the clinic API:

```bash
curl -X PUT http://localhost:8082/api/clinic \
  -H "Authorization: Bearer <TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Demo Dental Almaty",
    "timezone": "Asia/Almaty",
    "scheduler_type": "mock",
    "slot_duration_min": 30,
    "whatsapp_phone_id": "<your phone number id>"
  }'
```

If you prefer to inspect the current clinic object first:

```bash
curl http://localhost:8082/api/clinic \
  -H "Authorization: Bearer <TOKEN>"
```

## 5. Test inbound and outbound

Inbound:

- send a real WhatsApp message from your test phone to the business number
- the message should appear in CRM chat list

Outbound from CRM:

- open the chat in CRM
- send a message as operator
- the backend will call WhatsApp Cloud API and the patient should receive it

## 6. Reminder behavior

The worker now sends:

- 24-hour reminder
- 1-hour reminder
- follow-up after completed visits

Start the worker with Docker Compose; it runs automatically as the `worker` service.

## 7. Production note

This repo currently sends plain text WhatsApp messages through the Cloud API text endpoint. For production reminder flows, approved WhatsApp templates are the safer next step for appointment reminders and other business-initiated messages.
