import { useCallback, useEffect, useMemo, useState } from "react";
import { api } from "../api/client";
import { useTranslation } from "../hooks/useTranslation";

// ── period helpers ─────────────────────────────────────────────────────────────

function periodBounds(key) {
  const now = new Date();
  const today = new Date(now.getFullYear(), now.getMonth(), now.getDate());
  if (key === "today") {
    return { from: today, to: new Date(today.getTime() + 86400000) };
  }
  if (key === "week") {
    const from = new Date(today.getTime() - 6 * 86400000);
    return { from, to: new Date(today.getTime() + 86400000) };
  }
  // month
  const from = new Date(now.getFullYear(), now.getMonth(), 1);
  const to = new Date(now.getFullYear(), now.getMonth() + 1, 1);
  return { from, to };
}

function fmtISO(d) {
  return d.toISOString();
}

function fmtMoney(n) {
  if (!n && n !== 0) return "—";
  return new Intl.NumberFormat("ru-KZ", { maximumFractionDigits: 0 }).format(n) + " ₸";
}

function fmtPct(n) {
  return (n * 100).toFixed(0) + "%";
}

// ── status metadata ────────────────────────────────────────────────────────────

const STATUS_META = {
  scheduled:  { bg: "bg-blue-50 dark:bg-blue-900/30",    text: "text-blue-700 dark:text-blue-300",    dot: "bg-blue-500" },
  confirmed:  { bg: "bg-emerald-50 dark:bg-emerald-900/30", text: "text-emerald-700 dark:text-emerald-300", dot: "bg-emerald-500" },
  in_process: { bg: "bg-violet-50 dark:bg-violet-900/30", text: "text-violet-700 dark:text-violet-300", dot: "bg-violet-500" },
  came:       { bg: "bg-teal-50 dark:bg-teal-900/30",    text: "text-teal-700 dark:text-teal-300",    dot: "bg-teal-500" },
  late:       { bg: "bg-amber-50 dark:bg-amber-900/30",  text: "text-amber-700 dark:text-amber-300",  dot: "bg-amber-500" },
  cancelled:  { bg: "bg-red-50 dark:bg-red-900/30",      text: "text-red-600 dark:text-red-400",      dot: "bg-red-400" },
  completed:  { bg: "bg-slate-50 dark:bg-slate-700/40",  text: "text-slate-500 dark:text-slate-400",  dot: "bg-slate-400" },
};

// ── sub-components ─────────────────────────────────────────────────────────────

function KpiCard({ label, value, sub, accent = "blue" }) {
  const accents = {
    blue:    "from-blue-500 to-blue-600",
    emerald: "from-emerald-500 to-emerald-600",
    violet:  "from-violet-500 to-violet-600",
    amber:   "from-amber-500 to-amber-600",
    red:     "from-red-400 to-red-500",
    slate:   "from-slate-400 to-slate-500",
    teal:    "from-teal-500 to-teal-600",
  };
  return (
    <div className="bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700 p-4 shadow-sm flex flex-col gap-1">
      <div className={`w-8 h-1 rounded-full bg-gradient-to-r ${accents[accent] ?? accents.blue} mb-1`} />
      <div className="text-2xl font-bold text-slate-900 dark:text-slate-100 tabular-nums">{value}</div>
      <div className="text-xs font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wide">{label}</div>
      {sub && <div className="text-[10px] text-slate-400 dark:text-slate-500 mt-0.5">{sub}</div>}
    </div>
  );
}

function SectionHeader({ title, right }) {
  return (
    <div className="flex items-center justify-between mb-3">
      <h2 className="text-sm font-bold text-slate-700 dark:text-slate-200 uppercase tracking-wider">{title}</h2>
      {right}
    </div>
  );
}

function Card({ children, className = "" }) {
  return (
    <div className={`bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700 shadow-sm ${className}`}>
      {children}
    </div>
  );
}

