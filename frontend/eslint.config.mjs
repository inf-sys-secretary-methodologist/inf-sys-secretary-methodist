import nextCoreWebVitals from 'eslint-config-next/core-web-vitals'
import nextTypescript from 'eslint-config-next/typescript'
import eslintConfigPrettier from 'eslint-config-prettier'
import eslintPluginPrettier from 'eslint-plugin-prettier'

const eslintConfig = [
  ...nextCoreWebVitals,
  ...nextTypescript,
  eslintConfigPrettier,
  {
    plugins: {
      prettier: eslintPluginPrettier,
    },
    rules: {
      'prettier/prettier': 'error',
      '@typescript-eslint/no-unused-vars': [
        'error',
        {
          argsIgnorePattern: '^_',
          varsIgnorePattern: '^_',
          caughtErrorsIgnorePattern: '^_',
        },
      ],
      // React Compiler rules from Next.js 16 — too strict for common patterns
      // like setState in useEffect for initialization / hydration
      'react-hooks/set-state-in-effect': 'off',
    },
  },
  {
    // Disable React Compiler rules in test files — mock components
    // legitimately reassign outer variables for test assertions
    files: ['**/__tests__/**', '**/*.test.*'],
    rules: {
      'react-hooks/globals': 'off',
    },
  },
  {
    ignores: [
      '.next/',
      'node_modules/',
      'coverage/',
      'public/',
      'storybook-static/',
      '*.config.js',
      '*.config.mjs',
      'jest.setup.ts',
    ],
  },
]

export default eslintConfig
