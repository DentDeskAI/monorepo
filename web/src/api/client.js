const BASE = import.meta.env.VITE_API_URL || "/api";

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

  // doctors
  doctors: () => request("/doctors"),
  doctor: (id) => request(`/doctors/${id}`),

  // patients
  patients: () => request("/patients"),
  patient: (id) => request(`/patients/${id}`),

  // clinic
  clinic: () => request("/clinic"),

  // stats
  stats: () => request("/stats"),

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
