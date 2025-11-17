import { NextResponse } from 'next/server'
import type { NextRequest } from 'next/server'
import { decodeJWT, isTokenExpired } from '@/lib/auth/jwt'
import { publicRoutes, authRoutes, hasRouteAccess } from '@/lib/auth/route-config'
import type { UserRole } from '@/types/auth'

/**
 * Middleware to protect routes and handle authentication with RBAC
 */
export function middleware(request: NextRequest) {
  const { pathname } = request.nextUrl

  // Check if it's a public route
  if (publicRoutes.includes(pathname)) {
    return NextResponse.next()
  }

  const cookieObject = request.cookies.get('auth-storage')
  const token = cookieObject?.value

  // Extract authentication data
  let isAuthenticated = false
  let userToken: string | null = null
  let userRole: UserRole | null = null

  if (token) {
    try {
      const authData = JSON.parse(token)
      userToken = authData.state?.token
      userRole = authData.state?.user?.role

      if (userToken) {
        // Check if token is expired
        if (isTokenExpired(userToken)) {
          // Token expired - redirect to login
          const loginUrl = new URL('/login', request.url)
          loginUrl.searchParams.set('redirect', pathname)
          loginUrl.searchParams.set('session_expired', 'true')

          const response = NextResponse.redirect(loginUrl)
          response.cookies.delete('auth-storage')
          return response
        }

        // Decode token to verify structure
        const payload = decodeJWT(userToken)
        if (payload) {
          isAuthenticated = true
          // Use role from payload if not in cookie
          if (!userRole) {
            userRole = payload.role
          }
        }
      }
    } catch (error) {
      // Invalid token format - clear cookie and redirect
      console.error('Invalid auth token format:', error)
      isAuthenticated = false
    }
  }

  // If user is authenticated and tries to access auth pages, redirect to dashboard
  if (isAuthenticated && authRoutes.includes(pathname)) {
    return NextResponse.redirect(new URL('/dashboard', request.url))
  }

  // If user is not authenticated and tries to access protected routes, redirect to login
  if (!isAuthenticated) {
    const loginUrl = new URL('/login', request.url)
    loginUrl.searchParams.set('redirect', pathname)
    return NextResponse.redirect(loginUrl)
  }

  // Check role-based access control
  if (userRole && !hasRouteAccess(pathname, userRole)) {
    // User doesn't have permission - redirect to forbidden page
    return NextResponse.redirect(new URL('/forbidden', request.url))
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
