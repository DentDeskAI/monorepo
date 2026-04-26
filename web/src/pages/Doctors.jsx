import { useEffect, useState } from "react";
import { api } from "../api/client";

export default function Doctors() {
    const [list, setList] = useState([]);
    const [q, setQ] = useState("");
    const [selected, setSelected] = useState(null);
    const [appts, setAppts] = useState([]);
    const [loading, setLoading] = useState(false);

    // 1. Fetch the full list of doctors
    useEffect(() => {
        api.doctors()
            .then(setList)
            .catch((err) => console.error("Failed to fetch doctors", err));
    }, []);

    // 2. Fetch details/history when a doctor is selected
    useEffect(() => {
        if (!selected) {
            setAppts([]);
            return;
        }

        setLoading(true);
        // Using the doctor ID to get specific info/appointments
        api.doctor(selected.id)
            .then((data) => {
                // If the response is a single doctor object, we wrap it in an array
                // or handle it based on your API's behavior.
                // Note: If api.doctor returns history, this map will work.
                // If it returns a single object, we check for that.
                setAppts(Array.isArray(data) ? data : []);
                setLoading(false);
            })
            .catch(() => {
                setAppts([]);
                setLoading(false);
            });
    }, [selected]);

    const filtered = list.filter((p) => {
        if (!q) return true;
        const s = q.toLowerCase();
        return (p.name || "").toLowerCase().includes(s) || (p.phone || "").includes(s);
    });

    return (
        <div className="h-full flex bg-slate-50">

            {/* LEFT SIDEBAR: This is where your list lives */}
            <div className="w-96 shrink-0 border-r border-slate-200 bg-white flex flex-col">
                <div className="p-4 border-b border-slate-200">
                    <h2 className="font-semibold text-slate-900 mb-2">Докторы</h2>
                    <input
                        className="w-full px-3 py-2 border border-slate-200 rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-brand-500"
                        placeholder="Поиск по имени или телефону"
                        value={q}
                        onChange={(e) => setQ(e.target.value)}
                    />
                </div>

                <div className="flex-1 overflow-auto">
                    {filtered.length === 0 ? (
                        <div className="p-4 text-sm text-slate-500 text-center">Ничего не найдено</div>
                    ) : (
                        filtered.map((d) => (
                            <button
                                key={d.id}
                                onClick={() => setSelected(d)}
                                className={`w-full text-left p-3 border-b border-slate-100 transition ${
                                    selected?.id === d.id ? "bg-blue-50" : "hover:bg-slate-50"
                                }`}
                            >
                                <div className="flex items-center gap-3">
                                    <div className="w-9 h-9 rounded-full bg-slate-200 grid place-items-center text-slate-600 font-medium shrink-0">
                                        {(d.name || "?").charAt(0).toUpperCase()}
                                    </div>
                                    <div className="flex-1 min-w-0">
                                        <div className="font-medium text-sm text-slate-900 truncate">
                                            {d.name || "—"}
                                        </div>
                                        <div className="text-xs text-slate-500 truncate">{d.phone || "Нет телефона"}</div>
                                    </div>
                                </div>
                            </button>
                        ))
                    )}
                </div>
            </div>

            {/* RIGHT SIDE: Detail View */}
            <div className="flex-1 overflow-auto p-6">
                {!selected ? (
                    <div className="h-full grid place-items-center text-slate-400 text-center">
                        <div>
                            <div className="text-4xl mb-2">🧑‍⚕️</div>
                            <p>Выберите доктора слева для просмотра деталей</p>
                        </div>
                    </div>
                ) : (
                    <div className="max-w-3xl mx-auto space-y-4">
                        {/* Profile Card */}
                        <div className="bg-white rounded-xl shadow-sm border border-slate-200 p-5">
                            <div className="flex items-center gap-4">
                                <div className="w-16 h-16 rounded-full bg-blue-100 text-blue-700 grid place-items-center font-bold text-2xl">
                                    {(selected.name || "?").charAt(0).toUpperCase()}
                                </div>
                                <div>
                                    <h1 className="text-xl font-bold text-slate-900">
                                        {selected.name || "Имя не указано"}
                                    </h1>
                                </div>
                            </div>
                        </div>

                        {/* History or Info Card */}
                        <div className="bg-white rounded-xl shadow-sm border border-slate-200 p-5">
                            <h2 className="font-semibold text-slate-900 mb-4 flex items-center justify-between">
                                История или детали
                                {loading && <span className="text-xs font-normal text-slate-400 animate-pulse">Загрузка...</span>}
                            </h2>

                            {appts.length === 0 ? (
                                <div className="py-4 text-sm text-slate-500">
                                    <div className="grid grid-cols-2 gap-4">
                                        <div className="p-3 bg-slate-50 rounded-lg">
                                            <div className="text-slate-400 text-[10px] uppercase font-bold tracking-wider">Специальносьти:</div>
                                            <div className="text-slate-700">{(selected.specialties || []).join(", ")}</div>
                                        </div>
                                    </div>
                                </div>
                            ) : (
                                <ul className="divide-y divide-slate-100">
                                    {appts.map((a) => (
                                        <li key={a.id} className="py-3 flex justify-between items-center">
                                            <div>
                                                <div className="font-medium text-slate-900">{new Date(a.starts_at).toLocaleDateString()}</div>
                                                <div className="text-xs text-slate-500">{a.service}</div>
                                            </div>
                                            <span className={`text-[10px] px-2 py-1 rounded-full ${statusColor(a.status)}`}>
                                                {statusLabel(a.status)}
                                            </span>
                                        </li>
                                    ))}
                                </ul>
                            )}
                        </div>
                    </div>
                )}
            </div>
        </div>
    );
}

function statusColor(s) {
    const colors = {
        scheduled: "bg-blue-50 text-blue-700",
        confirmed: "bg-emerald-50 text-emerald-700",
        completed: "bg-slate-100 text-slate-600",
        cancelled: "bg-red-50 text-red-600",
    };
    return colors[s] || "bg-slate-50 text-slate-500";
}

function statusLabel(s) {
    const labels = {
        scheduled: "Запланировано",
        confirmed: "Подтверждено",
        completed: "Завершено",
        cancelled: "Отменено",
    };
    return labels[s] || s;
}