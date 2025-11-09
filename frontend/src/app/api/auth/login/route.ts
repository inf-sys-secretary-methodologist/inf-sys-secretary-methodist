import { NextRequest, NextResponse } from 'next/server'

/**
 * Mock login endpoint for testing
 * POST /api/auth/login
 */
export async function POST(request: NextRequest) {
  try {
    const body = await request.json()
    const { email, password } = body

    // Simulate network delay
    await new Promise((resolve) => setTimeout(resolve, 500))

    // Mock validation
    if (!email || !password) {
      return NextResponse.json(
        { message: 'Email и пароль обязательны' },
        { status: 400 }
      )
    }

    if (password.length < 8) {
      return NextResponse.json(
        { message: 'Неверный пароль' },
        { status: 401 }
      )
    }

    // Mock successful login
    const mockUser = {
      id: '1',
      email,
      name: 'Test User',
      role: 'STUDENT',
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
    }

    const mockResponse = {
      user: mockUser,
      token: 'mock-jwt-token-' + Date.now(),
      refreshToken: 'mock-refresh-token-' + Date.now(),
      expiresIn: 3600,
    }

    return NextResponse.json(mockResponse, { status: 200 })
  } catch (error) {
    console.error('Login error:', error)
    return NextResponse.json(
      { message: 'Внутренняя ошибка сервера' },
      { status: 500 }
    )
  }
}
