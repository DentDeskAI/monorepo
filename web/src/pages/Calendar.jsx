import { Fragment, useEffect, useMemo, useState } from "react";
import { api } from "../api/client";

const HOURS = Array.from({ length: 12 }, (_, i) => 9 + i); // 9..20
const DAYS = ["Пн", "Вт", "Ср", "Чт", "Пт", "Сб", "Вс"];

function startOfWeek(d) {
  const x = new Date(d);
  const day = (x.getDay() + 6) % 7; // Mon=0
  x.setHours(0, 0, 0, 0);
  x.setDate(x.getDate() - day);
  return x;
}

export default function Calendar() {
  const [weekStart, setWeekStart] = useState(() => startOfWeek(new Date()));
  const [items, setItems] = useState([]);
  const [loading, setLoading] = useState(false);

  const weekEnd = useMemo(() => {
    const e = new Date(weekStart);
    e.setDate(e.getDate() + 7);
    return e;
  }, [weekStart]);

  useEffect(() => {
    setLoading(true);
    api.calendar(weekStart.toISOString(), weekEnd.toISOString())
      .then(setItems)
      .catch(() => setItems([]))
      .finally(() => setLoading(false));
  }, [weekStart, weekEnd]);

  const byDayHour = useMemo(() => {
    const m = {};
    for (const a of items) {
      const d = new Date(a.starts_at);
      const dayIdx = (d.getDay() + 6) % 7;
      const key = `${dayIdx}-${d.getHours()}`;
      if (!m[key]) m[key] = [];
      m[key].push(a);
    }
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
    <div className="h-full flex flex-col">
      <div className="px-6 py-4 border-b border-slate-200 bg-white flex items-center justify-between">
        <div>
          <h1 className="text-lg font-semibold text-slate-900">Календарь</h1>
          <p className="text-xs text-slate-500">Записи в клинике за неделю</p>
        </div>
        <div className="flex items-center gap-2">
          <button onClick={() => shift(-1)} className="btn btn-secondary text-xs">← Неделя</button>
          <div className="text-sm font-medium text-slate-700 px-3">{fmtRange()}</div>
          <button onClick={() => shift(1)} className="btn btn-secondary text-xs">Неделя →</button>
          <button onClick={() => setWeekStart(startOfWeek(new Date()))} className="btn btn-secondary text-xs">
            Сегодня
          </button>
        </div>
      </div>

      <div className="flex-1 overflow-auto p-6">
        {loading && <div className="text-sm text-slate-400 mb-3">Загружаем...</div>}
        <div className="card overflow-hidden">
          <div className="grid" style={{ gridTemplateColumns: "60px repeat(7, 1fr)" }}>
            <div className="bg-slate-50 border-b border-r border-slate-200" />
            {DAYS.map((d, i) => {
              const date = new Date(weekStart);
              date.setDate(date.getDate() + i);
              const isToday = new Date().toDateString() === date.toDateString();
              return (
                <div key={d} className={`text-center py-2 border-b border-r border-slate-200 text-xs font-medium ${
                  isToday ? "bg-brand-50 text-brand-700" : "bg-slate-50 text-slate-600"
                }`}>
                  <div>{d}</div>
                  <div className="text-slate-400">{date.getDate()}.{String(date.getMonth()+1).padStart(2,"0")}</div>
                </div>
              );
            })}

            {HOURS.map((h) => (
              <Fragment key={`row-${h}`}>
                <div className="text-xs text-slate-400 text-center py-4 border-r border-b border-slate-200">
                  {h}:00
                </div>
                {DAYS.map((_, dayIdx) => {
                  const cellItems = byDayHour[`${dayIdx}-${h}`] || [];
                  return (
                    <div
                      key={`${h}-${dayIdx}`}
                      className="min-h-[60px] border-r border-b border-slate-200 p-1 space-y-1"
                    >
                      {cellItems.map((a) => {
                        const d = new Date(a.starts_at);
                        return (
                          <div
                            key={a.id}
                            className={`text-[11px] px-2 py-1 rounded ${
                              a.status === "cancelled"
                                ? "bg-slate-100 text-slate-500 line-through"
                                : "bg-brand-50 text-brand-700"
                            }`}
                          >
                            <div className="font-medium">
                              {String(d.getHours()).padStart(2,"0")}:{String(d.getMinutes()).padStart(2,"0")}
                              {" · "}
                              {a.patient_name || a.patient_phone || "Пациент"}
                            </div>
                            {a.doctor_name && <div className="text-slate-500 truncate">{a.doctor_name}</div>}
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
    </div>
  );
}
