.PHONY: help build test lint statictest fmt vet check pre-commit clean install-hooks autotest iter1 iter2 iter3 iter4 download-metricstest

# Цвета для вывода
GREEN=\033[0;32m
RED=\033[0;31m
YELLOW=\033[0;33m
NC=\033[0m # No Color

# Пути к бинарникам
SERVER_BINARY=cmd/server/server
AGENT_BINARY=cmd/agent/agent
METRICSTEST=./metricstest

help: ## Показать справку
	@echo "$(GREEN)Доступные команды:$(NC)"
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(GREEN)%-20s$(NC) %s\n", $$1, $$2}'

build: ## Собрать все бинарники
	@echo "$(GREEN)Building binaries...$(NC)"
	go build -o $(SERVER_BINARY) ./cmd/server
	go build -o $(AGENT_BINARY) ./cmd/agent
	go build -o ./bin/statictest ./cmd/statictest
	@echo "$(GREEN)✅ Build complete!$(NC)"

test: ## Запустить тесты
	@echo "$(GREEN)Running tests...$(NC)"
	go test -v -race -coverprofile=coverage.out ./...
	@echo "$(GREEN)✅ Tests passed!$(NC)"

test-short: ## Запустить быстрые тесты
	@echo "$(GREEN)Running short tests...$(NC)"
	go test -short ./...
	@echo "$(GREEN)✅ Short tests passed!$(NC)"

coverage: test ## Показать покрытие кода тестами
	@echo "$(GREEN)Generating coverage report...$(NC)"
	go tool cover -html=coverage.out

download-metricstest: ## Скачать автотесты с GitHub
	@echo "$(GREEN)Downloading metricstest...$(NC)"
	@if [ -f "./scripts/download-metricstest.sh" ]; then \
		./scripts/download-metricstest.sh; \
	else \
		echo "$(RED)❌ Script not found: ./scripts/download-metricstest.sh$(NC)"; \
		echo "$(YELLOW)Download manually from:$(NC)"; \
		echo "  https://github.com/Yandex-Practicum/go-autotests/releases"; \
		exit 1; \
	fi

# Автотесты по итерациям
check-metricstest:
	@if [ ! -f "$(METRICSTEST)" ]; then \
		echo "$(RED)❌ metricstest не найден!$(NC)"; \
		echo "$(YELLOW)Скачайте его командой:$(NC)"; \
		echo "  make download-metricstest"; \
		echo "$(YELLOW)Или вручную с:$(NC)"; \
		echo "  https://github.com/Yandex-Practicum/go-autotests/releases"; \
		exit 1; \
	fi
	@chmod +x $(METRICSTEST)

autotest: check-metricstest ## Запустить автотесты (все инкременты)
	@echo "$(GREEN)Running autotests...$(NC)"
	$(METRICSTEST)
	@echo "$(GREEN)✅ Autotests passed!$(NC)"

iter1: build check-metricstest ## Автотесты итерации 1
	@echo "$(GREEN)Running iteration 1 tests...$(NC)"
	$(METRICSTEST) -test.v -test.run='^TestIteration1$$' \
		-binary-path=$(SERVER_BINARY)
	@echo "$(GREEN)✅ Iteration 1 passed!$(NC)"

iter2: build check-metricstest ## Автотесты итерации 2
	@echo "$(GREEN)Running iteration 2 tests...$(NC)"
	$(METRICSTEST) -test.v -test.run='^TestIteration2[AB]*$$' \
		-source-path=. \
		-agent-binary-path=$(AGENT_BINARY)
	@echo "$(GREEN)✅ Iteration 2 passed!$(NC)"

iter3: build check-metricstest ## Автотесты итерации 3
	@echo "$(GREEN)Running iteration 3 tests...$(NC)"
	$(METRICSTEST) -test.v -test.run='^TestIteration3[AB]*$$' \
		-source-path=. \
		-agent-binary-path=$(AGENT_BINARY) \
		-binary-path=$(SERVER_BINARY)
	@echo "$(GREEN)✅ Iteration 3 passed!$(NC)"

iter4: build check-metricstest ## Автотесты итерации 4
	@echo "$(GREEN)Running iteration 4 tests...$(NC)"
	$(METRICSTEST) -test.v -test.run='^TestIteration4$$' \
		-agent-binary-path=$(AGENT_BINARY) \
		-binary-path=$(SERVER_BINARY) \
		-server-port=8080 \
		-source-path=.
	@echo "$(GREEN)✅ Iteration 4 passed!$(NC)"


fmt: ## Форматировать код
	@echo "$(GREEN)Formatting code...$(NC)"
	gofmt -w .
	@echo "$(GREEN)✅ Code formatted!$(NC)"

fmt-check: ## Проверить форматирование кода
	@echo "$(GREEN)Checking code formatting...$(NC)"
	@UNFORMATTED=$$(gofmt -l . 2>&1 | grep -v "^vendor/" | grep ".go$$" || true); \
	if [ -n "$$UNFORMATTED" ]; then \
		echo "$(RED)❌ The following files are not formatted:$(NC)"; \
		echo "$$UNFORMATTED"; \
		echo "$(YELLOW)Run: make fmt$(NC)"; \
		exit 1; \
	fi
	@echo "$(GREEN)✅ Code formatting OK!$(NC)"

vet: ## Запустить go vet
	@echo "$(GREEN)Running go vet...$(NC)"
	go vet ./...
	@echo "$(GREEN)✅ go vet passed!$(NC)"

statictest: ## Запустить statictest
	@echo "$(GREEN)Running statictest...$(NC)"
	@if [ ! -f "./bin/statictest" ]; then \
		echo "$(YELLOW)Building statictest...$(NC)"; \
		go build -o ./bin/statictest ./cmd/statictest; \
	fi
	go vet -vettool=./bin/statictest ./...
	@echo "$(GREEN)✅ statictest passed!$(NC)"

lint: fmt-check vet statictest ## Запустить все линтеры

check: lint test-short ## Полная проверка перед коммитом (быстрая)

pre-commit: lint test ## Полная проверка перед коммитом (с тестами)

install-hooks: ## Установить git hooks
	@echo "$(GREEN)Installing git hooks...$(NC)"
	@chmod +x .git/hooks/pre-commit
	@chmod +x .git/hooks/pre-commit.light
	@echo "$(GREEN)✅ Git hooks installed!$(NC)"
	@echo "$(YELLOW)Tip: To use light version, run:$(NC)"
	@echo "  cp .git/hooks/pre-commit.light .git/hooks/pre-commit"

clean: ## Очистить сгенерированные файлы
	@echo "$(GREEN)Cleaning...$(NC)"
	rm -f coverage.out
	rm -f $(SERVER_BINARY)
	rm -f $(AGENT_BINARY)
	rm -f ./bin/statictest
	@echo "$(GREEN)✅ Cleaned!$(NC)"

.DEFAULT_GOAL := help

