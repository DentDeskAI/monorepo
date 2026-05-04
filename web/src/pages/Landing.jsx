import { useState, useEffect, useRef, useCallback } from "react";
import { Link } from "react-router-dom";
import T from "../../landing-main/translations.json";

const GLASS = "bg-white/[0.82] backdrop-blur-lg border border-black/[0.14]";
const PSHADOW = { boxShadow: "0 28px 70px rgba(35,44,70,0.14)" };
const ACCENT = "#c9ff58";
const ACCENT_STRONG = "#b8f82d";
const INK = "#111216";
const MUTED = "#666d78";

// i18n
function useLang() {
  const [lang, setLangState] = useState(() => localStorage.getItem("dentdesk-language") || "ru");
  const t = useCallback((key) => T[key]?.[lang] || T[key]?.ru || key, [lang]);
  const setLang = (l) => { localStorage.setItem("dentdesk-language", l); setLangState(l); };
  return { lang, t, setLang };
}

function Reveal({ as: Tag = "div", children, className = "", style = {}, delay = 0, ...rest }) {
  const ref = useRef(null);
  const [visible, setVisible] = useState(false);

  useEffect(() => {
    const el = ref.current;
    if (!el) return;
    if (!("IntersectionObserver" in window)) { setVisible(true); return; }
    const obs = new IntersectionObserver(
      ([e]) => { if (e.isIntersecting) { setVisible(true); obs.disconnect(); } },
      { threshold: 0.18, rootMargin: "0px 0px -40px 0px" }
    );
    obs.observe(el);
    return () => obs.disconnect();
  }, []);

  return (
    <Tag
      ref={ref}
      className={`transition-[opacity,transform] duration-700 ease-[cubic-bezier(0.16,1,0.3,1)]
        ${visible ? "opacity-100 translate-y-0" : "opacity-0 translate-y-6"} ${className}`}
      style={{ transitionDelay: `${delay}ms`, ...style }}
      {...rest}
    />
  );
}

// ── PrimaryButton ──────────────────────────────────────────────────────────────

function PrimaryBtn({ children, onClick, className = "" }) {
  const [hovered, setHovered] = useState(false);
  return (
    <button
      type="button"
      onClick={onClick}
      onMouseEnter={() => setHovered(true)}
      onMouseLeave={() => setHovered(false)}
      style={{
        background: hovered ? ACCENT_STRONG : ACCENT,
        boxShadow: hovered ? "12px 12px 0 rgba(17,18,22,0.10)" : "8px 8px 0 rgba(17,18,22,0.08)",
        transform: hovered ? "translate(-4px,-4px)" : "none",
        border: `1px solid rgba(17,18,22,0.24)`,
        transition: "transform 200ms ease, box-shadow 200ms ease, background-color 200ms ease",
      }}
      className={`px-6 py-[18px] font-extrabold leading-snug text-[${INK}] cursor-pointer ${className}`}
    >
      {children}
    </button>
  );
}

// ── Lead Form ──────────────────────────────────────────────────────────────────

