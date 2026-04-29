import { Fragment, useEffect, useMemo, useState } from "react";
import { api } from "../api/client";
import { useTranslation } from "../hooks/useTranslation";
import Modal, { FormField, inputCls, btnPrimary, btnSecondary, btnDanger, btnGhost } from "../components/Modal";

// i18n Keys
const i18nKeys = {
  calendar: {
    title: "calendar.title",
    subtitle: "calendar.subtitle",
    loading: "calendar.loading",
    today: "calendar.today",
    prevWeek: "calendar.prev_week",
    nextWeek: "calendar.next_week",
    no_patient: "calendar.no_patient",
  },
  days: {
    mon: "calendar.days.mon",
    tue: "calendar.days.tue",
    wed: "calendar.days.wed",
    thu: "calendar.days.thu",
    fri: "calendar.days.fri",
    sat: "calendar.days.sat",
    sun: "calendar.days.sun",
  },
};

// Parses MacDent date strings like "28.04.2026 10:45:00"
function parseMacdentDate(s) {
  const [datePart, timePart = "00:00:00"] = s.split(" ");
  const [d, m, y] = datePart.split(".");
  const [h, min, sec] = timePart.split(":");
  return new Date(+y, +m - 1, +d, +h, +min, +sec);
}

const MD_STATUS = { 0: "scheduled", 1: "confirmed", 2: "cancelled" };

function normalizeMacdent(a, doctorMap, patientMap) {
  const patient = patientMap[a.patient];
  return {
    id: a.id,
    starts_at: parseMacdentDate(a.start),
    ends_at: parseMacdentDate(a.end),
    patient_name: patient?.name ?? null,
    patient_phone: patient?.phone ?? null,
    patient_is_first: a.isFirst,
    doctor_name: doctorMap[String(a.doctor)] ?? null,
    cabinet: a.cabinet || null,
    zhaloba: a.zhaloba || null,
    status: MD_STATUS[a.status] ?? "completed",
  };
}

function fmtTime(date) {
  return `${String(date.getHours()).padStart(2, "0")}:${String(date.getMinutes()).padStart(2, "0")}`;
}

const STATUS_CARD = {
  scheduled: "bg-blue-50 dark:bg-blue-900/30 text-blue-800 dark:text-blue-200 border-blue-200 dark:border-blue-800",
  confirmed: "bg-emerald-50 dark:bg-emerald-900/30 text-emerald-800 dark:text-emerald-200 border-emerald-200 dark:border-emerald-800",
  completed: "bg-slate-50 dark:bg-slate-700/50 text-slate-500 dark:text-slate-400 border-slate-200 dark:border-slate-600",
  cancelled: "bg-slate-50 dark:bg-slate-700/30 text-slate-400 dark:text-slate-500 border-slate-100 dark:border-slate-700 opacity-60",
};

// Module-level cache — survives component unmount/remount (cleared only on hard refresh).
const _refCache = { doctors: null, patients: null };

function fetchDoctors() {
  if (_refCache.doctors) return _refCache.doctors;
  _refCache.doctors = api.doctors().catch(() => []);
  return _refCache.doctors;
}

function fetchPatients() {
  if (_refCache.patients) return _refCache.patients;
  _refCache.patients = api.patients().catch(() => []);
  return _refCache.patients;
}

function startOfWeek(d) {
  const x = new Date(d);
  const day = (x.getDay() + 6) % 7; // Mon=0
  x.setHours(0, 0, 0, 0);
  x.setDate(x.getDate() - day);
  return x;
}

