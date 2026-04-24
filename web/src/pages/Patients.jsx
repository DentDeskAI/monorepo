import { useEffect, useState } from "react";
import { api } from "../api/client";

export default function Patients() {
  const [list, setList] = useState([]);
  const [q, setQ] = useState("");
  const [selected, setSelected] = useState(null);
  const [appts, setAppts] = useState([]);

  useEffect(() => {
    api.patients().then(setList).catch(() => {});
  }, []);

  useEffect(() => {
    if (!selected) { setAppts([]); return; }
    api.patientAppts(selected.id).then(setAppts).catch(() => setAppts([]));
  }, [selected]);

  const filtered = list.filter((p) => {
    if (!q) return true;
    const s = q.toLowerCase();
    return (p.name || "").toLowerCase().includes(s) || p.phone.includes(s);
  });

  return (
    <div className="h-full flex">
      <div className="w-96 shrink-0 border-r border-slate-200 bg-white flex flex-col">
        <div className="p-4 border-b border-slate-200">
          <h2 className="font-semibold text-slate-900 mb-2">Пациенты</h2>
          <input
            className="input"
            placeholder="Поиск по имени или телефону"
            value={q}
            onChange={(e) => setQ(e.target.value)}
          />
        </div>
        <div className="flex-1 overflow-auto">
          {filtered.length === 0 && (
            <div className="p-4 text-sm text-slate-500">Ничего не найдено</div>
          )}
          {filtered.map((p) => (
            <button
              key={p.id}
              onClick={() => setSelected(p)}
              className={`w-full text-left p-3 border-b border-slate-100 transition ${
                selected?.id === p.id ? "bg-brand-50" : "hover:bg-slate-50"
              }`}
            >
              <div className="flex items-center gap-3">
                <div className="w-9 h-9 rounded-full bg-slate-100 grid place-items-center text-slate-600 font-medium shrink-0">
                  {(p.name || p.phone).charAt(0).toUpperCase()}
                </div>
                <div className="flex-1 min-w-0">
                  <div className="font-medium text-sm text-slate-900 truncate">
                    {p.name || "—"}
                  </div>
                  <div className="text-xs text-slate-500 truncate">{p.phone}</div>
                </div>
              </div>
            </button>
          ))}
        </div>
      </div>

      <div className="flex-1 overflow-auto p-6">
        {!selected ? (
          <div className="h-full grid place-items-center text-slate-400">
            Выберите пациента слева
          </div>
        ) : (
          <div className="max-w-3xl space-y-4">
            <div className="card p-5">
              <div className="flex items-center gap-4">
                <div className="w-14 h-14 rounded-full bg-slate-100 grid place-items-center text-slate-600 font-medium text-xl">
                  {(selected.name || selected.phone).charAt(0).toUpperCase()}
                </div>
                <div>
                  <h1 className="text-lg font-semibold text-slate-900">
                    {selected.name || "Имя не указано"}
                  </h1>
                  <div className="text-sm text-slate-500">{selected.phone}</div>
                  <div className="text-xs text-slate-400 mt-1">
                    Язык: {selected.language === "kk" ? "казахский" : "русский"}
                  </div>
                </div>
              </div>
            </div>

            <div className="card p-5">
              <h2 className="font-semibold text-slate-900 mb-3">История записей</h2>
              {appts.length === 0 ? (
                <div className="text-sm text-slate-500">Записей ещё нет.</div>
              ) : (
                <ul className="divide-y divide-slate-100">
                  {appts.map((a) => {
                    const d = new Date(a.starts_at);
                    return (
                      <li key={a.id} className="py-3 flex items-center justify-between">
                        <div>
                          <div className="font-medium text-slate-900">
                            {d.toLocaleDateString("ru-RU", {
                              day: "2-digit", month: "long", year: "numeric",
                            })}{" "}
                            в {String(d.getHours()).padStart(2,"0")}:{String(d.getMinutes()).padStart(2,"0")}
                          </div>
                          <div className="text-sm text-slate-500">
                            {a.doctor_name || "Врач"} {a.service ? `· ${a.service}` : ""}
                          </div>
                        </div>
                        <span className={`text-xs px-2 py-1 rounded ${statusColor(a.status)}`}>
                          {statusLabel(a.status)}
                        </span>
                      </li>
                    );
                  })}
                </ul>
              )}
            </div>
          </div>
        )}
      </div>
    </div>
  );
}

function statusLabel(s) {
  return {
    scheduled: "запланирован",
    confirmed: "подтверждён",
    completed: "завершён",
    cancelled: "отменён",
    no_show:   "не пришёл",
  }[s] || s;
}
function statusColor(s) {
  return {
    scheduled: "bg-blue-50 text-blue-700",
    confirmed: "bg-emerald-50 text-emerald-700",
    completed: "bg-slate-100 text-slate-600",
    cancelled: "bg-slate-100 text-slate-500",
    no_show:   "bg-red-50 text-red-700",
  }[s] || "bg-slate-100 text-slate-600";
}
