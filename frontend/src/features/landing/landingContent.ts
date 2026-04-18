export type LandingLanguage = 'ru' | 'en'

type MessageTone = 'incoming' | 'outgoing' | 'accent'

interface LandingMessage {
  name: string
  text: string
  tone: MessageTone
}

interface LandingMetric {
  value: string
  label: string
}

interface LandingFeature {
  index: string
  title: string
  text: string
}

interface LandingWorkflowStep {
  number: string
  title: string
  text: string
}

export interface LandingCopy {
  meta: {
    title: string
    description: string
  }
  nav: {
    aria: string
    platform: string
    contact: string
    login: string
    openApp: string
    bookDemo: string
    switchLanguage: string
    openMenu: string
    closeMenu: string
  }
  hero: {
    eyebrow: string
    title: string
    text: string
    primaryCta: string
    secondaryCta: string
    note: string
    panelLabel: string
    panelTitle: string
    online: string
    messages: LandingMessage[]
    metrics: LandingMetric[]
  }
  proof: {
    kicker: string
    title: string
    nowTitle: string
    nowItems: string[]
    futureTitle: string
    futureItems: string[]
  }
  features: {
    kicker: string
    title: string
    items: LandingFeature[]
    ticker: string[]
  }
  workflow: {
    kicker: string
    title: string
    steps: LandingWorkflowStep[]
  }
  contact: {
    kicker: string
    title: string
    text: string
    responseWindow: string
  }
  form: {
    kicker: string
    modalKicker: string
    modalTitle: string
    phoneLabel: string
    nameLabel: string
    clinicLabel: string
    phonePlaceholder: string
    namePlaceholder: string
    clinicPlaceholder: string
    submit: string
    close: string
    requiredError: string
    phoneError: string
    success: string
  }
  footer: {
    tagline: string
  }
}

