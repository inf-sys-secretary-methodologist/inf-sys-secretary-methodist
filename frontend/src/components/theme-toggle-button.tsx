"use client"

import * as React from "react"
import { Moon, Sun } from "lucide-react"
import { useTheme } from "@/hooks/use-theme"
import { cn } from "@/lib/utils"

export function ThemeToggleButton() {
  const { resolvedTheme, toggleTheme } = useTheme()
  const [mounted, setMounted] = React.useState(false)

  React.useEffect(() => {
    setMounted(true)
  }, [])

  // Render a placeholder with the same dimensions to prevent layout shift
  if (!mounted) {
    return (
      <div
        className="flex w-16 h-8 p-1 rounded-full bg-zinc-200 dark:bg-zinc-800 border border-zinc-300 dark:border-zinc-700 opacity-50"
        aria-hidden="true"
      />
    )
  }

  const isDark = resolvedTheme === "dark"

  return (
    <div
      className={cn(
        "flex w-16 h-8 p-1 rounded-full cursor-pointer transition-all duration-300",
        "hover:scale-105 hover:shadow-lg active:scale-95",
        "focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-offset-2",
        isDark
          ? "bg-zinc-950 border border-zinc-800 hover:border-zinc-700 focus-visible:ring-zinc-600"
          : "bg-white border border-zinc-200 hover:border-zinc-300 hover:shadow-md focus-visible:ring-blue-500"
      )}
      onClick={toggleTheme}
      role="button"
      tabIndex={0}
      aria-label={`Switch to ${isDark ? "light" : "dark"} theme`}
      onKeyDown={(e) => {
        if (e.key === "Enter" || e.key === " ") {
          e.preventDefault()
          toggleTheme()
        }
      }}
    >
      <div className="flex justify-between items-center w-full">
        <div
          className={cn(
            "flex justify-center items-center w-6 h-6 rounded-full",
            "transition-all duration-300 ease-in-out",
            "group-hover:scale-110",
            isDark
              ? "transform translate-x-0 bg-zinc-800 shadow-md"
              : "transform translate-x-8 bg-gray-200 shadow-sm"
          )}
        >
          {isDark ? (
            <Moon
              className={cn(
                "w-4 h-4 text-white transition-all duration-300",
                "animate-in fade-in zoom-in"
              )}
              strokeWidth={1.5}
            />
          ) : (
            <Sun
              className={cn(
                "w-4 h-4 text-gray-700 transition-all duration-300",
                "animate-in fade-in zoom-in spin-in-0"
              )}
              strokeWidth={1.5}
            />
          )}
        </div>
        <div
          className={cn(
            "flex justify-center items-center w-6 h-6 rounded-full",
            "transition-all duration-300 ease-in-out",
            isDark
              ? "bg-transparent"
              : "transform -translate-x-8"
          )}
        >
          {isDark ? (
            <Sun
              className={cn(
                "w-4 h-4 text-gray-500 transition-all duration-300",
                "opacity-50 hover:opacity-70"
              )}
              strokeWidth={1.5}
            />
          ) : (
            <Moon
              className={cn(
                "w-4 h-4 text-black transition-all duration-300",
                "opacity-50 hover:opacity-70"
              )}
              strokeWidth={1.5}
            />
          )}
        </div>
      </div>
    </div>
  )
}
