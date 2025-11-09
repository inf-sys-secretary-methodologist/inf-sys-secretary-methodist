import { NextRequest, NextResponse } from 'next/server'

/**
 * Mock register endpoint for testing
 * POST /api/auth/register
 */
export async function POST(request: NextRequest) {
  try {
    const body = await request.json()
    const { email, password, name, role } = body

    // Simulate network delay
    await new Promise((resolve) => setTimeout(resolve, 800))

    // Mock validation
    if (!email || !password || !name || !role) {
      return NextResponse.json(
        { message: 'Все поля обязательны для заполнения' },
        { status: 400 }
      )
    }

    // Mock email already exists check
    if (email === 'test@example.com') {
      return NextResponse.json(
        { message: 'Пользователь с таким email уже существует' },
        { status: 409 }
      )
    }

    // Mock successful registration
    const mockUser = {
      id: '2',
      email,
      name,
      role,
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
    }

    const mockResponse = {
      user: mockUser,
      token: 'mock-jwt-token-' + Date.now(),
      refreshToken: 'mock-refresh-token-' + Date.now(),
      expiresIn: 3600,
    }

    return NextResponse.json(mockResponse, { status: 201 })
  } catch (error) {
    console.error('Registration error:', error)
    return NextResponse.json(
      { message: 'Внутренняя ошибка сервера' },
      { status: 500 }
    )
  }
}
