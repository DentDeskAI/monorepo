import { useState, useEffect } from "react";
import { api } from "../api/client";
import { useTranslation } from "../hooks/useTranslation";

// i18n Keys
const i18nKeys = {
    doctors: {
        title: "doctors.title",
        search: "doctors.search_placeholder",
        empty: "doctors.empty_state",
        no_phone: "common.no_phone",
        profile_info: "doctors.profile_info",
        history: "doctors.history_title",
        history_empty: "doctors.history_empty",
        loading: "common.loading",
        specialty: "doctors.specialty",
        no_specialty: "doctors.no_specialty",
        unknown_doctor: "doctors.unknown_doctor",
    },
    status: {
        scheduled: "common.status.scheduled",
        confirmed: "common.status.confirmed",
        completed: "common.status.completed",
        cancelled: "common.status.cancelled",
    },
};

export default function Doctors() {
    const { t } = useTranslation();
    const [list, setList] = useState([]);
    const [q, setQ] = useState("");
    const [selected, setSelected] = useState(null);
    const [appts, setAppts] = useState([]);
    const [loading, setLoading] = useState(false);

    useEffect(() => {
        setLoading(true);
        api.doctors()
            .then((data) => {
                setList(data || []);
            })
            .catch((err) => console.error("Failed to fetch doctors", err))
            .finally(() => setLoading(false));
    }, []);

    useEffect(() => {
        if (!selected) {
            setAppts([]);
            return;
        }

        setLoading(true);
        api.doctor(selected.id)
            .then((data) => {
                setAppts(Array.isArray(data) ? data : []);
            })
            .catch(() => {
                setAppts([]);
            })
            .finally(() => setLoading(false));
    }, [selected]);

    const filtered = list.filter((p) => {
        if (!q) return true;
        const s = q.toLowerCase();
        return (p.name || "").toLowerCase().includes(s) || (p.phone || "").includes(s);
    });

    return (
        <div className="h-screen flex flex-col bg-[#F7F8FA] dark:bg-slate-900">

            {/* TOP BAR */}
            <div className="h-16 bg-white dark:bg-slate-950 border-b border-slate-200 dark:border-slate-800 flex items-center justify-between px-6 shrink-0">
                <div className="text-sm font-medium text-slate-400 uppercase tracking-wider">
                    {t(i18nKeys.doctors.title)}
                </div>
                <div className="w-8 h-8 rounded-full bg-slate-100 dark:bg-slate-800 flex items-center justify-center text-slate-600 dark:text-slate-300 text-xs font-bold">
                    D
                </div>
            </div>

            {/* MAIN CONTENT */}
            <div className="flex flex-1 overflow-hidden">

                {/* LEFT SIDEBAR: LIST */}
                <div className="w-80 bg-white dark:bg-slate-950 border-r border-slate-200 dark:border-slate-800 flex flex-col shrink-0">
                    <div className="p-4 border-b border-slate-100 dark:border-slate-800">
                        <h2 className="text-sm font-semibold text-slate-900 dark:text-slate-100 mb-3">
                            {t(i18nKeys.doctors.title)}
                        </h2>
                        <input
                            className="w-full px-3 py-2 bg-slate-50 dark:bg-slate-900 border border-slate-200 dark:border-slate-700 rounded-lg text-sm text-slate-900 dark:text-slate-100 placeholder-slate-400 focus:outline-none focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 transition-all"
                            placeholder={t(i18nKeys.doctors.search)}
                            value={q}
                            onChange={(e) => setQ(e.target.value)}
                        />
                    </div>

                    <div className="flex-1 overflow-y-auto">
                        {loading ? (
                            <div className="p-4 text-xs text-slate-400 text-center">
                                {t(i18nKeys.doctors.loading)}
                            </div>
                        ) : filtered.length === 0 ? (
                            <div className="p-4 text-xs text-slate-400 text-center">
                                {t(i18nKeys.doctors.empty)}
                            </div>
                        ) : (
                            filtered.map((d) => (
                                <button
                                    key={d.id}
                                    onClick={() => setSelected(d)}
                                    className={`w-full text-left px-4 py-3 border-b border-slate-50 dark:border-slate-800 transition-colors flex items-center gap-3 ${
                                        selected?.id === d.id
                                            ? "bg-blue-50/50 dark:bg-blue-600/20 border-l-4 border-l-blue-600 pl-[11px]"
                                            : "hover:bg-slate-50 dark:hover:bg-slate-900 border-l-4 border-l-transparent pl-[15px]"
                                    }`}
                                >
                                    <div className="w-9 h-9 rounded-full bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 grid place-items-center text-slate-600 dark:text-slate-300 font-medium text-sm shrink-0 shadow-sm">
                                        {(d.name || "?").charAt(0).toUpperCase()}
                                    </div>
                                    <div className="flex-1 min-w-0">
                                        <div className="font-medium text-sm text-slate-900 dark:text-slate-100 truncate">
                                            {d.name || "—"}
                                        </div>
                                        <div className="text-xs text-slate-500 dark:text-slate-400 truncate mt-0.5">
                                            {d.phone || t(i18nKeys.doctors.no_phone)}
                                        </div>
                                    </div>
                                </button>
                            ))
                        )}
                    </div>
                </div>

                {/* RIGHT PANEL: DETAILS */}
                <div className="flex-1 overflow-y-auto bg-[#F7F8FA] dark:bg-slate-900 p-6 lg:p-8">
                    {!selected ? (
                        <div className="h-full flex flex-col items-center justify-center text-slate-400">
                            <div className="w-16 h-16 bg-slate-100 dark:bg-slate-800 rounded-full grid place-items-center text-2xl mb-4">
                                👨‍⚕️
                            </div>
                            <p className="text-sm font-medium">
                                {t(i18nKeys.doctors.empty)}
                            </p>
                        </div>
                    ) : (
                        <div className="max-w-4xl mx-auto space-y-6">

                            {/* PROFILE HEADER */}
                            <div className="bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700 shadow-sm p-5 flex items-center gap-5">
                                <div className="w-16 h-16 rounded-full bg-emerald-50 dark:bg-emerald-900/30 text-emerald-600 dark:text-emerald-400 border border-emerald-100 dark:border-emerald-800 grid place-items-center font-bold text-2xl shrink-0">
                                    {(selected.name || "?").charAt(0).toUpperCase()}
                                </div>
                                <div>
                                    <h1 className="text-xl font-bold text-slate-900 dark:text-slate-100 leading-tight">
                                        {selected.name || t(i18nKeys.doctors.unknown_doctor)}
                                    </h1>
                                    <div className="text-sm text-slate-500 dark:text-slate-400 mt-1">
                                        {selected.phone || t(i18nKeys.doctors.no_phone)}
                                    </div>
                                </div>
                            </div>

                            {/* SPECIALTIES & HISTORY */}
                            <div className="bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700 shadow-sm overflow-hidden">
                                <div className="px-6 py-4 border-b border-slate-100 dark:border-slate-700 bg-slate-50/50 dark:bg-slate-900/30">
                                    <h2 className="text-sm font-bold text-slate-900 dark:text-slate-100 uppercase tracking-wide">
                                        {t(i18nKeys.doctors.profile_info)}
                                    </h2>
                                </div>

                                <div className="p-6">
                                    {/* Specialties */}
                                    <div className="mb-6">
                                        <div className="text-[10px] uppercase text-slate-400 font-bold tracking-wider mb-2">
                                            {t(i18nKeys.doctors.specialty)}
                                        </div>
                                        <div className="flex flex-wrap gap-2">
                                            {(selected.specialties || []).length > 0 ? (
                                                selected.specialties.map((spec, idx) => (
                                                    <span key={idx} className="px-3 py-1 bg-slate-100 dark:bg-slate-700 text-slate-600 dark:text-slate-300 text-xs font-medium rounded-full">
                                                        {spec}
                                                    </span>
                                                ))
                                            ) : (
                                                <span className="text-sm text-slate-400 italic">
                                                    {t(i18nKeys.doctors.no_specialty)}
                                                </span>
                                            )}
                                        </div>
                                    </div>

                                    <div className="border-t border-slate-100 dark:border-slate-700 pt-4">
                                        <div className="flex items-center justify-between mb-4">
                                            <h3 className="text-sm font-bold text-slate-900 dark:text-slate-100 uppercase tracking-wide">
                                                {t(i18nKeys.doctors.history)}
                                            </h3>
                                            {loading && (
                                                <span className="text-xs text-slate-400 animate-pulse">
                                                    {t(i18nKeys.doctors.loading)}
                                                </span>
                                            )}
                                        </div>

                                        {appts.length === 0 ? (
                                            <div className="py-4 text-sm text-slate-400 text-center">
                                                {t(i18nKeys.doctors.history_empty)}
                                            </div>
                                        ) : (
                                            <ul className="divide-y divide-slate-100 dark:divide-slate-700">
                                                {appts.map((a) => (
                                                    <li key={a.id} className="py-3 flex justify-between items-center">
                                                        <div>
                                                            <div className="font-medium text-slate-900 dark:text-slate-100 text-sm">
                                                                {formatDate(a.starts_at)}
                                                            </div>
                                                            <div className="text-xs text-slate-500 dark:text-slate-400 mt-0.5">
                                                                {a.service}
                                                            </div>
                                                        </div>
                                                        <StatusBadge status={a.status} t={t} />
                                                    </li>
                                                ))}
                                            </ul>
                                        )}
                                    </div>
                                </div>
                            </div>

                        </div>
                    )}
                </div>
            </div>
        </div>
    );
}

function StatusBadge({ status, t }) {
    const styles = {
        scheduled: "bg-blue-50 text-blue-700 border-blue-100",
        confirmed: "bg-emerald-50 text-emerald-700 border-emerald-100",
        completed: "bg-slate-100 text-slate-600 border-slate-200",
        cancelled: "bg-red-50 text-red-600 border-red-100",
    };

    const defaultStyle = "bg-slate-50 text-slate-500 border-slate-100";

    return (
        <span className={`text-[10px] font-bold px-2.5 py-1 rounded-full uppercase tracking-wide border ${styles[status] || defaultStyle}`}>
            {statusLabel(status, t)}
        </span>
    );
}

function formatDate(v) {
    try {
        return new Date(v).toLocaleDateString(undefined, {
            year: 'numeric',
            month: 'short',
            day: 'numeric'
        });
    } catch {
        return v;
    }
}

function statusLabel(s, t) {
    const labels = {
        scheduled: t(i18nKeys.status.scheduled),
        confirmed: t(i18nKeys.status.confirmed),
        completed: t(i18nKeys.status.completed),
        cancelled: t(i18nKeys.status.cancelled),
    };
    return labels[s] || s;
}
