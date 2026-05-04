// Mock API — used when VITE_MOCK=true
// Simulates all backend endpoints with realistic data.

const delay = (ms = 300) => new Promise((r) => setTimeout(r, ms));

const CLINIC_ID = "11111111-1111-1111-1111-111111111111";
const USER_ID   = "22222222-2222-2222-2222-222222222222";

// ── shared fixtures ────────────────────────────────────────────────────────────

const DOCTORS = [
  { id: "d1", clinic_id: CLINIC_ID, name: "Айгерим Касымова",  specialty: "therapist",    active: true },
  { id: "d2", clinic_id: CLINIC_ID, name: "Бауыржан Ахметов",  specialty: "surgeon",      active: true },
  { id: "d3", clinic_id: CLINIC_ID, name: "Динара Серикова",   specialty: "orthodontist", active: true },
  { id: "d4", clinic_id: CLINIC_ID, name: "Марат Жанболатов",  specialty: "therapist",    active: false },
];

const PATIENTS = [
  { id: "p1", clinic_id: CLINIC_ID, phone: "+77011234567", name: "Алибек Досанов",    language: "ru" },
  { id: "p2", clinic_id: CLINIC_ID, phone: "+77029876543", name: "Жазира Нурланова",  language: "kk" },
  { id: "p3", clinic_id: CLINIC_ID, phone: "+77055551234", name: "Серик Бектуров",    language: "ru" },
  { id: "p4", clinic_id: CLINIC_ID, phone: "+77031112233", name: "Акмарал Сейткали",  language: "ru" },
  { id: "p5", clinic_id: CLINIC_ID, phone: "+77077778899", name: null,                language: "ru" },
];

const today = new Date();
const iso = (d) => d.toISOString();
const addH = (d, h) => new Date(d.getTime() + h * 3600_000);
const addD = (d, n) => new Date(d.getFullYear(), d.getMonth(), d.getDate() + n);

const APPTS = [
  { id: "a1", clinic_id: CLINIC_ID, patient_id: "p1", doctor_id: "d1", starts_at: iso(addH(today, 1)),  ends_at: iso(addH(today, 1.5)),  status: "scheduled", service: "Осмотр" },
  { id: "a2", clinic_id: CLINIC_ID, patient_id: "p2", doctor_id: "d2", starts_at: iso(addH(today, 2)),  ends_at: iso(addH(today, 2.5)),  status: "confirmed", service: "Удаление зуба" },
  { id: "a3", clinic_id: CLINIC_ID, patient_id: "p3", doctor_id: "d1", starts_at: iso(addH(today, 3)),  ends_at: iso(addH(today, 3.5)),  status: "completed", service: "Чистка" },
  { id: "a4", clinic_id: CLINIC_ID, patient_id: "p4", doctor_id: "d3", starts_at: iso(addH(today, 4)),  ends_at: iso(addH(today, 4.5)),  status: "cancelled", service: "Брекеты" },
  { id: "a5", clinic_id: CLINIC_ID, patient_id: "p1", doctor_id: "d2", starts_at: iso(addD(today, -1)), ends_at: iso(addH(addD(today, -1), 0.5)), status: "completed", service: "Пломба" },
  { id: "a6", clinic_id: CLINIC_ID, patient_id: "p2", doctor_id: "d1", starts_at: iso(addD(today, 1)),  ends_at: iso(addH(addD(today, 1), 0.5)), status: "scheduled", service: "Осмотр" },
];

const CONV_ID_1 = "c1";
const CONV_ID_2 = "c2";
const CONV_ID_3 = "c3";

const CHATS = [
  {
    conversation: { id: CONV_ID_1, status: "handoff",  last_message_at: iso(addH(today, -0.5)) },
    patient:      PATIENTS[0],
    last_message: { body: "Можно перенести на завтра?" },
  },
  {
    conversation: { id: CONV_ID_2, status: "active",   last_message_at: iso(addH(today, -1)) },
    patient:      PATIENTS[1],
    last_message: { body: "Хорошо, спасибо!" },
  },
  {
    conversation: { id: CONV_ID_3, status: "active",   last_message_at: iso(addH(today, -2)) },
    patient:      PATIENTS[2],
    last_message: { body: "Записали на 10:00" },
  },
];

