# Frontend - Information System Secretary/Methodologist

Next.js 14 frontend application with TypeScript, Material-UI, and modern React patterns.

## 📁 Project Structure

```
frontend/
├── src/
│   ├── app/                 # Next.js App Router
│   │   ├── layout.tsx       # Root layout
│   │   ├── page.tsx         # Home page
│   │   └── globals.css      # Global styles
│   ├── components/          # React components
│   │   ├── common/          # Shared components (Button, Input, etc.)
│   │   ├── layouts/         # Layout components (Header, Sidebar)
│   │   └── features/        # Feature-specific components
│   ├── lib/                 # Utilities and helpers
│   │   ├── api.ts           # API client
│   │   └── utils.ts         # Helper functions
│   ├── types/               # TypeScript type definitions
│   │   ├── api.ts           # API types
│   │   └── models.ts        # Domain models
│   ├── hooks/               # Custom React hooks
│   │   ├── useAuth.ts       # Authentication hook
│   │   └── useApi.ts        # API data fetching hook
│   └── styles/              # Additional styles
├── public/                  # Static assets
├── .env.example            # Environment variables template
├── next.config.js          # Next.js configuration
├── tsconfig.json           # TypeScript configuration
├── package.json            # Dependencies and scripts
└── .eslintrc.json          # ESLint configuration
```

## 🏗️ Architecture

### Tech Stack

- **Framework**: Next.js 14 (App Router)
- **Language**: TypeScript 5.4
- **UI Library**: Material-UI (MUI) 5.15
- **State Management**: Zustand 4.5
- **Data Fetching**: SWR 2.2 + axios 1.6
- **Styling**: CSS Modules + MUI emotion
- **Testing**: Jest + React Testing Library + Playwright

### Key Features

- **Server-Side Rendering (SSR)**: Fast initial page loads
- **Static Site Generation (SSG)**: Pre-rendered pages where applicable
- **API Routes**: Backend proxy via Next.js rewrites
- **Type Safety**: Full TypeScript coverage
- **Responsive Design**: Mobile-first approach with MUI
- **Optimistic Updates**: SWR for cache management
- **Authentication**: JWT-based with automatic token refresh

### Design Patterns

- **Component Composition**: Reusable, composable components
- **Custom Hooks**: Encapsulate logic (useAuth, useApi, etc.)
- **API Client**: Centralized axios instance with interceptors
- **Error Boundaries**: Graceful error handling
- **Loading States**: Skeleton screens and spinners

## 🚀 Getting Started

### Prerequisites

- Node.js 18 or higher
- npm or yarn
- Backend server running on `http://localhost:8080`

### Installation

1. **Navigate to frontend directory**:
   ```bash
   cd frontend
   ```

2. **Install dependencies**:
   ```bash
   npm install
   # or
   yarn install
   ```

3. **Configure environment**:
   ```bash
   cp .env.example .env.local
   # Edit .env.local with your configuration
   ```

4. **Run development server**:
   ```bash
   npm run dev
   # or
   yarn dev
   ```

   Application will start on `http://localhost:3000`

### Development Commands

```bash
# Development server with hot reload
npm run dev

# Production build
npm run build

# Start production server
npm start

# Run linter
npm run lint

# Fix linting issues
npm run lint --fix

# Type checking
npm run type-check

# Run unit tests
npm run test:unit

# Run E2E tests
npm run test:e2e

# Run E2E tests in UI mode
npx playwright test --ui
```

## 🔧 Configuration

### Environment Variables

Create `.env.local` file:

```env
# Backend API URL
NEXT_PUBLIC_API_URL=http://localhost:8080

# Environment
NODE_ENV=development
```

See `.env.example` for all available options.

### API Proxy Configuration

The Next.js config proxies API requests to the backend:

```javascript
// next.config.js
async rewrites() {
  return [
    {
      source: '/api/:path*',
      destination: 'http://localhost:8080/api/:path*',
    },
  ];
}
```

This allows you to call `/api/auth/login` from the frontend, and it will be proxied to `http://localhost:8080/api/auth/login`.

## 🎨 Styling

### Material-UI Theme

Customize theme in `src/app/theme.ts` (planned):

