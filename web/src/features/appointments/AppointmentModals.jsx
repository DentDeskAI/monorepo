import { useEffect, useMemo, useState } from "react";
import { api } from "../../api/client";
import Modal, { FormField, inputCls, btnPrimary, btnSecondary, btnDanger, btnGhost } from "../../components/Modal";
import { fmtDateForInput, fmtTime } from "./schedule";

const STATUS_CODES = {
  confirm: 1,
  came: 3,
  in_process: 5,
  late: 6,
  cancel: 2,
};

export function CreateAppointmentModal({ open, defaultDate, doctors, patients, onClose, onCreated, t }) {
  const [mode, setMode] = useState("book");
  const [form, setForm] = useState({
    doctor_id: "",
    patient_id: "",
    patient_search: "",
    patient_name: "",
    patient_phone: "",
    start: "",
    duration: 30,
    zhaloba: "",
    cabinet: "",
    is_first: false,
  });
  const [saving, setSaving] = useState(false);
  const [err, setErr] = useState(null);

  useEffect(() => {
    if (open) {
      setForm((f) => ({
        ...f,
        doctor_id: "",
        patient_id: "",
        patient_search: "",
        patient_name: "",
        patient_phone: "",
        start: fmtDateForInput(defaultDate),
        duration: 30,
        zhaloba: "",
        cabinet: "",
        is_first: false,
      }));
      setErr(null);
      setMode("book");
    }
  }, [open, defaultDate]);

  const set = (k) => (e) =>
    setForm({ ...form, [k]: e.target.type === "checkbox" ? e.target.checked : e.target.value });

  const matchedPatients = useMemo(() => {
    const q = form.patient_search.trim().toLowerCase();
    if (!q) return [];
    return patients
      .filter((p) =>
        (p.name || "").toLowerCase().includes(q) ||
        (p.phone || "").includes(q) ||
        String(p.id).includes(q),
      )
      .slice(0, 6);
  }, [form.patient_search, patients]);

  const submit = async () => {
    setErr(null);
    if (!form.start) {
      setErr(t("forms.start") + " ?");
      return;
    }
    const start = new Date(form.start);
    const end = new Date(start.getTime() + Number(form.duration) * 60_000);

    setSaving(true);
    try {
      if (mode === "book") {
        if (!form.doctor_id) {
          setErr(t("forms.select_doctor"));
          setSaving(false);
          return;
        }
        if (!form.patient_id) {
          setErr(t("forms.select_patient"));
          setSaving(false);
          return;
        }
        await api.createScheduleAppointment({
          doctor_id: Number(form.doctor_id),
          patient_id: Number(form.patient_id),
          starts_at: start.toISOString(),
          ends_at: end.toISOString(),
          zhaloba: form.zhaloba,
          cabinet: form.cabinet,
          is_first: form.is_first,
        });
      } else {
        if (!form.patient_name.trim() || !form.patient_phone.trim()) {
          setErr(`${t("forms.name")} + ${t("forms.phone")}`);
          setSaving(false);
          return;
        }
        await api.sendAppointmentRequest({
          patient_name: form.patient_name,
          patient_phone: form.patient_phone,
          starts_at: start.toISOString(),
          ends_at: end.toISOString(),
        });
      }
      onCreated();
    } catch (e) {
      setErr(e.message || t("forms.action_failed"));
    } finally {
      setSaving(false);
    }
  };

  const selectedPatient = patients.find((p) => String(p.id) === String(form.patient_id));

  return (
    <Modal
      open={open}
      onClose={onClose}
      title={t("forms.new_appointment")}
      width="max-w-xl"
      footer={
        <>
          <button onClick={onClose} className={btnSecondary} disabled={saving}>
            {t("forms.cancel")}
          </button>
          <button onClick={submit} disabled={saving} className={btnPrimary}>
            {saving ? t("forms.saving") : (mode === "book" ? t("forms.book") : t("forms.request"))}
          </button>
        </>
      }
    >
      {err && (
        <div className="mb-3 px-3 py-2 text-xs text-red-600 bg-red-50 dark:bg-red-900/30 border border-red-100 dark:border-red-800 rounded-lg">
          {err}
        </div>
      )}

      <div className="flex gap-2 mb-4">
        <button
          onClick={() => setMode("book")}
          className={`flex-1 px-3 py-2 text-xs font-bold rounded-lg border transition-colors ${
            mode === "book"
              ? "bg-blue-600 text-white border-blue-600"
              : "bg-slate-50 dark:bg-slate-900 text-slate-600 dark:text-slate-300 border-slate-200 dark:border-slate-700"
          }`}
        >
          {t("forms.book")}
        </button>
        <button
          onClick={() => setMode("request")}
          className={`flex-1 px-3 py-2 text-xs font-bold rounded-lg border transition-colors ${
            mode === "request"
              ? "bg-amber-500 text-white border-amber-500"
              : "bg-slate-50 dark:bg-slate-900 text-slate-600 dark:text-slate-300 border-slate-200 dark:border-slate-700"
          }`}
        >
          {t("forms.request")}
        </button>
      </div>

      <div className="grid grid-cols-2 gap-3">
        <FormField label={t("forms.start")} full>
          <input type="datetime-local" className={inputCls} value={form.start} onChange={set("start")} />
        </FormField>
        <FormField label={`${t("forms.duration")} (${t("forms.min")})`} full>
          <select className={inputCls} value={form.duration} onChange={set("duration")}>
            {[15, 30, 45, 60, 90, 120].map((d) => (
              <option key={d} value={d}>{d}</option>
            ))}
          </select>
        </FormField>

        {mode === "book" ? (
          <>
            <FormField label={t("forms.doctor")} full>
              <select className={inputCls} value={form.doctor_id} onChange={set("doctor_id")}>
                <option value="">- {t("forms.select_doctor")} -</option>
                {doctors.map((d) => (
                  <option key={d.id} value={d.id}>{d.name}</option>
                ))}
              </select>
            </FormField>

            <FormField label={t("forms.patient")} full>
              {selectedPatient ? (
                <div className="flex items-center justify-between px-3 py-2 bg-blue-50 dark:bg-blue-900/30 border border-blue-200 dark:border-blue-800 rounded-lg">
                  <div className="text-sm text-slate-800 dark:text-slate-100">
                    <div className="font-bold">{selectedPatient.name}</div>
                    <div className="text-xs text-slate-500 dark:text-slate-400">{selectedPatient.phone || `#${selectedPatient.id}`}</div>
                  </div>
                  <button
                    onClick={() => setForm({ ...form, patient_id: "", patient_search: "" })}
                    className="text-slate-400 hover:text-slate-600 text-lg leading-none"
                  >x</button>
                </div>
              ) : (
                <>
                  <input
                    className={inputCls}
                    placeholder={t("forms.patient_search")}
                    value={form.patient_search}
                    onChange={set("patient_search")}
                  />
                  {matchedPatients.length > 0 && (
                    <div className="mt-1 max-h-40 overflow-y-auto bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-lg shadow-sm">
                      {matchedPatients.map((p) => (
                        <button
                          key={p.id}
                          onClick={() => setForm({ ...form, patient_id: p.id, patient_search: "" })}
                          className="w-full text-left px-3 py-1.5 text-xs hover:bg-blue-50 dark:hover:bg-blue-900/30 text-slate-700 dark:text-slate-200 border-b border-slate-50 dark:border-slate-700 last:border-0"
                        >
                          <span className="font-medium">{p.name || "-"}</span>
                          {p.phone && <span className="text-slate-400 ml-2">{p.phone}</span>}
                        </button>
                      ))}
                    </div>
                  )}
                </>
              )}
            </FormField>

            <FormField label={t("forms.cabinet")}>
              <input className={inputCls} value={form.cabinet} onChange={set("cabinet")} />
            </FormField>
            <FormField label={t("forms.first_visit")}>
              <label className="flex items-center gap-2 mt-1.5">
                <input type="checkbox" checked={form.is_first} onChange={set("is_first")} className="w-4 h-4 accent-blue-600" />
                <span className="text-xs text-slate-600 dark:text-slate-300">{t("common.yes")}</span>
              </label>
            </FormField>

            <FormField label={t("forms.complaint")} full>
              <textarea className={inputCls} rows={2} value={form.zhaloba} onChange={set("zhaloba")} />
            </FormField>
          </>
        ) : (
          <>
            <FormField label={t("forms.name")} full>
              <input className={inputCls} value={form.patient_name} onChange={set("patient_name")} autoFocus />
            </FormField>
            <FormField label={t("forms.phone")} full>
              <input className={inputCls} value={form.patient_phone} onChange={set("patient_phone")} placeholder="+77..." />
            </FormField>
          </>
        )}
      </div>
    </Modal>
  );
}

