# Security Policy

Этот документ описывает, как сообщать об уязвимостях в **inf-sys-secretary-methodist** и какие версии получают security updates.

> **Контекст**: проект учебный (дипломный), не развёрнут в production.
> Security posture направлен на лучшие практики разработки, не на soft real-time SLA.

## Поддерживаемые версии

Security updates выходят только для последней minor-линии. Старые versions не патчатся — рекомендуется обновляться.

| Версия      | Поддержка             |
| ----------- | --------------------- |
| 0.128.x     | ✅ Supported          |
| 0.127.x     | ⚠️ Best effort        |
| < 0.127     | ❌ Not supported      |

Источник истины — файл [`VERSION`](./VERSION) в корне репо.

## Как сообщить об уязвимости

**Не открывай public issue с деталями уязвимости.** Используй один из приватных каналов ниже.

### Предпочтительный способ — GitHub Private Vulnerability Reporting

[Создать private security advisory →](https://github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/security/advisories/new)

Только maintainers увидят сообщение до публичного disclosure.

### Альтернатива — email

`daniilvdovin4@gmail.com` (только для security issues; для общих вопросов и багов — issues tracker).

## Что включить в отчёт

- Версия проекта и/или commit hash из `VERSION` или `git log`
- Описание уязвимости с PoC, если возможно
- Шаги для воспроизведения
- Потенциальный impact (что атакующий может получить / нарушить)
- (Опционально) предложение по fix

## Что ожидать

| Этап                              | Срок                               |
| --------------------------------- | ---------------------------------- |
| Подтверждение получения           | В течение 5 рабочих дней           |
| Первая оценка severity            | В течение 14 дней                  |
| Patch / response                  | По договорённости в зависимости от severity |
| Public disclosure                 | После patch + 30 дней grace period |

## Out of scope

Поскольку проект учебный, следующее **не считается уязвимостями**:

- Default values в `.env.example` или fixtures — это шаблоны, реальные секреты не commit'ятся
- DoS на dev-сборках (`npm run dev`, `go run`) без production hardening
- Mock-данные или test fixtures с заведомо слабыми credentials
- Уязвимости в зависимостях, для которых нет fix в latest version, и которые не impact production code path (см. CHANGELOG.md секцию "Out of scope" в latest release notes)

## Текущие accepted residual

После масштабной security cleanup (v0.128.6–v0.128.9 cluster) остаются 2 transitive alert в `node_modules`:

- **postcss XSS via unescaped `</style>`** (Moderate) — transitive в `next/node_modules`. Закрытие требует Next.js major downgrade (breaking). Impact — dev-tooling only, не production code path.
- **@tootallnate/once Incorrect Control Flow Scoping** (Low) — deeply transitive. Negligible attack surface.

Tracking — см. https://github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/security/dependabot

## Security practices в проекте

- **Dependabot alerts + automatic security updates** (включено)
- **Secret scanning + push protection** (включено) — блокирует commit accidentally committed credentials
- **Code quality findings** (включено) — automatic detection кода с проблемами
- **Pre-commit hook** (`.husky/pre-commit`) — eslint + prettier + golangci-lint + AmE/BrE word check
- **Branch protection** на `main` — required PR review, no direct push
- **`min-release-age=7`** в `.npmrc` — supply chain defence, новые npm пакеты не устанавливаются раньше 7 дней с момента публикации
- **Reviewer round** через `superpowers:code-reviewer` agent для feature releases

## Lifecycle уязвимостей

1. **Discovery** — researcher reports privately (через формы выше)
2. **Triage** — maintainer оценивает severity (CVSS) в течение 14 дней
3. **Fix** — patch разрабатывается на private branch
4. **Disclosure** — после release нового patch'а, advisory публикуется через GitHub Security Advisories
5. **Credit** — researcher получает credit в advisory (если хочет)

---

Спасибо за помощь в улучшении безопасности проекта!
