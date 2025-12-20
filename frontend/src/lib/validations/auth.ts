import { z } from 'zod'
import { UserRole } from '@/types/auth'

/**
 * Translation function type for validation messages
 */
type TranslationFn = (key: string) => string

/**
 * Creates login form validation schema with translations
 */
export const createLoginSchema = (t: TranslationFn) =>
  z.object({
    email: z
      .string({ message: t('emailRequired') })
      .min(1, t('emailRequired'))
      .email(t('emailInvalid')),
    password: z
      .string({ message: t('passwordRequired') })
      .min(1, t('passwordRequired'))
      .min(8, t('passwordMinLength')),
  })

export type LoginFormData = z.infer<ReturnType<typeof createLoginSchema>>

/**
 * Creates register form validation schema with translations
 */
export const createRegisterSchema = (t: TranslationFn) =>
  z
    .object({
      name: z
        .string({ message: t('nameRequired') })
        .min(2, t('nameMinLength'))
        .max(50, t('nameMaxLength'))
        .trim(),
      email: z
        .string({ message: t('emailRequired') })
        .min(1, t('emailRequired'))
        .email(t('emailInvalid'))
        .trim(),
      password: z
        .string({ message: t('passwordRequired') })
        .min(8, t('passwordMinLength'))
        .regex(/[A-Z]/, t('passwordUppercase'))
        .regex(/[a-z]/, t('passwordLowercase'))
        .regex(/[0-9]/, t('passwordDigit'))
        .regex(/[^a-zA-Z0-9]/, t('passwordSpecial'))
        .trim(),
      confirmPassword: z
        .string({ message: t('confirmPasswordRequired') })
        .min(1, t('confirmPassword')),
      role: z.nativeEnum(UserRole, {
        message: t('roleInvalid'),
      }),
    })
    .refine((data) => data.password === data.confirmPassword, {
      message: t('passwordsMismatch'),
      path: ['confirmPassword'],
    })

export type RegisterFormData = z.infer<ReturnType<typeof createRegisterSchema>>

/**
 * Creates password recovery validation schema with translations
 */
export const createPasswordRecoverySchema = (t: TranslationFn) =>
  z.object({
    email: z
      .string({ message: t('emailRequired') })
      .min(1, t('emailRequired'))
      .email(t('emailInvalid'))
      .trim(),
  })

export type PasswordRecoveryFormData = z.infer<ReturnType<typeof createPasswordRecoverySchema>>

/**
 * Creates password reset validation schema with translations
 */
export const createPasswordResetSchema = (t: TranslationFn) =>
  z
    .object({
      password: z
        .string({ message: t('passwordRequired') })
        .min(8, t('passwordMinLength'))
        .regex(/[A-Z]/, t('passwordUppercase'))
        .regex(/[a-z]/, t('passwordLowercase'))
        .regex(/[0-9]/, t('passwordDigit'))
        .regex(/[^a-zA-Z0-9]/, t('passwordSpecial'))
        .trim(),
      confirmPassword: z
        .string({ message: t('confirmPasswordRequired') })
        .min(1, t('confirmPassword')),
    })
    .refine((data) => data.password === data.confirmPassword, {
      message: t('passwordsMismatch'),
      path: ['confirmPassword'],
    })

export type PasswordResetFormData = z.infer<ReturnType<typeof createPasswordResetSchema>>
