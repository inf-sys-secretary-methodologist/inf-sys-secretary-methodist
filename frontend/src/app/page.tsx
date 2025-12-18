'use client'

import * as React from 'react'
import {
  Calendar,
  ChevronDown,
  ChevronLeft,
  ChevronRight,
  Shield,
  GraduationCap,
  UserCog,
  BookOpen,
  Users,
  FileText,
  Bell,
  RefreshCw,
  Palette,
  Search,
  Archive,
  BarChart3,
  Briefcase,
  MessageSquare,
} from 'lucide-react'
import { useRouter } from 'next/navigation'
import { cn } from '@/lib/utils'
import { GlowingEffect } from '@/components/ui/glowing-effect-lazy'
import { ThemeSettingsPopover } from '@/components/theme-settings-popover'
import { UserMenu } from '@/components/UserMenu'
import { Button } from '@/components/ui/button'
import { useAuthCheck } from '@/hooks/useAuth'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'

// Module data with detailed descriptions
const MODULES_DATA = [
  // ✅ Реализованные модули
  {
    icon: <FileText className="h-6 w-6" />,
    title: 'Документооборот',
    description: 'Полный цикл работы с документами: создание, согласование, архивирование',
    details: {
      features: [
        'Создание документов по шаблонам',
        'Автоматическая регистрация и нумерация',
        'Workflow согласования документов',
        'Версионирование и история изменений',
        'Контроль сроков исполнения',
      ],
      status: 'Доступен',
    },
  },
  {
    icon: <Bell className="h-6 w-6" />,
    title: 'Система уведомлений',
    description: 'Многоканальные уведомления: Email, Telegram, In-app',
    details: {
      features: [
        'Email-уведомления через Gmail API',
        'Telegram-бот с привязкой аккаунта',
        'Уведомления внутри приложения',
        'Настройка тихих часов и таймзон',
        'Приоритеты и фильтрация уведомлений',
      ],
      status: 'Доступен',
    },
  },
  {
    icon: <Calendar className="h-6 w-6" />,
    title: 'Календарь событий',
    description: 'Интерактивный календарь с напоминаниями о важных событиях',
    details: {
      features: [
        'Просмотр событий по дням, неделям и месяцам',
        'Создание повторяющихся мероприятий',
        'Push-уведомления о предстоящих событиях',
        'Синхронизация с внешними календарями',
        'Фильтрация по типам событий',
      ],
      status: 'Доступен',
    },
  },
  {
    icon: <RefreshCw className="h-6 w-6" />,
    title: 'Интеграция с 1С',
    description: 'Синхронизация данных сотрудников и студентов с системой 1С',
    details: {
      features: [
        'Импорт сотрудников из 1С',
        'Импорт студентов из 1С',
        'Автоматическое разрешение конфликтов',
        'Логирование операций синхронизации',
        'Гибкие настройки интеграции',
      ],
      status: 'Доступен',
    },
  },
  {
    icon: <Palette className="h-6 w-6" />,
    title: 'Персонализация интерфейса',
    description: 'Настройка внешнего вида: темы, шейдерные фоны, размер текста',
    details: {
      features: [
        'Светлая и тёмная темы оформления',
        'Анимированные шейдерные фоны',
        'Настройка размера интерфейса',
        'Режим высокой контрастности',
        'Сохранение настроек в профиле',
      ],
      status: 'Доступен',
    },
  },
  {
    icon: <Search className="h-6 w-6" />,
    title: 'Полнотекстовый поиск',
    description: 'Быстрый поиск по документам, пользователям и событиям',
    details: {
      features: [
        'Поиск по содержимому документов',
        'Фильтрация по типу, дате, автору',
        'Поиск по реквизитам документов',
        'Мгновенные результаты с подсветкой',
        'История поисковых запросов',
      ],
      status: 'Доступен',
    },
  },
  // 🚧 Планируемые модули
  {
    icon: <Archive className="h-6 w-6" />,
    title: 'Система архивирования',
    description: 'Автоматическое архивирование документов с возможностью восстановления',
    details: {
      features: [
        'Автоматическое резервное копирование всех документов',
        'Версионирование файлов с историей изменений',
        'Быстрый поиск в архиве по ключевым словам',
        'Восстановление удалённых документов в один клик',
        'Настраиваемые политики хранения данных',
      ],
      status: 'В разработке',
    },
  },
  {
    icon: <BarChart3 className="h-6 w-6" />,
    title: 'Отчеты и аналитика',
    description: 'Детальная статистика и генерация отчётов по всем направлениям',
    details: {
      features: [
        'Отслеживание посещаемости в реальном времени',
        'Генерация отчётов по группам и периодам',
        'Визуализация данных в графиках и диаграммах',
        'Экспорт в Excel и PDF форматы',
        'Автоматические уведомления о важных метриках',
      ],
      status: 'В разработке',
    },
  },
  {
    icon: <Briefcase className="h-6 w-6" />,
    title: 'Управление проектами',
    description: 'Планирование и контроль выполнения учебных проектов',
    details: {
      features: [
        'Создание проектов с этапами и дедлайнами',
        'Назначение ответственных исполнителей',
        'Отслеживание прогресса выполнения',
        'Канбан-доска для визуализации задач',
        'Совместная работа над проектами',
      ],
      status: 'В разработке',
    },
  },
  {
    icon: <MessageSquare className="h-6 w-6" />,
    title: 'Коммуникационный центр',
    description: 'Централизованная система для общения преподавателей и студентов',
    details: {
      features: [
        'Внутренний чат между пользователями',
        'Групповые обсуждения по темам',
        'Рассылка объявлений и уведомлений',
        'История переписки с поиском',
        'Интеграция со Slack и другими мессенджерами',
      ],
      status: 'В разработке',
    },
  },
]

