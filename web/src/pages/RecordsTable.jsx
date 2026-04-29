import { Fragment, useEffect, useMemo, useState } from "react";
import { api } from "../api/client";
import { useTranslation } from "../hooks/useTranslation";
import Modal, { btnPrimary, btnSecondary, btnDanger, btnGhost } from "../components/Modal";

// Reusing i18n keys (assuming they are defined in your translation file,
// or you can import them if exported from the previous file)
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

// ── Helpers ───────────────────────────────────────────────────────────────────

function parseMacdentDate(s) {
    if (!s) return new Date();
    const [datePart, timePart = "00:00:00"] = s.split(" ");
    const [d, m, y] = datePart.split(".");
    const [h, min, sec] = timePart.split(":");
    return new Date(+y, +m - 1, +d, +h, +min, +sec);
}

function fmtTime(date) {
    return `${String(date.getHours()).padStart(2, "0")}:${String(date.getMinutes()).padStart(2, "0")}`;
}

function fmtDateLong(date) {
    return date.toLocaleDateString("ru-RU", { day: "numeric", month: "long", year: "numeric" });
}

function fmtDateShort(date) {
    return date.toLocaleDateString("ru-RU", { day: "2-digit", month: "2-digit" });
}

const STATUS_CARD = {
    scheduled: "bg-blue-50 dark:bg-blue-900/30 text-blue-800 dark:text-blue-200 border-blue-200 dark:border-blue-800",
    confirmed: "bg-emerald-50 dark:bg-emerald-900/30 text-emerald-800 dark:text-emerald-200 border-emerald-200 dark:border-emerald-800",
    completed: "bg-slate-50 dark:bg-slate-700/50 text-slate-500 dark:text-slate-400 border-slate-200 dark:border-slate-600",
    cancelled: "bg-slate-50 dark:bg-slate-700/30 text-slate-400 dark:text-slate-500 border-slate-100 dark:border-slate-700 opacity-60",
};

// ── Main Component ────────────────────────────────────────────────────────────

