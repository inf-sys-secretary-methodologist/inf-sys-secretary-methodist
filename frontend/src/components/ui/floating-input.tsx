"use client"

import { Input } from "@/components/ui/input"
import { cn } from "@/lib/utils"
import { useId } from "react"

interface FloatingInputProps
  extends Omit<React.ComponentProps<typeof Input>, "id"> {
  label: string
  containerClassName?: string
}

export function FloatingInput({
  label,
  className,
  containerClassName,
  ...props
}: FloatingInputProps) {
  const id = useId()

  return (
    <div className={cn("group relative min-w-[200px]", containerClassName)}>
      <label
        htmlFor={id}
        className={cn(
          "origin-start absolute top-1/2 block -translate-y-1/2 cursor-text px-1 text-sm text-muted-foreground/70",
          "transition-all duration-200 ease-out",
          "group-focus-within:pointer-events-none group-focus-within:top-0 group-focus-within:cursor-default group-focus-within:text-xs group-focus-within:font-medium group-focus-within:text-foreground",
          "has-[+input:not(:placeholder-shown)]:pointer-events-none has-[+input:not(:placeholder-shown)]:top-0 has-[+input:not(:placeholder-shown)]:cursor-default has-[+input:not(:placeholder-shown)]:text-xs has-[+input:not(:placeholder-shown)]:font-medium has-[+input:not(:placeholder-shown)]:text-foreground"
        )}
      >
        <span className="inline-flex bg-background px-2">{label}</span>
      </label>
      <Input
        id={id}
        placeholder=" "
        className={cn(
          "peer",
          className
        )}
        {...props}
      />
    </div>
  )
}