const MESSAGES = {
  [CONV_ID_1]: [
    { id: "m1", conversation_id: CONV_ID_1, direction: "inbound",  sender: "patient",  body: "Здравствуйте, я хочу записаться к врачу",         created_at: iso(addH(today, -2)) },
    { id: "m2", conversation_id: CONV_ID_1, direction: "outbound", sender: "bot",      body: "Здравствуйте! Выберите удобное время: 10:00 или 14:00?", created_at: iso(addH(today, -1.9)) },
    { id: "m3", conversation_id: CONV_ID_1, direction: "inbound",  sender: "patient",  body: "Лучше 10:00",                                      created_at: iso(addH(today, -1.8)) },
    { id: "m4", conversation_id: CONV_ID_1, direction: "outbound", sender: "bot",      body: "Записал вас на 10:00 к Айгерим Касымовой. Подтверждаете?", created_at: iso(addH(today, -1.7)) },
    { id: "m5", conversation_id: CONV_ID_1, direction: "inbound",  sender: "patient",  body: "Можно перенести на завтра?",                       created_at: iso(addH(today, -0.5)) },
  ],
  [CONV_ID_2]: [
    { id: "m6", conversation_id: CONV_ID_2, direction: "inbound",  sender: "patient",  body: "Добрый день",                                      created_at: iso(addH(today, -2)) },
    { id: "m7", conversation_id: CONV_ID_2, direction: "outbound", sender: "bot",      body: "Добрый день! Чем могу помочь?",                    created_at: iso(addH(today, -1.95)) },
    { id: "m8", conversation_id: CONV_ID_2, direction: "inbound",  sender: "patient",  body: "Хорошо, спасибо!",                                 created_at: iso(addH(today, -1)) },
  ],
  [CONV_ID_3]: [
    { id: "m9",  conversation_id: CONV_ID_3, direction: "outbound", sender: "bot",     body: "Здравствуйте! Вы записаны на завтра в 10:00",      created_at: iso(addH(today, -3)) },
    { id: "m10", conversation_id: CONV_ID_3, direction: "inbound",  sender: "patient", body: "Записали на 10:00",                                created_at: iso(addH(today, -2)) },
  ],
};

// ── mock API object ─────────────────────────────────────────────────────────────

let _msgCounter = 100;