export default function RecordsTable() {
    const { t } = useTranslation();
    const [weekStart, setWeekStart] = useState(() => {
        const d = new Date();
        const day = (d.getDay() + 6) % 7;
        d.setDate(d.getDate() - day);
        return d;
    });
    const [items, setItems] = useState([]);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState(null);
    const [tick, setTick] = useState(0);

    // Modal states
    const [createCtx, setCreateCtx] = useState(null);
    const [detailCtx, setDetailCtx] = useState(null);
    const [doctors, setDoctors] = useState([]);
    const [patients, setPatients] = useState([]);

    // Cache logic (same as original)
    const _refCache = { doctors: null, patients: null };
    const fetchDoctors = () => {
        if (_refCache.doctors) return _refCache.doctors;
        _refCache.doctors = api.doctors().catch(() => []);
        return _refCache.doctors;
    };
    const fetchPatients = () => {
        if (_refCache.patients) return _refCache.patients;
        _refCache.patients = api.patients().catch(() => []);
        return _refCache.patients;
    };

    const weekEnd = useMemo(() => {
        const e = new Date(weekStart);
        e.setDate(e.getDate() + 7);
        return e;
    }, [weekStart]);

    useEffect(() => {
        setLoading(true);
        setError(null);

        Promise.all([
            api.history(
                weekStart.toISOString(),
                weekEnd.toISOString()
            ),
            fetchDoctors(),
            fetchPatients(),
        ])
            .then(([resp, docs, pats]) => {
                setDoctors(docs);
                setPatients(pats);
                const doctorMap = Object.fromEntries(docs.map((d) => [d.id, d.name]));
                const patientMap = Object.fromEntries(pats.map((p) => [p.id, p]));

                const normalized = (resp?.appointments ?? [])
                    .map((a) => {
                        // Inline normalize logic to keep self-contained if needed,
                        // or reuse the helper if exported.
                        // Here I inline the logic for safety in a new file:
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
                            status: { 0: "scheduled", 1: "confirmed", 2: "cancelled", 3: "completed", 5: "completed" }[a.status] ?? "completed",
                        };
                    })
                    .filter(({ starts_at }) => starts_at >= weekStart && starts_at < weekEnd);

                // Sort all records by date descending for the table
                normalized.sort((a, b) => b.starts_at - a.starts_at);

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
                        onClick={() => setWeekStart(() => {
                            const d = new Date();
                            const day = (d.getDay() + 6) % 7;
                            d.setDate(d.getDate() - day);
                            return d;
                        })}
                        className="px-3 py-1.5 text-xs font-medium text-blue-600 bg-blue-50 dark:bg-blue-900/30 border border-blue-100 dark:border-blue-800 rounded-lg hover:bg-blue-100 dark:hover:bg-blue-900/50 transition-colors"
                    >
                        {t(i18nKeys.calendar.today)}
                    </button>
                </div>
            </div>

            {/* TABLE AREA */}
            <div className="flex-1 overflow-auto p-4 lg:p-6">
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
                    <div className="overflow-x-auto">
                        <table className="w-full text-left border-collapse">
                            <thead>
                            <tr className="bg-slate-50 dark:bg-slate-900/50 border-b border-slate-200 dark:border-slate-700 text-[10px] uppercase tracking-wider font-bold text-slate-500 dark:text-slate-400">
                                <th className="px-4 py-3">Date & Time</th>
                                <th className="px-4 py-3">Patient</th>
                                <th className="px-4 py-3">Doctor</th>
                                <th className="px-4 py-3">Cabinet</th>
                                <th className="px-4 py-3">Status</th>
                                <th className="px-4 py-3">Complaint</th>
                                <th className="px-4 py-3 text-right">Actions</th>
                            </tr>
                            </thead>
                            <tbody className="divide-y divide-slate-100 dark:divide-slate-700/50">
                            {items.length === 0 ? (
                                <tr>
                                    <td colSpan="7" className="px-4 py-8 text-center text-slate-400 text-sm">
                                        No appointments found for this week.
                                    </td>
                                </tr>
                            ) : (
                                items.map((a) => {
                                    const cardCls = STATUS_CARD[a.status] ?? STATUS_CARD.scheduled;
                                    const isCancelled = a.status === "cancelled";
                                    const secondary = [a.doctor_name, a.cabinet].filter(Boolean).join(" · ");

                                    return (
                                        <tr
                                            key={a.id}
                                            className={`hover:bg-slate-50 dark:hover:bg-slate-700/30 transition-colors ${isCancelled ? "bg-slate-50/50 dark:bg-slate-800/50" : ""}`}
                                        >
                                            <td className="px-4 py-3 align-top">
                                                <div className="flex items-start gap-2">
                            <span className="text-sm font-bold text-slate-700 dark:text-slate-200">
                              {fmtTime(a.starts_at)}
                            </span>
                                                    <span className="text-[10px] text-slate-400 mt-1">
                              → {fmtTime(a.ends_at)}
                            </span>
                                                </div>
                                                <div className="text-[10px] text-slate-400 mt-0.5">
                                                    {fmtDateLong(a.starts_at)}
                                                </div>
                                            </td>
                                            <td className="px-4 py-3 align-top">
                                                <div className="flex flex-col">
                            <span className={`text-sm font-medium ${isCancelled ? "line-through text-slate-400" : "text-slate-700 dark:text-slate-200"}`}>
                              {a.patient_name || a.patient_phone || t(i18nKeys.calendar.no_patient)}
                            </span>
                                                    {a.patient_is_first && (
                                                        <span className="inline-block mt-1 text-[9px] bg-amber-100 dark:bg-amber-900/40 text-amber-700 dark:text-amber-400 px-1 rounded font-bold">
                                NEW
                              </span>
                                                    )}
                                                </div>
                                            </td>
                                            <td className="px-4 py-3 align-top text-sm text-slate-600 dark:text-slate-300">
                                                {a.doctor_name || "—"}
                                            </td>
                                            <td className="px-4 py-3 align-top text-sm text-slate-600 dark:text-slate-300">
                                                {a.cabinet || "—"}
                                            </td>
                                            <td className="px-4 py-3 align-top">
                          <span className={`inline-flex items-center px-2 py-0.5 rounded text-[10px] font-bold border ${cardCls}`}>
                            {a.status.toUpperCase()}
                          </span>
                                            </td>
                                            <td className="px-4 py-3 align-top text-xs text-slate-500 dark:text-slate-400 max-w-[200px] truncate">
                                                {a.zhaloba || "—"}
                                            </td>
                                            <td className="px-4 py-3 align-top text-right">
                                                <button
                                                    onClick={() => setDetailCtx(a)}
                                                    className="text-blue-600 hover:text-blue-800 dark:text-blue-400 dark:hover:text-blue-300 text-xs font-medium hover:underline"
                                                >
                                                    Details
                                                </button>
                                            </td>
                                        </tr>
                                    );
                                })
                            )}
                            </tbody>
                        </table>
                    </div>
                </div>
            </div>

            {/* ── Modals (Reused from original) ─────────────────────────────────────── */}

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

// ── Modal Components (Copied from original file) ──────────────────────────────

function fmtDateForInput(d) {
    if (!d) return "";
    const pad = (n) => String(n).padStart(2, "0");
    return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`;
}

function CreateAppointmentModal({ open, defaultDate, doctors, patients, onClose, onCreated, t }) {
    const [mode, setMode] = useState("book");
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
                <div className="col-span-2">
                    <label className="block text-xs font-medium text-slate-600 dark:text-slate-300 mb-1">{t("forms.start")}</label>
                    <input type="datetime-local" className="w-full px-3 py-2 bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500" value={form.start} onChange={set("start")} />
                </div>
                <div className="col-span-2">
                    <label className="block text-xs font-medium text-slate-600 dark:text-slate-300 mb-1">{t("forms.duration")} ({t("forms.min")})</label>
                    <select className="w-full px-3 py-2 bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500" value={form.duration} onChange={set("duration")}>
                        {[15, 30, 45, 60, 90, 120].map((d) => (
                            <option key={d} value={d}>{d}</option>
                        ))}
                    </select>
                </div>

                {mode === "book" ? (
                    <>
                        <div className="col-span-2">
                            <label className="block text-xs font-medium text-slate-600 dark:text-slate-300 mb-1">{t("forms.doctor")}</label>
                            <select className="w-full px-3 py-2 bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500" value={form.doctor_id} onChange={set("doctor_id")}>
                                <option value="">— {t("forms.select_doctor")} —</option>
                                {doctors.map((d) => (
                                    <option key={d.id} value={d.id}>{d.name}</option>
                                ))}
                            </select>
                        </div>

                        <div className="col-span-2">
                            <label className="block text-xs font-medium text-slate-600 dark:text-slate-300 mb-1">{t("forms.patient")}</label>
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
                                        className="w-full px-3 py-2 bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
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
                        </div>

                        <div>
                            <label className="block text-xs font-medium text-slate-600 dark:text-slate-300 mb-1">{t("forms.cabinet")}</label>
                            <input className="w-full px-3 py-2 bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500" value={form.cabinet} onChange={set("cabinet")} />
                        </div>
                        <div>
                            <label className="flex items-center gap-2 mt-3 cursor-pointer">
                                <input type="checkbox" checked={form.is_first} onChange={set("is_first")} className="w-4 h-4 accent-blue-600" />
                                <span className="text-xs text-slate-600 dark:text-slate-300">{t("forms.first_visit")}</span>
                            </label>
                        </div>

                        <div className="col-span-2">
                            <label className="block text-xs font-medium text-slate-600 dark:text-slate-300 mb-1">{t("forms.complaint")}</label>
                            <textarea className="w-full px-3 py-2 bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500" rows={2} value={form.zhaloba} onChange={set("zhaloba")} />
                        </div>
                    </>
                ) : (
                    <>
                        <div className="col-span-2">
                            <label className="block text-xs font-medium text-slate-600 dark:text-slate-300 mb-1">{t("forms.name")}</label>
                            <input className="w-full px-3 py-2 bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500" value={form.patient_name} onChange={set("patient_name")} autoFocus />
                        </div>
                        <div className="col-span-2">
                            <label className="block text-xs font-medium text-slate-600 dark:text-slate-300 mb-1">{t("forms.phone")}</label>
                            <input className="w-full px-3 py-2 bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500" value={form.patient_phone} onChange={set("patient_phone")} placeholder="+77..." />
                        </div>
                    </>
                )}
            </div>
        </Modal>
    );
}

const STATUS_CODES = {
    confirm: 1,
    came: 3,
    in_process: 5,
    late: 6,
    cancel: 2,
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
                    <DetailRow label={t("forms.patient")} value={appt.patient_name || appt.patient_phone || "—"} />
                    <DetailRow label={t("forms.phone")} value={appt.patient_phone || "—"} />
                    <DetailRow label={t("forms.doctor")} value={appt.doctor_name || "—"} />
                    <DetailRow label={t("forms.cabinet")} value={appt.cabinet || "—"} />
                    <DetailRow label={t("forms.complaint")} value={appt.zhaloba || "—"} full />
                </div>
            </div>

            <div>
                <div className="text-[10px] font-bold uppercase tracking-wider text-slate-500 dark:text-slate-400 mb-2">
                    {t("forms.status")}
                </div>
                <div className="grid grid-cols-2 sm:grid-cols-3 gap-2">
                    <button onClick={() => setStatus(STATUS_CODES.confirm)} disabled={busy} className={btnGhost}>{t("forms.status_confirm")}</button>
                    <button onClick={() => setStatus(STATUS_CODES.came)} disabled={busy} className={btnGhost}>{t("forms.status_came")}</button>
                    <button onClick={() => setStatus(STATUS_CODES.in_process)} disabled={busy} className={btnGhost}>{t("forms.status_in_process")}</button>
                    <button onClick={() => setStatus(STATUS_CODES.late)} disabled={busy} className={btnGhost}>{t("forms.status_late")}</button>
                    <button onClick={() => setStatus(STATUS_CODES.cancel)} disabled={busy} className={btnDanger}>{t("forms.status_cancel")}</button>
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

