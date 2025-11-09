import { NextResponse } from 'next/server'
import type { NextRequest } from 'next/server'

/**
 * Public routes that don't require authentication
 */
const publicRoutes = ['/login', '/register', '/forgot-password', '/reset-password']

/**
 * Routes that should redirect to home if already authenticated
 */
const authRoutes = ['/login', '/register']

/**
 * Middleware to protect routes and handle authentication
 */
export function middleware(request: NextRequest) {
  const { pathname } = request.nextUrl
  const token = request.cookies.get('auth-storage')?.value

  // Check if user has auth token
  let isAuthenticated = false
  if (token) {
    try {
      const authData = JSON.parse(token)
      isAuthenticated = authData.state?.isAuthenticated && authData.state?.token
    } catch (error) {
      // Invalid token format
      isAuthenticated = false
    }
  }

  // If user is authenticated and tries to access auth pages, redirect to home
  if (isAuthenticated && authRoutes.includes(pathname)) {
    return NextResponse.redirect(new URL('/', request.url))
  }

  // If user is not authenticated and tries to access protected routes, redirect to login
  if (!isAuthenticated && !publicRoutes.includes(pathname)) {
    const loginUrl = new URL('/login', request.url)
    // Add redirect param to return to original URL after login
    loginUrl.searchParams.set('redirect', pathname)
    return NextResponse.redirect(loginUrl)
  }

  return NextResponse.next()
}

/**
 * Matcher config to specify which routes to run middleware on
 */
export const config = {
  matcher: [
    /*
     * Match all request paths except for the ones starting with:
     * - api (API routes)
     * - _next/static (static files)
     * - _next/image (image optimization files)
     * - favicon.ico (favicon file)
     * - public files (public folder)
     */
    '/((?!api|_next/static|_next/image|favicon.ico|.*\\..*|_next).*)',
  ],
}
