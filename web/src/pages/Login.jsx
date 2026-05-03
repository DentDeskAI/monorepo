import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { api, saveAuth } from "../api/client";
import { useTranslation } from "../hooks/useTranslation";

// i18n Keys
const i18nKeys = {
  login: {
    title: "login.title",
    subtitle: "login.subtitle",
    email: "login.email_label",
    password: "login.password_label",
    submit: "login.submit",
    loading: "login.loading",
    error: "login.error",
    demo_hint: "login.demo_hint",
    demo_email: "login.demo_email",
    demo_pass: "login.demo_pass",
  },
};

export default function Login() {
  const { t } = useTranslation();
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
      nav("/app");
    } catch (e) {
      setErr(t(i18nKeys.login.error));
    } finally {
      setLoading(false);
    }
  };

  return (
      <div className="min-h-screen flex items-center justify-center bg-[#F7F8FA] dark:bg-slate-900">
        <div className="w-full max-w-md bg-white dark:bg-slate-800 rounded-xl shadow-sm border border-slate-200 dark:border-slate-700 p-8">

          {/* Header */}
          <div className="flex items-center gap-3 mb-8">
            <div className="w-10 h-10 rounded-lg bg-blue-600 text-white grid place-items-center font-bold text-lg shadow-sm">
              D
            </div>
            <div>
              <div className="font-bold text-slate-900 dark:text-slate-100 text-lg">
                {t(i18nKeys.login.title)}
              </div>
              <div className="text-xs text-slate-500 dark:text-slate-400 font-medium">
                {t(i18nKeys.login.subtitle)}
              </div>
            </div>
          </div>

          <form onSubmit={submit} className="space-y-4">
            {/* Email Input */}
            <div>
              <label className="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1.5">
                {t(i18nKeys.login.email)}
              </label>
              <input
                  type="email"
                  className="w-full px-3 py-2.5 bg-slate-50 dark:bg-slate-900 border border-slate-200 dark:border-slate-700 rounded-lg text-sm text-slate-900 dark:text-slate-100 placeholder-slate-400 focus:outline-none focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 transition-all"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  autoFocus
                  required
              />
            </div>

            {/* Password Input */}
            <div>
              <label className="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1.5">
                {t(i18nKeys.login.password)}
              </label>
              <input
                  type="password"
                  className="w-full px-3 py-2.5 bg-slate-50 dark:bg-slate-900 border border-slate-200 dark:border-slate-700 rounded-lg text-sm text-slate-900 dark:text-slate-100 placeholder-slate-400 focus:outline-none focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 transition-all"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  required
              />
            </div>

            {/* Error Message */}
            {err && (
                <div className="p-3 bg-red-50 border border-red-100 rounded-lg text-sm text-red-600 flex items-center gap-2">
                  <span>⚠️</span> {err}
                </div>
            )}

            {/* Submit Button */}
            <button
                type="submit"
                className="w-full py-2.5 bg-blue-600 hover:bg-blue-700 text-white text-sm font-medium rounded-lg transition-colors shadow-sm disabled:opacity-70 disabled:cursor-not-allowed"
                disabled={loading}
            >
              {loading ? t(i18nKeys.login.loading) : t(i18nKeys.login.submit)}
            </button>

            {/* Demo Hint */}
            <div className="pt-4 text-center">
              <p className="text-[11px] text-slate-400 font-medium">
                {t(i18nKeys.login.demo_hint)}
              </p>
              <div className="mt-1 text-[11px] text-slate-500 dark:text-slate-400 font-mono bg-slate-50 dark:bg-slate-900 inline-block px-2 py-1 rounded border border-slate-100 dark:border-slate-700">
                {t(i18nKeys.login.demo_email)} / {t(i18nKeys.login.demo_pass)}
              </div>
            </div>
          </form>
        </div>
      </div>
  );
}
