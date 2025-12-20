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
import { useTranslations } from 'next-intl'
import { cn } from '@/lib/utils'
import { GlowingEffect } from '@/components/ui/glowing-effect-lazy'
import { ThemeSettingsPopover } from '@/components/theme-settings-popover'
import { LanguageSwitcher } from '@/components/language-switcher'
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

// Module keys for translation
const MODULE_KEYS = [
  { key: 'documents', icon: <FileText className="h-6 w-6" />, available: true },
  { key: 'notifications', icon: <Bell className="h-6 w-6" />, available: true },
  { key: 'calendar', icon: <Calendar className="h-6 w-6" />, available: true },
  { key: 'integration', icon: <RefreshCw className="h-6 w-6" />, available: true },
  { key: 'personalization', icon: <Palette className="h-6 w-6" />, available: true },
  { key: 'search', icon: <Search className="h-6 w-6" />, available: true },
  { key: 'archive', icon: <Archive className="h-6 w-6" />, available: false },
  { key: 'reports', icon: <BarChart3 className="h-6 w-6" />, available: false },
  { key: 'projects', icon: <Briefcase className="h-6 w-6" />, available: false },
  { key: 'communication', icon: <MessageSquare className="h-6 w-6" />, available: false },
]

// Role keys for translation
const ROLE_KEYS = [
  { key: 'admin', icon: <Shield className="h-6 w-6" />, color: 'from-red-500 to-orange-500' },
  { key: 'methodist', icon: <BookOpen className="h-6 w-6" />, color: 'from-blue-500 to-cyan-500' },
  { key: 'secretary', icon: <UserCog className="h-6 w-6" />, color: 'from-purple-500 to-pink-500' },
  {
    key: 'teacher',
    icon: <GraduationCap className="h-6 w-6" />,
    color: 'from-green-500 to-emerald-500',
  },
  { key: 'student', icon: <Users className="h-6 w-6" />, color: 'from-amber-500 to-yellow-500' },
]

// Slide keys for translation
const SLIDE_KEYS = ['hero', 'modules', 'stats', 'roles', 'faq', 'cta'] as const

// FAQ keys for translation (6 questions)
const FAQ_KEYS = [0, 1, 2, 3, 4, 5] as const

