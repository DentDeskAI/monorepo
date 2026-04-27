import { useEffect, useState } from "react";

const DEFAULT_THEME = "light";
const THEME_STORAGE_KEY = "crm_theme";
const THEME_EVENT = "crm_theme_change";

function applyTheme(theme) {
  if (theme === "dark") {
    document.documentElement.classList.add("dark");
  } else {
    document.documentElement.classList.remove("dark");
  }
}

export function useTheme() {
  const [theme, setThemeState] = useState(() => {
    const saved = localStorage.getItem(THEME_STORAGE_KEY);
    return saved === "dark" ? "dark" : DEFAULT_THEME;
  });

  const setTheme = (nextTheme) => {
    const safeTheme = nextTheme === "dark" ? "dark" : "light";
    setThemeState(safeTheme);
    localStorage.setItem(THEME_STORAGE_KEY, safeTheme);
    applyTheme(safeTheme);
    window.dispatchEvent(new CustomEvent(THEME_EVENT, { detail: safeTheme }));
  };

  const toggleTheme = () => {
    setTheme(theme === "dark" ? "light" : "dark");
  };

  useEffect(() => {
    const savedTheme = localStorage.getItem(THEME_STORAGE_KEY);
    const initial = savedTheme === "dark" ? "dark" : "light";
    setThemeState(initial);
    applyTheme(initial);

    const onThemeChange = (event) => {
      const next = event?.detail;
      if (next === "dark" || next === "light") {
        setThemeState(next);
        applyTheme(next);
      }
    };

    const onStorage = (event) => {
      if (event.key !== THEME_STORAGE_KEY || !event.newValue) return;
      const next = event.newValue === "dark" ? "dark" : "light";
      setThemeState(next);
      applyTheme(next);
    };

    window.addEventListener(THEME_EVENT, onThemeChange);
    window.addEventListener("storage", onStorage);
    return () => {
      window.removeEventListener(THEME_EVENT, onThemeChange);
      window.removeEventListener("storage", onStorage);
    };
  }, []);

  return { theme, setTheme, toggleTheme };
}