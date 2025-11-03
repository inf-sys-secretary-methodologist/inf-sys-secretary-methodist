# Justfile для проекта inf-sys-secretary-methodist

# Показать список всех команд
default:
    @just --list

# Запустить все тесты
test:
    go test -v -race ./...

# Запустить только unit тесты (быстро, без БД)
test-unit:
    go test -v -race -short ./...

# Запустить integration тесты (требуют БД)
test-integration:
    go test -v -race -run Integration ./...

# Сгенерировать coverage report
test-coverage:
    go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
    go tool cover -func=coverage.out

# Сгенерировать HTML coverage report
test-coverage-html:
    go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
    go tool cover -html=coverage.out -o coverage.html
    @echo "Coverage report generated: coverage.html"

# Запустить тесты с подробным выводом
test-verbose:
    go test -v -race -count=1 ./...

# Запустить тесты конкретного пакета
test-package PKG:
    go test -v -race ./{{PKG}}

# Очистить кеш тестов и файлы coverage
clean:
    go clean -testcache
    rm -f coverage.out coverage.html

# Собрать приложение
build:
    go build -o bin/server ./cmd/server

# Запустить приложение
run:
    go run ./cmd/server

# Установить зависимости
deps:
    go mod download
    go mod tidy

# Lint код
lint:
    golangci-lint run

# Форматировать код
fmt:
    go fmt ./...

# Запустить миграции
migrate-up:
    @echo "Running migrations..."
    # TODO: Добавить команду миграций когда инструмент настроен

# Откатить миграции
migrate-down:
    @echo "Rolling back migrations..."
    # TODO: Добавить команду отката миграций

# Настроить тестовую БД
setup-test-db:
    @echo "Setting up test database..."
    createdb secretary_methodist_test || true
    # TODO: Добавить миграции для test DB

# Удалить тестовую БД
drop-test-db:
    @echo "Dropping test database..."
    dropdb secretary_methodist_test || true
