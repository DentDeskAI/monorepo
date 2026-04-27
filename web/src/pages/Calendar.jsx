import { Fragment, useEffect, useMemo, useState } from "react";
import { api } from "../api/client";
import { useTranslation } from "../hooks/useTranslation";

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

// MacDent integer status → local status string
const MD_STATUS = { 0: "scheduled", 1: "confirmed", 2: "cancelled" };

function normalizeMacdent(a) {
  return {
    id: a.id,
    starts_at: parseMacdentDate(a.start),
    patient_name: a.cabinet || null, // Using cabinet as location placeholder
    patient_phone: null,
    doctor_name: null,
    status: MD_STATUS[a.status] ?? "completed",
  };
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
    api
        .scheduleDoctors()
        .then((resp) => {
          if (resp && resp.appointments) {
            const items = resp.appointments
                .map(normalizeMacdent)
                .filter(({ starts_at }) => {
                  return starts_at >= weekStart && starts_at < weekEnd;
                });
            setItems(items);
          } else {
            setItems([]);
          }
        })
        .catch((err) => {
          console.error("Failed to fetch schedule:", err);
          setError("Failed to load schedule");
          setItems([]);
        })
        .finally(() => setLoading(false));
  }, [weekStart, weekEnd]);

  const byDayHour = useMemo(() => {
    const m = {};
    for (const a of items) {
      const d = a.starts_at;
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
        <div className="flex-1 overflow-auto p-6 relative">
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

          <div className="bg-white dark:bg-slate-800 rounded-xl shadow-sm border border-slate-200 dark:border-slate-700 overflow-hidden min-w-[900px]">
            <div className="grid" style={{ gridTemplateColumns: "60px repeat(7, 1fr)" }}>
              {/* Header Row */}
              <div className="bg-slate-50 dark:bg-slate-900 border-b border-r border-slate-200 dark:border-slate-700 p-2" />
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
                    {/* Time Label */}
                    <div className="text-xs text-slate-400 font-medium text-center py-4 border-r border-b border-slate-100 dark:border-slate-700 bg-slate-50/50 dark:bg-slate-900/30">
                      {h}:00
                    </div>

                    {/* Day Cells */}
                    {Array.from({ length: 7 }).map((_, dayIdx) => {
                      const cellItems = byDayHour[`${dayIdx}-${h}`] || [];
                      return (
                          <div
                              key={`${h}-${dayIdx}`}
                              className="min-h-[80px] border-r border-b border-slate-100 dark:border-slate-700 p-1 space-y-1.5 hover:bg-slate-50/30 dark:hover:bg-slate-700/20 transition-colors"
                          >
                            {cellItems.map((a) => {
                              const isCancelled = a.status === "cancelled";
                              const d = a.starts_at;

                              return (
                                  <div
                                      key={a.id}
                                      className={`text-[11px] px-2 py-1.5 rounded border transition-all cursor-pointer ${
                                          isCancelled
                                              ? "bg-slate-50 dark:bg-slate-700 text-slate-400 border-slate-100 dark:border-slate-600 line-through"
                                              : "bg-blue-50 dark:bg-blue-900/30 text-blue-700 dark:text-blue-300 border-blue-100 dark:border-blue-800 shadow-sm hover:shadow-md"
                                      }`}
                                  >
                                    <div className="font-bold mb-0.5">
                                      {String(d.getHours()).padStart(2, "0")}:{String(d.getMinutes()).padStart(2, "0")}
                                      {" · "}
                                      {a.patient_name || a.patient_phone || t(i18nKeys.calendar.no_patient)}
                                    </div>
                                    {a.doctor_name && (
                                        <div className="text-slate-500 dark:text-slate-400 truncate font-medium">
                                          {a.doctor_name}
                                        </div>
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
      </div>
  );
}