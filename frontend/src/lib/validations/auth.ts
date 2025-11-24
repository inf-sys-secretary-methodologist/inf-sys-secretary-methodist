import { z } from 'zod'
import { UserRole } from '@/types/auth'

/**
 * Login form validation schema
 */
export const loginSchema = z.object({
  email: z
    .string({ message: 'Email обязателен' })
    .min(1, 'Email обязателен')
    .email('Неверный формат email'),
  password: z
    .string({ message: 'Пароль обязателен' })
    .min(1, 'Пароль обязателен')
    .min(8, 'Пароль должен содержать минимум 8 символов'),
})

export type LoginFormData = z.infer<typeof loginSchema>

/**
 * Register form validation schema
 */
export const registerSchema = z
  .object({
    name: z
      .string({ message: 'Имя обязательно' })
      .min(2, 'Имя должно содержать минимум 2 символа')
      .max(50, 'Имя не должно превышать 50 символов')
      .trim(),
    email: z
      .string({ message: 'Email обязателен' })
      .min(1, 'Email обязателен')
      .email('Неверный формат email')
      .trim(),
    password: z
      .string({ message: 'Пароль обязателен' })
      .min(8, 'Пароль должен содержать минимум 8 символов')
      .regex(/[A-Z]/, 'Пароль должен содержать хотя бы одну заглавную букву')
      .regex(/[a-z]/, 'Пароль должен содержать хотя бы одну строчную букву')
      .regex(/[0-9]/, 'Пароль должен содержать хотя бы одну цифру')
      .regex(
        /[^a-zA-Z0-9]/,
        'Пароль должен содержать хотя бы один специальный символ'
      )
      .trim(),
    confirmPassword: z
      .string({ message: 'Подтверждение пароля обязательно' })
      .min(1, 'Подтвердите пароль'),
    role: z.nativeEnum(UserRole, {
      message: 'Выберите корректную роль',
    }),
  })
  .refine((data) => data.password === data.confirmPassword, {
    message: 'Пароли не совпадают',
    path: ['confirmPassword'],
  })

export type RegisterFormData = z.infer<typeof registerSchema>

/**
 * Password recovery validation schema
 */
export const passwordRecoverySchema = z.object({
  email: z
    .string({ message: 'Email обязателен' })
    .min(1, 'Email обязателен')
    .email('Неверный формат email')
    .trim(),
})

export type PasswordRecoveryFormData = z.infer<typeof passwordRecoverySchema>

/**
 * Password reset validation schema
 */
export const passwordResetSchema = z
  .object({
    password: z
      .string({ message: 'Пароль обязателен' })
      .min(8, 'Пароль должен содержать минимум 8 символов')
      .regex(/[A-Z]/, 'Пароль должен содержать хотя бы одну заглавную букву')
      .regex(/[a-z]/, 'Пароль должен содержать хотя бы одну строчную букву')
      .regex(/[0-9]/, 'Пароль должен содержать хотя бы одну цифру')
      .regex(
        /[^a-zA-Z0-9]/,
        'Пароль должен содержать хотя бы один специальный символ'
      )
      .trim(),
    confirmPassword: z
      .string({ message: 'Подтверждение пароля обязательно' })
      .min(1, 'Подтвердите пароль'),
  })
  .refine((data) => data.password === data.confirmPassword, {
    message: 'Пароли не совпадают',
    path: ['confirmPassword'],
  })

export type PasswordResetFormData = z.infer<typeof passwordResetSchema>