```typescript
import { createTheme } from '@mui/material/styles';

export const theme = createTheme({
  palette: {
    primary: {
      main: '#1976d2',
    },
    secondary: {
      main: '#dc004e',
    },
  },
  typography: {
    fontFamily: 'Roboto, Arial, sans-serif',
  },
});
```

### CSS Modules

```typescript
// Button.module.css
.button {
  padding: 10px 20px;
  border-radius: 4px;
}

// Button.tsx
import styles from './Button.module.css';

export function Button() {
  return <button className={styles.button}>Click me</button>;
}
```

## 🔐 Authentication

### API Client with Auth

The API client (`src/lib/api.ts`) automatically handles authentication:

```typescript
import { apiClient } from '@/lib/api';

// Login
const response = await apiClient.post('/api/auth/login', {
  email: 'user@example.com',
  password: 'password',
});

// Token is automatically stored
apiClient.setAuthToken(response.token);

// Authenticated requests automatically include token
const user = await apiClient.get('/api/auth/me');
```

### Protected Routes

Create protected route wrapper (planned):

```typescript
// middleware.ts
import { NextResponse } from 'next/server';

export function middleware(request) {
  const token = request.cookies.get('authToken');

  if (!token) {
    return NextResponse.redirect('/login');
  }

  return NextResponse.next();
}

export const config = {
  matcher: ['/dashboard/:path*', '/documents/:path*'],
};
```

## 🧪 Testing

### Unit Tests (Jest)

```typescript
// Button.test.tsx
import { render, screen } from '@testing-library/react';
import { Button } from './Button';

describe('Button', () => {
  it('renders button with text', () => {
    render(<Button>Click me</Button>);
    expect(screen.getByText('Click me')).toBeInTheDocument();
  });
});
```

Run tests:
```bash
npm run test:unit
```

### E2E Tests (Playwright)

```typescript
// tests/login.spec.ts
import { test, expect } from '@playwright/test';

test('user can login', async ({ page }) => {
  await page.goto('http://localhost:3000/login');

  await page.fill('input[name="email"]', 'user@example.com');
  await page.fill('input[name="password"]', 'password');
  await page.click('button[type="submit"]');

  await expect(page).toHaveURL('http://localhost:3000/dashboard');
});
```

Run E2E tests:
```bash
npm run test:e2e
```

### Test Coverage Goals

- **Components**: 80%+
- **Utilities**: 90%+
- **Hooks**: 85%+
- **Overall**: 80%+

## 📊 State Management

### Zustand Store

```typescript
// stores/authStore.ts
import { create } from 'zustand';

interface AuthState {
  user: User | null;
  token: string | null;
  login: (email: string, password: string) => Promise<void>;
  logout: () => void;
}

export const useAuthStore = create<AuthState>((set) => ({
  user: null,
  token: null,
  login: async (email, password) => {
    const response = await apiClient.post('/api/auth/login', { email, password });
    set({ user: response.user, token: response.token });
  },
  logout: () => {
    set({ user: null, token: null });
  },
}));
```

### SWR for Data Fetching

```typescript
import useSWR from 'swr';
import { apiClient } from '@/lib/api';

export function useDocuments() {
  const { data, error, isLoading } = useSWR('/api/documents', (url) =>
    apiClient.get(url)
  );

  return {
    documents: data,
    isLoading,
    isError: error,
  };
}
```

## 🎯 Code Quality

### TypeScript Configuration

Strict mode enabled in `tsconfig.json`:

```json
{
  "compilerOptions": {
    "strict": true,
    "noEmit": true,
    "esModuleInterop": true,
    "moduleResolution": "bundler",
    "resolveJsonModule": true,
    "isolatedModules": true,
    "jsx": "preserve"
  }
}
```

### ESLint Rules

Configured in `.eslintrc.json`:

```json
{
  "extends": [
    "next/core-web-vitals",
    "next/typescript"
  ],
  "rules": {
    "@typescript-eslint/no-unused-vars": "error",
    "@typescript-eslint/no-explicit-any": "warn"
  }
}
```

### Code Style Guidelines

- Use functional components with hooks
- Prefer named exports over default exports
- Use TypeScript interfaces for props
- Implement error boundaries for components
- Use semantic HTML elements
- Follow accessibility best practices (ARIA labels)

