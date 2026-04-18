import { useEffect, useRef, useState, type RefObject } from 'react'
import { Link } from 'react-router-dom'
import { ArrowRight, Globe, Menu, Stethoscope, X } from 'lucide-react'
import { useAuth } from '@/hooks/useAuth'
import { cn } from '@/lib/utils'
import { landingContent, type LandingCopy, type LandingLanguage } from './landingContent'
import './landing.css'

const LANGUAGE_STORAGE_KEY = 'dentdesk-landing-lang'
const LEAD_STORAGE_KEY = 'dentdesk-demo-leads'

interface LeadPayload {
  phone: string
  name: string
  clinic: string
}

interface LeadFormProps {
  copy: LandingCopy['form']
  compact?: boolean
  focusRef?: RefObject<HTMLInputElement | null>
  onSubmitted?: () => void
}

type FeedbackTone = 'error' | 'success'

interface FeedbackState {
  tone: FeedbackTone
  message: string
}

interface StoredLead extends LeadPayload {
  submittedAt: string
}

function getInitialLanguage(): LandingLanguage {
  if (typeof window === 'undefined') {
    return 'ru'
  }

  const stored = window.localStorage.getItem(LANGUAGE_STORAGE_KEY)
  if (stored === 'ru' || stored === 'en') {
    return stored
  }

  return window.navigator.language.toLowerCase().startsWith('en') ? 'en' : 'ru'
}

function phoneLooksValid(phone: string) {
  return phone.replace(/\D/g, '').length >= 10
}

function saveLead(payload: LeadPayload) {
  try {
    const raw = window.localStorage.getItem(LEAD_STORAGE_KEY)
    const existing = raw ? (JSON.parse(raw) as StoredLead[]) : []

    existing.push({
      ...payload,
      submittedAt: new Date().toISOString(),
    })

    window.localStorage.setItem(LEAD_STORAGE_KEY, JSON.stringify(existing))
  } catch {
    // The demo lead capture should stay non-blocking even if storage is unavailable.
  }
}

function LeadForm({ copy, compact = false, focusRef, onSubmitted }: LeadFormProps) {
  const [phone, setPhone] = useState('')
  const [name, setName] = useState('')
  const [clinic, setClinic] = useState('')
  const [feedback, setFeedback] = useState<FeedbackState | null>(null)

  function resetForm() {
    setPhone('')
    setName('')
    setClinic('')
  }

  function handleSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault()

    const payload = {
      phone: phone.trim(),
      name: name.trim(),
      clinic: clinic.trim(),
    }

    if (!payload.phone || !payload.name || !payload.clinic) {
      setFeedback({ tone: 'error', message: copy.requiredError })
      return
    }

    if (!phoneLooksValid(payload.phone)) {
      setFeedback({ tone: 'error', message: copy.phoneError })
      return
    }

    saveLead(payload)
    resetForm()
    setFeedback({ tone: 'success', message: copy.success })

    if (onSubmitted) {
      window.setTimeout(onSubmitted, 1300)
    }
  }

  return (
    <form className={cn('landing-lead-form', compact && 'is-compact')} onSubmit={handleSubmit} noValidate>
      {!compact && <p className="landing-form-kicker">{copy.kicker}</p>}

      <label>
        <span>{copy.phoneLabel}</span>
        <input
          ref={focusRef}
          type="tel"
          name="phone"
          inputMode="tel"
          placeholder={copy.phonePlaceholder}
          value={phone}
          onChange={(event) => setPhone(event.target.value)}
          required
        />
      </label>

      <label>
        <span>{copy.nameLabel}</span>
        <input
          type="text"
          name="name"
          placeholder={copy.namePlaceholder}
          value={name}
          onChange={(event) => setName(event.target.value)}
          required
        />
      </label>

      <label>
        <span>{copy.clinicLabel}</span>
        <input
          type="text"
          name="clinic"
          placeholder={copy.clinicPlaceholder}
          value={clinic}
          onChange={(event) => setClinic(event.target.value)}
          required
        />
      </label>

      <button className="landing-button-primary landing-submit-button" type="submit">
        {copy.submit}
      </button>

      <p
        className={cn(
          'landing-form-feedback',
          feedback?.tone === 'error' && 'is-error',
          feedback?.tone === 'success' && 'is-success',
        )}
        aria-live="polite"
      >
        {feedback?.message ?? ''}
      </p>
    </form>
  )
}

