import { api } from "../../api/client";

const STATUS_BY_CODE = {
  0: "scheduled",
  1: "confirmed",
  2: "cancelled",
  3: "completed",
  4: "completed",
  5: "confirmed",
  6: "scheduled",
};

export const STATUS_CARD = {
  scheduled: "bg-blue-50 dark:bg-blue-900/30 text-blue-800 dark:text-blue-200 border-blue-200 dark:border-blue-800",
  confirmed: "bg-emerald-50 dark:bg-emerald-900/30 text-emerald-800 dark:text-emerald-200 border-emerald-200 dark:border-emerald-800",
  completed: "bg-slate-50 dark:bg-slate-700/50 text-slate-500 dark:text-slate-400 border-slate-200 dark:border-slate-600",
  cancelled: "bg-slate-50 dark:bg-slate-700/30 text-slate-400 dark:text-slate-500 border-slate-100 dark:border-slate-700 opacity-60",
};

const refCache = { doctors: null, patients: null };

export function fetchDoctors() {
  if (refCache.doctors) return refCache.doctors;
  refCache.doctors = api.doctors().catch(() => []);
  return refCache.doctors;
}

export function fetchPatients() {
  if (refCache.patients) return refCache.patients;
  refCache.patients = api.patients().catch(() => []);
  return refCache.patients;
}

export function parseScheduleDate(s) {
  if (!s) return new Date();
  const [datePart, timePart = "00:00:00"] = s.split(" ");
  const [d, m, y] = datePart.split(".");
  const [h, min, sec] = timePart.split(":");
  return new Date(+y, +m - 1, +d, +h, +min, +sec);
}

export function normalizeScheduleAppointment(a, doctorMap, patientMap) {
  const patient = patientMap[a.patient];
  return {
    id: a.id,
    starts_at: parseScheduleDate(a.start),
    ends_at: parseScheduleDate(a.end),
    patient_name: patient?.name ?? null,
    patient_phone: patient?.phone ?? null,
    patient_is_first: a.isFirst,
    doctor_name: doctorMap[String(a.doctor)] ?? null,
    cabinet: a.cabinet || null,
    zhaloba: a.zhaloba || null,
    status: STATUS_BY_CODE[a.status] ?? "completed",
  };
}

export function startOfWeek(d) {
  const x = new Date(d);
  const day = (x.getDay() + 6) % 7;
  x.setHours(0, 0, 0, 0);
  x.setDate(x.getDate() - day);
  return x;
}

export function fmtTime(date) {
  return `${String(date.getHours()).padStart(2, "0")}:${String(date.getMinutes()).padStart(2, "0")}`;
}

export function fmtDateForInput(d) {
  if (!d) return "";
  const pad = (n) => String(n).padStart(2, "0");
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`;
}

export function fmtDateLong(date) {
  return date.toLocaleDateString("ru-RU", { day: "numeric", month: "long", year: "numeric" });
}