## 🚀 Performance Optimization

### Next.js Optimizations

- **Image Optimization**: Use `next/image` component
- **Font Optimization**: Use `next/font`
- **Code Splitting**: Automatic with dynamic imports
- **Static Generation**: Pre-render pages at build time
- **Incremental Static Regeneration**: Update static pages without rebuild

### Best Practices

```typescript
// Dynamic imports for code splitting
import dynamic from 'next/dynamic';

const HeavyComponent = dynamic(() => import('./HeavyComponent'), {
  loading: () => <p>Loading...</p>,
});

// Image optimization
import Image from 'next/image';

<Image
  src="/logo.png"
  alt="Logo"
  width={200}
  height={100}
  priority
/>
```

## 📦 Building for Production

### Build Process

```bash
# Create optimized production build
npm run build

# Output will be in .next/ directory
```

### Deployment

#### Vercel (Recommended)

```bash
# Install Vercel CLI
npm i -g vercel

# Deploy
vercel --prod
```

#### Docker

```dockerfile
# Dockerfile (planned)
FROM node:18-alpine

WORKDIR /app
COPY package*.json ./
RUN npm ci --only=production
COPY . .
RUN npm run build

EXPOSE 3000
CMD ["npm", "start"]
```

#### Static Export

```bash
# Add to package.json scripts
"export": "next build && next export"

# Run export
npm run export

# Output will be in out/ directory
```

## 🔄 CI/CD

Frontend CI/CD will be integrated into existing workflows:

### Planned Workflow (`.github/workflows/frontend-ci.yml`)

- **Linting**: ESLint checks
- **Type Checking**: TypeScript compilation
- **Unit Tests**: Jest with coverage
- **E2E Tests**: Playwright tests
- **Build**: Next.js production build
- **Lighthouse**: Performance and accessibility scores

See [CI/CD Workflows Documentation](../docs/development/ci-cd-workflows.md) for details.

## 📚 Component Library (Planned)

### Common Components

- `Button` - Customizable button with variants
- `Input` - Form input with validation
- `Card` - Content container
- `Modal` - Dialog/modal window
- `Table` - Data table with sorting and filtering
- `Sidebar` - Navigation sidebar
- `Header` - Application header with user menu

### Feature Components

- `DocumentList` - Document management interface
- `WorkflowDiagram` - Workflow visualization
- `ScheduleCalendar` - Schedule management
- `UserProfile` - User profile editor

## 🌐 Internationalization (Planned)

Using `next-intl` for i18n:

```typescript
import { useTranslations } from 'next-intl';

export function LoginForm() {
  const t = useTranslations('Auth');

  return (
    <form>
      <h1>{t('login')}</h1>
      <button>{t('submit')}</button>
    </form>
  );
}
```

## 📖 Additional Documentation

- [Project Overview](../docs/project-overview.md)
- [Development Guide](../docs/development/development-guide.md)
- [Pull Request Guide](../docs/development/pull-request-guide.md)
- [API Documentation](../docs/api/api-documentation.md)

## 🤝 Contributing

1. Read [Development Guide](../docs/development/development-guide.md)
2. Check [Pull Request Guide](../docs/development/pull-request-guide.md)
3. Create feature branch: `feature/issue-N-description`
4. Write tests for new features
5. Ensure all checks pass (lint, type-check, tests)
6. Submit PR with conventional commit format

## 🐛 Troubleshooting

### Common Issues

#### Port already in use
```bash
# Kill process on port 3000
lsof -ti:3000 | xargs kill -9
```

#### Module not found errors
```bash
# Clear node_modules and reinstall
rm -rf node_modules package-lock.json
npm install
```

#### TypeScript errors after update
```bash
# Regenerate TypeScript cache
rm -rf .next
npm run dev
```

#### API proxy not working
- Check `NEXT_PUBLIC_API_URL` in `.env.local`
- Ensure backend is running on the specified URL
- Check `next.config.js` rewrites configuration

## 📝 License

MIT License - see [LICENSE](../LICENSE) file for details.

---

**Tech Stack**: Next.js 14 • TypeScript 5.4 • React 18 • MUI 5 • Zustand • SWR • Playwright