// Animated Counter Component with intersection observer
interface CounterProps {
  end: number
  className?: string
  duration?: number
  suffix?: string
}

const Counter = ({ end, className, duration = 2000, suffix = '' }: CounterProps) => {
  const [count, setCount] = React.useState(0)
  const [hasAnimated, setHasAnimated] = React.useState(false)
  const ref = React.useRef<HTMLDivElement>(null)

  React.useEffect(() => {
    const observer = new IntersectionObserver(
      (entries) => {
        if (entries[0].isIntersecting && !hasAnimated) {
          setHasAnimated(true)
          let startTime: number
          const animate = (currentTime: number) => {
            if (!startTime) startTime = currentTime
            const progress = Math.min((currentTime - startTime) / duration, 1)
            // Easing function for smooth animation
            const easeOutQuart = 1 - Math.pow(1 - progress, 4)
            setCount(Math.floor(easeOutQuart * end))
            if (progress < 1) {
              requestAnimationFrame(animate)
            }
          }
          requestAnimationFrame(animate)
        }
      },
      { threshold: 0.1 }
    )

    if (ref.current) {
      observer.observe(ref.current)
    }

    return () => observer.disconnect()
  }, [end, duration, hasAnimated])

  return (
    <div ref={ref} className={cn('font-bold tabular-nums', className)}>
      {count.toLocaleString('ru-RU')}
      {suffix}
    </div>
  )
}

// Statistics data
const STATS_DATA = [
  { value: 15000, label: 'Документов обработано', suffix: '+' },
  { value: 2500, label: 'Активных пользователей', suffix: '+' },
  { value: 850, label: 'Мероприятий проведено', suffix: '+' },
  { value: 99, label: 'Удовлетворённость', suffix: '%' },
]

