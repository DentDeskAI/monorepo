import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { api, saveAuth } from "../api/client";

export default function Login() {
  const [email, setEmail] = useState("admin@demo.kz");
  const [password, setPassword] = useState("demo1234");
  const [err, setErr] = useState("");
  const [loading, setLoading] = useState(false);
  const nav = useNavigate();

  const submit = async (e) => {
    e.preventDefault();
    setErr("");
    setLoading(true);
    try {
      const { token, user } = await api.login(email, password);
      saveAuth({ token, user });
      nav("/");
    } catch (e) {
      setErr("Неверный email или пароль");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen grid place-items-center bg-gradient-to-br from-slate-50 to-brand-50">
      <div className="card w-full max-w-sm p-6">
        <div className="flex items-center gap-2 mb-6">
          <div className="w-10 h-10 rounded-xl bg-brand-600 text-white grid place-items-center font-bold text-lg">D</div>
          <div>
            <div className="font-semibold text-slate-900">DentDesk</div>
            <div className="text-xs text-slate-500">Mini-CRM для стоматологий</div>
          </div>
        </div>

        <form onSubmit={submit} className="space-y-3">
          <div>
            <label className="block text-sm font-medium text-slate-700 mb-1">Email</label>
            <input
              type="email"
              className="input"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              autoFocus
              required
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-slate-700 mb-1">Пароль</label>
            <input
              type="password"
              className="input"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              required
            />
          </div>
          {err && <div className="text-sm text-red-600">{err}</div>}
          <button type="submit" className="btn btn-primary w-full" disabled={loading}>
            {loading ? "Входим..." : "Войти"}
          </button>
          <div className="text-xs text-slate-400 text-center pt-2">
            Demo: admin@demo.kz / demo1234
          </div>
        </form>
      </div>
    </div>
  );
}