export function AppointmentDetailModal({ open, appt, onClose, onChanged, t }) {
  const [busy, setBusy] = useState(false);
  const [err, setErr] = useState(null);

  useEffect(() => { if (open) setErr(null); }, [open]);

  if (!appt) return null;

  const setStatus = async (code) => {
    setBusy(true);
    setErr(null);
    try {
      await api.setScheduleAppointmentStatus(appt.id, code);
      onChanged();
    } catch (e) {
      setErr(e.message || t("forms.action_failed"));
    } finally {
      setBusy(false);
    }
  };

  const remove = async () => {
    if (!window.confirm(t("forms.delete_confirm"))) return;
    setBusy(true);
    setErr(null);
    try {
      await api.deleteScheduleAppointment(appt.id);
      onChanged();
    } catch (e) {
      setErr(e.message || t("forms.action_failed"));
    } finally {
      setBusy(false);
    }
  };

  const date = appt.starts_at.toLocaleDateString();

  return (
    <Modal
      open={open}
      onClose={onClose}
      title={t("forms.appointment_detail")}
      footer={
        <>
          <button onClick={remove} disabled={busy} className={btnDanger}>
            {t("forms.delete")}
          </button>
          <button onClick={onClose} disabled={busy} className={btnSecondary}>
            {t("forms.close")}
          </button>
        </>
      }
    >
      {err && (
        <div className="mb-3 px-3 py-2 text-xs text-red-600 bg-red-50 dark:bg-red-900/30 border border-red-100 dark:border-red-800 rounded-lg">
          {err}
        </div>
      )}

      <div className="space-y-3 mb-4">
        <div className="text-xl font-bold text-slate-900 dark:text-slate-100">
          {fmtTime(appt.starts_at)} - {fmtTime(appt.ends_at)}
          <span className="ml-2 text-xs font-medium text-slate-400">{date}</span>
        </div>

        <div className="grid grid-cols-2 gap-3 text-xs">
          <DetailRow label={t("forms.patient")} value={appt.patient_name || appt.patient_phone || "-"} />
          <DetailRow label={t("forms.phone")} value={appt.patient_phone || "-"} />
          <DetailRow label={t("forms.doctor")} value={appt.doctor_name || "-"} />
          <DetailRow label={t("forms.cabinet")} value={appt.cabinet || "-"} />
          <DetailRow label={t("forms.complaint")} value={appt.zhaloba || "-"} full />
        </div>
      </div>

      <div>
        <div className="text-[10px] font-bold uppercase tracking-wider text-slate-500 dark:text-slate-400 mb-2">
          {t("forms.status")}
        </div>
        <div className="grid grid-cols-2 sm:grid-cols-3 gap-2">
          <button onClick={() => setStatus(STATUS_CODES.confirm)} disabled={busy} className={btnGhost}>{t("forms.status_confirm")}</button>
          <button onClick={() => setStatus(STATUS_CODES.came)} disabled={busy} className={btnGhost}>{t("forms.status_came")}</button>
          <button onClick={() => setStatus(STATUS_CODES.in_process)} disabled={busy} className={btnGhost}>{t("forms.status_in_process")}</button>
          <button onClick={() => setStatus(STATUS_CODES.late)} disabled={busy} className={btnGhost}>{t("forms.status_late")}</button>
          <button onClick={() => setStatus(STATUS_CODES.cancel)} disabled={busy} className={btnDanger}>{t("forms.status_cancel")}</button>
        </div>
      </div>
    </Modal>
  );
}

function DetailRow({ label, value, full }) {
  return (
    <div className={full ? "col-span-2" : ""}>
      <div className="text-[10px] font-bold uppercase tracking-wider text-slate-500 dark:text-slate-400 mb-1">
        {label}
      </div>
      <div className="text-sm text-slate-800 dark:text-slate-100 break-words">{value}</div>
    </div>
  );
}
