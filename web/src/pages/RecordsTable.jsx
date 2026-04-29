import { Fragment, useEffect, useMemo, useState } from "react";
import { api } from "../api/client";
import { AppointmentDetailModal, CreateAppointmentModal } from "../features/appointments/AppointmentModals";
import {
    STATUS_CARD,
    fetchDoctors,
    fetchPatients,
    fmtDateLong,
    fmtTime,
    normalizeScheduleAppointment,
    startOfWeek,
} from "../features/appointments/schedule";
import { useTranslation } from "../hooks/useTranslation";

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

// ── Main Component ────────────────────────────────────────────────────────────

export default function RecordsTable() {
    const { t } = useTranslation();
    const [weekStart, setWeekStart] = useState(() => startOfWeek(new Date()));
    const [items, setItems] = useState([]);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState(null);
    const [tick, setTick] = useState(0);

    // Modal states
    const [createCtx, setCreateCtx] = useState(null);
    const [detailCtx, setDetailCtx] = useState(null);
    const [doctors, setDoctors] = useState([]);
    const [patients, setPatients] = useState([]);

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
                    .map((a) => normalizeScheduleAppointment(a, doctorMap, patientMap))
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
                        onClick={() => setWeekStart(startOfWeek(new Date()))}
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