export default function Calendar() {
  const { t } = useTranslation();
  const [weekStart, setWeekStart] = useState(() => startOfWeek(new Date()));
  const [items, setItems] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [tick, setTick] = useState(0);
  const [createCtx, setCreateCtx] = useState(null);   // { date: Date }
  const [detailCtx, setDetailCtx] = useState(null);   // appt object
  const [doctors, setDoctors] = useState([]);
  const [patients, setPatients] = useState([]);

  const daysShort = useMemo(
      () => [
        t(i18nKeys.days.mon),
        t(i18nKeys.days.tue),
        t(i18nKeys.days.wed),
        t(i18nKeys.days.thu),
        t(i18nKeys.days.fri),
        t(i18nKeys.days.sat),
        t(i18nKeys.days.sun),
      ],
      [t]
  );

  const weekEnd = useMemo(() => {
    const e = new Date(weekStart);
    e.setDate(e.getDate() + 7);
    return e;
  }, [weekStart]);

  useEffect(() => {
    setLoading(true);
    setError(null);

    Promise.all([
      api.scheduleDoctors(weekStart.toISOString(), weekEnd.toISOString()),
      fetchDoctors(),
      fetchPatients(),
    ])
        .then(([resp, docs, pats]) => {
          setDoctors(docs);
          setPatients(pats);
          const doctorMap  = Object.fromEntries(docs.map((d) => [d.id, d.name]));
          const patientMap = Object.fromEntries(pats.map((p) => [p.id, p]));
          const normalized = (resp?.appointments ?? [])
              .map((a) => normalizeMacdent(a, doctorMap, patientMap))
              .filter(({ starts_at }) => starts_at >= weekStart && starts_at < weekEnd);
          setItems(normalized);
        })
        .catch((err) => {
          console.error("Failed to fetch schedule:", err);
          setError("Failed to load schedule");
          setItems([]);
        })
        .finally(() => setLoading(false));
  }, [weekStart, weekEnd, tick]);

  const refresh = () => setTick((n) => n + 1);

  const byDayHour = useMemo(() => {
    const m = {};
    for (const a of items) {
      const d = a.starts_at;
      const dayIdx = (d.getDay() + 6) % 7;
      const key = `${dayIdx}-${d.getHours()}`;
      if (!m[key]) m[key] = [];
      m[key].push(a);
    }

    // SORTING FIX: Sort appointments in each hour slot by minutes
    Object.keys(m).forEach(key => {
      m[key].sort((a, b) => a.starts_at.getTime() - b.starts_at.getTime());
    });

    return m;
  }, [items]);

  const shift = (delta) => {
    const d = new Date(weekStart);
    d.setDate(d.getDate() + delta * 7);
    setWeekStart(d);
  };

  const fmtRange = () => {
    const e = new Date(weekEnd);
    e.setDate(e.getDate() - 1);
    const f = (x) => x.toLocaleDateString("ru-RU", { day: "2-digit", month: "short" });
    return `${f(weekStart)} — ${f(e)}`;
  };

  return (
      <div className="h-full flex flex-col bg-[#F7F8FA] dark:bg-slate-900">
        {/* TOP BAR */}
        <div className="h-16 bg-white dark:bg-slate-950 border-b border-slate-200 dark:border-slate-800 flex items-center justify-between px-6 shrink-0">
          <div>
            <h1 className="text-lg font-bold text-slate-900 dark:text-slate-100">
              {t(i18nKeys.calendar.title)}
            </h1>
            <p className="text-xs text-slate-500 dark:text-slate-400 font-medium">
              {t(i18nKeys.calendar.subtitle)}
            </p>
          </div>
          <div className="flex items-center gap-2">
            <button
                onClick={() => shift(-1)}
                className="px-3 py-1.5 text-xs font-medium text-slate-600 dark:text-slate-300 bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-lg hover:bg-slate-50 dark:hover:bg-slate-700 transition-colors"
            >
              ← {t(i18nKeys.calendar.prevWeek)}
            </button>
            <div className="text-sm font-semibold text-slate-700 dark:text-slate-200 px-4 bg-slate-50 dark:bg-slate-800 rounded-lg border border-slate-100 dark:border-slate-700 min-w-[140px] text-center">
              {fmtRange()}
            </div>
            <button
                onClick={() => shift(1)}
                className="px-3 py-1.5 text-xs font-medium text-slate-600 dark:text-slate-300 bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-lg hover:bg-slate-50 dark:hover:bg-slate-700 transition-colors"
            >
              {t(i18nKeys.calendar.nextWeek)} →
            </button>
            <button
                onClick={() => setWeekStart(startOfWeek(new Date()))}
                className="px-3 py-1.5 text-xs font-medium text-blue-600 bg-blue-50 dark:bg-blue-900/30 border border-blue-100 dark:border-blue-800 rounded-lg hover:bg-blue-100 dark:hover:bg-blue-900/50 transition-colors"
            >
              {t(i18nKeys.calendar.today)}
            </button>
          </div>
        </div>

        {/* CALENDAR GRID */}
        <div className="flex-1 overflow-auto p-4 lg:p-6 relative">
          {loading && (
              <div className="absolute inset-0 flex items-center justify-center bg-white/50 dark:bg-slate-900/50 z-10">
                <div className="flex items-center gap-2 text-sm text-slate-500">
                  <svg className="animate-spin h-5 w-5" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                    <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                    <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                  </svg>
                  {t(i18nKeys.calendar.loading)}
                </div>
              </div>
          )}

          {error && (
              <div className="absolute inset-0 flex items-center justify-center bg-white dark:bg-slate-900 z-10">
                <div className="text-red-500 text-sm">{error}</div>
              </div>
          )}

          <div className="bg-white dark:bg-slate-800 rounded-xl shadow-sm border border-slate-200 dark:border-slate-700 overflow-hidden">
            <div className="grid" style={{ gridTemplateColumns: "50px repeat(7, minmax(120px, 1fr))" }}>
              {/* Header Row */}
              <div className="bg-slate-50 dark:bg-slate-900 border-b border-r border-slate-200 dark:border-slate-700 p-2 sticky left-0 z-20" />
              {daysShort.map((d, i) => {
                const date = new Date(weekStart);
                date.setDate(date.getDate() + i);
                const isToday = new Date().toDateString() === date.toDateString();

                return (
                    <div
                        key={d}
                        className={`text-center py-3 border-b border-r border-slate-200 dark:border-slate-700 text-xs font-medium ${
                            isToday
                                ? "bg-blue-50 dark:bg-blue-900/30 text-blue-700 dark:text-blue-400 border-blue-100 dark:border-blue-800"
                                : "bg-white dark:bg-slate-800 text-slate-600 dark:text-slate-400"
                        }`}
                    >
                      <div className="uppercase tracking-wider mb-1">{d}</div>
                      <div className={`text-sm font-bold ${isToday ? "text-blue-600 dark:text-blue-400" : "text-slate-900 dark:text-slate-100"}`}>
                        {date.getDate()}
                      </div>
                    </div>
                );
              })}

              {/* Time Rows */}
              {Array.from({ length: 12 }, (_, i) => 9 + i).map((h) => (
                  <Fragment key={`row-${h}`}>
                    <div className="text-[10px] text-slate-400 font-bold text-center py-4 border-r border-b border-slate-100 dark:border-slate-700 bg-slate-50 dark:bg-slate-900/90 sticky left-0 z-10">
                      {h}:00
                    </div>

                    {Array.from({ length: 7 }).map((_, dayIdx) => {
                      const cellItems = byDayHour[`${dayIdx}-${h}`] || [];
                      const cellDate = new Date(weekStart);
                      cellDate.setDate(cellDate.getDate() + dayIdx);
                      cellDate.setHours(h, 0, 0, 0);
                      return (
                          <div
                              key={`${h}-${dayIdx}`}
                              onClick={() => setCreateCtx({ date: cellDate })}
                              className="min-h-[90px] border-r border-b border-slate-100 dark:border-slate-700 p-1 space-y-1.5 hover:bg-blue-50/40 dark:hover:bg-blue-900/10 transition-colors cursor-pointer"
                          >
                            {cellItems.map((a) => {
                              const cardCls = STATUS_CARD[a.status] ?? STATUS_CARD.scheduled;
                              const isCancelled = a.status === "cancelled";
                              const secondary = [a.doctor_name, a.cabinet].filter(Boolean).join(" · ");

                              return (
                                  <div
                                      key={a.id}
                                      onClick={(e) => { e.stopPropagation(); setDetailCtx(a); }}
                                      className={`text-[10px] px-1.5 py-1 rounded border shadow-sm transition-all cursor-pointer hover:shadow-md ${cardCls} ${isCancelled ? "line-through" : ""}`}
                                  >
                                    <div className="flex items-center justify-between font-bold mb-0.5 gap-1">
                                      <span className="whitespace-nowrap">{fmtTime(a.starts_at)}</span>
                                      {a.patient_is_first && (
                                          <span className="text-[8px] bg-amber-100 dark:bg-amber-900/40 text-amber-700 dark:text-amber-400 rounded px-0.5 font-bold shrink-0">
                                  NEW
                                </span>
                                      )}
                                    </div>
                                    <div className="truncate font-bold leading-tight">
                                      {a.patient_name || a.patient_phone || t(i18nKeys.calendar.no_patient)}
                                    </div>
                                    {secondary && (
                                        <div className="truncate opacity-80 text-[9px] mt-0.5">{secondary}</div>
                                    )}
                                  </div>
                              );
                            })}
                          </div>
                      );
                    })}
                  </Fragment>
              ))}
            </div>
          </div>
        </div>

        <CreateAppointmentModal
            open={!!createCtx}
            defaultDate={createCtx?.date}
            doctors={doctors}
            patients={patients}
            onClose={() => setCreateCtx(null)}
            onCreated={() => { setCreateCtx(null); refresh(); }}
            t={t}
        />

        <AppointmentDetailModal
            open={!!detailCtx}
            appt={detailCtx}
            onClose={() => setDetailCtx(null)}
            onChanged={() => { setDetailCtx(null); refresh(); }}
            t={t}
        />
      </div>
  );
}

