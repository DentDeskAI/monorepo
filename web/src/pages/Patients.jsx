import { useEffect, useState } from "react";
import { api } from "../api/client";

export default function Patients() {
  const [list, setList] = useState([]);
  const [q, setQ] = useState("");
  const [selected, setSelected] = useState(null);
  const [appts, setAppts] = useState([]);

  const [loadingList, setLoadingList] = useState(false);
  const [loadingDetails, setLoadingDetails] = useState(false);

  // 1. Fetch patients
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

  // 2. Fetch patient history
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

  // 3. Filtering
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
      <div className="h-full flex bg-slate-50">

        {/* LEFT */}
        <div className="w-96 shrink-0 border-r border-slate-200 bg-white flex flex-col">
          <div className="p-4 border-b border-slate-200">
            <h2 className="font-semibold text-slate-900 mb-2">Пациенты</h2>
            <input
                className="w-full px-3 py-2 border border-slate-200 rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-brand-500"
                placeholder="Поиск по имени или телефону"
                value={q}
                onChange={(e) => setQ(e.target.value)}
            />
          </div>

          <div className="flex-1 overflow-auto">
            {loadingList ? (
                <div className="p-4 text-sm text-slate-400 text-center">Загрузка...</div>
            ) : filtered.length === 0 ? (
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
                          <div className="text-xs text-slate-500 truncate">
                            {d.phone || d.number || "Нет контакта"}
                          </div>
                        </div>
                      </div>
                    </button>
                ))
            )}
          </div>
        </div>

        {/* RIGHT */}
        <div className="flex-1 overflow-auto p-6">
          {!selected ? (
              <div className="h-full grid place-items-center text-slate-400 text-center">
                <div>
                  <div className="text-4xl mb-2">🧑‍⚕️</div>
                  <p>Выберите пациента слева</p>
                </div>
              </div>
          ) : (
              <div className="max-w-3xl mx-auto space-y-4">

                {/* Profile */}
                <div className="bg-white rounded-xl shadow-sm border border-slate-200 p-5">
                  <div className="flex items-center gap-4">
                    <div className="w-16 h-16 rounded-full bg-blue-100 text-blue-700 grid place-items-center font-bold text-2xl">
                      {(selected.name || "?").charAt(0).toUpperCase()}
                    </div>
                    <div>
                      <h1 className="text-xl font-bold text-slate-900">
                        {selected.name || "Имя не указано"}
                      </h1>
                      <div className="text-sm text-slate-500">
                        {selected.phone || selected.number || "Нет контакта"}
                      </div>
                    </div>
                  </div>
                </div>

                {/* Info + History */}
                <div className="bg-white rounded-xl shadow-sm border border-slate-200 p-5">

                  <h2 className="font-semibold text-slate-900 mb-4">
                    Информация о пациенте
                  </h2>

                  <div className="grid grid-cols-2 gap-4 mb-6 text-sm">
                    <Info label="ID" value={selected.id} />
                    <Info label="Телефон" value={selected.phone} />
                    <Info label="Номер" value={selected.number} />
                    <Info label="Пол" value={selected.gender} />
                    <Info label="Дата рождения" value={selected.birth} />
                    <Info label="Ребёнок" value={selected.isChild ? "Да" : "Нет"} />
                    <Info label="Источник" value={selected.whereKnow} />
                    <Info label="Комментарий" value={selected.comment} />
                  </div>

                  <h2 className="font-semibold text-slate-900 mb-4 flex items-center justify-between">
                    История
                    {loadingDetails && (
                        <span className="text-xs text-slate-400 animate-pulse">Загрузка...</span>
                    )}
                  </h2>

                  {appts.length === 0 ? (
                      <div className="py-4 text-sm text-slate-400">
                        История отсутствует
                      </div>
                  ) : (
                      <ul className="divide-y divide-slate-100">
                        {appts.map((a) => (
                            <li key={a.id} className="py-3 flex justify-between items-center">
                              <div>
                                <div className="font-medium text-slate-900">
                                  {formatDate(a.starts_at)}
                                </div>
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

// utils

function Info({ label, value }) {
  return (
      <div className="p-3 bg-slate-50 rounded-lg">
        <div className="text-[10px] uppercase text-slate-400 font-bold tracking-wide">
          {label}
        </div>
        <div className="text-slate-700">{value || "—"}</div>
      </div>
  );
}

function formatDate(v) {
  try {
    return new Date(v).toLocaleDateString();
  } catch {
    return v;
  }
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