export function LandingPage() {
  const pageRef = useRef<HTMLDivElement>(null)
  const firstModalInputRef = useRef<HTMLInputElement>(null)
  const user = useAuth((state) => state.user)

  const [language, setLanguage] = useState<LandingLanguage>(getInitialLanguage)
  const [isMenuOpen, setIsMenuOpen] = useState(false)
  const [isModalOpen, setIsModalOpen] = useState(false)

  const copy = landingContent[language]
  const appEntryPath = user ? '/app' : '/login'
  const appEntryLabel = user ? copy.nav.openApp : copy.nav.login

  useEffect(() => {
    window.localStorage.setItem(LANGUAGE_STORAGE_KEY, language)
    document.documentElement.lang = language
    document.title = copy.meta.title

    const descriptionTag = document.querySelector('meta[name="description"]')
    if (descriptionTag) {
      descriptionTag.setAttribute('content', copy.meta.description)
    }
  }, [copy.meta.description, copy.meta.title, language])

  useEffect(() => {
    const page = pageRef.current
    if (!page) {
      return
    }

    const revealItems = Array.from(page.querySelectorAll<HTMLElement>('.landing-reveal'))
    if (!('IntersectionObserver' in window)) {
      revealItems.forEach((item) => item.classList.add('is-visible'))
      return
    }

    const observer = new IntersectionObserver(
      (entries) => {
        entries.forEach((entry) => {
          if (entry.isIntersecting) {
            entry.target.classList.add('is-visible')
            observer.unobserve(entry.target)
          }
        })
      },
      {
        threshold: 0.18,
        rootMargin: '0px 0px -40px 0px',
      },
    )

    revealItems.forEach((item, index) => {
      item.style.transitionDelay = `${Math.min(index * 45, 280)}ms`
      observer.observe(item)
    })

    return () => observer.disconnect()
  }, [])

  useEffect(() => {
    if (!isModalOpen) {
      return
    }

    document.body.style.overflow = 'hidden'
    const focusTimer = window.setTimeout(() => firstModalInputRef.current?.focus(), 30)

    return () => {
      window.clearTimeout(focusTimer)
      document.body.style.overflow = ''
    }
  }, [isModalOpen])

  useEffect(() => {
    if (!isMenuOpen && !isModalOpen) {
      return
    }

    function handleKeyDown(event: KeyboardEvent) {
      if (event.key !== 'Escape') {
        return
      }

      if (isModalOpen) {
        setIsModalOpen(false)
        return
      }

      setIsMenuOpen(false)
    }

    window.addEventListener('keydown', handleKeyDown)
    return () => window.removeEventListener('keydown', handleKeyDown)
  }, [isMenuOpen, isModalOpen])

  function openModal() {
    setIsMenuOpen(false)
    setIsModalOpen(true)
  }

  function closeModal() {
    setIsModalOpen(false)
  }

  function closeMenu() {
    setIsMenuOpen(false)
  }

  return (
    <div ref={pageRef} className="landing-page">
      <div className="landing-shell">
        <header className="landing-topbar">
          <a className="landing-brand" href="#top" aria-label="DentDesk">
            <span className="landing-brand-mark" aria-hidden="true" />
            <span>DentDesk</span>
          </a>

          <div className="landing-topbar-actions">
            <div className="landing-language-switch" role="group" aria-label={copy.nav.switchLanguage}>
              {(['ru', 'en'] as LandingLanguage[]).map((nextLanguage) => (
                <button
                  key={nextLanguage}
                  type="button"
                  className={cn('landing-language-button', language === nextLanguage && 'is-active')}
                  onClick={() => setLanguage(nextLanguage)}
                >
                  <Globe className="landing-language-icon" />
                  <span>{nextLanguage.toUpperCase()}</span>
                </button>
              ))}
            </div>

            <button
              type="button"
              className="landing-menu-button"
              aria-expanded={isMenuOpen}
              aria-controls="landing-nav-panel"
              aria-label={isMenuOpen ? copy.nav.closeMenu : copy.nav.openMenu}
              onClick={() => setIsMenuOpen((open) => !open)}
            >
              {isMenuOpen ? <X /> : <Menu />}
            </button>

            <div id="landing-nav-panel" className={cn('landing-nav-panel', isMenuOpen && 'is-open')}>
              <nav className="landing-nav" aria-label={copy.nav.aria}>
                <a href="#proof" onClick={closeMenu}>
                  {copy.nav.platform}
                </a>
                <a href="#contact" onClick={closeMenu}>
                  {copy.nav.contact}
                </a>
                <Link className="landing-nav-login" to={appEntryPath} onClick={closeMenu}>
                  {appEntryLabel}
                </Link>
                <button className="landing-nav-cta" type="button" onClick={openModal}>
                  {copy.nav.bookDemo}
                </button>
              </nav>
            </div>
          </div>
        </header>

        <div className={cn('landing-nav-scrim', isMenuOpen && 'is-open')} onClick={closeMenu} aria-hidden="true" />

        <main>
          <section className="landing-section landing-hero" id="top">
            <div className="landing-hero-copy">
              <p className="landing-eyebrow landing-reveal">{copy.hero.eyebrow}</p>
              <h1 className="landing-hero-title landing-reveal">{copy.hero.title}</h1>
              <p className="landing-hero-text landing-reveal">{copy.hero.text}</p>

              <div className="landing-hero-actions landing-reveal">
                <button className="landing-button-primary" type="button" onClick={openModal}>
                  {copy.hero.primaryCta}
                </button>
                <a className="landing-text-link" href="#proof">
                  <span>{copy.hero.secondaryCta}</span>
                  <ArrowRight className="landing-arrow-icon" />
                </a>
              </div>

              <p className="landing-hero-note landing-reveal">{copy.hero.note}</p>
            </div>

            <div className="landing-hero-panel landing-reveal">
              <div className="landing-panel-head">
                <div>
                  <p className="landing-panel-label">{copy.hero.panelLabel}</p>
                  <p className="landing-panel-title">{copy.hero.panelTitle}</p>
                </div>
                <span className="landing-live-pill">{copy.hero.online}</span>
              </div>

              <div className="landing-message-stack" aria-hidden="true">
                {copy.hero.messages.map((message) => (
                  <article
                    key={`${message.name}-${message.text}`}
                    className={cn(
                      'landing-message-card',
                      message.tone === 'outgoing' && 'is-outgoing',
                      message.tone === 'accent' && 'is-accent',
                    )}
                  >
                    <p className="landing-message-name">{message.name}</p>
                    <p>{message.text}</p>
                  </article>
                ))}
              </div>

              <div className="landing-stats-grid">
                {copy.hero.metrics.map((metric) => (
                  <article key={`${metric.value}-${metric.label}`}>
                    <span>{metric.value}</span>
                    <p>{metric.label}</p>
                  </article>
                ))}
              </div>
            </div>
          </section>

          <section className="landing-section" id="proof">
            <div className="landing-section-heading landing-reveal">
              <p className="landing-section-kicker">{copy.proof.kicker}</p>
              <h2>{copy.proof.title}</h2>
            </div>

            <div className="landing-contrast-board">
              <article className="landing-contrast-column landing-reveal">
                <h3>{copy.proof.nowTitle}</h3>
                <ul>
                  {copy.proof.nowItems.map((item) => (
                    <li key={item}>{item}</li>
                  ))}
                </ul>
              </article>

              <article className="landing-contrast-column landing-contrast-column-positive landing-reveal">
                <h3>{copy.proof.futureTitle}</h3>
                <ul>
                  {copy.proof.futureItems.map((item) => (
                    <li key={item}>{item}</li>
                  ))}
                </ul>
              </article>
            </div>
          </section>

          <section className="landing-section">
            <div className="landing-section-heading landing-section-heading-narrow landing-reveal">
              <p className="landing-section-kicker">{copy.features.kicker}</p>
              <h2>{copy.features.title}</h2>
            </div>

            <div className="landing-feature-grid">
              {copy.features.items.map((feature) => (
                <article key={feature.index} className="landing-feature-card landing-reveal">
                  <p className="landing-feature-index">{feature.index}</p>
                  <h3>{feature.title}</h3>
                  <p>{feature.text}</p>
                </article>
              ))}
            </div>

            <div className="landing-ticker landing-reveal" aria-label={copy.features.kicker}>
              {copy.features.ticker.map((item) => (
                <span key={item}>{item}</span>
              ))}
            </div>
          </section>

          <section className="landing-section">
            <div className="landing-section-heading landing-reveal">
              <p className="landing-section-kicker">{copy.workflow.kicker}</p>
              <h2>{copy.workflow.title}</h2>
            </div>

            <div className="landing-workflow-grid">
              {copy.workflow.steps.map((step) => (
                <article key={step.number} className="landing-workflow-step landing-reveal">
                  <span className="landing-step-number">{step.number}</span>
                  <h3>{step.title}</h3>
                  <p>{step.text}</p>
                </article>
              ))}
            </div>
          </section>

          <section className="landing-section" id="contact">
            <div className="landing-contact-section">
              <div className="landing-contact-copy landing-reveal">
                <p className="landing-section-kicker">{copy.contact.kicker}</p>
                <h2>{copy.contact.title}</h2>
                <p>{copy.contact.text}</p>

                <div className="landing-contact-meta">
                  <a href="mailto:hello@dentdesk.ai">hello@dentdesk.ai</a>
                  <span>{copy.contact.responseWindow}</span>
                </div>
              </div>

              <div className="landing-reveal">
                <LeadForm copy={copy.form} />
              </div>
            </div>
          </section>
        </main>

        <footer className="landing-footer">
          <p>DentDesk</p>
          <p>{copy.footer.tagline}</p>
        </footer>
      </div>

      <div
        className={cn('landing-modal', isModalOpen && 'is-open')}
        role="dialog"
        aria-modal="true"
        aria-hidden={!isModalOpen}
        aria-labelledby="landing-modal-title"
      >
        <div className="landing-modal-backdrop" onClick={closeModal} />
        <div className="landing-modal-dialog">
          <button
            className="landing-modal-close"
            type="button"
            aria-label={copy.form.close}
            onClick={closeModal}
          >
            &times;
          </button>

          <p className="landing-form-kicker">{copy.form.modalKicker}</p>
          <h2 id="landing-modal-title">{copy.form.modalTitle}</h2>
          <LeadForm
            copy={copy.form}
            compact
            focusRef={firstModalInputRef}
            onSubmitted={closeModal}
          />
        </div>
      </div>

      <Link className="landing-floating-login" to={appEntryPath}>
        <Stethoscope className="landing-floating-login-icon" />
        <span>{appEntryLabel}</span>
      </Link>
    </div>
  )
}
