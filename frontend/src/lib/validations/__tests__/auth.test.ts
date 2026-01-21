import {
  createLoginSchema,
  createRegisterSchema,
  createPasswordRecoverySchema,
  createPasswordResetSchema,
} from '../auth'
import { UserRole } from '@/types/auth'

// Mock translation function that returns the key
const t = (key: string) => key

describe('createLoginSchema', () => {
  const schema = createLoginSchema(t)

  it('validates correct login data', () => {
    const result = schema.safeParse({
      email: 'test@example.com',
      password: 'password123',
    })

    expect(result.success).toBe(true)
  })

  it('rejects empty email', () => {
    const result = schema.safeParse({
      email: '',
      password: 'password123',
    })

    expect(result.success).toBe(false)
    if (!result.success) {
      expect(result.error.issues[0].message).toBe('emailRequired')
    }
  })

  it('rejects invalid email format', () => {
    const result = schema.safeParse({
      email: 'invalid-email',
      password: 'password123',
    })

    expect(result.success).toBe(false)
    if (!result.success) {
      expect(result.error.issues[0].message).toBe('emailInvalid')
    }
  })

  it('rejects empty password', () => {
    const result = schema.safeParse({
      email: 'test@example.com',
      password: '',
    })

    expect(result.success).toBe(false)
    if (!result.success) {
      expect(result.error.issues[0].message).toBe('passwordRequired')
    }
  })

  it('rejects password shorter than 8 characters', () => {
    const result = schema.safeParse({
      email: 'test@example.com',
      password: 'short',
    })

    expect(result.success).toBe(false)
    if (!result.success) {
      expect(result.error.issues[0].message).toBe('passwordMinLength')
    }
  })
})

describe('createRegisterSchema', () => {
  const schema = createRegisterSchema(t)

  it('validates correct registration data', () => {
    const result = schema.safeParse({
      name: 'Test User',
      email: 'test@example.com',
      password: 'Password1!',
      confirmPassword: 'Password1!',
      role: UserRole.STUDENT,
    })

    expect(result.success).toBe(true)
  })

  it('rejects short name', () => {
    const result = schema.safeParse({
      name: 'A',
      email: 'test@example.com',
      password: 'Password1!',
      confirmPassword: 'Password1!',
      role: UserRole.STUDENT,
    })

    expect(result.success).toBe(false)
    if (!result.success) {
      expect(result.error.issues[0].message).toBe('nameMinLength')
    }
  })

  it('rejects password without uppercase', () => {
    const result = schema.safeParse({
      name: 'Test User',
      email: 'test@example.com',
      password: 'password1!',
      confirmPassword: 'password1!',
      role: UserRole.STUDENT,
    })

    expect(result.success).toBe(false)
    if (!result.success) {
      expect(result.error.issues[0].message).toBe('passwordUppercase')
    }
  })

  it('rejects password without lowercase', () => {
    const result = schema.safeParse({
      name: 'Test User',
      email: 'test@example.com',
      password: 'PASSWORD1!',
      confirmPassword: 'PASSWORD1!',
      role: UserRole.STUDENT,
    })

    expect(result.success).toBe(false)
    if (!result.success) {
      expect(result.error.issues[0].message).toBe('passwordLowercase')
    }
  })

  it('rejects password without digit', () => {
    const result = schema.safeParse({
      name: 'Test User',
      email: 'test@example.com',
      password: 'Password!',
      confirmPassword: 'Password!',
      role: UserRole.STUDENT,
    })

    expect(result.success).toBe(false)
    if (!result.success) {
      expect(result.error.issues[0].message).toBe('passwordDigit')
    }
  })

  it('rejects password without special character', () => {
    const result = schema.safeParse({
      name: 'Test User',
      email: 'test@example.com',
      password: 'Password1',
      confirmPassword: 'Password1',
      role: UserRole.STUDENT,
    })

    expect(result.success).toBe(false)
    if (!result.success) {
      expect(result.error.issues[0].message).toBe('passwordSpecial')
    }
  })

  it('rejects mismatched passwords', () => {
    const result = schema.safeParse({
      name: 'Test User',
      email: 'test@example.com',
      password: 'Password1!',
      confirmPassword: 'DifferentPassword1!',
      role: UserRole.STUDENT,
    })

    expect(result.success).toBe(false)
    if (!result.success) {
      const mismatchError = result.error.issues.find((i) => i.path.includes('confirmPassword'))
      expect(mismatchError?.message).toBe('passwordsMismatch')
    }
  })

  it('rejects invalid role', () => {
    const result = schema.safeParse({
      name: 'Test User',
      email: 'test@example.com',
      password: 'Password1!',
      confirmPassword: 'Password1!',
      role: 'invalid_role',
    })

    expect(result.success).toBe(false)
  })

  it('accepts all valid roles', () => {
    const roles = [
      UserRole.STUDENT,
      UserRole.TEACHER,
      UserRole.ACADEMIC_SECRETARY,
      UserRole.METHODIST,
      UserRole.SYSTEM_ADMIN,
    ]

    roles.forEach((role) => {
      const result = schema.safeParse({
        name: 'Test User',
        email: 'test@example.com',
        password: 'Password1!',
        confirmPassword: 'Password1!',
        role,
      })
      expect(result.success).toBe(true)
    })
  })
})

describe('createPasswordRecoverySchema', () => {
  const schema = createPasswordRecoverySchema(t)

  it('validates correct email', () => {
    const result = schema.safeParse({
      email: 'test@example.com',
    })

    expect(result.success).toBe(true)
  })

  it('rejects empty email', () => {
    const result = schema.safeParse({
      email: '',
    })

    expect(result.success).toBe(false)
  })

  it('rejects invalid email', () => {
    const result = schema.safeParse({
      email: 'invalid',
    })

    expect(result.success).toBe(false)
  })

  it('rejects email with leading/trailing spaces before validation', () => {
    // Note: Zod validates before trim(), so spaces cause invalid email error
    const result = schema.safeParse({
      email: '  test@example.com  ',
    })

    expect(result.success).toBe(false)
  })
})

describe('createPasswordResetSchema', () => {
  const schema = createPasswordResetSchema(t)

  it('validates correct password reset data', () => {
    const result = schema.safeParse({
      password: 'NewPassword1!',
      confirmPassword: 'NewPassword1!',
    })

    expect(result.success).toBe(true)
  })

  it('rejects weak password', () => {
    const result = schema.safeParse({
      password: 'weak',
      confirmPassword: 'weak',
    })

    expect(result.success).toBe(false)
  })

  it('rejects mismatched passwords', () => {
    const result = schema.safeParse({
      password: 'NewPassword1!',
      confirmPassword: 'DifferentPassword1!',
    })

    expect(result.success).toBe(false)
    if (!result.success) {
      const mismatchError = result.error.issues.find((i) => i.path.includes('confirmPassword'))
      expect(mismatchError?.message).toBe('passwordsMismatch')
    }
  })
})
