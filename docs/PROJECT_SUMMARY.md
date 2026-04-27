# DentDesk Project Summary

## What DentDesk Is

DentDesk is a dental CRM focused on inbound patient communication, appointment booking, and operator workflow. The product combines a WhatsApp-based assistant with an internal CRM so clinics can capture leads, answer routine questions, and move patients into confirmed appointments.

## Core Product Scope

- WhatsApp chatbot for first-line communication with patients
- Operator CRM for live chat handoff and conversation history
- Scheduling layer that can work with local data, mock data, or MacDent
- Patient, doctor, chair, and appointment management
- Real-time updates in the CRM via SSE
- Reminder and follow-up workflows handled by a background worker

## Current Technical Stack

- Backend: Go, Gin, sqlx, PostgreSQL, Redis
- Frontend: React, Vite, Tailwind
- AI orchestration: Anthropic or Groq, with a mock fallback for local development
- External integrations: WhatsApp Cloud API and MacDent
- Runtime: Docker Compose

## How The System Works

1. A patient writes to the clinic in WhatsApp.
2. The webhook creates or updates the patient and conversation in DentDesk.
3. The LLM orchestrator classifies intent and, when needed, asks the scheduler for free slots.
4. DentDesk replies automatically or hands the conversation to an operator.
5. When an appointment is created, the CRM and reminder worker keep the clinic workflow in sync.

## MacDent In This Project

MacDent acts as the external scheduling and clinic-management source. DentDesk already uses a MacDent scheduler adapter for doctor, patient, free-slot, and appointment operations. The next integration layer is inbound MacDent webhooks so DentDesk can react when records change on the MacDent side.

## Near-Term Integration Priorities

- Receive MacDent webhook events on a clinic-specific public URL
- Validate and deduplicate incoming webhook deliveries
- Decide which MacDent entities should trigger local synchronization
- Add per-entity sync handlers for appointments, patients, and status changes
- Expose the generated webhook URL so each clinic can register it in MacDent