function LeadForm({ t, compact = false, onSuccess }) {
  const [fields, setFields] = useState({ phone: "", name: "", clinic: "" });
  const [fb, setFb] = useState({ msg: "", ok: null });
  const set = (k) => (e) => setFields((p) => ({ ...p, [k]: e.target.value }));

  const submit = (e) => {
    e.preventDefault();
    setFb({ msg: "", ok: null });
    const { phone, name, clinic } = fields;
    if (!phone.trim() || !name.trim() || !clinic.trim()) { setFb({ msg: t("formErrorMissingFields"), ok: false }); return; }
    if (phone.replace(/\D/g, "").length < 10) { setFb({ msg: t("formErrorInvalidPhone"), ok: false }); return; }
    const p = { phone: phone.trim(), name: name.trim(), clinic: clinic.trim() };
    try {
      const prev = JSON.parse(localStorage.getItem("dentdesk-demo-leads") || "[]");
      prev.push({ ...p, submittedAt: new Date().toISOString() });
      localStorage.setItem("dentdesk-demo-leads", JSON.stringify(prev));
    } catch {}
    const msg = encodeURIComponent(`Здравствуйте!\nЗапрос с сайта DentDesk\n\nНомер WhatsApp: ${p.phone}\nИмя: ${p.name}\nКлиника: ${p.clinic}\n\nСтраница: ${window.location.href}`);
    window.open(`https://wa.me/77058106425?text=${msg}`, "_blank");
    setFb({ msg: t("formSuccess"), ok: true });
    setFields({ phone: "", name: "", clinic: "" });
    if (onSuccess) setTimeout(onSuccess, 1300);
  };

  const inputCls = "w-full px-4 py-[15px] border border-black/[0.24] bg-white/80 focus:outline-none focus:border-blue-500/50 focus:ring-[3px] focus:ring-blue-400/10 transition-[border-color,box-shadow] duration-[180ms]";

  return (
    <form onSubmit={submit} noValidate className="grid gap-[18px]">
      {!compact && <p className="text-[0.86rem] font-extrabold uppercase mb-0" style={{ color: MUTED }}>{t("formKicker")}</p>}
      {[
        { k: "phone",  label: t("inputPhoneLabel"),  ph: t("inputPhonePlaceholder"),  type: "tel",  mode: "tel" },
        { k: "name",   label: t("inputNameLabel"),   ph: t("inputNamePlaceholder"),   type: "text" },
        { k: "clinic", label: t("inputClinicLabel"), ph: t("inputClinicPlaceholder"), type: "text" },
      ].map(({ k, label, ph, type, mode }) => (
        <label key={k} className="grid gap-[10px] text-[0.96rem] font-bold leading-[1.65]" style={{ color: MUTED }}>
          <span>{label}</span>
          <input type={type} inputMode={mode} placeholder={ph} value={fields[k]} onChange={set(k)} required className={inputCls} style={{ color: INK }} />
        </label>
      ))}
      <PrimaryBtn className="justify-self-center min-w-[200px] mt-2 max-[720px]:w-full max-[720px]:min-w-0">{t("formSubmit")}</PrimaryBtn>
      {fb.msg && <p className={`min-h-6 text-[0.95rem] font-bold ${fb.ok ? "text-green-700" : "text-red-600"}`} aria-live="polite">{fb.msg}</p>}
    </form>
  );
}

// ── Modal ──────────────────────────────────────────────────────────────────────

function Modal({ open, onClose, t }) {
  useEffect(() => { document.body.style.overflow = open ? "hidden" : ""; return () => { document.body.style.overflow = ""; }; }, [open]);
  useEffect(() => {
    const h = (e) => { if (e.key === "Escape") onClose(); };
    document.addEventListener("keydown", h);
    return () => document.removeEventListener("keydown", h);
  }, [onClose]);

  return (
    <div
      className={`fixed inset-0 grid place-items-center p-5 z-20 transition-[opacity,visibility] duration-[220ms] ${open ? "opacity-100 visible pointer-events-auto" : "opacity-0 invisible pointer-events-none"}`}
      aria-hidden={!open} role="dialog"
    >
      <div className="absolute inset-0 bg-black/30 backdrop-blur-sm" onClick={onClose} />
      <div className="relative w-full max-w-[420px] max-h-[92vh] overflow-auto p-7 rounded-[32px] bg-white/95 border border-black/[0.14] backdrop-blur-xl" style={PSHADOW}>
        <button onClick={onClose} aria-label={t("modalCloseLabel")}
          className="absolute right-[18px] top-[14px] w-[42px] h-[42px] rounded-full border-0 bg-transparent text-[1.8rem] leading-none cursor-pointer transition-colors hover:opacity-60"
          style={{ color: MUTED }}>
          &times;
        </button>
        <p className="text-[0.86rem] font-extrabold uppercase mb-4" style={{ color: MUTED }}>{t("modalKicker")}</p>
        <h2 className="m-0 mb-[22px] font-unbounded font-bold text-[2.2rem] leading-[1.04] max-w-[14ch] text-balance" style={{ color: INK }}>
          {t("modalTitle")}
        </h2>
        <LeadForm t={t} compact onSuccess={onClose} />
      </div>
    </div>
  );
}

// ── Kicker ─────────────────────────────────────────────────────────────────────

