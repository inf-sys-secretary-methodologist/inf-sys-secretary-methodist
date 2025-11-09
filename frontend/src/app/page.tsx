"use client"

import * as React from "react"
import { FileText, Users, Calendar, TrendingUp, Plus, Globe, BarChart3, ArrowUp } from "lucide-react"
import { cn } from "@/lib/utils"
import { GlowingEffect } from "@/components/ui/glowing-effect"
import { ThemeToggleButton } from "@/components/theme-toggle-button"
import { TubelightNavbar } from "@/components/ui/tubelight-navbar"
import { navigationItems } from "@/config/navigation"

// Simple Counter Component
interface CounterProps {
  end: number
  className?: string
  fontSize?: number
}

const Counter = ({ end, className, fontSize = 30 }: CounterProps) => {
  return (
    <div style={{ fontSize }} className={cn("font-bold", className)}>
      {end}
    </div>
  )
}

// Feature Card Component
interface FeatureCardProps extends React.HTMLAttributes<HTMLDivElement> {
  title: string
  description: string
  icon: React.ReactNode
}

const FeatureCard = React.forwardRef<HTMLDivElement, FeatureCardProps>(
  ({ className, title, description, icon, ...props }, ref) => {
    return (
      <div
        ref={ref}
        className={cn(
          "relative flex flex-col overflow-hidden rounded-2xl p-6 transition-all",
          "bg-[#C5CBE3] border-2 border-[#D79922]",
          "dark:bg-[#312e81] dark:border-[#a5b4fc]",
          className
        )}
        {...props}
      >
        <GlowingEffect spread={40} glow={true} disabled={false} proximity={64} inactiveZone={0.01} borderWidth={3} />
        <div className="mb-4 flex h-12 w-12 items-center justify-center rounded-full border-2 bg-[#D79922] text-white border-[#D79922] dark:bg-[#c4b5fd] dark:text-[#1e1b4b] dark:border-[#c4b5fd]">
          {icon}
        </div>
        <h3 className="text-lg font-semibold mb-2 text-[#4056A1] dark:text-white">{title}</h3>
        <p className="text-sm leading-relaxed text-[#2e3b72] dark:text-[#e0e7ff]">{description}</p>
      </div>
    )
  }
)

FeatureCard.displayName = "FeatureCard"

// Stat Card Component
interface StatCardProps extends React.HTMLAttributes<HTMLDivElement> {
  title: string
  value: number
  icon: React.ReactNode
  trend?: string
}

const StatCard = React.forwardRef<HTMLDivElement, StatCardProps>(
  ({ className, title, value, icon, trend, ...props }, ref) => {
    return (
      <div
        ref={ref}
        className={cn(
          "relative overflow-hidden rounded-2xl p-6 transition-all",
          "border-2 border-[#D79922] bg-gradient-to-br from-[#4056A1] to-[#2e3b72]",
          "dark:border-[#818cf8] dark:bg-[#1e1b4b]",
          className
        )}
        {...props}
      >
        <GlowingEffect spread={40} glow={true} disabled={false} proximity={64} inactiveZone={0.01} borderWidth={3} />
        <div className="relative z-10">
          <div className="flex items-center justify-between mb-4">
            <div className="flex h-10 w-10 items-center justify-center rounded-lg border-2 bg-[#D79922]/30 text-[#D79922] border-[#D79922] dark:bg-[#818cf8] dark:text-white dark:border-[#818cf8]">
              {icon}
            </div>
            {trend && (
              <span className="text-xs px-2 py-1 rounded-full font-medium border bg-green-500/20 text-green-300 border-green-500/50 dark:bg-[#fbbf24]/20 dark:text-[#fbbf24] dark:border-[#fbbf24]/50">
                {trend}
              </span>
            )}
          </div>
          <h3 className="text-sm font-medium mb-2 text-gray-300 dark:text-[#e0e7ff]">{title}</h3>
          <div className="text-3xl font-bold text-white dark:text-white">
            <Counter end={value} fontSize={32} />
          </div>
        </div>
      </div>
    )
  }
)

StatCard.displayName = "StatCard"

