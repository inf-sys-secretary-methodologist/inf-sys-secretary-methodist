# 🔧 Git Terminal Guide

Полное руководство по работе с Git через терминал для проекта "Информационная система академического секретаря/методиста"

## 📋 Содержание

1. [Настройка Git](#-настройка-git)
2. [Основы работы с репозиторием](#-основы-работы-с-репозиторием)
3. [Работа с ветками](#-работа-с-ветками)
4. [Коммиты и история](#-коммиты-и-история)
5. [Работа с удаленными репозиториями](#-работа-с-удаленными-репозиториями)
6. [Слияние и конфликты](#-слияние-и-конфликты)
7. [Отмена изменений](#-отмена-изменений)
8. [GitHub CLI (gh)](#-github-cli-gh)
9. [Продвинутые команды](#-продвинутые-команды)
10. [Полезные алиасы](#-полезные-алиасы)
11. [Troubleshooting](#-troubleshooting)

---

## ⚙️ Настройка Git

### Первичная настройка
```bash
# Установка имени и email (обязательно)
git config --global user.name "Ваше Имя"
git config --global user.email "your.email@example.com"

# Настройка редактора по умолчанию
git config --global core.editor "code --wait"  # VS Code
git config --global core.editor "vim"          # Vim
git config --global core.editor "nano"         # Nano

# Настройка окончаний строк
git config --global core.autocrlf input   # macOS/Linux
git config --global core.autocrlf true    # Windows

# Включение цветного вывода
git config --global color.ui auto
```

### Просмотр конфигурации
```bash
# Показать всю конфигурацию
git config --list

# Показать конкретную настройку
git config user.name
git config user.email

# Показать расположение файлов конфигурации
git config --list --show-origin
```

### SSH ключи для GitHub
```bash
# Генерация SSH ключа
ssh-keygen -t ed25519 -C "your.email@example.com"

# Добавление ключа в ssh-agent
eval "$(ssh-agent -s)"
ssh-add ~/.ssh/id_ed25519

# Копирование публичного ключа (для добавления в GitHub)
cat ~/.ssh/id_ed25519.pub | pbcopy  # macOS
cat ~/.ssh/id_ed25519.pub           # Linux/Windows

# Проверка соединения с GitHub
ssh -T git@github.com
```

---

## 📁 Основы работы с репозиторием

### Создание репозитория
```bash
# Инициализация нового репозитория
mkdir my-project
cd my-project
git init

# Клонирование существующего репозитория
git clone https://github.com/username/repository.git
git clone git@github.com:username/repository.git  # SSH

# Клонирование с другим именем папки
git clone https://github.com/username/repo.git my-folder

# Клонирование только последних коммитов (shallow clone)
git clone --depth 1 https://github.com/username/repo.git
```

### Статус и информация
```bash
# Текущий статус репозитория
git status

# Краткий статус
git status -s
git status --short

# Показать удаленные репозитории
git remote -v

# Информация о ветках
git branch -a              # Все ветки (локальные и удаленные)
git branch -r              # Только удаленные ветки
git branch -vv             # Ветки с информацией о tracking

# Информация о коммитах
git log --oneline          # Краткая история
git log --graph --oneline  # История с графом
git show HEAD              # Последний коммит
```

---

## 🌿 Работа с ветками

### Создание и переключение веток
```bash
# Создание новой ветки
git branch feature/new-feature
git branch hotfix/critical-bug

# Переключение на ветку
git checkout feature/new-feature
git switch feature/new-feature     # Новая команда (Git 2.23+)

# Создание и переключение одновременно
git checkout -b feature/user-auth
git switch -c feature/user-auth     # Новая команда

# Создание ветки от определенного коммита
git checkout -b hotfix/bug-123 abc1234

# Создание ветки от удаленной ветки
git checkout -b local-branch origin/remote-branch
```

### Управление ветками
```bash
# Переименование текущей ветки
git branch -m new-branch-name

# Переименование любой ветки
git branch -m old-name new-name

# Удаление ветки (безопасное)
git branch -d feature/completed-feature

# Принудительное удаление ветки
git branch -D feature/abandoned-feature

# Удаление удаленной ветки
git push origin --delete feature/old-feature

# Показать последний коммит каждой ветки
git branch -v

# Показать слитые ветки
git branch --merged
git branch --no-merged
```

### Отслеживание удаленных веток
```bash
# Установка upstream для ветки
git push -u origin feature/new-feature
git branch --set-upstream-to=origin/main main

# Создание локальной ветки для отслеживания удаленной
git checkout --track origin/feature/remote-feature

# Обновление информации об удаленных ветках
git fetch origin
git fetch --all

# Синхронизация с удаленным репозиторием
git pull origin main
git pull --rebase origin main  # С rebase вместо merge
```

---

## 💾 Коммиты и история

### Добавление файлов в индекс
```bash
# Добавление конкретных файлов
git add file1.txt file2.txt
git add src/components/Button.tsx

# Добавление всех файлов в текущей папке
git add .

# Добавление всех файлов в репозитории
git add -A
git add --all

# Интерактивное добавление
git add -i
git add -p file.txt  # Добавление по частям (hunks)

# Добавление только отслеживаемых файлов
git add -u
git add --update
```

### Создание коммитов
```bash
# Обычный коммит
git commit -m "feat: add user authentication"

# Коммит с детальным описанием
git commit -m "feat: add user authentication

- Implement OAuth 2.0 integration
- Add JWT token validation
- Create login/logout components
- Add user session management

Closes #123"

# Коммит всех изменений (add + commit)
git commit -am "fix: resolve merge conflict in user service"

# Изменение последнего коммита
git commit --amend -m "Updated commit message"

# Добавление файлов к последнему коммиту
git add forgotten-file.txt
git commit --amend --no-edit
```

### Типы коммитов (Conventional Commits)
```bash
# Новая функциональность
git commit -m "feat: add document export functionality"

# Исправление бага
git commit -m "fix: resolve login redirect issue"

# Документация
git commit -m "docs: update API documentation"

# Стилизация кода
git commit -m "style: format code according to ESLint rules"

# Рефакторинг
git commit -m "refactor: extract common utility functions"

# Тесты
git commit -m "test: add unit tests for user service"

# Производительность
git commit -m "perf: optimize database queries"

# CI/CD
git commit -m "ci: update GitHub Actions workflow"

# Обновление зависимостей
git commit -m "chore: update dependencies to latest versions"
```

### Просмотр истории
```bash
# Стандартная история
git log

# Краткая история (одна строка на коммит)
git log --oneline

# История с графом веток
git log --graph --oneline --all

# Детальная история с изменениями
git log --stat
git log -p

# История за определенный период
git log --since="2024-01-01"
git log --until="2024-12-31"
git log --since="2 weeks ago"

# История конкретного файла
git log -- path/to/file.txt
git log -p -- src/components/Button.tsx

# История с поиском
git log --grep="authentication"
git log --author="John Doe"
git log --grep="fix" --author="Jane"

# Показать изменения между коммитами
git diff HEAD~1..HEAD
git diff abc1234..def5678
git diff main..feature/new-feature
```

---

## 🌐 Работа с удаленными репозиториями

### Настройка удаленных репозиториев
```bash
# Добавление удаленного репозитория
git remote add origin https://github.com/username/repo.git
git remote add upstream https://github.com/original/repo.git

# Изменение URL удаленного репозитория
git remote set-url origin git@github.com:username/repo.git

# Удаление удаленного репозитория
git remote remove upstream

# Переименование удаленного репозитория
git remote rename origin new-origin
```

### Синхронизация с удаленным репозиторием
```bash
# Загрузка изменений (без слияния)
git fetch origin
git fetch --all

# Загрузка и слияние изменений
git pull origin main
git pull origin feature/user-auth

# Pull с rebase (линейная история)
git pull --rebase origin main

# Отправка изменений
git push origin main
git push origin feature/new-feature

# Первая отправка ветки
git push -u origin feature/new-feature

# Принудительная отправка (осторожно!)
git push --force origin feature/rebased-branch
git push --force-with-lease origin feature/rebased-branch  # Безопаснее

# Отправка всех веток
git push --all origin

# Отправка тегов
git push --tags origin
```

### Работа с форками
```bash
# Настройка для работы с форком
git remote add upstream https://github.com/original/repo.git
git fetch upstream

# Синхронизация форка с оригинальным репозиторием
git checkout main
git pull upstream main
git push origin main

# Создание PR-ветки от актуального main
git checkout main
git pull upstream main
git checkout -b feature/my-contribution
```

---

## 🔀 Слияние и конфликты

### Слияние веток
```bash
# Обычное слияние
git checkout main
git merge feature/new-feature

# Слияние без fast-forward (создает merge commit)
git merge --no-ff feature/new-feature

# Слияние с сообщением
git merge feature/new-feature -m "Merge feature: user authentication"

# Слияние только если нет конфликтов
git merge --no-commit feature/new-feature
git commit -m "Custom merge message"
```

### Rebase
```bash
# Rebase текущей ветки на main
git rebase main

# Интерактивный rebase (изменение истории)
git rebase -i HEAD~3
git rebase -i abc1234

# Rebase при pull
git pull --rebase origin main

# Продолжение rebase после разрешения конфликтов
git rebase --continue

# Отмена rebase
git rebase --abort

# Пропуск текущего коммита при rebase
git rebase --skip
```

### Разрешение конфликтов
```bash
# Показать конфликтующие файлы
git status

# Открыть инструмент для разрешения конфликтов
git mergetool

# После разрешения конфликтов
git add resolved-file.txt
git commit

# Показать различия для конфликтов
git diff
git diff --conflict=diff3

# Выбрать версию файла целиком
git checkout --ours file.txt    # Наша версия
git checkout --theirs file.txt  # Их версия
```

---

## ↩️ Отмена изменений

### Отмена изменений в рабочей директории
```bash
# Отмена изменений в конкретном файле
git checkout -- file.txt
git restore file.txt            # Новая команда (Git 2.23+)

# Отмена всех изменений в рабочей директории
git checkout -- .
git restore .

# Удаление неотслеживаемых файлов
git clean -f               # Удалить файлы
git clean -fd              # Удалить файлы и папки
git clean -n               # Показать что будет удалено (dry run)
```

### Отмена изменений в индексе
```bash
# Убрать файл из индекса (unstage)
git reset HEAD file.txt
git restore --staged file.txt   # Новая команда

# Убрать все файлы из индекса
git reset HEAD
git restore --staged .
```

### Отмена коммитов
```bash
# Отмена последнего коммита (soft reset)
git reset --soft HEAD~1    # Изменения остаются в индексе

# Отмена последнего коммита (mixed reset, по умолчанию)
git reset HEAD~1           # Изменения остаются в рабочей директории

# Полная отмена последнего коммита (hard reset)
git reset --hard HEAD~1   # Изменения удаляются

# Отмена до конкретного коммита
git reset --hard abc1234

# Создание reverting коммита
git revert HEAD            # Отменить последний коммит
git revert abc1234         # Отменить конкретный коммит
git revert HEAD~3..HEAD    # Отменить диапазон коммитов
```

### Временное сохранение изменений
```bash
# Сохранить изменения в stash
git stash
git stash save "Work in progress on feature X"

# Показать список stash'ей
git stash list

# Применить последний stash
git stash apply
git stash pop              # Применить и удалить

# Применить конкретный stash
git stash apply stash@{2}

# Удалить stash
git stash drop stash@{1}
git stash clear            # Удалить все stash'и

# Создать ветку из stash
git stash branch feature/temp-work stash@{1}
```

---

## 🐙 GitHub CLI (gh)

GitHub CLI - официальный инструмент командной строки для работы с GitHub, который значительно ускоряет работу с репозиториями, issues, PR и другими функциями GitHub.

### Установка GitHub CLI
```bash
# macOS (Homebrew)
brew install gh

# Ubuntu/Debian
sudo apt install gh
# или через snap
sudo snap install gh

# Windows (Chocolatey)
choco install gh
# или через Scoop
scoop install gh

# Windows (Winget)
winget install --id GitHub.cli

# Arch Linux
sudo pacman -S github-cli
```

### Первичная настройка
```bash
# Авторизация в GitHub
gh auth login

# Выберите:
# - GitHub.com
# - HTTPS или SSH
# - Авторизация через браузер или токен

# Проверка авторизации
gh auth status

# Настройка Git для использования gh как credential helper
gh auth setup-git

# Выбор редактора по умолчанию
gh config set editor "code --wait"  # VS Code
gh config set editor "vim"          # Vim
```

### Работа с репозиториями
```bash
# Клонирование репозитория
gh repo clone username/repository
gh repo clone inf-sys-secretary-methodologist/inf-sys-secretary-methodist

# Создание нового репозитория
gh repo create my-new-repo
gh repo create my-new-repo --public --description "My awesome project"
gh repo create my-new-repo --private --clone

# Просмотр информации о репозитории
gh repo view
gh repo view username/repository

# Форк репозитория
gh repo fork username/repository
gh repo fork username/repository --clone

# Удаление репозитория (осторожно!)
gh repo delete username/repository

# Архивирование репозитория
gh repo archive username/repository

# Список репозиториев пользователя
gh repo list
gh repo list username
gh repo list --limit 50 --visibility public
```

### Работа с Issues
```bash
# Просмотр всех issues
gh issue list
gh issue list --state open
gh issue list --state closed
gh issue list --assignee @me

# Фильтрация issues
gh issue list --label "bug"
gh issue list --label "enhancement,documentation"
gh issue list --author "username"
gh issue list --milestone "v1.0"

# Просмотр конкретного issue
gh issue view 123
gh issue view 123 --web  # Открыть в браузере

# Создание нового issue
gh issue create
gh issue create --title "Bug: Login not working" --body "Description of the bug"
gh issue create --title "Feature request" --body "$(cat issue-template.md)"

# Создание issue с метками и исполнителями
gh issue create --title "Documentation update" --label "documentation" --assignee "username"

# Редактирование issue
gh issue edit 123 --title "New title"
gh issue edit 123 --body "New description"
gh issue edit 123 --add-label "priority:high"
gh issue edit 123 --remove-label "bug"
gh issue edit 123 --add-assignee "username"

# Изменение статуса issue
gh issue close 123
gh issue close 123 --comment "Fixed in PR #456"
gh issue reopen 123

# Комментирование issues
gh issue comment 123 --body "Thanks for reporting this!"
gh issue comment 123 --body "$(cat comment.md)"

# Поиск issues
gh search issues "authentication bug" --repo username/repository
gh search issues "is:open label:bug author:username"
```

### Работа с Pull Requests
```bash
# Просмотр всех PR
gh pr list
gh pr list --state open
gh pr list --state merged
gh pr list --author @me

# Просмотр конкретного PR
gh pr view 456
gh pr view 456 --web

# Создание PR
gh pr create
gh pr create --title "feat: add user authentication" --body "Description"
gh pr create --draft  # Черновик PR

# Создание PR с автозаполнением
gh pr create --fill  # Использует заголовок и описание из коммитов

# Создание PR в другой репозиторий (для форков)
gh pr create --repo upstream-owner/repository

# Checkout PR локально
gh pr checkout 456
gh pr checkout https://github.com/owner/repo/pull/456

# Просмотр изменений в PR
gh pr diff 456
gh pr diff 456 --name-only

# Обновление PR
gh pr edit 456 --title "New title"
gh pr edit 456 --body "New description"
gh pr edit 456 --add-reviewer "username"
gh pr edit 456 --add-assignee "username"

# Слияние PR
gh pr merge 456
gh pr merge 456 --merge     # Создать merge commit
gh pr merge 456 --squash    # Squash and merge
gh pr merge 456 --rebase    # Rebase and merge

# Закрытие PR без слияния
gh pr close 456
gh pr close 456 --comment "Closing due to..."

# Переоткрытие PR
gh pr reopen 456

# Комментирование PR
gh pr comment 456 --body "Looks good to me!"
gh pr review 456 --approve --body "LGTM!"
gh pr review 456 --request-changes --body "Please fix..."
gh pr review 456 --comment --body "Question about line 15"

# Синхронизация с upstream (для форков)
gh pr checkout 456
git fetch upstream
git merge upstream/main
git push

# Проверка статуса CI/CD
gh pr checks 456
gh pr checks 456 --watch  # Отслеживание в реальном времени
```

### Работа с Releases
```bash
# Просмотр релизов
gh release list
gh release list --limit 10

# Просмотр конкретного релиза
gh release view v1.0.0
gh release view latest

# Создание релиза
gh release create v1.0.0
gh release create v1.0.0 --title "Version 1.0.0" --notes "Release notes"
gh release create v1.0.0 --draft  # Черновик релиза
gh release create v1.0.0 --prerelease  # Предварительный релиз

# Создание релиза с файлами
gh release create v1.0.0 ./dist/*.zip ./dist/*.tar.gz

# Автогенерация release notes
gh release create v1.0.0 --generate-notes

# Загрузка файлов к существующему релизу
gh release upload v1.0.0 ./dist/app.zip

# Скачивание релиза
gh release download v1.0.0
gh release download v1.0.0 --pattern "*.zip"

# Удаление релиза
gh release delete v1.0.0
```

### Работа с Gists
```bash
# Создание gist
gh gist create file.txt
gh gist create file1.txt file2.txt --public
gh gist create --desc "My awesome script" script.sh

# Список gists
gh gist list
gh gist list --public
gh gist list --secret

# Просмотр gist
gh gist view abc123
gh gist view abc123 --web

# Редактирование gist
gh gist edit abc123

# Клонирование gist
gh gist clone abc123

# Удаление gist
gh gist delete abc123
```

### Работа с GitHub Actions
```bash
# Просмотр workflow runs
gh run list
gh run list --workflow "CI"
gh run list --branch main

# Просмотр конкретного run
gh run view 123456789
gh run view 123456789 --log

# Скачивание артефактов
gh run download 123456789

# Отмена run
gh run cancel 123456789

# Перезапуск failed run
gh run rerun 123456789

# Просмотр workflow файлов
gh workflow list
gh workflow view "CI"
gh workflow run "CI"  # Ручной запуск workflow
```

### Работа с организациями и командами
```bash
# Просмотр организаций
gh api user/orgs

# Список участников организации
gh api orgs/org-name/members

# Работа с командами
gh api orgs/org-name/teams
gh api teams/team-id/members
```

### Секреты и переменные
```bash
# Просмотр секретов репозитория
gh secret list

# Установка секрета
gh secret set SECRET_NAME --body "secret-value"
gh secret set SECRET_NAME < secret-file.txt

# Удаление секрета
gh secret remove SECRET_NAME

# Работа с переменными
gh variable list
gh variable set VAR_NAME --body "value"
gh variable delete VAR_NAME
```

### Полезные комбинации команд
```bash
# Быстрое создание issue из файла
gh issue create --title "$(head -1 issue.md)" --body "$(tail -n +2 issue.md)"

# Создание PR после пуша ветки
git push -u origin feature/new-feature
gh pr create --fill

# Checkout PR, тестирование и merge
gh pr checkout 123
npm test
gh pr merge 123 --squash

# Создание релиза с тегом
git tag v1.0.0
git push origin v1.0.0
gh release create v1.0.0 --generate-notes

# Быстрый форк и клонирование
gh repo fork username/repository --clone
cd repository
gh repo set-default

# Массовое закрытие старых issues
gh issue list --state open --json number --jq '.[].number' | head -5 | xargs -I {} gh issue close {}
```

### Конфигурация и настройки
```bash
# Просмотр текущей конфигурации
gh config list

# Установка настроек
gh config set git_protocol ssh
gh config set editor "code --wait"
gh config set prompt enabled
gh config set pager less

# Просмотр алиасов
gh alias list

# Создание алиасов
gh alias set prc 'pr create --fill'
gh alias set prs 'pr status'
gh alias set il 'issue list'
gh alias set rv 'repo view'

# Использование алиасов
gh prc  # Эквивалент: gh pr create --fill
gh il   # Эквивалент: gh issue list
```

### API запросы
```bash
# Прямые API запросы
gh api user
gh api repos/:owner/:repo
gh api repos/:owner/:repo/issues

# POST запросы
gh api repos/:owner/:repo/issues --field title="Bug report" --field body="Description"

# Обработка JSON с jq
gh api repos/:owner/:repo/issues | jq '.[].title'
gh api user --jq '.login'

# Пагинация
gh api repos/:owner/:repo/issues --paginate

# Загрузка файлов
gh api repos/:owner/:repo/releases/tags/v1.0.0/assets --field name=app.zip --field data=@app.zip
```

### Полезные советы
```bash
# Установка репозитория по умолчанию (избегает необходимости указывать --repo)
gh repo set-default

# Использование с другими инструментами
gh issue list --json number,title | jq -r '.[] | "\(.number): \(.title)"'

# Быстрое создание PR из коммитов
git log --oneline main..HEAD | gh pr create --title "$(git log -1 --format=%s)" --body "$(git log main..HEAD --format='- %s')"

# Backup issues в JSON
gh issue list --state all --limit 1000 --json number,title,body,state > issues-backup.json

# Синхронизация labels между репозиториями
gh label list --json name,color,description | gh label create --repo target/repo
```

---

## 🔍 Продвинутые команды

### Поиск и фильтрация
```bash
# Поиск текста в файлах
git grep "function"
git grep -n "TODO"          # С номерами строк
git grep -i "error"         # Регистронезависимый поиск

# Поиск в истории коммитов
git log -S "function_name"   # Поиск добавления/удаления кода
git log -G "pattern"         # Поиск по регулярному выражению

# Показать коммиты, которые изменили конкретную строку
git blame file.txt
git blame -L 10,20 file.txt  # Конкретные строки

# Найти коммит, который ввел баг
git bisect start
git bisect bad               # Текущий коммит плохой
git bisect good abc1234      # Этот коммит хороший
# Git будет предлагать коммиты для тестирования
git bisect good              # Этот коммит хороший
git bisect bad               # Этот коммит плохой
git bisect reset             # Завершить bisect
```

### Работа с тегами
```bash
# Создание тега
git tag v1.0.0
git tag -a v1.0.0 -m "Release version 1.0.0"

# Создание тега для конкретного коммита
git tag v0.9.0 abc1234

# Показать теги
git tag
git tag -l "v1.*"           # Фильтр по шаблону

# Показать информацию о теге
git show v1.0.0

# Отправка тегов
git push origin v1.0.0
git push origin --tags

# Удаление тега
git tag -d v1.0.0           # Локально
git push origin :refs/tags/v1.0.0  # С удаленного репозитория
```

### Подмодули
```bash
# Добавление подмодуля
git submodule add https://github.com/user/repo.git path/to/submodule

# Инициализация подмодулей после клонирования
git submodule init
git submodule update

# Клонирование с подмодулями
git clone --recursive https://github.com/user/repo.git

# Обновление подмодулей
git submodule update --remote

# Удаление подмодуля
git submodule deinit path/to/submodule
git rm path/to/submodule
```

### Архивирование
```bash
# Создание архива
git archive --format=zip --output=project.zip HEAD

# Архив конкретной ветки
git archive --format=tar.gz --output=release.tar.gz v1.0.0

# Архив конкретной папки
git archive --format=zip --output=src.zip HEAD:src/
```

---

## ⚡ Полезные алиасы

### Настройка алиасов
```bash
# Короткие команды
git config --global alias.st status
git config --global alias.co checkout
git config --global alias.br branch
git config --global alias.ci commit

# Расширенные алиасы
git config --global alias.unstage 'reset HEAD --'
git config --global alias.last 'log -1 HEAD'
git config --global alias.visual '!gitk'

# Красивая история
git config --global alias.lg "log --color --graph --pretty=format:'%Cred%h%Creset -%C(yellow)%d%Creset %s %Cgreen(%cr) %C(bold blue)<%an>%Creset' --abbrev-commit"

# Показать ветки с последними коммитами
git config --global alias.branches "branch -v"

# Быстрый push
git config --global alias.pushup "push -u origin HEAD"

# Синхронизация с main
git config --global alias.sync "!git checkout main && git pull origin main && git checkout - && git rebase main"
```

### Использование алиасов
```bash
# Вместо git status
git st

# Вместо git checkout
git co feature/new-branch

# Красивая история
git lg

# Быстрый push новой ветки
git pushup

# Синхронизация с main
git sync
```

---

## 🛠️ Troubleshooting

### Частые проблемы и решения

#### "Detached HEAD state"
```bash
# Создать ветку из текущего состояния
git checkout -b new-branch-name

# Вернуться к последней ветке
git checkout -
```

#### Случайно удалили коммиты
```bash
# Найти потерянные коммиты
git reflog

# Восстановить коммит
git checkout abc1234
git checkout -b recovered-branch
```

#### Проблемы с line endings
```bash
# Настройка для команды
git config core.autocrlf false
git config core.eol lf

# Переконвертировать все файлы
git add --renormalize .
git commit -m "Normalize line endings"
```

#### Большие файлы в истории
```bash
# Найти большие файлы
git rev-list --objects --all | git cat-file --batch-check='%(objecttype) %(objectname) %(objectsize) %(rest)' | sed -n 's/^blob //p' | sort --numeric-sort --key=2 | tail -10

# Удалить файл из истории (осторожно!)
git filter-branch --force --index-filter 'git rm --cached --ignore-unmatch path/to/large/file' --prune-empty --tag-name-filter cat -- --all
```

#### Проблемы с SSL сертификатами
```bash
# Отключить проверку SSL (не рекомендуется)
git config --global http.sslVerify false

# Указать путь к сертификатам
git config --global http.sslCAInfo /path/to/certificate.pem
```

### Восстановление и диагностика
```bash
# Проверка целостности репозитория
git fsck

# Сбор мусора и оптимизация
git gc
git gc --aggressive

# Показать размер репозитория
git count-objects -vH

# Очистка недостижимых объектов
git prune

# Показать конфигурацию для диагностики
git config --list --show-origin
```

---

## 📚 Дополнительные ресурсы

### Полезные ссылки
- [Pro Git Book](https://git-scm.com/book) - Официальная документация
- [Git Cheat Sheet](https://education.github.com/git-cheat-sheet-education.pdf) - Шпаргалка
- [Conventional Commits](https://www.conventionalcommits.org/) - Стандарт оформления коммитов
- [Git Flow](https://nvie.com/posts/a-successful-git-branching-model/) - Модель ветвления

### Рекомендуемые инструменты
- **GitKraken** - GUI клиент для Git
- **Sourcetree** - Бесплатный GUI от Atlassian
- **VS Code** - Встроенная поддержка Git
- **oh-my-zsh** - Плагины для терминала с Git shortcuts

---

*"Мастерство приходит с практикой. Используйте Git ежедневно и экспериментируйте с командами в тестовых репозиториях!"* 🚀
---

**📅 Актуальность документа**  
**Последнее обновление**: 2025-01-15  
**Версия проекта**: 0.2.0  
**Статус**: Актуальный

