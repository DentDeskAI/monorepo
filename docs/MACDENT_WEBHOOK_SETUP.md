# MacDent Webhook Setup

## New DentDesk Endpoints

- Public receiver: `POST /webhook/macdent/:clinicID`
- Helper endpoint for authenticated owners/admins: `GET /api/clinic/macdent/webhook-url`

The helper endpoint returns the exact URL that should be registered in MacDent for the current clinic. If `MACDENT_WEBHOOK_TOKEN` is configured, the URL includes `?token=...`.

## What The Current Implementation Does

- Accepts MacDent webhook calls on a clinic-specific URL
- Optionally validates a shared token from the query string or `X-MacDent-Token`
- Deduplicates identical deliveries for 15 minutes using Redis
- Logs the clinic, event name, entity type, and object identifier when present
- Always returns `HTTP 200` for accepted deliveries so MacDent stops retrying

This is the correct first step for integration because MacDent retries every 2 minutes on non-200 responses and stops only after repeated failures.

## Recommended Setup Steps

1. Set `MACDENT_WEBHOOK_TOKEN` in `.env` to a long random string.
2. Make sure your public domain points to the DentDesk web/API entrypoint.
3. Start the stack and log in as an owner or admin.
4. Call `GET /api/clinic/macdent/webhook-url` with your Bearer token.
5. Copy the returned URL into MacDent using `webhook/set`.
6. Run a debug delivery from MacDent with `webhook/send_debug`.
7. Check that DentDesk returns `200` and logs the event.

## Example

If the clinic ID is `11111111-1111-1111-1111-111111111111` and your public host is `https://crm.example.com`, the webhook URL will look like:

```text
https://crm.example.com/webhook/macdent/11111111-1111-1111-1111-111111111111?token=change-me
```

## What You Still Need To Add Next

- Entity-specific sync logic for `patient`, `zapis`, `appointment`, and other MacDent objects you care about
- A reconciliation strategy: pull fresh data from MacDent after `onCreate`, `onChange`, and `onRemove`
- Optional persistence of raw webhook payloads if you need an audit trail
- UI or admin settings for MacDent credentials and webhook status inspection

## Suggested Event Handling Plan

- `onCreate` for `patient`: refresh or create the patient locally
- `onChange` for `patient`: update patient profile fields
- `onCreate` / `onChange` for `zapis`: sync appointments and statuses
- `onRemove` for `zapis`: cancel or archive the corresponding local appointment
- Financial entities such as `payment`: only sync if they are part of CRM workflows you expose in DentDesk

## Operational Notes

- MacDent stops retrying only when your endpoint responds with `200`
- Invalid token or invalid clinic URL will cause MacDent to keep retrying until fixed
- Keep Redis enabled in all environments where you expect webhook retries, otherwise duplicate deliveries may be processed more than once