// Main Dashboard Component
const SecretaryMethodistDashboard = () => {
  return (
    <div className="min-h-screen bg-background p-8">
      {/* Navigation Bar - Fixed Position at Top */}
      <TubelightNavbar items={navigationItems} />

      {/* Theme Toggle - Fixed Position with isolation */}
      <div className="fixed top-8 right-8 z-50 pointer-events-auto" style={{ isolation: 'isolate' }}>
        <ThemeToggleButton />
      </div>

      <div className="max-w-7xl mx-auto space-y-8 pt-20">
        {/* Header */}
        <div className="text-center space-y-2">
          <h1 className="text-4xl font-bold text-gray-900 dark:text-white">Информационная система секретаря-методиста</h1>
          <p className="text-sm text-gray-600 dark:text-gray-300">Современная панель управления для учебной части и управления документами</p>
        </div>

        {/* Search Input Bar */}
        <div className="relative overflow-hidden rounded-2xl bg-white dark:bg-gradient-to-br dark:from-gray-900 dark:to-black border border-gray-200 dark:border-gray-800">
          <GlowingEffect spread={40} glow={true} disabled={false} proximity={64} inactiveZone={0.01} borderWidth={3} />
          <div className="relative z-10 p-4">
            <div className="flex items-center gap-3">
              {/* Plus Button */}
              <button className="flex items-center justify-center w-10 h-10 rounded-lg bg-gray-100 hover:bg-gray-200 dark:bg-gray-800 dark:hover:bg-gray-700 text-gray-700 dark:text-gray-300 hover:text-gray-900 dark:hover:text-white transition-all duration-200 hover:scale-105 active:scale-95">
                <Plus className="w-5 h-5" />
              </button>

              {/* Input Field */}
              <input
                type="text"
                placeholder="Попросите создать дашборд для..."
                className="flex-1 bg-transparent text-gray-900 dark:text-white placeholder:text-gray-500 dark:placeholder:text-gray-400 text-base outline-none px-2"
              />

              {/* Right Side Buttons */}
              <div className="flex items-center gap-2">
                {/* Public Badge */}
                <div className="flex items-center gap-2 px-3 py-1.5 rounded-lg bg-gray-100 dark:bg-gray-800 text-gray-700 dark:text-gray-300">
                  <Globe className="w-4 h-4" />
                  <span className="text-sm font-medium">Public</span>
                </div>

                {/* Analytics Button */}
                <button className="flex items-center justify-center w-10 h-10 rounded-lg bg-gray-100 hover:bg-gray-200 dark:bg-gray-800 dark:hover:bg-gray-700 text-gray-700 dark:text-gray-300 hover:text-gray-900 dark:hover:text-white transition-all duration-200 hover:scale-105 active:scale-95">
                  <BarChart3 className="w-5 h-5" />
                </button>

                {/* Submit Button */}
                <button className="flex items-center justify-center w-10 h-10 rounded-lg bg-gray-100 hover:bg-gray-200 dark:bg-gray-800 dark:hover:bg-gray-700 text-gray-700 dark:text-gray-300 hover:text-gray-900 dark:hover:text-white transition-all duration-200 hover:scale-105 active:scale-95">
                  <ArrowUp className="w-5 h-5" />
                </button>
              </div>
            </div>
          </div>
        </div>

        {/* Animated Glowing Cards Grid */}
        <div className="space-y-4">
          <h2 className="text-2xl font-semibold text-gray-900 dark:text-white">Дополнительные модули</h2>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {[
              {
                icon: <FileText className="h-6 w-6" />,
                title: "Система архивирования",
                description: "Автоматическое архивирование документов с возможностью восстановления"
              },
              {
                icon: <Users className="h-6 w-6" />,
                title: "Отчеты по посещаемости",
                description: "Детальная статистика посещений и участия студентов"
              },
              {
                icon: <Calendar className="h-6 w-6" />,
                title: "Календарь событий",
                description: "Интерактивный календарь с напоминаниями о важных событиях"
              },
              {
                icon: <TrendingUp className="h-6 w-6" />,
                title: "Финансовая аналитика",
                description: "Отслеживание финансовых показателей и генерация отчетов"
              },
              {
                icon: <FileText className="h-6 w-6" />,
                title: "Управление проектами",
                description: "Планирование и контроль выполнения учебных проектов"
              },
              {
                icon: <Users className="h-6 w-6" />,
                title: "Коммуникационный центр",
                description: "Централизованная система для общения с преподавателями и студентами"
              }
            ].map((item, index) => (
              <div
                key={index}
                className="group relative overflow-hidden rounded-2xl p-6 bg-white dark:bg-black/95 border border-gray-200 dark:border-gray-700 transition-all duration-300 hover:scale-[1.02] hover:shadow-xl hover:bg-gray-50 dark:hover:bg-black cursor-pointer"
              >
                <GlowingEffect spread={40} glow={true} disabled={false} proximity={64} inactiveZone={0.01} borderWidth={3} />
                <div className="relative z-10 space-y-4">
                  <div className="flex h-12 w-12 items-center justify-center rounded-lg bg-gray-100 dark:bg-white/10 text-gray-900 dark:text-white transition-all duration-300 group-hover:scale-110 group-hover:bg-gray-200 dark:group-hover:bg-white/20">
                    {item.icon}
                  </div>
                  <div>
                    <h3 className="text-xl font-semibold text-gray-900 dark:text-white mb-2 transition-colors duration-300 group-hover:text-gray-700 dark:group-hover:text-gray-100">{item.title}</h3>
                    <p className="text-sm text-gray-600 dark:text-gray-400 leading-relaxed transition-colors duration-300 group-hover:text-gray-800 dark:group-hover:text-gray-300">{item.description}</p>
                  </div>
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* Quick Actions */}
        <div className="relative overflow-hidden rounded-2xl p-8 bg-white dark:bg-black/95 border border-gray-200 dark:border-gray-700">
          <GlowingEffect spread={40} glow={true} disabled={false} proximity={64} inactiveZone={0.01} borderWidth={3} />
          <div className="relative z-10 space-y-6">
            <h2 className="text-2xl font-semibold text-gray-900 dark:text-white">Быстрые действия</h2>
            <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
              <button className="px-6 py-4 rounded-lg font-medium transition-all duration-300 bg-white dark:bg-white text-gray-900 hover:bg-gray-900 dark:hover:bg-gray-900 hover:text-white dark:hover:text-white border border-gray-200 hover:border-gray-900 dark:hover:border-gray-700 hover:scale-105 active:scale-95 hover:shadow-lg">
                Загрузить документ
              </button>
              <button className="px-6 py-4 rounded-lg font-medium transition-all duration-300 bg-white dark:bg-white text-gray-900 hover:bg-gray-900 dark:hover:bg-gray-900 hover:text-white dark:hover:text-white border border-gray-200 hover:border-gray-900 dark:hover:border-gray-700 hover:scale-105 active:scale-95 hover:shadow-lg">
                Добавить студента
              </button>
              <button className="px-6 py-4 rounded-lg font-medium transition-all duration-300 bg-white dark:bg-white text-gray-900 hover:bg-gray-900 dark:hover:bg-gray-900 hover:text-white dark:hover:text-white border border-gray-200 hover:border-gray-900 dark:hover:border-gray-700 hover:scale-105 active:scale-95 hover:shadow-lg">
                Создать мероприятие
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}

export default SecretaryMethodistDashboard