function FunnelBar({ label, count, max, accent }) {
  const pct = max > 0 ? Math.round((count / max) * 100) : 0;
  const colors = {
    blue: "bg-blue-500",
    emerald: "bg-emerald-500",
    teal: "bg-teal-500",
    violet: "bg-violet-500",
  };
  return (
    <div className="flex items-center gap-3">
      <div className="w-24 text-[11px] text-slate-500 dark:text-slate-400 text-right shrink-0">{label}</div>
      <div className="flex-1 bg-slate-100 dark:bg-slate-700 rounded-full h-5 relative overflow-hidden">
        <div
          className={`h-full rounded-full transition-all duration-500 ${colors[accent] ?? colors.blue}`}
          style={{ width: `${pct}%` }}
        />
        <span className="absolute inset-0 flex items-center px-2 text-[10px] font-bold text-white mix-blend-overlay">
          {count}
        </span>
      </div>
      <div className="w-10 text-[11px] font-semibold text-slate-600 dark:text-slate-300 text-right tabular-nums shrink-0">
        {pct}%
      </div>
    </div>
  );
}

function TrendChart({ trend }) {
  if (!trend || trend.length === 0) return null;
  const maxIncome = Math.max(...trend.map((p) => p.income), 1);
  return (
    <div className="flex items-end gap-1 h-24 w-full">
      {trend.map((p) => {
        const h = Math.max(4, Math.round((p.income / maxIncome) * 88));
        return (
          <div
            key={p.date}
            className="flex-1 group relative"
            title={`${p.date}: ${fmtMoney(p.income)}`}
          >
            <div
              className="w-full bg-blue-400 dark:bg-blue-500 rounded-sm transition-all hover:bg-blue-500 dark:hover:bg-blue-400"
              style={{ height: `${h}px` }}
            />
            {p.expense > 0 && (
              <div
                className="absolute bottom-0 w-full bg-red-300 dark:bg-red-500/50 rounded-sm opacity-60"
                style={{ height: `${Math.max(2, Math.round((p.expense / maxIncome) * 88))}px` }}
              />
            )}
          </div>
        );
      })}
    </div>
  );
}

function Heatmap({ heatmap }) {
  const days = ["Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"];
  const hours = Array.from({ length: 12 }, (_, i) => 9 + i); // 9-20
  const allVals = days.flatMap((d) => hours.map((h) => heatmap?.[d]?.[h] ?? 0));
  const maxVal = Math.max(...allVals, 1);

  function cellColor(v) {
    if (!v) return "bg-slate-100 dark:bg-slate-700";
    const pct = v / maxVal;
    if (pct < 0.25) return "bg-blue-100 dark:bg-blue-900/40";
    if (pct < 0.5)  return "bg-blue-300 dark:bg-blue-700/60";
    if (pct < 0.75) return "bg-blue-500 dark:bg-blue-500/80";
    return "bg-blue-700 dark:bg-blue-400";
  }

  return (
    <div className="overflow-x-auto">
      <div className="inline-grid gap-0.5" style={{ gridTemplateColumns: `28px repeat(${hours.length}, 1fr)` }}>
        {/* corner */}
        <div />
        {hours.map((h) => (
          <div key={h} className="text-[9px] text-slate-400 text-center pb-1 font-medium">{h}</div>
        ))}
        {days.map((d) => (
          <>
            <div key={`d-${d}`} className="text-[9px] text-slate-400 flex items-center justify-end pr-1 font-medium">{d}</div>
            {hours.map((h) => {
              const v = heatmap?.[d]?.[h] ?? 0;
              return (
                <div
                  key={`${d}-${h}`}
                  title={`${d} ${h}:00 — ${v}`}
                  className={`h-5 w-full rounded-sm transition-colors ${cellColor(v)}`}
                />
              );
            })}
          </>
        ))}
      </div>
    </div>
  );
}

function PaymentTypeBar({ name, amount, max }) {
  const pct = max > 0 ? Math.round((amount / max) * 100) : 0;
  return (
    <div>
      <div className="flex justify-between text-[11px] mb-0.5">
        <span className="text-slate-600 dark:text-slate-300 font-medium truncate">{name}</span>
        <span className="text-slate-500 dark:text-slate-400 tabular-nums shrink-0 ml-2">{fmtMoney(amount)}</span>
      </div>
      <div className="h-1.5 bg-slate-100 dark:bg-slate-700 rounded-full overflow-hidden">
        <div className="h-full bg-emerald-500 dark:bg-emerald-400 rounded-full" style={{ width: `${pct}%` }} />
      </div>
    </div>
  );
}

// ── main component ─────────────────────────────────────────────────────────────

