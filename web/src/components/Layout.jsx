import { NavLink, Outlet, useNavigate } from "react-router-dom";
import { getUser, logout } from "../api/client";
import { useTranslation } from "../hooks/useTranslation";
import { useTheme } from "../hooks/useTheme";

// i18n Keys for Layout
const i18nKeys = {
  layout: {
    brand: "DentDesk",
    online: "layout.online",
    language: "layout.language",
    themeDark: "layout.theme_dark",
    themeLight: "layout.theme_light",
    logout: "layout.logout",
    nav: {
      dashboard: "layout.nav.dashboard",
      chats: "layout.nav.chats",
      calendar: "layout.nav.calendar",
      patients: "layout.nav.patients",
      doctors: "layout.nav.doctors",
      records: "layout.nav.records",
    },
  },
};

export default function Layout() {
  const { t, lang, setLanguage } = useTranslation();
  const { theme, toggleTheme } = useTheme();
  const user = getUser();
  const nav = useNavigate();

  const items = [
    { to: "/", label: t(i18nKeys.layout.nav.dashboard), icon: "📊" },
    { to: "/chats", label: t(i18nKeys.layout.nav.chats), icon: "💬" },
    { to: "/calendar", label: t(i18nKeys.layout.nav.calendar), icon: "📅" },
    { to: "/patients", label: t(i18nKeys.layout.nav.patients), icon: "👤" },
    { to: "/whatsapp-web", label: "WhatsApp Web", icon: "WA" },
    { to: "/doctors", label: t(i18nKeys.layout.nav.doctors), icon: "👨‍⚕️" },
    { to: "/records", label: t(i18nKeys.layout.nav.records), icon: "📋" },
  ];

  const onLogout = () => {
    logout();
    nav("/login");
  };

  // Helper to determine text colors based on theme
  const textPrimary = theme === "dark" ? "text-slate-100" : "text-slate-900";
  const textSecondary = theme === "dark" ? "text-slate-400" : "text-slate-500";
  const textTertiary = theme === "dark" ? "text-slate-500" : "text-slate-400";
  const bgPrimary = theme === "dark" ? "bg-slate-950" : "bg-white";
  const bgSecondary = theme === "dark" ? "bg-slate-900/50" : "bg-slate-50/50";
  const borderPrimary = theme === "dark" ? "border-slate-800" : "border-slate-200";
  const borderSecondary = theme === "dark" ? "border-slate-700" : "border-slate-200";

  return (
      <div className={`flex h-screen ${theme === "dark" ? "bg-slate-900" : "bg-[#F7F8FA]"}`}>
        {/* SIDEBAR */}
        <aside className={`w-64 border-r flex flex-col shrink-0 shadow-sm z-10 ${bgPrimary} ${borderPrimary}`}>
          {/* Brand Header */}
          <div className={`h-16 flex items-center px-5 border-b ${borderPrimary}`}>
            <div className="w-8 h-8 rounded-lg bg-blue-600 text-white grid place-items-center font-bold text-sm shadow-sm">
              D
            </div>
            <div className="ml-3">
              <div className={`font-bold text-sm ${textPrimary}`}>
                {t(i18nKeys.layout.brand)}
              </div>
              <div className={`text-[10px] font-medium uppercase tracking-wider ${textTertiary}`}>
                {t(i18nKeys.layout.online)}
              </div>
            </div>
          </div>

          {/* Navigation */}
          <nav className="flex-1 p-3 space-y-1 overflow-y-auto">
            {items.map((it) => (
                <NavLink
                    key={it.to}
                    to={it.to}
                    end={it.to === "/"}
                    className={({ isActive }) =>
                        `flex items-center gap-3 px-3 py-2.5 rounded-lg text-sm font-medium transition-all duration-200 ${
                            isActive
                                ? "bg-blue-50 text-blue-700 shadow-sm"
                                : theme === "dark"
                                    ? "text-slate-300 hover:bg-slate-800 hover:text-slate-100"
                                    : "text-slate-600 hover:bg-slate-50 hover:text-slate-900"
                        }`
                    }
                >
                  <span className="text-lg leading-none">{it.icon}</span>
                  <span>{it.label}</span>
                </NavLink>
            ))}
          </nav>

          {/* User Profile & Controls */}
          <div className={`p-3 border-t ${borderPrimary} ${bgSecondary}`}>

            {/* Language Selector */}
            <div className="px-2 py-2 mb-2">
              <label className={`block text-[10px] mb-1 font-medium uppercase tracking-wider ${textTertiary}`}>
                {t(i18nKeys.layout.language)}
              </label>
              <select
                  value={lang}
                  onChange={(e) => setLanguage(e.target.value)}
                  className={`w-full px-2 py-1.5 rounded-lg text-xs border ${
                      theme === "dark"
                          ? "bg-slate-900 border-slate-700 text-slate-200"
                          : "bg-white border-slate-200 text-slate-700"
                  }`}
              >
                <option value="ru">Русский</option>
                <option value="en">English</option>
                <option value="kz">Қазақша</option>
              </select>
            </div>

            {/* Theme Toggle */}
            <button
                onClick={toggleTheme}
                className={`mt-2 w-full px-3 py-2 text-xs font-medium rounded-lg transition-colors ${
                    theme === "dark"
                        ? "bg-slate-800 text-slate-100 border border-slate-700 hover:bg-slate-700"
                        : "bg-white text-slate-700 border border-slate-200 hover:bg-slate-100"
                }`}
            >
              {theme === "dark" ? t(i18nKeys.layout.themeLight) : t(i18nKeys.layout.themeDark)}
            </button>
          </div>

          {/* User Info & Logout */}
          <div className={`p-3 border-t ${borderPrimary} ${bgPrimary}`}>
            <div className="px-2 py-2 mb-2">
              <div className={`font-medium text-sm truncate ${textPrimary}`}>
                {user?.name || "User"}
              </div>
              <div className={`text-xs truncate ${textSecondary}`}>
                {user?.email || "user@demo.kz"}
              </div>
            </div>
            <button
                onClick={onLogout}
                className={`w-full px-3 py-2 text-xs font-medium rounded-lg transition-colors ${
                    theme === "dark"
                        ? "text-slate-200 bg-slate-900 border border-slate-700 hover:bg-slate-800"
                        : "text-slate-600 bg-white border border-slate-200 hover:bg-slate-100 hover:text-slate-900"
                }`}
            >
              {t(i18nKeys.layout.logout)}
            </button>
          </div>
        </aside>

        {/* MAIN CONTENT AREA */}
        <main className={`flex-1 overflow-hidden relative ${theme === "dark" ? "bg-slate-900" : ""}`}>
          <div className="h-full w-full">
            <Outlet />
          </div>
        </main>
      </div>
  );
}