// ── modal components ──────────────────────────────────────────────────────────

function fmtDateForInput(d) {
  if (!d) return "";
  const pad = (n) => String(n).padStart(2, "0");
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`;
}

function CreateAppointmentModal({ open, defaultDate, doctors, patients, onClose, onCreated, t }) {
  const [mode, setMode] = useState("book"); // "book" or "request"
  const [form, setForm] = useState({
    doctor_id: "",
    patient_id: "",
    patient_search: "",
    patient_name: "",
    patient_phone: "",
    start: "",
    duration: 30,
    zhaloba: "",
    cabinet: "",
    is_first: false,
  });
  const [saving, setSaving] = useState(false);
  const [err, setErr] = useState(null);

  useEffect(() => {
    if (open) {
      setForm((f) => ({
        ...f,
        doctor_id: "",
        patient_id: "",
        patient_search: "",
        patient_name: "",
        patient_phone: "",
        start: fmtDateForInput(defaultDate),
        duration: 30,
        zhaloba: "",
        cabinet: "",
        is_first: false,
      }));
      setErr(null);
      setMode("book");
    }
  }, [open, defaultDate]);

  const set = (k) => (e) =>
    setForm({ ...form, [k]: e.target.type === "checkbox" ? e.target.checked : e.target.value });

  const matchedPatients = useMemo(() => {
    const q = form.patient_search.trim().toLowerCase();
    if (!q) return [];
    return patients
      .filter((p) =>
        (p.name || "").toLowerCase().includes(q) ||
        (p.phone || "").includes(q) ||
        String(p.id).includes(q),
      )
      .slice(0, 6);
  }, [form.patient_search, patients]);

  const submit = async () => {
    setErr(null);
    if (!form.start) { setErr(t("forms.start") + " ?"); return; }
    const start = new Date(form.start);
    const end = new Date(start.getTime() + Number(form.duration) * 60_000);

    setSaving(true);
    try {
      if (mode === "book") {
        if (!form.doctor_id) { setErr(t("forms.select_doctor")); setSaving(false); return; }
        if (!form.patient_id) { setErr(t("forms.select_patient")); setSaving(false); return; }
        await api.createScheduleAppointment({
          doctor_id: Number(form.doctor_id),
          patient_id: Number(form.patient_id),
          starts_at: start.toISOString(),
          ends_at: end.toISOString(),
          zhaloba: form.zhaloba,
          cabinet: form.cabinet,
          is_first: form.is_first,
        });
      } else {
        if (!form.patient_name.trim() || !form.patient_phone.trim()) {
          setErr(`${t("forms.name")} + ${t("forms.phone")}`);
          setSaving(false);
          return;
        }
        await api.sendAppointmentRequest({
          patient_name: form.patient_name,
          patient_phone: form.patient_phone,
          starts_at: start.toISOString(),
          ends_at: end.toISOString(),
        });
      }
      onCreated();
    } catch (e) {
      setErr(e.message || t("forms.action_failed"));
    } finally {
      setSaving(false);
    }
  };

  const selectedPatient = patients.find((p) => String(p.id) === String(form.patient_id));

  return (
    <Modal
      open={open}
      onClose={onClose}
      title={t("forms.new_appointment")}
      width="max-w-xl"
      footer={
        <>
          <button onClick={onClose} className={btnSecondary} disabled={saving}>
            {t("forms.cancel")}
          </button>
          <button onClick={submit} disabled={saving} className={btnPrimary}>
            {saving ? t("forms.saving") : (mode === "book" ? t("forms.book") : t("forms.request"))}
          </button>
        </>
      }
    >
      {err && (
        <div className="mb-3 px-3 py-2 text-xs text-red-600 bg-red-50 dark:bg-red-900/30 border border-red-100 dark:border-red-800 rounded-lg">
          {err}
        </div>
      )}

      {/* mode toggle */}
      <div className="flex gap-2 mb-4">
        <button
          onClick={() => setMode("book")}
          className={`flex-1 px-3 py-2 text-xs font-bold rounded-lg border transition-colors ${
            mode === "book"
              ? "bg-blue-600 text-white border-blue-600"
              : "bg-slate-50 dark:bg-slate-900 text-slate-600 dark:text-slate-300 border-slate-200 dark:border-slate-700"
          }`}
        >
          {t("forms.book")}
        </button>
        <button
          onClick={() => setMode("request")}
          className={`flex-1 px-3 py-2 text-xs font-bold rounded-lg border transition-colors ${
            mode === "request"
              ? "bg-amber-500 text-white border-amber-500"
              : "bg-slate-50 dark:bg-slate-900 text-slate-600 dark:text-slate-300 border-slate-200 dark:border-slate-700"
          }`}
        >
          {t("forms.request")}
        </button>
      </div>

      <div className="grid grid-cols-2 gap-3">
        <FormField label={t("forms.start")}>
          <input type="datetime-local" className={inputCls} value={form.start} onChange={set("start")} />
        </FormField>
        <FormField label={`${t("forms.duration")} (${t("forms.min")})`}>
          <select className={inputCls} value={form.duration} onChange={set("duration")}>
            {[15, 30, 45, 60, 90, 120].map((d) => (
              <option key={d} value={d}>{d}</option>
            ))}
          </select>
        </FormField>

        {mode === "book" ? (
          <>
            <FormField label={t("forms.doctor")} full>
              <select className={inputCls} value={form.doctor_id} onChange={set("doctor_id")}>
                <option value="">— {t("forms.select_doctor")} —</option>
                {doctors.map((d) => (
                  <option key={d.id} value={d.id}>{d.name}</option>
                ))}
              </select>
            </FormField>

            <FormField label={t("forms.patient")} full>
              {selectedPatient ? (
                <div className="flex items-center justify-between px-3 py-2 bg-blue-50 dark:bg-blue-900/30 border border-blue-200 dark:border-blue-800 rounded-lg">
                  <div className="text-sm text-slate-800 dark:text-slate-100">
                    <div className="font-bold">{selectedPatient.name}</div>
                    <div className="text-xs text-slate-500 dark:text-slate-400">{selectedPatient.phone || `#${selectedPatient.id}`}</div>
                  </div>
                  <button
                    onClick={() => setForm({ ...form, patient_id: "", patient_search: "" })}
                    className="text-slate-400 hover:text-slate-600 text-lg leading-none"
                  >×</button>
                </div>
              ) : (
                <>
                  <input
                    className={inputCls}
                    placeholder={t("forms.patient_search")}
                    value={form.patient_search}
                    onChange={set("patient_search")}
                  />
                  {matchedPatients.length > 0 && (
                    <div className="mt-1 max-h-40 overflow-y-auto bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-lg shadow-sm">
                      {matchedPatients.map((p) => (
                        <button
                          key={p.id}
                          onClick={() => setForm({ ...form, patient_id: p.id, patient_search: "" })}
                          className="w-full text-left px-3 py-1.5 text-xs hover:bg-blue-50 dark:hover:bg-blue-900/30 text-slate-700 dark:text-slate-200 border-b border-slate-50 dark:border-slate-700 last:border-0"
                        >
                          <span className="font-medium">{p.name || "—"}</span>
                          {p.phone && <span className="text-slate-400 ml-2">{p.phone}</span>}
                        </button>
                      ))}
                    </div>
                  )}
                </>
              )}
            </FormField>

            <FormField label={t("forms.cabinet")}>
              <input className={inputCls} value={form.cabinet} onChange={set("cabinet")} />
            </FormField>
            <FormField label={t("forms.first_visit")}>
              <label className="flex items-center gap-2 mt-1.5">
                <input type="checkbox" checked={form.is_first} onChange={set("is_first")} className="w-4 h-4 accent-blue-600" />
                <span className="text-xs text-slate-600 dark:text-slate-300">{t("common.yes")}</span>
              </label>
            </FormField>

            <FormField label={t("forms.complaint")} full>
              <textarea className={inputCls} rows={2} value={form.zhaloba} onChange={set("zhaloba")} />
            </FormField>
          </>
        ) : (
          <>
            <FormField label={t("forms.name")} full>
              <input className={inputCls} value={form.patient_name} onChange={set("patient_name")} autoFocus />
            </FormField>
            <FormField label={t("forms.phone")} full>
              <input className={inputCls} value={form.patient_phone} onChange={set("patient_phone")} placeholder="+77..." />
            </FormField>
          </>
        )}
      </div>
    </Modal>
  );
}