export default function Dashboard() {
  const { t } = useTranslation();

  const [period, setPeriod] = useState("month");
  const [todayData, setTodayData]     = useState(null);
  const [statsData, setStatsData]     = useState(null);
  const [revenueData, setRevenueData] = useState(null);
  const [loading, setLoading]         = useState(false);
  const [error, setError]             = useState(null);
  const [tick, setTick]               = useState(0);

  const { from, to } = useMemo(() => periodBounds(period), [period, tick]); // tick forces re-fetch on refresh

  const refresh = useCallback(() => setTick((n) => n + 1), []);

  useEffect(() => {
    setLoading(true);
    setError(null);

    const fromISO = fmtISO(from);
    const toISO   = fmtISO(to);

    Promise.all([
      api.dashboardToday(),
      api.dashboardStats(fromISO, toISO),
      api.dashboardRevenue(fromISO, toISO),
    ])
      .then(([td, st, rev]) => {
        setTodayData(td);
        setStatsData(st);
        setRevenueData(rev);
      })
      .catch((err) => {
        console.error("dashboard fetch", err);
        setError(err.message);
      })
      .finally(() => setLoading(false));
  }, [from, to]);

  const today = todayData;
  const stats = statsData;
  const rev   = revenueData;

  const funnelMax = stats?.funnel?.booked ?? 0;

  const maxPayType = rev?.by_type?.length
    ? Math.max(...rev.by_type.map((b) => b.amount))
    : 1;

  return (
    <div className="h-full flex flex-col bg-[#F7F8FA] dark:bg-slate-900">
      {/* TOP BAR */}
      <div className="h-16 bg-white dark:bg-slate-950 border-b border-slate-200 dark:border-slate-800 flex items-center justify-between px-6 shrink-0">
        <div>
          <h1 className="text-lg font-bold text-slate-900 dark:text-slate-100">{t("dash.title")}</h1>
          <p className="text-xs text-slate-500 dark:text-slate-400 font-medium">{t("dash.subtitle")}</p>
        </div>
        <div className="flex items-center gap-2">
          {/* Period selector */}
          {["month", "week", "today"].map((p) => (
            <button
              key={p}
              onClick={() => setPeriod(p)}
              className={`px-3 py-1.5 text-xs font-medium rounded-lg border transition-colors ${
                period === p
                  ? "bg-blue-600 text-white border-blue-600"
                  : "bg-white dark:bg-slate-800 text-slate-600 dark:text-slate-300 border-slate-200 dark:border-slate-700 hover:bg-slate-50 dark:hover:bg-slate-700"
              }`}
            >
              {t(`dash.period_${p}`)}
            </button>
          ))}
          <button
            onClick={refresh}
            disabled={loading}
            className="px-3 py-1.5 text-xs font-medium text-blue-600 dark:text-blue-400 bg-blue-50 dark:bg-blue-900/30 border border-blue-100 dark:border-blue-800 rounded-lg hover:bg-blue-100 dark:hover:bg-blue-900/50 transition-colors disabled:opacity-50"
          >
            {loading ? "…" : t("dash.refresh")}
          </button>
        </div>
      </div>

      {/* BODY */}
      <div className="flex-1 overflow-auto p-4 lg:p-6 space-y-6">

        {error && (
          <div className="text-sm text-red-500 bg-red-50 dark:bg-red-900/20 border border-red-100 dark:border-red-800 rounded-xl px-4 py-3">
            {error}
          </div>
        )}

        {/* ── TODAY SECTION ── */}
        <section>
          <SectionHeader title={t("dash.today_section")} right={
            today && (
              <span className="text-xs text-slate-400 dark:text-slate-500">
                {today.date}
              </span>
            )
          } />

          {/* Status KPI strip */}
          <div className="grid grid-cols-2 sm:grid-cols-4 lg:grid-cols-8 gap-3 mb-4">
            {[
              { key: "total",      val: today?.total ?? "—",                     accent: "blue" },
              { key: "confirmed",  val: today?.counts?.confirmed ?? "—",          accent: "emerald" },
              { key: "in_process", val: today?.counts?.in_process ?? "—",         accent: "violet" },
              { key: "came",       val: today?.counts?.came ?? "—",               accent: "teal" },
              { key: "scheduled",  val: today?.counts?.scheduled ?? "—",          accent: "blue" },
              { key: "late",       val: today?.counts?.late ?? "—",               accent: "amber" },
              { key: "cancelled",  val: today?.counts?.cancelled ?? "—",          accent: "red" },
              { key: "new_patients", val: today?.new_patients_today ?? "—",       accent: "slate" },
            ].map(({ key, val, accent }) => (
              <KpiCard key={key} label={t(`dash.${key}`)} value={val} accent={accent} />
            ))}
          </div>

          {/* Upcoming queue */}
          {today?.upcoming?.length > 0 && (
            <Card>
              <div className="px-4 py-3 border-b border-slate-100 dark:border-slate-700">
                <span className="text-xs font-bold text-slate-600 dark:text-slate-300 uppercase tracking-wide">
                  {t("dash.upcoming")} ({today.upcoming.length})
                </span>
              </div>
              <div className="divide-y divide-slate-50 dark:divide-slate-700 max-h-52 overflow-y-auto">
                {today.upcoming.map((a) => {
                  const statusKey = a.status === 0 ? "scheduled" : a.status === 1 ? "confirmed" : a.status === 6 ? "late" : "scheduled";
                  const meta = STATUS_META[statusKey] ?? STATUS_META.scheduled;
                  return (
                    <div key={a.id} className="flex items-center gap-3 px-4 py-2.5 hover:bg-slate-50/50 dark:hover:bg-slate-700/20 transition-colors">
                      <div className={`w-2 h-2 rounded-full shrink-0 ${meta.dot}`} />
                      <span className="text-sm font-bold text-slate-800 dark:text-slate-100 tabular-nums w-12 shrink-0">
                        {a.start?.slice(11, 16)}
                      </span>
                      {a.is_first && (
                        <span className="text-[9px] font-bold px-1 bg-amber-100 dark:bg-amber-900/40 text-amber-700 dark:text-amber-400 rounded shrink-0">
                          NEW
                        </span>
                      )}
                      <span className="text-xs text-slate-500 dark:text-slate-400 truncate flex-1">
                        {a.cabinet ? `Каб. ${a.cabinet}` : `Dr #${a.doctor_id}`}
                      </span>
                      <span className={`text-[10px] font-bold px-2 py-0.5 rounded-full ${meta.bg} ${meta.text}`}>
                        {t(`dash.${statusKey}`)}
                      </span>
                    </div>
                  );
                })}
              </div>
            </Card>
          )}
        </section>

        {/* ── STATS + HEATMAP ── */}
        <section>
          <SectionHeader title={t("dash.stats_section")} right={
            stats && (
              <span className="text-xs text-slate-400 dark:text-slate-500">
                {t("dash.completion_rate")}: {fmtPct(stats.completion_rate)}
                {" · "}
                {t("dash.new_patient_rate")}: {fmtPct(stats.new_patient_rate)}
              </span>
            )
          } />

          <div className="grid grid-cols-1 lg:grid-cols-3 gap-4">
            {/* Funnel */}
            <Card className="p-4 lg:col-span-1">
              <div className="text-xs font-bold text-slate-600 dark:text-slate-300 uppercase tracking-wide mb-3">
                {t("dash.funnel")}
              </div>
              {stats ? (
                <div className="space-y-2.5">
                  <FunnelBar label={t("dash.booked")}           count={stats.funnel.booked}    max={funnelMax} accent="blue" />
                  <FunnelBar label={t("dash.funnel_confirmed")} count={stats.funnel.confirmed} max={funnelMax} accent="teal" />
                  <FunnelBar label={t("dash.funnel_came")}      count={stats.funnel.came}      max={funnelMax} accent="emerald" />
                  <FunnelBar label={t("dash.funnel_completed")} count={stats.funnel.completed} max={funnelMax} accent="violet" />
                </div>
              ) : (
                <Skeleton />
              )}
            </Card>

            {/* By doctor */}
            <Card className="p-4 lg:col-span-2 overflow-hidden">
              <div className="text-xs font-bold text-slate-600 dark:text-slate-300 uppercase tracking-wide mb-3">
                {t("dash.by_doctor")}
              </div>
              {stats?.by_doctor?.length > 0 ? (
                <div className="overflow-x-auto">
                  <table className="w-full text-xs">
                    <thead>
                      <tr className="text-[10px] uppercase text-slate-400 dark:text-slate-500 border-b border-slate-100 dark:border-slate-700">
                        <th className="text-left pb-2 font-semibold">{t("dash.doctor_id")}</th>
                        <th className="text-right pb-2 font-semibold">{t("dash.total")}</th>
                        <th className="text-right pb-2 font-semibold">{t("dash.completed")}</th>
                        <th className="text-right pb-2 font-semibold">{t("dash.cancelled")}</th>
                        <th className="text-right pb-2 font-semibold">{t("dash.new_pats")}</th>
                      </tr>
                    </thead>
                    <tbody className="divide-y divide-slate-50 dark:divide-slate-700">
                      {stats.by_doctor.map((d) => (
                        <tr key={d.doctor_id} className="hover:bg-slate-50/50 dark:hover:bg-slate-700/20 transition-colors">
                          <td className="py-2 font-medium text-slate-700 dark:text-slate-200">#{d.doctor_id}</td>
                          <td className="py-2 text-right font-bold tabular-nums text-slate-900 dark:text-slate-100">{d.total}</td>
                          <td className="py-2 text-right tabular-nums text-emerald-600 dark:text-emerald-400">{d.completed}</td>
                          <td className="py-2 text-right tabular-nums text-red-500 dark:text-red-400">{d.cancelled}</td>
                          <td className="py-2 text-right tabular-nums text-amber-600 dark:text-amber-400">{d.new_patients}</td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              ) : (
                <Skeleton />
              )}
            </Card>
          </div>

          {/* Heatmap */}
          {stats?.heatmap && (
            <Card className="p-4 mt-4">
              <div className="text-xs font-bold text-slate-600 dark:text-slate-300 uppercase tracking-wide mb-3">
                {t("dash.heatmap")}
              </div>
              <Heatmap heatmap={stats.heatmap} />
            </Card>
          )}
        </section>

        {/* ── REVENUE SECTION ── */}
        <section>
          <SectionHeader title={t("dash.revenue_section")} />

          <div className="grid grid-cols-1 sm:grid-cols-3 gap-3 mb-4">
            <KpiCard label={t("dash.income")}  value={fmtMoney(rev?.total_income)}  accent="emerald" />
            <KpiCard label={t("dash.expense")} value={fmtMoney(rev?.total_expense)} accent="red" />
            <KpiCard
              label={t("dash.net")}
              value={fmtMoney(rev?.net)}
              accent={(rev?.net ?? 0) >= 0 ? "teal" : "red"}
              sub={(rev?.net ?? 0) >= 0 ? "↑" : "↓"}
            />
          </div>

          <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
            {/* Payment type breakdown */}
            <Card className="p-4">
              <div className="text-xs font-bold text-slate-600 dark:text-slate-300 uppercase tracking-wide mb-3">
                {t("dash.by_type")}
              </div>
              {rev?.by_type?.length > 0 ? (
                <div className="space-y-3">
                  {rev.by_type.slice(0, 8).map((bt) => (
                    <PaymentTypeBar key={bt.name} name={bt.name} amount={bt.amount} max={maxPayType} />
                  ))}
                </div>
              ) : (
                <div className="text-xs text-slate-400 dark:text-slate-500">{t("dash.no_data")}</div>
              )}
            </Card>

            {/* Daily trend */}
            <Card className="p-4">
              <div className="text-xs font-bold text-slate-600 dark:text-slate-300 uppercase tracking-wide mb-3">
                {t("dash.trend")}
              </div>
              {rev?.trend?.length > 0 ? (
                <>
                  <TrendChart trend={rev.trend} />
                  <div className="flex gap-4 mt-2">
                    <div className="flex items-center gap-1.5 text-[10px] text-slate-500 dark:text-slate-400">
                      <span className="w-3 h-1.5 rounded-sm bg-blue-400 inline-block" />
                      {t("dash.income")}
                    </div>
                    <div className="flex items-center gap-1.5 text-[10px] text-slate-500 dark:text-slate-400">
                      <span className="w-3 h-1.5 rounded-sm bg-red-300 inline-block" />
                      {t("dash.expense")}
                    </div>
                  </div>
                </>
              ) : (
                <div className="text-xs text-slate-400 dark:text-slate-500">{t("dash.no_data")}</div>
              )}
            </Card>
          </div>
        </section>
      </div>
    </div>
  );
}

function Skeleton() {
  return (
    <div className="space-y-2 animate-pulse">
      {[80, 60, 90, 40].map((w) => (
        <div key={w} className="h-3 bg-slate-100 dark:bg-slate-700 rounded" style={{ width: `${w}%` }} />
      ))}
    </div>
  );
}
