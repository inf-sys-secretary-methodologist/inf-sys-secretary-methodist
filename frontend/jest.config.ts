import type { Config } from 'jest'
import nextJest from 'next/jest.js'

const createJestConfig = nextJest({
  // Provide the path to your Next.js app to load next.config.js and .env files in your test environment
  dir: './',
})

// Add any custom config to be passed to Jest
const config: Config = {
  coverageProvider: 'v8',
  testEnvironment: 'jsdom',
  // Add more setup options before each test is run
  setupFilesAfterEnv: ['<rootDir>/jest.setup.ts'],

  // Игнорируем E2E тесты Playwright
  testPathIgnorePatterns: ['<rootDir>/.next/', '<rootDir>/node_modules/', '<rootDir>/tests/e2e/'],

  // Transform ESM modules from node_modules
  transformIgnorePatterns: ['/node_modules/(?!(next-intl|use-intl)/)'],

  // Coverage настройки
  collectCoverageFrom: [
    'src/**/*.{js,jsx,ts,tsx}',
    '!src/**/*.d.ts',
    '!src/**/*.stories.{js,jsx,ts,tsx}',
    '!src/**/__tests__/**',
    // Exclude type-only files (no runtime code)
    '!src/types/**/*.ts',
    // Exclude re-export index files (barrel exports)
    '!src/components/**/index.ts',
    // Exclude Next.js app directory files (tested via E2E/integration)
    '!src/app/**/*.tsx',
    '!src/app/**/*.ts',
    // Exclude middleware (tested separately)
    '!src/middleware.ts',
  ],

  coverageThreshold: {
    global: {
      branches: 30,
      functions: 15,
      lines: 5,
      statements: 5,
    },
  },

  moduleNameMapper: {
    // Handle module aliases (this will be automatically configured for you soon)
    '^@/(.*)$': '<rootDir>/src/$1',
  },
}

// createJestConfig is exported this way to ensure that next/jest can load the Next.js config which is async
export default createJestConfig(config)