export const mockApi = {
  login: async () => {
    await delay(400);
    return {
      token: "mock-jwt-token",
      user: { id: USER_ID, email: "admin@demo.kz", name: "Демо Админ", role: "owner", clinic_id: CLINIC_ID },
    };
  },

  me: async () => {
    await delay(100);
    return { user_id: USER_ID, clinic_id: CLINIC_ID, role: "owner" };
  },

  // chats
  chats: async () => { await delay(200); return CHATS; },
  messages: async (id) => { await delay(150); return MESSAGES[id] ?? []; },
  send: async (id, body) => {
    await delay(200);
    const msg = {
      id: `m${++_msgCounter}`,
      conversation_id: id,
      direction: "outbound",
      sender: "operator",
      body,
      created_at: new Date().toISOString(),
    };
    (MESSAGES[id] = MESSAGES[id] ?? []).push(msg);
    return msg;
  },
  release: async () => { await delay(100); },
  subscribe: () => () => {},  // no-op, returns unsubscribe fn

  // calendar
  scheduleDoctors: async () => {
    await delay(200);
    return DOCTORS.filter((d) => d.active).map((doc) => ({
      doctor: doc,
      appointments: APPTS.filter((a) => a.doctor_id === doc.id).map((a) => ({
        ...a,
        patient: PATIENTS.find((p) => p.id === a.patient_id),
      })),
    }));
  },

  scheduleAppointment: async (id) => {
    await delay(100);
    const a = APPTS.find((x) => x.id === id);
    return a ? { ...a, patient: PATIENTS.find((p) => p.id === a.patient_id) } : null;
  },
  updateScheduleAppointment: async () => { await delay(200); },
  deleteScheduleAppointment: async () => { await delay(200); },
  schedulePatient: async (id) => { await delay(100); return PATIENTS.find((p) => p.id === id) ?? null; },
  sendAppointmentRequest: async () => { await delay(300); },
  createSchedulePatient: async (body) => {
    await delay(300);
    return { id: `p${Date.now()}`, clinic_id: CLINIC_ID, ...body };
  },
  createScheduleAppointment: async (body) => {
    await delay(300);
    return { id: `a${Date.now()}`, clinic_id: CLINIC_ID, status: "scheduled", ...body };
  },
  setScheduleAppointmentStatus: async () => { await delay(200); },

  // history
  history: async () => {
    await delay(200);
    return APPTS.map((a) => ({
      ...a,
      patient: PATIENTS.find((p) => p.id === a.patient_id),
      doctor:  DOCTORS.find((d)  => d.id === a.doctor_id),
    }));
  },

  // doctors
  doctors: async () => { await delay(150); return DOCTORS; },
  doctor:  async (id) => { await delay(100); return DOCTORS.find((d) => d.id === id) ?? null; },

  // patients
  patients: async () => { await delay(150); return PATIENTS; },
  patient:  async (id) => { await delay(100); return PATIENTS.find((p) => p.id === id) ?? null; },
  patientAppointments: async (id) => {
    await delay(150);
    return APPTS.filter((a) => a.patient_id === id);
  },

  // clinic
  clinic: async () => {
    await delay(100);
    return { id: CLINIC_ID, name: "Demo Dental Almaty", timezone: "Asia/Almaty", scheduler_type: "mock", slot_duration_min: 30 };
  },

  // stats
  stats: async () => { await delay(200); return { total: 5, completed: 3, cancelled: 1 }; },

  // dashboard
  dashboardToday: async () => {
    await delay(200);
    return {
      date: today.toISOString().slice(0, 10),
      total: 4,
      new_patients_today: 1,
      counts: { scheduled: 1, confirmed: 1, in_process: 0, came: 0, late: 0, cancelled: 1, completed: 1 },
      upcoming: [
        { id: "a1", start: iso(addH(today, 1)), doctor_id: "d1", cabinet: "1", status: 0, is_first: true },
        { id: "a2", start: iso(addH(today, 2)), doctor_id: "d2", cabinet: "2", status: 1, is_first: false },
      ],
    };
  },
  dashboardStats: async () => {
    await delay(250);
    return {
      completion_rate: 0.75,
      new_patient_rate: 0.3,
      funnel: { booked: 40, confirmed: 32, came: 28, completed: 25 },
      by_doctor: [
        { doctor_id: "d1", total: 18, completed: 14, cancelled: 2, new_patients: 5 },
        { doctor_id: "d2", total: 12, completed: 9,  cancelled: 1, new_patients: 3 },
        { doctor_id: "d3", total: 10, completed: 7,  cancelled: 2, new_patients: 2 },
      ],
      heatmap: {
        Mon: { 9: 3, 10: 5, 11: 4, 14: 6, 15: 5 },
        Tue: { 9: 2, 10: 7, 11: 6, 14: 4, 15: 3 },
        Wed: { 10: 4, 11: 3, 14: 5, 16: 2 },
        Thu: { 9: 5, 10: 6, 11: 5, 14: 7, 15: 4 },
        Fri: { 9: 3, 10: 4, 14: 3 },
        Sat: { 10: 6, 11: 5, 12: 4 },
        Sun: {},
      },
    };
  },
  dashboardRevenue: async () => {
    await delay(250);
    const trend = Array.from({ length: 14 }, (_, i) => ({
      date: addD(today, i - 13).toISOString().slice(0, 10),
      income:  Math.round(50000 + Math.random() * 80000),
      expense: Math.round(10000 + Math.random() * 20000),
    }));
    const total_income  = trend.reduce((s, t) => s + t.income, 0);
    const total_expense = trend.reduce((s, t) => s + t.expense, 0);
    return {
      total_income,
      total_expense,
      net: total_income - total_expense,
      trend,
      by_type: [
        { name: "Наличные",     amount: Math.round(total_income * 0.4) },
        { name: "Kaspi",        amount: Math.round(total_income * 0.35) },
        { name: "Карта",        amount: Math.round(total_income * 0.15) },
        { name: "Перевод",      amount: Math.round(total_income * 0.1) },
      ],
    };
  },
};
