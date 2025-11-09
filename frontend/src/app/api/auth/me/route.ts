import { NextRequest, NextResponse } from 'next/server'

/**
 * Mock get current user endpoint
 * GET /api/auth/me
 */
export async function GET(request: NextRequest) {
  try {
    const authHeader = request.headers.get('authorization')

    if (!authHeader || !authHeader.startsWith('Bearer ')) {
      return NextResponse.json(
        { message: 'Unauthorized' },
        { status: 401 }
      )
    }

    // Simulate network delay
    await new Promise((resolve) => setTimeout(resolve, 200))

    // Mock current user
    const mockUser = {
      id: '1',
      email: 'user@example.com',
      name: 'Test User',
      role: 'STUDENT',
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
    }

    return NextResponse.json(mockUser, { status: 200 })
  } catch (error) {
    console.error('Get current user error:', error)
    return NextResponse.json(
      { message: 'Внутренняя ошибка сервера' },
      { status: 500 }
    )
  }
}
