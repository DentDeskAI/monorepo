# DentDesk Architecture

## Vision (target)

DentDesk is a standalone CRM. Its **canonical data lives in our PostgreSQL**:
clinics, doctors, patients, appointments, conversations, chairs. External
systems (MacDent today; IDENT, Dental4Windows, etc. later) are **optional
integrations** that synchronise data with our DB through background workers.

```
            HTTP request                     write
Frontend ─────────────► API handlers ─────────────► PostgreSQL (source of truth)
                            │                            ▲
                            └───── read ─────────────────┘
                                                         │
                                       ┌─────────────────┴─────────────────┐
                                       ▼                                   ▼
                           Sync worker: MacDent                Sync worker: IDENT (future)
                           internal/integrations/macdent       internal/integrations/ident
```

User-facing requests **never** hit an integration synchronously. If MacDent
goes down, DentDesk continues to serve doctors, patients, appointments from
our DB.

## Where we are now (Phase 1.5)

We are not at the target yet. Today, several read endpoints (`/api/doctors`,
`/api/patients`, `/api/schedule/*`) read through the configured scheduling
backend for the clinic. For `scheduler_type='macdent'` that still means live
MacDent calls; for `scheduler_type='local'` and `mock`, it means local
PostgreSQL repos. This keeps the MVP usable for both staged demo clinics and
clinics already populated in MacDent.

What has been cleaned up:

- The fake `Scheduler` interface with three implementations (`LocalAdapter`,
  `MockAdapter`, `MacDentAdapter`) was deleted. There was only ever one real
  implementation; the others were dead code.
- `internal/macdent/` moved to `internal/integrations/macdent/` to make it
  visible that this is **one of N possible** integrations.
- `internal/scheduler/` owns the app-level scheduling contract. `Registry`
  dispatches to either the MacDent-backed `Service` or `LocalScheduler` based
  on `clinics.scheduler_type`.
- The public scheduler contract uses DentDesk domain DTOs; MacDent response
  shapes are translated inside the MacDent-backed implementation.
- `services.SchedulingService` was trimmed to methods that carry actual
  business logic (`CreateAppointment`, `UpdateAppointmentStatus`,
  `SyncDoctors`, conversation lifecycle). Pass-through reads were deleted.
- Duplicate routes (`/api/schedule/doctors` ⇆ `/api/doctors`,
  `/api/schedule/patients` ⇆ `/api/patients`,
  `/api/schedule/clinic` ⇆ `/api/clinic`) were collapsed to one canonical
  path each.
- Frontend (`web/src/api/client.js`) was updated to use the canonical paths.

## Conventions

- **`internal/integrations/<name>/`** — pure clients to external systems.
  No DB access. No business logic. Just typed wrappers around HTTP/SDK calls.
  `ClientFor(ctx, db, http, clinicID)` is the single entry point that resolves
  per-tenant credentials and returns a ready-to-use client.
- **`internal/scheduler/`** — domain types (`Doctor`, `Patient`, `Slot`,
  `Appointment`, `AppointmentDetail`, `RevenueRecord`) and the scheduling
  backend contract. `Registry` is the entry point used by handlers and services.
- **`internal/services/`** — business-logic services. Validation, orchestration,
  cross-repo writes. No pass-through methods.
- **`internal/<entity>/`** (e.g. `doctors/`, `patients/`, `appointments/`) —
  repositories over PostgreSQL. These are already the primary read source for
  local/mock clinics and will become the primary read source for MacDent clinics
  after sync is introduced.

## Phase 2 (next, when justified)

Triggers: first paying clinic, OR first integration besides MacDent, whichever
comes first.

- Add a sync worker (`cmd/worker/`) that periodically pulls from MacDent into
  our `doctors` and `patients` tables.
- Switch handlers to read from local repos (`doctors.Repo.ListByClinic`, etc.).
- Move appointment writes to: write local first → enqueue push to MacDent.
- Keep the scheduler contract stable; only the MacDent implementation and sync
  worker internals change.

## Phase 3 (when there are two integrations)

- Introduce an `Integration` interface designed against two real
  external integrations.
- Reorganise `internal/integrations/` so each integration implements that
  interface.

Until then: keep external DTOs inside integration packages and expose only
DentDesk domain types to handlers, services, and the frontend API.
