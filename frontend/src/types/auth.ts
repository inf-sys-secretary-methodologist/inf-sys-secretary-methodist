// User types
export interface User {
  id: string
  email: string
  name: string
  role: UserRole
  createdAt: string
  updatedAt: string
}

export enum UserRole {
  SYSTEM_ADMIN = 'system_admin',
  METHODIST = 'methodist',
  ACADEMIC_SECRETARY = 'academic_secretary',
  TEACHER = 'teacher',
  STUDENT = 'student',
}

// Auth request types
export interface LoginRequest {
  email: string
  password: string
}

export interface RegisterRequest {
  email: string
  password: string
  name: string
  role: UserRole
}

export interface RefreshTokenRequest {
  refreshToken: string
}

// Auth response types
export interface AuthResponse {
  user: User
  token: string
  refreshToken: string
  expiresIn: number
}

export interface RefreshTokenResponse {
  token: string
  refreshToken: string
  expiresIn: number
}

// Auth error types
export interface AuthError {
  message: string
  code: AuthErrorCode
}

export enum AuthErrorCode {
  INVALID_CREDENTIALS = 'INVALID_CREDENTIALS',
  USER_NOT_FOUND = 'USER_NOT_FOUND',
  EMAIL_ALREADY_EXISTS = 'EMAIL_ALREADY_EXISTS',
  WEAK_PASSWORD = 'WEAK_PASSWORD',
  INVALID_TOKEN = 'INVALID_TOKEN',
  TOKEN_EXPIRED = 'TOKEN_EXPIRED',
  UNAUTHORIZED = 'UNAUTHORIZED',
}
