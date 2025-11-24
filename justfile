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
    golangci-lint run --config=.github/workflows/golangci.yml

# Форматировать код
fmt:
    go fmt ./...

# Запустить миграции
migrate-up:
    @echo "Запуск миграций..."
    docker run --rm -v "$(pwd)/migrations:/migrations" --network host migrate/migrate:latest -path=/migrations -database "postgres://$(grep DB_USER .env | cut -d= -f2):$(grep DB_PASSWORD .env | cut -d= -f2)@$(grep DB_HOST .env | cut -d= -f2):$(grep DB_PORT .env | cut -d= -f2)/$(grep DB_NAME .env | cut -d= -f2)?sslmode=disable" up

# Откатить все миграции
migrate-down:
    @echo "Откат всех миграций..."
    docker run --rm -v "$(pwd)/migrations:/migrations" --network host migrate/migrate:latest -path=/migrations -database "postgres://$(grep DB_USER .env | cut -d= -f2):$(grep DB_PASSWORD .env | cut -d= -f2)@$(grep DB_HOST .env | cut -d= -f2):$(grep DB_PORT .env | cut -d= -f2)/$(grep DB_NAME .env | cut -d= -f2)?sslmode=disable" down

# Откатить одну последнюю миграцию
migrate-down-one:
    @echo "Откат одной миграции..."
    docker run --rm -v "$(pwd)/migrations:/migrations" --network host migrate/migrate:latest -path=/migrations -database "postgres://$(grep DB_USER .env | cut -d= -f2):$(grep DB_PASSWORD .env | cut -d= -f2)@$(grep DB_HOST .env | cut -d= -f2):$(grep DB_PORT .env | cut -d= -f2)/$(grep DB_NAME .env | cut -d= -f2)?sslmode=disable" down 1

# Откатить N миграций
migrate-down-n N:
    @echo "Откат {{N}} миграций..."
    docker run --rm -v "$(pwd)/migrations:/migrations" --network host migrate/migrate:latest -path=/migrations -database "postgres://$(grep DB_USER .env | cut -d= -f2):$(grep DB_PASSWORD .env | cut -d= -f2)@$(grep DB_HOST .env | cut -d= -f2):$(grep DB_PORT .env | cut -d= -f2)/$(grep DB_NAME .env | cut -d= -f2)?sslmode=disable" down {{N}}

# Применить конкретную версию миграции
migrate-goto VERSION:
    @echo "Переход к версии {{VERSION}}..."
    docker run --rm -v "$(pwd)/migrations:/migrations" --network host migrate/migrate:latest -path=/migrations -database "postgres://$(grep DB_USER .env | cut -d= -f2):$(grep DB_PASSWORD .env | cut -d= -f2)@$(grep DB_HOST .env | cut -d= -f2):$(grep DB_PORT .env | cut -d= -f2)/$(grep DB_NAME .env | cut -d= -f2)?sslmode=disable" goto {{VERSION}}

# Форсировать версию (для восстановления после ошибок)
migrate-force VERSION:
    @echo "Установка версии {{VERSION}} без применения миграции..."
    docker run --rm -v "$(pwd)/migrations:/migrations" --network host migrate/migrate:latest -path=/migrations -database "postgres://$(grep DB_USER .env | cut -d= -f2):$(grep DB_PASSWORD .env | cut -d= -f2)@$(grep DB_HOST .env | cut -d= -f2):$(grep DB_PORT .env | cut -d= -f2)/$(grep DB_NAME .env | cut -d= -f2)?sslmode=disable" force {{VERSION}}

# Показать текущую версию БД
migrate-version:
    @echo "Текущая версия миграции:"
    docker run --rm -v "$(pwd)/migrations:/migrations" --network host migrate/migrate:latest -path=/migrations -database "postgres://$(grep DB_USER .env | cut -d= -f2):$(grep DB_PASSWORD .env | cut -d= -f2)@$(grep DB_HOST .env | cut -d= -f2):$(grep DB_PORT .env | cut -d= -f2)/$(grep DB_NAME .env | cut -d= -f2)?sslmode=disable" version

# Создать новую миграцию
migrate-create NAME:
    @echo "Создание миграции: {{NAME}}"
    docker run --rm -v "$(pwd)/migrations:/migrations" migrate/migrate:latest create -ext sql -dir /migrations -seq {{NAME}}

# Настроить тестовую БД
setup-test-db:
    @echo "Настройка тестовой БД..."
    docker exec postgres-dev psql -U postgres -c "CREATE DATABASE inf_sys_db_test;" || true
    migrate -path migrations -database "postgres://$(grep DB_USER .env | cut -d= -f2):$(grep DB_PASSWORD .env | cut -d= -f2)@$(grep DB_HOST .env | cut -d= -f2):$(grep DB_PORT .env | cut -d= -f2)/inf_sys_db_test?sslmode=disable" up

# Удалить тестовую БД
drop-test-db:
    @echo "Удаление тестовой БД..."
    docker exec postgres-dev psql -U postgres -c "DROP DATABASE IF EXISTS inf_sys_db_test;"