// FAQ data
const FAQ_DATA = [
  {
    question: 'Как начать работу с системой?',
    answer:
      'Для начала работы зарегистрируйтесь в системе или войдите с помощью учётных данных, предоставленных администратором. После входа вы получите доступ к функциям согласно вашей роли.',
  },
  {
    question: 'Какие роли пользователей существуют в системе?',
    answer:
      'В системе предусмотрено 5 ролей: Администратор (полный доступ), Методист (управление документами и шаблонами), Секретарь (работа с расписанием и отчётами), Преподаватель (просмотр расписания и студентов) и Студент (ограниченный доступ для просмотра).',
  },
  {
    question: 'Как загрузить документы?',
    answer:
      'Перейдите в раздел "Документы" и нажмите кнопку "Загрузить документ". Поддерживаются форматы PDF, DOC, DOCX, XLS, XLSX. Максимальный размер файла — 50 МБ.',
  },
  {
    question: 'Можно ли восстановить удалённые документы?',
    answer:
      'Да, все удалённые документы перемещаются в архив, где хранятся 30 дней. В течение этого времени их можно восстановить через раздел "Архив".',
  },
  {
    question: 'Как настроить уведомления?',
    answer:
      'В разделе "Настройки" профиля вы можете выбрать типы уведомлений (email, push) и события, о которых хотите получать оповещения: новые документы, приближающиеся дедлайны, изменения в расписании.',
  },
  {
    question: 'Безопасны ли мои данные?',
    answer:
      'Да, мы используем современные методы шифрования (TLS 1.3, AES-256), двухфакторную аутентификацию и регулярное резервное копирование. Все данные хранятся на защищённых серверах в соответствии с требованиями законодательства.',
  },
]

// Roles data
const ROLES_DATA = [
  {
    icon: <Shield className="h-6 w-6" />,
    title: 'Администратор',
    description: 'Полный контроль над системой',
    features: ['Управление пользователями', 'Настройка системы', 'Все операции'],
    color: 'from-red-500 to-orange-500',
  },
  {
    icon: <BookOpen className="h-6 w-6" />,
    title: 'Методист',
    description: 'Работа с документами и шаблонами',
    features: ['Создание документов', 'Управление шаблонами', 'Генерация отчётов'],
    color: 'from-blue-500 to-cyan-500',
  },
  {
    icon: <UserCog className="h-6 w-6" />,
    title: 'Секретарь',
    description: 'Организация учебного процесса',
    features: ['Расписание занятий', 'Учёт посещаемости', 'Работа со студентами'],
    color: 'from-purple-500 to-pink-500',
  },
  {
    icon: <GraduationCap className="h-6 w-6" />,
    title: 'Преподаватель',
    description: 'Доступ к учебным материалам',
    features: ['Просмотр расписания', 'Списки студентов', 'Календарь событий'],
    color: 'from-green-500 to-emerald-500',
  },
  {
    icon: <Users className="h-6 w-6" />,
    title: 'Студент',
    description: 'Просмотр информации',
    features: ['Личное расписание', 'Документы', 'Объявления'],
    color: 'from-amber-500 to-yellow-500',
  },
]

// Slide names for navigation
const SLIDES = [
  { id: 'hero', name: 'Главная' },
  { id: 'modules', name: 'Модули' },
  { id: 'stats', name: 'Статистика' },
  { id: 'roles', name: 'Роли' },
  { id: 'faq', name: 'FAQ' },
  { id: 'cta', name: 'Начать' },
]

