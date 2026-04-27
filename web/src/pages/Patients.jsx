import { useState, useEffect } from "react";
import { api } from "../api/client";
import { useTranslation } from "../hooks/useTranslation";

// i18n Keys
const i18nKeys = {
    patients: {
        title: "patients.title",
        search: "patients.search_placeholder",
        empty: "patients.empty_state",
        no_contact: "common.no_contact",
        profile_info: "patients.profile_info",
        history: "patients.history_title",
        history_empty: "patients.history_empty",
        loading: "common.loading",
    },
    info: {
        id: "common.id",
        phone: "common.phone",
        number: "patients.number",
        gender: "common.gender",
        dob: "common.dob",
        is_child: "patients.is_child",
        source: "patients.source",
        comment: "patients.comment",
    },
    status: {
        scheduled: "common.status.scheduled",
        confirmed: "common.status.confirmed",
        completed: "common.status.completed",
        cancelled: "common.status.cancelled",
    },
};

export default function Patients() {
    const { t } = useTranslation();
    const [list, setList] = useState([]);
    const [q, setQ] = useState("");
    const [selected, setSelected] = useState(null);
    const [appts, setAppts] = useState([]);

    const [loadingList, setLoadingList] = useState(false);
    const [loadingDetails, setLoadingDetails] = useState(false);

    useEffect(() => {
        setLoadingList(true);
        api.patients()
            .then((data) => {
                const normalized = (data || []).map((p) => ({
                    id: p.id,
                    name: p.name ?? "",
                    phone: p.phone ?? "",
                    number: p.number ?? "",
                    gender: p.gender ?? null,
                    birth: p.birth ?? null,
                    isChild: p.isChild ?? false,
                    comment: p.comment ?? "",
                    whereKnow: p.whereKnow ?? "",
                }));
                setList(normalized);
            })
            .catch((err) => console.error("Failed to fetch patients", err))
            .finally(() => setLoadingList(false));
    }, []);

    useEffect(() => {
        if (!selected) {
            setAppts([]);
            return;
        }

        setLoadingDetails(true);
        api.patient(selected.id)
            .then((data) => {
                setAppts(Array.isArray(data) ? data : []);
            })
            .catch(() => {
                setAppts([]);
            })
            .finally(() => setLoadingDetails(false));
    }, [selected]);

    const filtered = list.filter((p) => {
        if (!q) return true;
        const s = q.toLowerCase().trim();
        return (
            (p.name || "").toLowerCase().includes(s) ||
            (p.phone || "").includes(s) ||
            (p.number || "").includes(s)
        );
    });

    return (
        <div className="h-screen flex flex-col bg-[#F7F8FA] dark:bg-slate-900">

            {/* TOP BAR */}
            <div className="h-16 bg-white dark:bg-slate-950 border-b border-slate-200 dark:border-slate-800 flex items-center justify-between px-6 shrink-0">
                <div className="text-sm font-medium text-slate-400 uppercase tracking-wider">
                    {t(i18nKeys.patients.title)}
                </div>
                <div className="w-8 h-8 rounded-full bg-slate-100 dark:bg-slate-800 flex items-center justify-center text-slate-600 dark:text-slate-300 text-xs font-bold">
                    P
                </div>
            </div>

            {/* MAIN CONTENT */}
            <div className="flex flex-1 overflow-hidden">

                {/* LEFT SIDEBAR: LIST */}
                <div className="w-80 bg-white dark:bg-slate-950 border-r border-slate-200 dark:border-slate-800 flex flex-col shrink-0">
                    <div className="p-4 border-b border-slate-100 dark:border-slate-800">
                        <h2 className="text-sm font-semibold text-slate-900 dark:text-slate-100 mb-3">
                            {t(i18nKeys.patients.title)}
                        </h2>
                        <input
                            className="w-full px-3 py-2 bg-slate-50 dark:bg-slate-900 border border-slate-200 dark:border-slate-700 rounded-lg text-sm text-slate-900 dark:text-slate-100 placeholder-slate-400 focus:outline-none focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 transition-all"
                            placeholder={t(i18nKeys.patients.search)}
                            value={q}
                            onChange={(e) => setQ(e.target.value)}
                        />
                    </div>

                    <div className="flex-1 overflow-y-auto">
                        {loadingList ? (
                            <div className="p-4 text-xs text-slate-400 text-center">
                                {t(i18nKeys.patients.loading)}
                            </div>
                        ) : filtered.length === 0 ? (
                            <div className="p-4 text-xs text-slate-400 text-center">
                                {t(i18nKeys.patients.empty)}
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
                                            {d.phone || d.number || t(i18nKeys.patients.no_contact)}
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
                                👤
                            </div>
                            <p className="text-sm font-medium">
                                {t(i18nKeys.patients.empty)}
                            </p>
                        </div>
                    ) : (
                        <div className="max-w-4xl mx-auto space-y-6">

                            {/* PROFILE HEADER */}
                            <div className="bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700 shadow-sm p-5 flex items-center gap-5">
                                <div className="w-16 h-16 rounded-full bg-blue-50 dark:bg-blue-900/30 text-blue-600 dark:text-blue-400 border border-blue-100 dark:border-blue-800 grid place-items-center font-bold text-2xl shrink-0">
                                    {(selected.name || "?").charAt(0).toUpperCase()}
                                </div>
                                <div>
                                    <h1 className="text-xl font-bold text-slate-900 dark:text-slate-100 leading-tight">
                                        {selected.name || t("patients.unknown_patient")}
                                    </h1>
                                    <div className="text-sm text-slate-500 dark:text-slate-400 mt-1 flex items-center gap-2">
                                        <span>{selected.phone || selected.number || "—"}</span>
                                        {selected.isChild && (
                                            <span className="px-2 py-0.5 rounded-full bg-purple-50 dark:bg-purple-900/30 text-purple-600 dark:text-purple-400 text-[10px] font-bold uppercase tracking-wide">
                                                {t("patients.child")}
                                            </span>
                                        )}
                                    </div>
                                </div>
                            </div>

                            {/* INFO GRID */}
                            <div className="bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700 shadow-sm p-6">
                                <h2 className="text-sm font-bold text-slate-900 dark:text-slate-100 uppercase tracking-wide mb-5 pb-2 border-b border-slate-50 dark:border-slate-700">
                                    {t(i18nKeys.patients.profile_info)}
                                </h2>

                                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                                    <InfoItem label={t(i18nKeys.info.id)} value={selected.id} />
                                    <InfoItem label={t(i18nKeys.info.phone)} value={selected.phone} />
                                    <InfoItem label={t(i18nKeys.info.number)} value={selected.number} />
                                    <InfoItem label={t(i18nKeys.info.gender)} value={selected.gender} />
                                    <InfoItem label={t(i18nKeys.info.dob)} value={selected.birth} />
                                    <InfoItem
                                        label={t(i18nKeys.info.is_child)}
                                        value={selected.isChild ? t("common.yes") : t("common.no")}
                                    />
                                    <InfoItem label={t(i18nKeys.info.source)} value={selected.whereKnow} />
                                    <InfoItem label={t(i18nKeys.info.comment)} value={selected.comment} isFullWidth />
                                </div>
                            </div>

                            {/* HISTORY LIST */}
                            <div className="bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700 shadow-sm overflow-hidden">
                                <div className="px-6 py-4 border-b border-slate-100 dark:border-slate-700 flex items-center justify-between bg-slate-50/50 dark:bg-slate-900/30">
                                    <h2 className="text-sm font-bold text-slate-900 dark:text-slate-100 uppercase tracking-wide">
                                        {t(i18nKeys.patients.history)}
                                    </h2>
                                    {loadingDetails && (
                                        <span className="text-xs text-slate-400 animate-pulse">
                                            {t(i18nKeys.patients.loading)}
                                        </span>
                                    )}
                                </div>

                                <div className="divide-y divide-slate-100 dark:divide-slate-700">
                                    {appts.length === 0 ? (
                                        <div className="p-8 text-center text-sm text-slate-400">
                                            {t(i18nKeys.patients.history_empty)}
                                        </div>
                                    ) : (
                                        appts.map((a) => (
                                            <div key={a.id} className="px-6 py-4 flex items-center justify-between hover:bg-slate-50 dark:hover:bg-slate-700/30 transition-colors">
                                                <div className="flex flex-col">
                                                    <span className="text-sm font-medium text-slate-900 dark:text-slate-100">
                                                        {formatDate(a.starts_at)}
                                                    </span>
                                                    <span className="text-xs text-slate-500 dark:text-slate-400 mt-0.5">
                                                        {a.service}
                                                    </span>
                                                </div>
                                                <StatusBadge status={a.status} t={t} />
                                            </div>
                                        ))
                                    )}
                                </div>
                            </div>

                        </div>
                    )}
                </div>
            </div>
        </div>
    );
}

function InfoItem({ label, value, isFullWidth }) {
    return (
        <div className={`p-3 bg-slate-50 dark:bg-slate-900/50 rounded-lg border border-slate-100 dark:border-slate-700 ${isFullWidth ? 'md:col-span-2' : ''}`}>
            <div className="text-[10px] uppercase text-slate-400 font-bold tracking-wider mb-1">
                {label}
            </div>
            <div className="text-sm text-slate-700 dark:text-slate-300 font-medium">
                {value || "—"}
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
