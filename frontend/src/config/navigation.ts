import { Home, FileText, Users, Calendar, Settings } from "lucide-react"
import { LucideIcon } from "lucide-react"

export interface NavItem {
  name: string
  url: string
  icon: LucideIcon
}

export const navigationItems: NavItem[] = [
  {
    name: "Главная",
    url: "/",
    icon: Home,
  },
  {
    name: "Документы",
    url: "/documents",
    icon: FileText,
  },
  {
    name: "Студенты",
    url: "/students",
    icon: Users,
  },
  {
    name: "Расписание",
    url: "/schedule",
    icon: Calendar,
  },
  {
    name: "Настройки",
    url: "/settings",
    icon: Settings,
  },
]