// Main Dashboard Component
const SecretaryMethodistDashboard = () => {
  const router = useRouter()
  const { isAuthenticated } = useAuthCheck()
  const [selectedModule, setSelectedModule] = React.useState<(typeof MODULES_DATA)[0] | null>(null)
  const [openFaqIndex, setOpenFaqIndex] = React.useState<number | null>(null)
  const [currentSlide, setCurrentSlide] = React.useState(0)
  const carouselRef = React.useRef<HTMLDivElement>(null)

  const scrollToSlide = (index: number) => {
    if (carouselRef.current) {
      const slideWidth = carouselRef.current.offsetWidth
      carouselRef.current.scrollTo({
        left: slideWidth * index,
        behavior: 'smooth',
      })
      setCurrentSlide(index)
    }
  }

  const handleScroll = () => {
    if (carouselRef.current) {
      const slideWidth = carouselRef.current.offsetWidth
      const scrollLeft = carouselRef.current.scrollLeft
      const newSlide = Math.round(scrollLeft / slideWidth)
      if (newSlide !== currentSlide) {
        setCurrentSlide(newSlide)
      }
    }
  }

  const nextSlide = () => {
    if (currentSlide < SLIDES.length - 1) {
      scrollToSlide(currentSlide + 1)
    }
  }

  const prevSlide = () => {
    if (currentSlide > 0) {
      scrollToSlide(currentSlide - 1)
    }
  }

  return (
    <div className="h-screen overflow-hidden flex flex-col">
      {/* Top Navigation - Fixed */}
      <div className="flex-shrink-0 bg-background/80 backdrop-blur-md border-b border-border/40">
        <div className="flex items-center justify-between px-4 py-4 sm:px-8 sm:py-5">
          {/* Slide indicators */}
          <div className="hidden sm:flex items-center gap-3">
            {SLIDES.map((slide, index) => (
              <button
                key={slide.id}
                onClick={() => scrollToSlide(index)}
                className={cn(
                  'px-4 py-2 rounded-full text-sm font-medium transition-all',
                  currentSlide === index
                    ? 'bg-gray-900 text-white dark:bg-white dark:text-gray-900'
                    : 'text-gray-500 hover:text-gray-900 dark:hover:text-white hover:bg-gray-100 dark:hover:bg-gray-800'
                )}
              >
                {slide.name}
              </button>
            ))}
          </div>
          {/* Mobile slide indicator */}
          <div className="sm:hidden flex items-center gap-1">
            {SLIDES.map((_, index) => (
              <button
                key={index}
                onClick={() => scrollToSlide(index)}
                className={cn(
                  'w-2 h-2 rounded-full transition-all',
                  currentSlide === index
                    ? 'bg-gray-900 dark:bg-white w-6'
                    : 'bg-gray-300 dark:bg-gray-600'
                )}
              />
            ))}
          </div>
          <div className="flex items-center gap-3">
            {isAuthenticated ? (
              <UserMenu />
            ) : (
              <Button onClick={() => router.push('/login')} variant="default" size="sm">
                Войти
              </Button>
            )}
            <ThemeSettingsPopover />
          </div>
        </div>
      </div>

      {/* Carousel Container */}
      <div className="flex-1 relative overflow-hidden">
        {/* Navigation Arrows */}
        <button
          onClick={prevSlide}
          disabled={currentSlide === 0}
          className={cn(
            'absolute left-4 sm:left-6 top-1/2 -translate-y-1/2 z-10 p-3 sm:p-4 rounded-full bg-white dark:bg-gray-800 shadow-lg transition-all border border-gray-200 dark:border-gray-700',
            currentSlide === 0 ? 'opacity-30 cursor-not-allowed' : 'hover:scale-110 hover:shadow-xl'
          )}
        >
          <ChevronLeft className="h-6 w-6 sm:h-7 sm:w-7" />
        </button>
        <button
          onClick={nextSlide}
          disabled={currentSlide === SLIDES.length - 1}
          className={cn(
            'absolute right-4 sm:right-6 top-1/2 -translate-y-1/2 z-10 p-3 sm:p-4 rounded-full bg-white dark:bg-gray-800 shadow-lg transition-all border border-gray-200 dark:border-gray-700',
            currentSlide === SLIDES.length - 1
              ? 'opacity-30 cursor-not-allowed'
              : 'hover:scale-110 hover:shadow-xl'
          )}
        >
          <ChevronRight className="h-6 w-6 sm:h-7 sm:w-7" />
        </button>

        {/* Slides */}
        <div
          ref={carouselRef}
          onScroll={handleScroll}
          className="h-full flex overflow-x-auto snap-x snap-mandatory scrollbar-hide"
          style={{ scrollbarWidth: 'none', msOverflowStyle: 'none' }}
        >
          {/* Slide 1: Hero */}
          <div className="flex-shrink-0 w-full h-full snap-start overflow-y-auto">
            <div className="min-h-full flex flex-col items-center justify-center p-8">
              <div className="max-w-4xl text-center space-y-8">
                <h1 className="text-4xl sm:text-5xl lg:text-6xl font-bold text-gray-900 dark:text-white leading-tight">
                  Информационная система
                  <br />
                  секретаря-методиста
                </h1>
                <p className="text-lg sm:text-xl text-gray-600 dark:text-gray-300 max-w-2xl mx-auto">
                  Современная панель управления для учебной части и управления документами
                </p>
                <div className="flex flex-col sm:flex-row items-center justify-center gap-4">
                  <Button
                    onClick={() => router.push('/register')}
                    size="lg"
                    className="w-full sm:w-auto px-8"
                  >
                    Начать работу
                  </Button>
                  <Button
                    onClick={() => scrollToSlide(1)}
                    variant="outline"
                    size="lg"
                    className="w-full sm:w-auto px-8"
                  >
                    Узнать больше
                  </Button>
                </div>
              </div>
            </div>
          </div>

          {/* Slide 2: Modules */}
          <div className="flex-shrink-0 w-full h-full snap-start overflow-y-auto">
            <div className="min-h-full flex flex-col justify-center p-6 sm:p-8 lg:p-12">
              <div className="max-w-7xl mx-auto w-full space-y-6">
                <div className="text-center space-y-2">
                  <h2 className="text-2xl sm:text-3xl font-bold text-gray-900 dark:text-white">
                    Возможности системы
                  </h2>
                  <p className="text-gray-600 dark:text-gray-400">
                    Нажмите на карточку, чтобы узнать больше о модуле
                  </p>
                </div>
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 sm:gap-6">
                  {MODULES_DATA.map((item, index) => (
                    <button
                      key={index}
                      onClick={() => setSelectedModule(item)}
                      className="group relative overflow-hidden rounded-2xl p-5 sm:p-6 bg-white dark:bg-black/95 border border-gray-200 dark:border-gray-700 transition-all duration-300 hover:scale-[1.02] hover:shadow-xl hover:bg-gray-50 dark:hover:bg-black cursor-pointer text-left"
                    >
                      <GlowingEffect
                        spread={40}
                        glow={true}
                        disabled={false}
                        proximity={64}
                        inactiveZone={0.01}
                        borderWidth={3}
                      />
                      <div className="relative z-10 space-y-3">
                        <div className="flex items-center justify-between">
                          <div className="flex h-10 w-10 sm:h-12 sm:w-12 items-center justify-center rounded-lg bg-gray-100 dark:bg-white/10 text-gray-900 dark:text-white transition-all duration-300 group-hover:scale-110 group-hover:bg-gray-200 dark:group-hover:bg-white/20">
                            {item.icon}
                          </div>
                          <span
                            className={cn(
                              'text-xs px-2 py-1 rounded-full font-medium',
                              item.details.status === 'Доступен'
                                ? 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400'
                                : 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-400'
                            )}
                          >
                            {item.details.status}
                          </span>
                        </div>
                        <div>
                          <h3 className="text-lg sm:text-xl font-semibold text-gray-900 dark:text-white mb-1 transition-colors duration-300 group-hover:text-gray-700 dark:group-hover:text-gray-100">
                            {item.title}
                          </h3>
                          <p className="text-sm text-gray-600 dark:text-gray-400 leading-relaxed transition-colors duration-300 group-hover:text-gray-800 dark:group-hover:text-gray-300">
                            {item.description}
                          </p>
                        </div>
                      </div>
                    </button>
                  ))}
                </div>
              </div>
            </div>
          </div>

          {/* Slide 3: Statistics */}
          <div className="flex-shrink-0 w-full h-full snap-start overflow-y-auto">
            <div className="min-h-full flex flex-col justify-center p-6 sm:p-8 lg:p-12">
              <div className="max-w-5xl mx-auto w-full">
                <div className="relative overflow-hidden rounded-2xl p-6 sm:p-8 lg:p-12 bg-white dark:bg-black/95 border border-gray-200 dark:border-gray-700">
                  <GlowingEffect
                    spread={60}
                    glow={true}
                    disabled={false}
                    proximity={80}
                    inactiveZone={0.01}
                    borderWidth={3}
                  />
                  <div className="relative z-10 space-y-8">
                    <div className="text-center space-y-2">
                      <h2 className="text-2xl sm:text-3xl font-bold text-gray-900 dark:text-white">
                        Наши достижения в цифрах
                      </h2>
                      <p className="text-gray-600 dark:text-gray-400 max-w-2xl mx-auto">
                        Система непрерывно развивается и помогает тысячам пользователей
                        оптимизировать рабочие процессы
                      </p>
                    </div>
                    <div className="grid grid-cols-2 lg:grid-cols-4 gap-4 sm:gap-6">
                      {STATS_DATA.map((stat, index) => (
                        <div
                          key={index}
                          className="text-center p-4 sm:p-6 rounded-xl bg-gray-50 dark:bg-white/5 border border-gray-200 dark:border-white/10 hover:bg-gray-100 dark:hover:bg-white/10 transition-all duration-300 hover:scale-105"
                        >
                          <Counter
                            end={stat.value}
                            suffix={stat.suffix}
                            className="text-2xl sm:text-3xl lg:text-4xl text-gray-900 dark:text-white mb-2"
                            duration={2500}
                          />
                          <p className="text-xs sm:text-sm text-gray-600 dark:text-gray-400">
                            {stat.label}
                          </p>
                        </div>
                      ))}
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>

          {/* Slide 4: Roles */}
          <div className="flex-shrink-0 w-full h-full snap-start overflow-y-auto">
            <div className="min-h-full flex flex-col justify-center p-6 sm:p-8 lg:p-12">
              <div className="max-w-7xl mx-auto w-full space-y-6">
                <div className="text-center space-y-2">
                  <h2 className="text-2xl sm:text-3xl font-bold text-gray-900 dark:text-white">
                    Роли в системе
                  </h2>
                  <p className="text-gray-600 dark:text-gray-400 max-w-2xl mx-auto">
                    Каждый пользователь получает доступ к функциям согласно своей роли
                  </p>
                </div>
                <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-5 gap-4">
                  {ROLES_DATA.map((role, index) => (
                    <div
                      key={index}
                      className="group relative overflow-hidden rounded-2xl p-5 sm:p-6 bg-white dark:bg-black/95 border border-gray-200 dark:border-gray-700 transition-all duration-300 hover:scale-[1.02] hover:shadow-xl"
                    >
                      <GlowingEffect
                        spread={40}
                        glow={true}
                        disabled={false}
                        proximity={64}
                        inactiveZone={0.01}
                        borderWidth={3}
                      />
                      <div className="relative z-10 space-y-3">
                        <div
                          className={cn(
                            'flex h-10 w-10 sm:h-12 sm:w-12 items-center justify-center rounded-xl bg-gradient-to-br text-white transition-transform duration-300 group-hover:scale-110',
                            role.color
                          )}
                        >
                          {role.icon}
                        </div>
                        <div>
                          <h3 className="text-base sm:text-lg font-semibold text-gray-900 dark:text-white mb-1">
                            {role.title}
                          </h3>
                          <p className="text-xs sm:text-sm text-gray-600 dark:text-gray-400 mb-2">
                            {role.description}
                          </p>
                          <ul className="space-y-1">
                            {role.features.map((feature, fIndex) => (
                              <li
                                key={fIndex}
                                className="flex items-center gap-2 text-xs text-gray-500 dark:text-gray-500"
                              >
                                <span className="h-1 w-1 rounded-full bg-gray-400" />
                                {feature}
                              </li>
                            ))}
                          </ul>
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            </div>
          </div>

          {/* Slide 5: FAQ */}
          <div className="flex-shrink-0 w-full h-full snap-start overflow-y-auto">
            <div className="min-h-full flex flex-col justify-center p-6 sm:p-8 lg:p-12">
              <div className="max-w-3xl mx-auto w-full space-y-6">
                <div className="text-center space-y-2">
                  <h2 className="text-2xl sm:text-3xl font-bold text-gray-900 dark:text-white">
                    Часто задаваемые вопросы
                  </h2>
                  <p className="text-gray-600 dark:text-gray-400">
                    Ответы на популярные вопросы о работе системы
                  </p>
                </div>
                <div className="space-y-3">
                  {FAQ_DATA.map((faq, index) => (
                    <div
                      key={index}
                      className="relative overflow-hidden rounded-xl bg-white dark:bg-black/95 border border-gray-200 dark:border-gray-700 transition-all duration-300"
                    >
                      <GlowingEffect
                        spread={30}
                        glow={true}
                        disabled={false}
                        proximity={50}
                        inactiveZone={0.01}
                        borderWidth={2}
                      />
                      <button
                        onClick={() => setOpenFaqIndex(openFaqIndex === index ? null : index)}
                        className="relative z-10 w-full flex items-center justify-between p-4 sm:p-5 text-left hover:bg-gray-50 dark:hover:bg-white/5 transition-colors"
                      >
                        <span className="font-medium text-sm sm:text-base text-gray-900 dark:text-white pr-4">
                          {faq.question}
                        </span>
                        <ChevronDown
                          className={cn(
                            'h-5 w-5 text-gray-500 transition-transform duration-300 flex-shrink-0',
                            openFaqIndex === index && 'rotate-180'
                          )}
                        />
                      </button>
                      <div
                        className={cn(
                          'relative z-10 overflow-hidden transition-all duration-300',
                          openFaqIndex === index ? 'max-h-96 opacity-100' : 'max-h-0 opacity-0'
                        )}
                      >
                        <p className="px-4 sm:px-5 pb-4 sm:pb-5 text-sm text-gray-600 dark:text-gray-400 leading-relaxed">
                          {faq.answer}
                        </p>
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            </div>
          </div>

          {/* Slide 6: CTA */}
          <div className="flex-shrink-0 w-full h-full snap-start overflow-y-auto">
            <div className="min-h-full flex flex-col items-center justify-center p-6 sm:p-8 lg:p-12">
              <div className="max-w-3xl w-full">
                <div className="relative overflow-hidden rounded-2xl p-8 sm:p-12 bg-white dark:bg-black/95 border border-gray-200 dark:border-gray-700">
                  <GlowingEffect
                    spread={80}
                    glow={true}
                    disabled={false}
                    proximity={100}
                    inactiveZone={0.01}
                    borderWidth={3}
                  />
                  <div className="relative z-10 text-center space-y-6">
                    <h2 className="text-2xl sm:text-3xl lg:text-4xl font-bold text-gray-900 dark:text-white">
                      Готовы начать работу?
                    </h2>
                    <p className="text-base sm:text-lg text-gray-600 dark:text-gray-400 max-w-2xl mx-auto">
                      Присоединяйтесь к тысячам пользователей, которые уже оптимизировали свои
                      рабочие процессы с помощью нашей системы
                    </p>
                    <div className="flex flex-col sm:flex-row items-center justify-center gap-4">
                      <Button
                        onClick={() => router.push('/register')}
                        size="lg"
                        className="w-full sm:w-auto font-semibold px-8"
                      >
                        Зарегистрироваться бесплатно
                      </Button>
                      <Button
                        onClick={() => router.push('/login')}
                        variant="outline"
                        size="lg"
                        className="w-full sm:w-auto font-semibold px-8"
                      >
                        Войти в систему
                      </Button>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* Module Details Dialog */}
      <Dialog open={!!selectedModule} onOpenChange={() => setSelectedModule(null)}>
        <DialogContent className="sm:max-w-[500px]">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-3">
              {selectedModule && (
                <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-gray-100 dark:bg-white/10 text-gray-900 dark:text-white">
                  {selectedModule.icon}
                </div>
              )}
              {selectedModule?.title}
            </DialogTitle>
            <DialogDescription>{selectedModule?.description}</DialogDescription>
          </DialogHeader>
          {selectedModule && (
            <div className="space-y-4">
              <div className="flex items-center gap-2">
                <span className="text-sm font-medium text-gray-700 dark:text-gray-300">
                  Статус:
                </span>
                <span
                  className={cn(
                    'text-xs px-2 py-1 rounded-full font-medium',
                    selectedModule.details.status === 'Доступен'
                      ? 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400'
                      : 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-400'
                  )}
                >
                  {selectedModule.details.status}
                </span>
              </div>
              <div>
                <h4 className="text-sm font-semibold text-gray-900 dark:text-white mb-2">
                  Возможности модуля:
                </h4>
                <ul className="space-y-2">
                  {selectedModule.details.features.map((feature, index) => (
                    <li
                      key={index}
                      className="flex items-start gap-2 text-sm text-gray-600 dark:text-gray-400"
                    >
                      <span className="text-green-500 mt-0.5">✓</span>
                      {feature}
                    </li>
                  ))}
                </ul>
              </div>
              {selectedModule.details.status === 'Доступен' && (
                <Button
                  className="w-full mt-4"
                  onClick={() => {
                    setSelectedModule(null)
                    router.push('/calendar')
                  }}
                >
                  Перейти к модулю
                </Button>
              )}
            </div>
          )}
        </DialogContent>
      </Dialog>
    </div>
  )
}

export default SecretaryMethodistDashboard
