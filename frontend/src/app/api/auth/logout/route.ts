import { NextRequest, NextResponse } from 'next/server'

/**
 * Mock logout endpoint
 * POST /api/auth/logout
 */
export async function POST(request: NextRequest) {
  try {
    // Simulate network delay
    await new Promise((resolve) => setTimeout(resolve, 200))

    return NextResponse.json({ message: 'Успешный выход' }, { status: 200 })
  } catch (error) {
    console.error('Logout error:', error)
    return NextResponse.json(
      { message: 'Внутренняя ошибка сервера' },
      { status: 500 }
    )
  }
}
