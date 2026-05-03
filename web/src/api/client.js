const BASE = import.meta.env.VITE_API_URL || "/api";
const WA_WEB_BASE = import.meta.env.VITE_WA_WEB_URL || "/waweb/api";

function getToken() {
  return localStorage.getItem("dd_token") || "";
}

async function request(path, { method = "GET", body } = {}) {
  const headers = { "Content-Type": "application/json" };
  const token = getToken();
  if (token) headers.Authorization = `Bearer ${token}`;

  const res = await fetch(BASE + path, {
    method,
    headers,
    body: body ? JSON.stringify(body) : undefined,
  });

  if (res.status === 401) {
    localStorage.removeItem("dd_token");
    localStorage.removeItem("dd_user");
    window.location.assign("/login");
    throw new Error("unauthorized");
  }
  if (!res.ok) {
    const text = await res.text();
    throw new Error(`HTTP ${res.status}: ${text}`);
  }
  if (res.status === 204) return null;
  return res.json();
}

async function waWebRequest(path, { method = "GET", body } = {}) {
  const res = await fetch(WA_WEB_BASE + path, {
    method,
    headers: { "Content-Type": "application/json" },
    body: body ? JSON.stringify(body) : undefined,
  });

  const data = await res.json().catch(() => ({}));
  if (!res.ok) {
    throw new Error(data.error || `WhatsApp Web HTTP ${res.status}`);
  }
  return data;
}

export const api = {
  // auth
  login: (email, password) =>
    request("/auth/login", { method: "POST", body: { email, password } }),
  me: () => request("/auth/me"),

  // chats
  chats: () => request("/chats"),
  messages: (id) => request(`/chats/${id}/messages`),
  send: (id, body) => request(`/chats/${id}/send`, { method: "POST", body: { body } }),
  release: (id) => request(`/chats/${id}/release`, { method: "POST" }),

  // calendar
  calendar: (from, to) =>
    request(`/calendar?from=${encodeURIComponent(from)}&to=${encodeURIComponent(to)}`),
  scheduleDoctors: () => request("/schedule/doctors"),

  // history = record table
  history: (from, to) =>
      request(`/history?from=${encodeURIComponent(from)}&to=${encodeURIComponent(to)}`),

  // schedule — MacDent live data
  scheduleDoctors: (from, to) => {
    const q = from && to
      ? `?from=${encodeURIComponent(from)}&to=${encodeURIComponent(to)}`
      : "";
    return request(`/schedule/doctors${q}`);
  },
  scheduleAppointment: (id) => request(`/schedule/appointments/${id}`),
  updateScheduleAppointment: (id, body) =>
    request(`/schedule/appointments/${id}`, { method: "PUT", body }),
  deleteScheduleAppointment: (id) =>
    request(`/schedule/appointments/${id}`, { method: "DELETE" }),
  schedulePatient: (id) => request(`/schedule/patients/${id}`),
  sendAppointmentRequest: (body) =>
    request("/schedule/appointment-requests", { method: "POST", body }),
  createSchedulePatient: (body) =>
    request("/schedule/patients", { method: "POST", body }),
  createScheduleAppointment: (body) =>
    request("/schedule/appointments", { method: "POST", body }),
  setScheduleAppointmentStatus: (id, status) =>
    request(`/schedule/appointments/${id}/status`, { method: "PUT", body: { status } }),

  // doctors
  doctors: () => request("/doctors"),
  doctor: (id) => request(`/doctors/${id}`),

  // patients
  patients: () => request("/patients"),
  patient: (id) => request(`/patients/${id}`),
  patientAppointments: (id) => request(`/patients/${id}/appointments`),

  // clinic
  clinic: () => request("/clinic"),

  // stats
  stats: () => request("/stats"),

  // dashboard analytics
  dashboardToday: () => request("/dashboard/today"),
  dashboardStats: (from, to) =>
    request(`/dashboard/stats?from=${encodeURIComponent(from)}&to=${encodeURIComponent(to)}`),
  dashboardRevenue: (from, to) =>
    request(`/dashboard/revenue?from=${encodeURIComponent(from)}&to=${encodeURIComponent(to)}`),

  // SSE
  subscribe: (onEvent) => {
    const token = getToken();
    const url = `${BASE}/events?token=${encodeURIComponent(token)}`;
    const es = new EventSource(url);
    es.addEventListener("message", (e) => onEvent("message", JSON.parse(e.data)));
    es.addEventListener("appointment", (e) => onEvent("appointment", JSON.parse(e.data)));
    es.onerror = () => {
      // браузер сам переподключится
    };
    return () => es.close();
  },
};

export const waWeb = {
  status: () => waWebRequest("/status"),
  qr: () => waWebRequest("/qr"),
  chats: (limit = 80) => waWebRequest(`/chats?limit=${limit}`),
  messages: (chatId, limit = 80) =>
    waWebRequest(`/chats/${encodeURIComponent(chatId)}/messages?limit=${limit}`),
  send: (chatId, body) =>
    waWebRequest(`/chats/${encodeURIComponent(chatId)}/send`, {
      method: "POST",
      body: { body },
    }),
  logout: () => waWebRequest("/logout", { method: "POST" }),
  resetSession: () => waWebRequest("/session/reset", { method: "POST" }),
  eventsUrl: () => `${WA_WEB_BASE}/events`,
};

export function saveAuth({ token, user }) {
  localStorage.setItem("dd_token", token);
  localStorage.setItem("dd_user", JSON.stringify(user));
}
export function getUser() {
  try { return JSON.parse(localStorage.getItem("dd_user") || "null"); }
  catch { return null; }
}
export function logout() {
  localStorage.removeItem("dd_token");
  localStorage.removeItem("dd_user");
}