// Main Dashboard Component
const SecretaryMethodistDashboard = () => {
  const router = useRouter()
  const { isAuthenticated } = useAuthCheck()
  const t = useTranslations('landing')
  const tAuth = useTranslations('auth')
  const [selectedModule, setSelectedModule] = React.useState<{
    key: string
    icon: React.ReactNode
    available: boolean
  } | null>(null)
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
    if (currentSlide < SLIDE_KEYS.length - 1) {
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
            {SLIDE_KEYS.map((slideKey, index) => (
              <button
                key={slideKey}
                onClick={() => scrollToSlide(index)}
                className={cn(
                  'px-4 py-2 rounded-full text-sm font-medium transition-all',
                  currentSlide === index
                    ? 'bg-gray-900 text-white dark:bg-white dark:text-gray-900'
                    : 'text-gray-500 hover:text-gray-900 dark:hover:text-white hover:bg-gray-100 dark:hover:bg-gray-800'
                )}
              >
                {t(`slides.${slideKey}`)}
              </button>
            ))}
          </div>
          {/* Mobile slide indicator */}
          <div className="sm:hidden flex items-center gap-1">
            {SLIDE_KEYS.map((_, index) => (
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
                {tAuth('login')}
              </Button>
            )}
            <LanguageSwitcher />
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
          disabled={currentSlide === SLIDE_KEYS.length - 1}
          className={cn(
            'absolute right-4 sm:right-6 top-1/2 -translate-y-1/2 z-10 p-3 sm:p-4 rounded-full bg-white dark:bg-gray-800 shadow-lg transition-all border border-gray-200 dark:border-gray-700',
            currentSlide === SLIDE_KEYS.length - 1
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
                  {t('title')}
                  <br />
                  {t('subtitle')}
                </h1>
                <p className="text-lg sm:text-xl text-gray-600 dark:text-gray-300 max-w-2xl mx-auto">
                  {t('description')}
                </p>
                <div className="flex flex-col sm:flex-row items-center justify-center gap-4">
                  <Button
                    onClick={() => router.push('/register')}
                    size="lg"
                    className="w-full sm:w-auto px-8"
                  >
                    {t('getStarted')}
                  </Button>
                  <Button
                    onClick={() => scrollToSlide(1)}
                    variant="outline"
                    size="lg"
                    className="w-full sm:w-auto px-8"
                  >
                    {t('learnMore')}
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
                    {t('modules.title')}
                  </h2>
                  <p className="text-gray-600 dark:text-gray-400">{t('modules.subtitle')}</p>
                </div>
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 sm:gap-6">
                  {MODULE_KEYS.map((item, index) => (
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
                              item.available
                                ? 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400'
                                : 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-400'
                            )}
                          >
                            {item.available ? t('modules.available') : t('modules.inDevelopment')}
                          </span>
                        </div>
                        <div>
                          <h3 className="text-lg sm:text-xl font-semibold text-gray-900 dark:text-white mb-1 transition-colors duration-300 group-hover:text-gray-700 dark:group-hover:text-gray-100">
                            {t(`modules.${item.key}.title`)}
                          </h3>
                          <p className="text-sm text-gray-600 dark:text-gray-400 leading-relaxed transition-colors duration-300 group-hover:text-gray-800 dark:group-hover:text-gray-300">
                            {t(`modules.${item.key}.description`)}
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
                        {t('stats.title')}
                      </h2>
                      <p className="text-gray-600 dark:text-gray-400 max-w-2xl mx-auto">
                        {t('stats.subtitle')}
                      </p>
                    </div>
                    <div className="grid grid-cols-2 lg:grid-cols-4 gap-4 sm:gap-6">
                      {[
                        { value: 15000, labelKey: 'documentsProcessed', suffix: '+' },
                        { value: 2500, labelKey: 'activeUsers', suffix: '+' },
                        { value: 850, labelKey: 'eventsHeld', suffix: '+' },
                        { value: 99, labelKey: 'satisfaction', suffix: '%' },
                      ].map((stat, index) => (
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
                            {t(`stats.${stat.labelKey}`)}
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
                    {t('roleSection.title')}
                  </h2>
                  <p className="text-gray-600 dark:text-gray-400 max-w-2xl mx-auto">
                    {t('roleSection.subtitle')}
                  </p>
                </div>
                <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-5 gap-4">
                  {ROLE_KEYS.map((role, index) => (
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
                            {t(`roleSection.${role.key}.title`)}
                          </h3>
                          <p className="text-xs sm:text-sm text-gray-600 dark:text-gray-400 mb-2">
                            {t(`roleSection.${role.key}.description`)}
                          </p>
                          <ul className="space-y-1">
                            {(t.raw(`roleSection.${role.key}.features`) as string[]).map(
                              (feature, fIndex) => (
                                <li
                                  key={fIndex}
                                  className="flex items-center gap-2 text-xs text-gray-500 dark:text-gray-500"
                                >
                                  <span className="h-1 w-1 rounded-full bg-gray-400" />
                                  {feature}
                                </li>
                              )
                            )}
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
                    {t('faq.title')}
                  </h2>
                  <p className="text-gray-600 dark:text-gray-400">{t('faq.subtitle')}</p>
                </div>
                <div className="space-y-3">
                  {FAQ_KEYS.map((faqIndex) => (
                    <div
                      key={faqIndex}
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
                        onClick={() => setOpenFaqIndex(openFaqIndex === faqIndex ? null : faqIndex)}
                        className="relative z-10 w-full flex items-center justify-between p-4 sm:p-5 text-left hover:bg-gray-50 dark:hover:bg-white/5 transition-colors"
                      >
                        <span className="font-medium text-sm sm:text-base text-gray-900 dark:text-white pr-4">
                          {t(`faq.items.${faqIndex}.question`)}
                        </span>
                        <ChevronDown
                          className={cn(
                            'h-5 w-5 text-gray-500 transition-transform duration-300 flex-shrink-0',
                            openFaqIndex === faqIndex && 'rotate-180'
                          )}
                        />
                      </button>
                      <div
                        className={cn(
                          'relative z-10 overflow-hidden transition-all duration-300',
                          openFaqIndex === faqIndex ? 'max-h-96 opacity-100' : 'max-h-0 opacity-0'
                        )}
                      >
                        <p className="px-4 sm:px-5 pb-4 sm:pb-5 text-sm text-gray-600 dark:text-gray-400 leading-relaxed">
                          {t(`faq.items.${faqIndex}.answer`)}
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
                      {t('cta.title')}
                    </h2>
                    <p className="text-base sm:text-lg text-gray-600 dark:text-gray-400 max-w-2xl mx-auto">
                      {t('cta.subtitle')}
                    </p>
                    <div className="flex flex-col sm:flex-row items-center justify-center gap-4">
                      <Button
                        onClick={() => router.push('/register')}
                        size="lg"
                        className="w-full sm:w-auto font-semibold px-8"
                      >
                        {tAuth('registerFree')}
                      </Button>
                      <Button
                        onClick={() => router.push('/login')}
                        variant="outline"
                        size="lg"
                        className="w-full sm:w-auto font-semibold px-8"
                      >
                        {tAuth('loginToSystem')}
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
              {selectedModule && t(`modules.${selectedModule.key}.title`)}
            </DialogTitle>
            <DialogDescription>
              {selectedModule && t(`modules.${selectedModule.key}.description`)}
            </DialogDescription>
          </DialogHeader>
          {selectedModule && (
            <div className="space-y-4">
              <div className="flex items-center gap-2">
                <span className="text-sm font-medium text-gray-700 dark:text-gray-300">
                  {t('modules.features').replace(':', '')}:
                </span>
                <span
                  className={cn(
                    'text-xs px-2 py-1 rounded-full font-medium',
                    selectedModule.available
                      ? 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400'
                      : 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-400'
                  )}
                >
                  {selectedModule.available ? t('modules.available') : t('modules.inDevelopment')}
                </span>
              </div>
              {selectedModule.available && (
                <Button
                  className="w-full mt-4"
                  onClick={() => {
                    setSelectedModule(null)
                    router.push('/calendar')
                  }}
                >
                  {t('modules.goToModule')}
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
