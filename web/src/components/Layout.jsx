import { NavLink, Outlet, useNavigate } from "react-router-dom";
import { getUser, logout } from "../api/client";

const items = [
  { to: "/",          label: "Главная",   icon: "📊" },
  { to: "/chats",     label: "Чаты",      icon: "💬" },
  { to: "/calendar",  label: "Календарь", icon: "📅" },
  { to: "/patients",  label: "Пациенты",  icon: "👤" },
];

export default function Layout() {
  const user = getUser();
  const nav = useNavigate();
  const onLogout = () => {
    logout();
    nav("/login");
  };

  return (
    <div className="flex h-screen">
      <aside className="w-60 shrink-0 bg-white border-r border-slate-200 flex flex-col">
        <div className="p-4 border-b border-slate-200">
          <div className="flex items-center gap-2">
            <div className="w-8 h-8 rounded-lg bg-brand-600 text-white grid place-items-center font-bold">D</div>
            <div>
              <div className="font-semibold text-slate-900">DentDesk</div>
              <div className="text-xs text-slate-500">Айгуль online</div>
            </div>
          </div>
        </div>

        <nav className="flex-1 p-3 space-y-1">
          {items.map((it) => (
            <NavLink
              key={it.to}
              to={it.to}
              end={it.to === "/"}
              className={({ isActive }) =>
                `flex items-center gap-3 px-3 py-2 rounded-lg text-sm font-medium transition ${
                  isActive
                    ? "bg-brand-50 text-brand-700"
                    : "text-slate-600 hover:bg-slate-50"
                }`
              }
            >
              <span>{it.icon}</span>
              <span>{it.label}</span>
            </NavLink>
          ))}
        </nav>

        <div className="p-3 border-t border-slate-200">
          <div className="px-3 py-2 text-sm">
            <div className="font-medium text-slate-900 truncate">{user?.name}</div>
            <div className="text-xs text-slate-500 truncate">{user?.email}</div>
          </div>
          <button onClick={onLogout} className="btn btn-secondary w-full mt-2">
            Выйти
          </button>
        </div>
      </aside>

      <main className="flex-1 overflow-hidden">
        <Outlet />
      </main>
    </div>
  );
}