const STATUS_CODES = {
  confirm:    1,
  came:       3,
  in_process: 5,
  late:       6,
  cancel:     2,
};

function AppointmentDetailModal({ open, appt, onClose, onChanged, t }) {
  const [busy, setBusy] = useState(false);
  const [err, setErr] = useState(null);

  useEffect(() => { if (open) setErr(null); }, [open]);

  if (!appt) return null;

  const setStatus = async (code) => {
    setBusy(true); setErr(null);
    try {
      await api.setScheduleAppointmentStatus(appt.id, code);
      onChanged();
    } catch (e) {
      setErr(e.message || t("forms.action_failed"));
    } finally { setBusy(false); }
  };

  const remove = async () => {
    if (!window.confirm(t("forms.delete_confirm"))) return;
    setBusy(true); setErr(null);
    try {
      await api.deleteScheduleAppointment(appt.id);
      onChanged();
    } catch (e) {
      setErr(e.message || t("forms.action_failed"));
    } finally { setBusy(false); }
  };

  const time = (d) => `${String(d.getHours()).padStart(2, "0")}:${String(d.getMinutes()).padStart(2, "0")}`;
  const date = appt.starts_at.toLocaleDateString();

  return (
    <Modal
      open={open}
      onClose={onClose}
      title={t("forms.appointment_detail")}
      footer={
        <>
          <button onClick={remove} disabled={busy} className={btnDanger}>
            {t("forms.delete")}
          </button>
          <button onClick={onClose} disabled={busy} className={btnSecondary}>
            {t("forms.close")}
          </button>
        </>
      }
    >
      {err && (
        <div className="mb-3 px-3 py-2 text-xs text-red-600 bg-red-50 dark:bg-red-900/30 border border-red-100 dark:border-red-800 rounded-lg">
          {err}
        </div>
      )}

      <div className="space-y-3 mb-4">
        <div className="text-xl font-bold text-slate-900 dark:text-slate-100">
          {time(appt.starts_at)} – {time(appt.ends_at)}
          <span className="ml-2 text-xs font-medium text-slate-400">{date}</span>
        </div>

        <div className="grid grid-cols-2 gap-3 text-xs">
          <DetailRow label={t("forms.patient")}   value={appt.patient_name || appt.patient_phone || "—"} />
          <DetailRow label={t("forms.phone")}     value={appt.patient_phone || "—"} />
          <DetailRow label={t("forms.doctor")}    value={appt.doctor_name || "—"} />
          <DetailRow label={t("forms.cabinet")}   value={appt.cabinet || "—"} />
          <DetailRow label={t("forms.complaint")} value={appt.zhaloba || "—"} full />
        </div>
      </div>

      <div>
        <div className="text-[10px] font-bold uppercase tracking-wider text-slate-500 dark:text-slate-400 mb-2">
          {t("forms.status")}
        </div>
        <div className="grid grid-cols-2 sm:grid-cols-3 gap-2">
          <button onClick={() => setStatus(STATUS_CODES.confirm)}    disabled={busy} className={btnGhost}>{t("forms.status_confirm")}</button>
          <button onClick={() => setStatus(STATUS_CODES.came)}       disabled={busy} className={btnGhost}>{t("forms.status_came")}</button>
          <button onClick={() => setStatus(STATUS_CODES.in_process)} disabled={busy} className={btnGhost}>{t("forms.status_in_process")}</button>
          <button onClick={() => setStatus(STATUS_CODES.late)}       disabled={busy} className={btnGhost}>{t("forms.status_late")}</button>
          <button onClick={() => setStatus(STATUS_CODES.cancel)}     disabled={busy} className={btnDanger}>{t("forms.status_cancel")}</button>
        </div>
      </div>
    </Modal>
  );
}

function DetailRow({ label, value, full }) {
  return (
    <div className={`p-2 bg-slate-50 dark:bg-slate-900/50 rounded-lg border border-slate-100 dark:border-slate-700 ${full ? "col-span-2" : ""}`}>
      <div className="text-[9px] uppercase text-slate-400 font-bold tracking-wider mb-0.5">{label}</div>
      <div className="text-sm text-slate-700 dark:text-slate-200 font-medium">{value}</div>
    </div>
  );
}