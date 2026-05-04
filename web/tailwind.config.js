/** @type {import('tailwindcss').Config} */
export default {
  darkMode: "class",
  content: ["./index.html", "./src/**/*.{js,jsx}"],
  theme: {
    extend: {
      colors: {
        brand: {
          50:  "#eefaf3",
          100: "#d6f2e1",
          500: "#10b981",
          600: "#059669",
          700: "#047857",
        },
        accent:  { DEFAULT: "#c9ff58", strong: "#b8f82d" },
        ink:     "#111216",
        muted:   "#666d78",
        line:    "rgba(17,18,22,0.14)",
        "line-s":"rgba(17,18,22,0.24)",
      },
      fontFamily: {
        sans:      ["Inter","system-ui","-apple-system","Segoe UI","Roboto","sans-serif"],
        manrope:   ['"Manrope"', '"Segoe UI Variable"', "sans-serif"],
        unbounded: ['"Unbounded"', "sans-serif"],
      },
      boxShadow: {
        panel:      "0 28px 70px rgba(35,44,70,0.14)",
        btn:        "8px 8px 0 rgba(17,18,22,0.08)",
        "btn-hover":"12px 12px 0 rgba(17,18,22,0.10)",
      },
      borderRadius: {
        xl2: "36px",
        xl3: "26px",
        xl4: "18px",
      },
    },
  },
  plugins: [],
};
