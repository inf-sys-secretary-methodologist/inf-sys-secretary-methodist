# frontend/Dockerfile

# Stage 1: Dependencies
FROM node:25-alpine AS deps

WORKDIR /app

# Копируем только package файлы для кэширования
COPY package.json package-lock.json ./
RUN npm ci --production=false --ignore-scripts

# Stage 2: Builder
FROM node:25-alpine AS builder

WORKDIR /app

# Копируем установленные зависимости
COPY --from=deps /app/node_modules ./node_modules

# Копируем исходники
COPY . .

# Build-time переменные
ENV NEXT_PUBLIC_API_URL=http://localhost:8080
ENV NEXT_TELEMETRY_DISABLED=1

# Собираем приложение
RUN npm run build

# Stage 3: Runner (минимальный production образ)
FROM node:25-alpine AS runner

WORKDIR /app

# Создаем пользователя для безопасности
RUN addgroup --system --gid 1001 nodejs && \
    adduser --system --uid 1001 nextjs

# Копируем собранное приложение (standalone уже содержит минимальный набор зависимостей)
COPY --from=builder --chown=nextjs:nodejs /app/.next/standalone ./
COPY --from=builder --chown=nextjs:nodejs /app/.next/static ./.next/static
COPY --from=builder --chown=nextjs:nodejs /app/public ./public

# Используем непривилегированного пользователя
USER nextjs

# Отключаем телеметрию
ENV NEXT_TELEMETRY_DISABLED=1
ENV NODE_ENV=production
ENV PORT=3000

EXPOSE 3000

# Healthcheck для мониторинга
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD node -e "require('http').get('http://localhost:3000/', (r) => {process.exit(r.statusCode === 200 ? 0 : 1)})" || exit 1

CMD ["node", "server.js"]