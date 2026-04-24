import { useEffect, useState } from "react";
import { api } from "../api/client";
import { Link } from "react-router-dom";

export default function Dashboard() {
  const [stats, setStats] = useState(null);
  const [chats, setChats] = useState([]);

  useEffect(() => {
    api.stats().then(setStats).catch(() => {});
    api.chats().then((cs) => setChats(cs.slice(0, 5))).catch(() => {});
  }, []);

  const tiles = [
    { label: "Активные чаты (24ч)", value: stats?.active_chats ?? "—", tint: "bg-blue-50 text-blue-700" },
    { label: "Записей сегодня",      value: stats?.today_appts ?? "—",   tint: "bg-emerald-50 text-emerald-700" },
    { label: "Пациентов всего",      value: stats?.total_patients ?? "—", tint: "bg-violet-50 text-violet-700" },
  ];

  return (
    <div className="h-full overflow-auto p-6 space-y-6">
      <div>
        <h1 className="text-2xl font-semibold text-slate-900">Здравствуйте 👋</h1>
        <p className="text-slate-500">Айгуль работает за вас — вот короткая сводка</p>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        {tiles.map((t) => (
          <div key={t.label} className="card p-5">
            <div className={`text-xs font-medium px-2 py-1 rounded-md inline-block ${t.tint}`}>
              {t.label}
            </div>
            <div className="text-3xl font-semibold text-slate-900 mt-3">{t.value}</div>
          </div>
        ))}
      </div>

      <div className="card p-5">
        <div className="flex items-center justify-between mb-4">
          <h2 className="font-semibold text-slate-900">Последние диалоги</h2>
          <Link to="/chats" className="text-sm text-brand-600 hover:underline">Все чаты →</Link>
        </div>
        {chats.length === 0 ? (
          <div className="text-sm text-slate-500">Пока тихо. Когда кто-то напишет — здесь появится.</div>
        ) : (
          <ul className="divide-y divide-slate-100">
            {chats.map((row) => (
              <li key={row.conversation.id} className="py-3 flex items-center gap-3">
                <div className="w-9 h-9 rounded-full bg-slate-100 grid place-items-center text-slate-600 font-medium">
                  {(row.patient?.name || row.patient?.phone || "?").charAt(0).toUpperCase()}
                </div>
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2">
                    <div className="font-medium text-slate-900 truncate">
                      {row.patient?.name || row.patient?.phone}
                    </div>
                    {row.conversation.status === "handoff" && (
                      <span className="text-xs px-1.5 py-0.5 rounded bg-amber-100 text-amber-700">
                        оператор
                      </span>
                    )}
                  </div>
                  <div className="text-sm text-slate-500 truncate">
                    {row.last_message?.body || "—"}
                  </div>
                </div>
                <Link to={`/chats/${row.conversation.id}`} className="btn btn-secondary text-xs">
                  Открыть
                </Link>
              </li>
            ))}
          </ul>
        )}
      </div>
    </div>
  );
}