export const landingContent: Record<LandingLanguage, LandingCopy> = {
  ru: {
    meta: {
      title: 'DentDesk — ИИ для стоматологий',
      description:
        'DentDesk — нишевый ИИ для стоматологий: запись пациентов 24/7, единое окно диалогов и умное распределение лидов.',
    },
    nav: {
      aria: 'Основная навигация',
      platform: 'Платформа',
      contact: 'Контакты',
      login: 'Войти',
      openApp: 'Открыть кабинет',
      bookDemo: 'Запросить демо',
      switchLanguage: 'Сменить язык',
      openMenu: 'Открыть меню',
      closeMenu: 'Закрыть меню',
    },
    hero: {
      eyebrow: 'Нишевый ИИ для стоматологий',
      title: 'Пациенты записываются без раздражения, а команда работает в одном спокойном окне.',
      text:
        'DentDesk берет на себя первичный диалог, квалификацию лида и маршрутизацию по креслам. Ваши администраторы не тонут в WhatsApp, а врачи видят ровный поток записей.',
      primaryCta: 'Записаться на консультацию',
      secondaryCta: 'Посмотреть сценарий работы',
      note:
        'Наши ИИ-агенты специализируются только на стоматологическом бизнесе и разговаривают языком клиники.',
      panelLabel: 'ИИ-регистратор',
      panelTitle: 'Живой прием заявок 24/7',
      online: 'online',
      messages: [
        {
          name: 'Новый лид',
          text: 'Здравствуйте, нужна запись на чистку завтра после 18:00.',
          tone: 'incoming',
        },
        {
          name: 'DentDesk',
          text: 'Подберу окно. Уточните филиал и удобно ли подтвердить запись в WhatsApp?',
          tone: 'outgoing',
        },
        {
          name: 'Маршрутизация',
          text: 'Терапия, филиал Абая, подтверждение получено, слот найден.',
          tone: 'accent',
        },
      ],
      metrics: [
        { value: '5 мин', label: 'на запуск первого агента' },
        { value: '1 окно', label: 'для всех новых диалогов' },
        { value: '24/7', label: 'ответ без пропущенных лидов' },
        { value: '0 хаоса', label: 'в ручном разборе переписок' },
      ],
    },
    proof: {
      kicker: 'До и после внедрения',
      title: 'Клиника перестает терять лиды между звонками, сменами и чатами.',
      nowTitle: 'Сейчас',
      nowItems: [
        'Лиды теряются, потому что не успели ответить вовремя',
        'Рабочий WhatsApp клиники превращается в бесконечный шум',
        'До 30% пациентов не получают ответ в тот же день',
        'Администратор вручную уточняет одно и то же десятки раз',
      ],
      futureTitle: 'DentDesk',
      futureItems: [
        'ИИ-агент запускается за 5 минут под вашу клинику',
        'Все диалоги собираются в одном окне без лишнего хаоса',
        'Лид автоматически попадает на нужное кресло и филиал',
        'Команда получает только уже подготовленного пациента',
      ],
    },
    features: {
      kicker: 'Что получает клиника',
      title:
        'Не просто чат-бот, а операционный слой для записи, подтверждения и перераспределения потока.',
      items: [
        {
          index: '01',
          title: 'Запись без очереди на ответ',
          text:
            'DentDesk ведет первый диалог, собирает контекст и доводит пациента до подтверждения, пока администратор занят живым приемом.',
        },
        {
          index: '02',
          title: 'Единый центр диалогов',
          text:
            'Все новые обращения, статусы и перезапуски сценариев собраны в одном интерфейсе вместо разбросанных чатов и заметок.',
        },
        {
          index: '03',
          title: 'Маршрутизация по профилю врача',
          text:
            'Система понимает тип услуги, филиал, срочность и корректно направляет лида туда, где его можно закрыть в запись быстрее.',
        },
      ],
      ticker: [
        'WhatsApp',
        'Сценарии записи',
        'Подтверждение визита',
        'Распределение по филиалам',
        'Повторный прогрев',
      ],
    },
    workflow: {
      kicker: 'Сценарий за один день',
      title: 'Команда видит только то, что требует решения человеком.',
      steps: [
        {
          number: '1',
          title: 'Пациент пишет',
          text:
            'DentDesk отвечает моментально, уточняет услугу, филиал и желаемое время, не заставляя человека ждать свободного администратора.',
        },
        {
          number: '2',
          title: 'Система квалифицирует лид',
          text:
            'Запрос раскладывается по типу услуги, врачу, срочности и этапу воронки. Повторные вопросы и неясные кейсы отдаются сотруднику.',
        },
        {
          number: '3',
          title: 'Кресло получает подготовленного пациента',
          text:
            'Команда видит собранный контекст, подтвержденный слот и историю диалога, а не сырой поток сообщений без структуры.',
        },
      ],
    },
    contact: {
      kicker: 'Контакты',
      title: 'Покажем, как DentDesk встраивается в вашу клинику без долгого запуска.',
      text:
        'Оставьте номер WhatsApp и имя. Мы свяжемся, покажем сценарий на ваших кейсах и соберем первую рабочую конфигурацию под филиалы, услуги и загрузку команды.',
      responseWindow: 'Ответ в течение рабочего дня',
    },
    form: {
      kicker: 'Стартовая консультация',
      modalKicker: 'Не упускайте лиды',
      modalTitle: 'Начните работу с DentDesk',
      phoneLabel: 'Номер WhatsApp',
      nameLabel: 'Имя',
      clinicLabel: 'Клиника',
      phonePlaceholder: '+7 (700) 123-45-67',
      namePlaceholder: 'Александр Пушкин',
      clinicPlaceholder: 'Dent Smile на Абая',
      submit: 'Отправить',
      close: 'Закрыть',
      requiredError: 'Заполните все поля, чтобы мы могли связаться с вами.',
      phoneError: 'Укажите корректный номер WhatsApp.',
      success: 'Заявка принята. Мы свяжемся с вами и покажем сценарий под вашу клинику.',
    },
    footer: {
      tagline: 'ИИ-инфраструктура для стоматологий',
    },
  },
  en: {
    meta: {
      title: 'DentDesk — AI for Dental Clinics',
      description:
        'DentDesk is a dental AI operations layer for 24/7 patient booking, unified conversations, and smarter lead routing.',
    },
    nav: {
      aria: 'Main navigation',
      platform: 'Platform',
      contact: 'Contact',
      login: 'Log in',
      openApp: 'Open app',
      bookDemo: 'Book a demo',
      switchLanguage: 'Switch language',
      openMenu: 'Open menu',
      closeMenu: 'Close menu',
    },
    hero: {
      eyebrow: 'Vertical AI for dental clinics',
      title: 'Patients book without friction, while your team works from one calm operating window.',
      text:
        'DentDesk handles the first conversation, qualifies leads, and routes patients to the right chair. Your admins stop drowning in WhatsApp and your doctors see a steady stream of confirmed bookings.',
      primaryCta: 'Book a consultation',
      secondaryCta: 'See the workflow',
      note:
        'Our AI agents are trained specifically for dental clinics and speak the language of your practice.',
      panelLabel: 'AI registrar',
      panelTitle: 'Live lead intake 24/7',
      online: 'online',
      messages: [
        {
          name: 'New lead',
          text: 'Hi, I need a teeth cleaning appointment tomorrow after 6 PM.',
          tone: 'incoming',
        },
        {
          name: 'DentDesk',
          text: 'I can find a slot. Which branch works for you, and is WhatsApp okay for confirmation?',
          tone: 'outgoing',
        },
        {
          name: 'Routing',
          text: 'Therapy, Abay branch, confirmation received, slot reserved.',
          tone: 'accent',
        },
      ],
      metrics: [
        { value: '5 min', label: 'to launch the first agent' },
        { value: '1 hub', label: 'for every new conversation' },
        { value: '24/7', label: 'response coverage without missed leads' },
        { value: '0 chaos', label: 'in manual inbox triage' },
      ],
    },
    proof: {
      kicker: 'Before and after rollout',
      title: 'The clinic stops losing leads between calls, shifts, and chat threads.',
      nowTitle: 'Today',
      nowItems: [
        'Leads get lost because nobody answered in time',
        'The clinic WhatsApp inbox turns into endless noise',
        'Up to 30% of patients never get a same-day reply',
        'Admins repeat the same qualification questions again and again',
      ],
      futureTitle: 'DentDesk',
      futureItems: [
        'An AI agent launches for your clinic in about five minutes',
        'Every conversation lands in one operating window without extra clutter',
        'Each lead is routed automatically to the right branch and chair',
        'Your team only sees patients who are already prepared for action',
      ],
    },
    features: {
      kicker: 'What the clinic gets',
      title:
        'Not just a chatbot, but an operating layer for booking, confirmation, and load rebalancing.',
      items: [
        {
          index: '01',
          title: 'Booking without a reply queue',
          text:
            'DentDesk runs the first conversation, gathers context, and moves the patient to confirmation while the admin stays focused on live reception.',
        },
        {
          index: '02',
          title: 'One conversation center',
          text:
            'New inquiries, statuses, and workflow restarts live in one interface instead of scattered chats and notes.',
        },
        {
          index: '03',
          title: 'Doctor-profile routing',
          text:
            'The system understands treatment type, branch, urgency, and routes the lead to the place where it can be converted fastest.',
        },
      ],
      ticker: [
        'WhatsApp',
        'Booking flows',
        'Visit confirmations',
        'Multi-branch routing',
        'Re-engagement',
      ],
    },
    workflow: {
      kicker: 'One-day workflow',
      title: 'The team only sees what still needs a human decision.',
      steps: [
        {
          number: '1',
          title: 'The patient reaches out',
          text:
            'DentDesk replies instantly, clarifies the treatment, branch, and preferred time, and removes the wait for a free administrator.',
        },
        {
          number: '2',
          title: 'The system qualifies the lead',
          text:
            'The request is mapped by service type, doctor, urgency, and funnel stage. Repeated questions and edge cases go to a staff member.',
        },
        {
          number: '3',
          title: 'The chair gets a prepared patient',
          text:
            'Your team sees the collected context, confirmed slot, and message history instead of a raw, unstructured inbox.',
        },
      ],
    },
    contact: {
      kicker: 'Contact',
      title: 'We will show how DentDesk fits into your clinic without a long rollout.',
      text:
        'Leave your WhatsApp number and name. We will reach out, walk through your own cases, and assemble the first working setup for your branches, services, and team workload.',
      responseWindow: 'Reply within one business day',
    },
    form: {
      kicker: 'Intro consultation',
      modalKicker: 'Do not lose more leads',
      modalTitle: 'Start with DentDesk',
      phoneLabel: 'WhatsApp number',
      nameLabel: 'Name',
      clinicLabel: 'Clinic',
      phonePlaceholder: '+1 (555) 123-45-67',
      namePlaceholder: 'Alex Morgan',
      clinicPlaceholder: 'Bright Smile Downtown',
      submit: 'Send request',
      close: 'Close',
      requiredError: 'Fill in every field so we can reach you.',
      phoneError: 'Enter a valid WhatsApp number.',
      success: 'Request received. We will contact you and map the right workflow for your clinic.',
    },
    footer: {
      tagline: 'AI infrastructure for dental clinics',
    },
  },
}
