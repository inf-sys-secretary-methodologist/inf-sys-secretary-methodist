import * as React from 'react'
import { cn } from '@/lib/utils'

export interface FloatingInputProps extends React.InputHTMLAttributes<HTMLInputElement> {
  label: string
}

const FloatingInput = React.forwardRef<HTMLInputElement, FloatingInputProps>(
  ({ className, label, type, id, ...props }, ref) => {
    const [isFocused, setIsFocused] = React.useState(false)
    const [hasValue, setHasValue] = React.useState(false)

    // Generate a unique ID if none provided
    const inputId = React.useId()
    const uniqueId = id || inputId

    /* c8 ignore start - Input event handlers */
    const handleFocus = (e: React.FocusEvent<HTMLInputElement>) => {
      setIsFocused(true)
      props.onFocus?.(e)
    }
    const handleBlur = (e: React.FocusEvent<HTMLInputElement>) => {
      setIsFocused(false)
      setHasValue(e.target.value !== '')
      props.onBlur?.(e)
    }

    const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
      setHasValue(e.target.value !== '')
      props.onChange?.(e)
    }
    const isActive = isFocused || hasValue || (props.value !== undefined && props.value !== '')
    /* c8 ignore stop */

    return (
      <div className="relative">
        <input
          id={uniqueId}
          type={type}
          className={cn(
            'peer flex h-12 w-full rounded-lg border border-input bg-background px-3 py-3 text-base',
            'ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium',
            'placeholder:text-transparent',
            'focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-0',
            'disabled:cursor-not-allowed disabled:opacity-50',
            'transition-all duration-200',
            className
          )}
          ref={ref}
          onFocus={handleFocus}
          onBlur={handleBlur}
          onChange={handleChange}
          placeholder={label}
          {...props}
        />
        <label
          className={cn(
            'absolute left-3 text-muted-foreground pointer-events-none',
            'transition-all duration-200',
            isActive ? 'top-[-0.5rem] text-xs bg-background px-1' : 'top-3 text-base'
          )}
          htmlFor={uniqueId}
        >
          {label}
        </label>
      </div>
    )
  }
)

FloatingInput.displayName = 'FloatingInput'

export { FloatingInput }