const Kicker = ({ children }) => (
  <p className="m-0 mb-4 text-[0.86rem] font-extrabold uppercase" style={{ color: MUTED }}>{children}</p>
);

const SectionH2 = ({ children, className = "" }) => (
  <h2 className={`m-0 font-unbounded font-bold text-[3.5rem] leading-[1.04] text-balance overflow-wrap-break-word max-[720px]:text-[2.45rem] max-[560px]:text-[2.05rem] ${className}`} style={{ color: INK }}>
    {children}
  </h2>
);

// ── Landing ────────────────────────────────────────────────────────────────────

export default function Landing() {
  const { lang, t, setLang } = useLang();
  const [navOpen, setNavOpen] = useState(false);
  const [modalOpen, setModalOpen] = useState(false);
  const navRef = useRef(null);
  const toggleRef = useRef(null);

  const openModal = () => { setNavOpen(false); setModalOpen(true); };

  useEffect(() => {
    const h = (e) => {
      if (!navOpen) return;
      if (navRef.current?.contains(e.target) || toggleRef.current?.contains(e.target)) return;
      setNavOpen(false);
    };
    document.addEventListener("click", h);
    return () => document.removeEventListener("click", h);
  }, [navOpen]);

  useEffect(() => {
    const h = (e) => { if (e.key === "Escape") setNavOpen(false); };
    document.addEventListener("keydown", h);
    return () => document.removeEventListener("keydown", h);
  }, []);

  const scrollTo = (id) => { setNavOpen(false); document.getElementById(id)?.scrollIntoView({ behavior: "smooth" }); };

  return (
    <div
      className="font-manrope min-h-screen relative overflow-x-hidden"
      style={{
        color: INK,
        background: "radial-gradient(circle at top left,rgba(136,170,255,0.22),transparent 28%),radial-gradient(circle at 85% 15%,rgba(201,255,88,0.18),transparent 24%),linear-gradient(180deg,#f8f9fd 0%,#f3f5fb 100%)",
      }}
      lang={lang}
    >
      {/* decorative grid */}
      <div className="fixed inset-0 pointer-events-none z-0" style={{
        backgroundImage: "linear-gradient(to right,rgba(136,148,194,0.12) 1px,transparent 1px),linear-gradient(to bottom,rgba(136,148,194,0.12) 1px,transparent 1px)",
        backgroundSize: "min(7vw,64px) min(7vw,64px)",
        WebkitMaskImage: "linear-gradient(180deg,rgba(0,0,0,0.7),transparent 92%)",
        maskImage: "linear-gradient(180deg,rgba(0,0,0,0.7),transparent 92%)",
      }} />
      {/* glow blob */}
      <div className="fixed pointer-events-none z-0 w-[42rem] h-[42rem] rounded-full" style={{ right: "-14rem", bottom: "-18rem", background: "radial-gradient(circle,rgba(160,190,255,0.28),transparent 70%)", filter: "blur(20px)" }} />

      <div className="relative z-[1] w-[min(calc(100%-28px),1220px)] mx-auto">

        {/* ── HEADER ── */}
        <header className="flex items-center justify-between gap-6 pt-[30px] pb-[18px] relative">
          <a
            href="#top"
            onClick={(e) => { e.preventDefault(); window.scrollTo({ top: 0, behavior: "smooth" }); }}
            className="inline-flex items-center gap-3 font-unbounded font-bold no-underline max-[560px]:text-[1.25rem]"
            style={{ fontSize: "1.55rem", color: INK }}
          >
            <span className="w-[15px] h-[15px] border-2 rounded-[4px] shrink-0" style={{ borderColor: INK, boxShadow: `8px 8px 0 ${ACCENT}` }} />
            DentDesk
          </a>

          <div className="flex items-center gap-3 ml-auto">
            {/* hamburger */}
            <button
              ref={toggleRef}
              type="button"
              aria-expanded={navOpen}
              aria-label={navOpen ? t("navCloseMenu") : t("navOpenMenu")}
              onClick={() => setNavOpen((v) => !v)}
              className="inline-flex flex-col justify-center gap-[5px] w-12 h-12 border border-black/[0.14] rounded-2xl bg-white/70 backdrop-blur-lg cursor-pointer"
            >
              <span className={`block w-[18px] h-[2px] mx-auto rounded-full transition-transform duration-[180ms]`} style={{ background: INK, transform: navOpen ? "translateY(7px) rotate(45deg)" : "none" }} />
              <span className={`block w-[18px] h-[2px] mx-auto rounded-full transition-opacity duration-[180ms] ${navOpen ? "opacity-0" : ""}`} style={{ background: INK }} />
              <span className={`block w-[18px] h-[2px] mx-auto rounded-full transition-transform duration-[180ms]`} style={{ background: INK, transform: navOpen ? "translateY(-7px) rotate(-45deg)" : "none" }} />
            </button>

            {/* dropdown nav */}
            <nav
              ref={navRef}
              className={`absolute top-[calc(100%+12px)] right-0 w-[min(360px,calc(100vw-28px))] grid gap-[14px] p-4 border border-black/[0.14] rounded-[28px] bg-white/95 backdrop-blur-xl z-10 transition-[opacity,visibility,transform] duration-[180ms] ${navOpen ? "opacity-100 visible pointer-events-auto translate-y-0" : "opacity-0 invisible pointer-events-none -translate-y-3"}`}
              style={PSHADOW}
            >
              <div className="grid gap-[6px]">
                {[
                  { label: t("navPlatform"), fn: () => scrollTo("proof") },
                  { label: t("navContact"),  fn: () => scrollTo("contact") },
                ].map(({ label, fn }) => (
                  <button key={label} type="button" onClick={fn}
                    className="w-full border-0 bg-transparent px-4 py-[14px] rounded-2xl text-[0.95rem] font-bold text-left cursor-pointer transition-[background-color,transform] duration-[180ms] hover:bg-black/[0.06] hover:-translate-y-px"
                    style={{ color: INK }}>
                    {label}
                  </button>
                ))}
                <Link to="/login" onClick={() => setNavOpen(false)}
                  className="block px-4 py-[14px] rounded-2xl text-[0.95rem] font-bold no-underline transition-[background-color,transform] duration-[180ms] hover:bg-black/[0.06] hover:-translate-y-px"
                  style={{ color: INK }}>
                  {t("navLogin")}
                </Link>
              </div>

              <div className="grid gap-[10px] pt-3 border-t border-black/[0.14]">
                <span className="px-[6px] text-[0.74rem] font-extrabold uppercase" style={{ color: MUTED }}>{t("languageMenuTitle")}</span>
                <ul className="grid gap-2 m-0 p-0 list-none">
                  {[["ru","Русский"],["en","English"],["kz","Қазақша"]].map(([code, label]) => (
                    <li key={code}>
                      <button type="button" onClick={() => setLang(code)}
                        className={`w-full flex items-center justify-between gap-3 border px-[14px] py-3 rounded-2xl text-[0.92rem] font-extrabold cursor-pointer transition-[background-color,border-color,transform] duration-[180ms] ${
                          lang === code
                            ? "text-white border-transparent"
                            : "border-black/[0.14] bg-white/80 hover:bg-black/[0.06] hover:-translate-y-px"
                        }`}
                        style={lang === code ? { background: INK, color: "#fff" } : { color: INK }}>
                        <span>{code.toUpperCase()}</span>
                        <small className="text-[0.82rem] font-bold" style={{ color: lang === code ? "rgba(255,255,255,0.72)" : MUTED }}>{label}</small>
                      </button>
                    </li>
                  ))}
                </ul>
              </div>
            </nav>
          </div>
        </header>

        <main>

          {/* ── HERO ── */}
          <section id="top" className="grid items-center justify-center justify-items-center gap-[72px] pt-6 pb-[72px] max-[1080px]:grid-cols-1 max-[1080px]:min-h-0 max-[1080px]:gap-10 max-[1080px]:pt-5"
            style={{ gridTemplateColumns: "minmax(0,760px) minmax(340px,410px)", minHeight: "calc(100vh - 110px)" }}>

            <div className="max-w-[760px] min-w-0 justify-self-center text-center">
              <Reveal as="p" className="text-[0.86rem] font-extrabold uppercase mb-4" style={{ color: MUTED }}>{t("heroEyebrow")}</Reveal>
              <Reveal as="h1" delay={45}
                className="m-0 font-unbounded font-bold leading-[1.04] max-w-[13ch] mx-auto text-balance hyphens-auto overflow-wrap-break-word max-[1240px]:text-[4.05rem] max-[720px]:text-[2.8rem] max-[560px]:text-[2.25rem]"
                style={{ fontSize: "4.85rem", color: INK }}>
                {t("heroTitle")}
              </Reveal>
              <Reveal as="p" delay={90} className="text-[1.06rem] leading-[1.65] max-w-[58ch] mt-6 mx-auto" style={{ color: MUTED }}>{t("heroText")}</Reveal>
              <Reveal className="flex flex-wrap items-center justify-center gap-[18px] mt-9 max-[720px]:flex-col max-[720px]:items-stretch" delay={135}>
                <PrimaryBtn onClick={openModal} className="max-[720px]:w-full">{t("heroCTA")}</PrimaryBtn>
                <a href="#proof" onClick={(e) => { e.preventDefault(); scrollTo("proof"); }}
                  className="font-extrabold inline-flex items-center gap-[10px] no-underline after:content-['→'] after:text-[1.1rem]" style={{ color: INK }}>
                  {t("heroSecondaryLink")}
                </a>
              </Reveal>
              <Reveal as="p" delay={180} className="mt-[52px] mx-auto max-w-[48ch]" style={{ color: MUTED }}>{t("heroNote")}</Reveal>
            </div>

            {/* panel */}
            <Reveal delay={225}
              className={`w-full max-w-[400px] justify-self-center p-7 rounded-[36px] max-[1080px]:max-w-[620px] max-[1080px]:w-full ${GLASS}`}
              style={PSHADOW}>
              <div className="flex justify-between items-start gap-4">
                <div>
                  <p className="m-0 mb-2 text-[0.86rem] font-extrabold uppercase" style={{ color: MUTED }}>{t("panelLabel")}</p>
                  <p className="m-0 text-[1.35rem] font-extrabold" style={{ color: INK }}>{t("panelTitle")}</p>
                </div>
                <span className="inline-flex items-center gap-2 px-[14px] py-[10px] rounded-full text-[0.84rem] font-extrabold uppercase" style={{ background: "rgba(17,18,22,0.06)", color: MUTED }}>
                  <span className="w-2 h-2 rounded-full shrink-0" style={{ background: "#42d978", boxShadow: "0 0 0 4px rgba(66,217,120,0.18)" }} />
                  {t("liveStatus")}
                </span>
              </div>

              <div className="mt-6 grid gap-[14px]" aria-hidden="true">
                <article className="p-4 border border-black/[0.14] rounded-[18px] bg-white/85">
                  <p className="m-0 mb-[6px] text-[0.82rem] font-extrabold uppercase" style={{ color: INK }}>{t("messageNameLead")}</p>
                  <p className="m-0 text-[0.98rem] leading-[1.55]" style={{ color: INK }}>{t("messageTextLead")}</p>
                </article>
                <article className="p-4 border border-black/[0.14] rounded-[18px] ml-8 max-[720px]:ml-0" style={{ background: "rgba(17,18,22,0.95)", color: "#f3f5fb" }}>
                  <p className="m-0 mb-[6px] text-[0.82rem] font-extrabold uppercase">{t("messageNameDentDesk")}</p>
                  <p className="m-0 text-[0.98rem] leading-[1.55]">{t("messageTextDentDesk")}</p>
                </article>
                <article className="p-4 rounded-[18px]" style={{ background: "rgba(201,255,88,0.32)", border: "1px solid rgba(134,173,38,0.5)" }}>
                  <p className="m-0 mb-[6px] text-[0.82rem] font-extrabold uppercase" style={{ color: INK }}>{t("messageNameRouting")}</p>
                  <p className="m-0 text-[0.98rem] leading-[1.55]" style={{ color: INK }}>{t("messageTextRouting")}</p>
                </article>
              </div>

              <div className="grid grid-cols-2 gap-3 mt-6 max-[720px]:grid-cols-1">
                {[["statsOneValue","statsOneLabel"],["statsTwoValue","statsTwoLabel"],["statsThreeValue","statsThreeLabel"],["statsFourValue","statsFourLabel"]].map(([v,l]) => (
                  <article key={v} className="p-4 rounded-[18px] border border-black/[0.14] bg-white/80">
                    <span className="block mb-[6px] font-unbounded text-[1.15rem]" style={{ color: INK }}>{t(v)}</span>
                    <p className="m-0 text-sm leading-[1.5]" style={{ color: MUTED }}>{t(l)}</p>
                  </article>
                ))}
              </div>
            </Reveal>
          </section>

          {/* ── PROOF ── */}
          <section id="proof" className="py-[52px] max-[720px]:py-[34px]">
            <div className="max-w-[900px] mx-auto mb-[34px] text-center">
              <Reveal><Kicker>{t("proofKicker")}</Kicker></Reveal>
              <Reveal delay={45}><SectionH2>{t("proofHeading")}</SectionH2></Reveal>
            </div>
            <div className="grid grid-cols-2 gap-[22px] max-[1080px]:grid-cols-1">
              {[
                { title: "proofNowTitle",  items: ["proofNowItem1","proofNowItem2","proofNowItem3","proofNowItem4"], bg: "bg-white/60" },
                { title: "proofThenTitle", items: ["proofThenItem1","proofThenItem2","proofThenItem3","proofThenItem4"], bg: "" },
              ].map(({ title, items, bg }, i) => (
                <Reveal key={title} as="article" delay={i * 90}
                  className={`p-[clamp(24px,3vw,36px)] rounded-[36px] ${GLASS} ${bg}`}
                  style={{ ...PSHADOW, ...(i === 1 ? { background: "linear-gradient(160deg,rgba(201,255,88,0.2),rgba(255,255,255,0.82))" } : {}) }}>
                  <h3 className="m-0 mb-[22px] font-unbounded font-bold text-[2rem] text-balance" style={{ color: INK }}>{t(title)}</h3>
                  <ul className="m-0 p-0 list-none grid gap-[18px]">
                    {items.map((k) => (
                      <li key={k} className="pb-[18px] border-b border-black/[0.14] last:pb-0 last:border-b-0 text-[1.06rem] leading-[1.65]" style={{ color: MUTED }}>{t(k)}</li>
                    ))}
                  </ul>
                </Reveal>
              ))}
            </div>
          </section>

          {/* ── FEATURES ── */}
          <section className="py-[52px] max-[720px]:py-[34px]">
            <div className="max-w-[860px] mx-auto mb-[34px] text-center">
              <Reveal><Kicker>{t("featuresKicker")}</Kicker></Reveal>
              <Reveal delay={45}><SectionH2>{t("featuresHeading")}</SectionH2></Reveal>
            </div>
            <div className="grid grid-cols-3 gap-5 max-[1080px]:grid-cols-1">
              {[["featureOneTitle","featureOneText"],["featureTwoTitle","featureTwoText"],["featureThreeTitle","featureThreeText"]].map(([title, text], i) => (
                <Reveal key={title} as="article" delay={i * 90} className={`p-7 rounded-[26px] ${GLASS}`} style={PSHADOW}>
                  <p className="m-0 mb-4 text-[0.86rem] font-extrabold uppercase" style={{ color: MUTED }}>{String(i + 1).padStart(2, "0")}</p>
                  <h3 className="m-0 mb-[14px] font-unbounded font-bold text-[1.24rem] text-balance" style={{ color: INK }}>{t(title)}</h3>
                  <p className="m-0 text-[1.06rem] leading-[1.65]" style={{ color: MUTED }}>{t(text)}</p>
                </Reveal>
              ))}
            </div>
            <Reveal className="mt-6 flex flex-wrap gap-3 max-[720px]:[&>span]:w-full max-[720px]:[&>span]:justify-center" delay={90}>
              {["tickerItemOne","tickerItemTwo","tickerItemThree","tickerItemFour","tickerItemFive"].map((k) => (
                <span key={k} className="inline-flex items-center px-[18px] py-3 border border-black/[0.14] rounded-full bg-white/65 font-extrabold text-sm" style={{ color: INK }}>{t(k)}</span>
              ))}
            </Reveal>
          </section>

          {/* ── WORKFLOW ── */}
          <section className="py-[52px] max-[720px]:py-[34px]">
            <div className="max-w-[900px] mx-auto mb-[34px] text-center">
              <Reveal><Kicker>{t("workflowKicker")}</Kicker></Reveal>
              <Reveal delay={45}><SectionH2>{t("workflowHeading")}</SectionH2></Reveal>
            </div>
            <div className="grid grid-cols-3 gap-5 max-[1080px]:grid-cols-1">
              {[["workflowStepOneTitle","workflowStepOneText"],["workflowStepTwoTitle","workflowStepTwoText"],["workflowStepThreeTitle","workflowStepThreeText"]].map(([title, text], i) => (
                <Reveal key={title} as="article" delay={i * 90} className={`p-7 rounded-[26px] relative overflow-hidden ${GLASS}`} style={PSHADOW}>
                  <div className="absolute w-[180px] h-[180px] -right-[70px] -top-[70px] rounded-full pointer-events-none" style={{ background: "radial-gradient(circle,rgba(201,255,88,0.2),transparent 70%)" }} />
                  <span className="inline-flex items-center justify-center w-[52px] h-[52px] mb-5 rounded-[14px] font-unbounded" style={{ background: "rgba(17,18,22,0.94)", color: "#f7f8fc" }}>{i + 1}</span>
                  <h3 className="m-0 mb-[14px] font-unbounded font-bold text-[1.24rem] text-balance" style={{ color: INK }}>{t(title)}</h3>
                  <p className="m-0 text-[1.06rem] leading-[1.65]" style={{ color: MUTED }}>{t(text)}</p>
                </Reveal>
              ))}
            </div>
          </section>

          {/* ── CONTACT ── */}
          <section id="contact" className="py-[52px] max-[720px]:py-[34px]">
            <div className="grid items-center gap-7 p-10 rounded-[44px] border border-black/[0.14] backdrop-blur-lg max-[1080px]:grid-cols-1 max-[560px]:p-5"
              style={{ gridTemplateColumns: "minmax(0,1.2fr) minmax(320px,430px)", background: "linear-gradient(135deg,rgba(255,255,255,0.92),rgba(231,235,251,0.88))", ...PSHADOW }}>
              <Reveal className="max-w-[52ch] justify-self-center text-center max-[1080px]:max-w-none">
                <Kicker>{t("contactKicker")}</Kicker>
                <SectionH2>{t("contactHeading")}</SectionH2>
                <p className="mt-5 mb-0 text-[1.06rem] leading-[1.65]" style={{ color: MUTED }}>{t("contactText")}</p>
                <div className="flex flex-wrap justify-center gap-x-6 gap-y-3 mt-[26px] font-bold max-[720px]:flex-col max-[720px]:items-center">
                  <a href="mailto:dentdesk.kz@gmail.com" className="underline decoration-[1px] underline-offset-4" style={{ color: INK }}>{t("contactEmailLabel")}</a>
                  <a href="https://wa.me/77058106425" target="_blank" rel="noreferrer" className="underline decoration-[1px] underline-offset-4" style={{ color: INK }}>{t("contactWhatsappLabel")}</a>
                  <span style={{ color: MUTED }}>{t("contactEmailNote")}</span>
                </div>
              </Reveal>
              <Reveal delay={90} className={`justify-self-center w-full max-w-[430px] p-7 rounded-[26px] max-[1080px]:max-w-[560px] ${GLASS}`} style={PSHADOW}>
                <LeadForm t={t} />
              </Reveal>
            </div>
          </section>

        </main>

        <footer className="flex justify-between gap-4 py-[26px] pb-[42px] text-[0.94rem] max-[720px]:flex-col max-[720px]:items-center" style={{ color: MUTED }}>
          <p className="m-0 font-bold">{t("footerTitle")}</p>
          <p className="m-0">{t("footerSubtitle")}</p>
        </footer>
      </div>

      <Modal open={modalOpen} onClose={() => setModalOpen(false)} t={t} />
    </div>
  );
}
