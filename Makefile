.PHONY: help build test test-integration test-coverpkg lint statictest linter staticcheck fmt vet check pre-commit clean install-hooks autotest iter1 iter2 iter3 iter4 iter5 iter6 iter7 iter8 iter9 iter10 iter11 iter12 iter13 iter14 download-metricstest coverage coverage-percent coverage-packages

# Цвета для вывода
GREEN=\033[0;32m
RED=\033[0;31m
YELLOW=\033[0;33m
NC=\033[0m # No Color

# Пути к бинарникам
SERVER_BINARY=cmd/server/server
AGENT_BINARY=cmd/agent/agent
METRICSTEST=metricstest
LINTER_BINARY=bin/linter

# DSN для локального запуска iter10/11/12 (переопредели: make iter12 DATABASE_DSN='...')
DATABASE_DSN ?= postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable

help: ## Показать справку
	@echo "$(GREEN)Доступные команды:$(NC)"
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(GREEN)%-20s$(NC) %s\n", $$1, $$2}'

build: ## Собрать все бинарники
	@echo "$(GREEN)Building binaries...$(NC)"
	go build -o $(SERVER_BINARY) ./cmd/server
	go build -o $(AGENT_BINARY) ./cmd/agent
	go build -o ./bin/statictest ./cmd/statictest
	go build -o ./$(LINTER_BINARY) ./cmd/linter
	@echo "$(GREEN)✅ Build complete!$(NC)"

COVER_EXCLUDE ?= cmd/statictest

test: ## Запустить тесты (без integration)
	@echo "$(GREEN)Running tests...$(NC)"
	go test -v -race -count=1 -coverprofile=coverage.out $$(go list ./... | grep -vE '$(COVER_EXCLUDE)')
	@echo "$(GREEN)✅ Tests passed!$(NC)"
test-integration: ## Тесты с тегом integration
	@echo "$(GREEN)Running tests (with integration)...$(NC)"
	go test -v -race -count=1 -tags=integration -coverprofile=coverage.out $$(go list ./... | grep -vE '$(COVER_EXCLUDE)')
	@echo "$(GREEN)✅ Tests passed!$(NC)"

# Тесты (с integration) с -coverpkg:
test-coverpkg: ## Тесты с тегом integration и -coverpkg
	@echo "$(GREEN)Running tests (integration + coverpkg)...$(NC)"
	@PKGS=$$(go list ./... | grep -vE '$(COVER_EXCLUDE)'); \
	COVERPKG=$$(echo "$$PKGS" | tr '\n' ',' | sed 's/,$$//'); \
	go test -v -race -count=1 -tags=integration -coverpkg="$$COVERPKG" -coverprofile=coverage.out $$PKGS
	@echo "$(GREEN)✅ Tests passed!$(NC)"

test-short: ## Запустить быстрые тесты
	@echo "$(GREEN)Running short tests...$(NC)"
	go test -short ./...
	@echo "$(GREEN)✅ Short tests passed!$(NC)"

coverage: test ## Показать покрытие
	@echo "$(GREEN)Generating coverage report...$(NC)"
	go tool cover  -html=coverage.out

coverage-percent: ## Показать общий процент покрытия
	@if [ ! -f coverage.out ]; then \
		echo "$(RED)❌ coverage.out не найден. Запустите: make test или make test-integration$(NC)"; \
		exit 1; \
	fi
	@echo "$(GREEN)Покрытие кода:$(NC)"
	@go tool cover -func=coverage.out | grep total | awk '{printf "  Всего: $(GREEN)%s$(NC)\n", $$3}'

coverage-packages: ## Показать процент покрытия по пакетам
	@if [ ! -f coverage.out ]; then \
		echo "$(RED)❌ coverage.out не найден. Запустите: make test или make test-integration$(NC)"; \
		exit 1; \
	fi
	@echo "$(GREEN)Покрытие по пакетам:$(NC)"
	@go tool cover -func=coverage.out | awk '$$0 !~ /^total/ { \
		path=$$1; sub(/:.*$$/, "", path); \
		match(path, /.*\//); pkg=(RLENGTH>0) ? substr(path, 1, RLENGTH-1) : "."; \
		gsub(/%/, "", $$3); sum[pkg]+=$$3; cnt[pkg]++ } \
		END { for (p in sum) printf "%6.1f%%  %s\n", sum[p]/cnt[p], p }' | sort -k1 -n
	@echo ""
	@echo "$(GREEN)Итого:$(NC)"
	@go tool cover -func=coverage.out | grep total | awk '{printf "  $(GREEN)%s$(NC)\n", $$3}'
# Автотесты по итерациям
check-metricstest:
	@if ! command -v metricstest >/dev/null 2>&1; then \
		echo "$(RED)❌ metricstest не найден в PATH!$(NC)"; \
		echo "$(YELLOW)Скачайте его командой:$(NC)"; \
		echo "  make download-metricstest"; \
		echo "$(YELLOW)Или вручную с:$(NC)"; \
		echo "  https://github.com/Yandex-Practicum/go-autotests/releases"; \
		exit 1; \
	fi

autotest: check-metricstest ## Запустить автотесты (все инкременты)
	@echo "$(GREEN)Running autotests...$(NC)"
	$(METRICSTEST)
	@echo "$(GREEN)✅ Autotests passed!$(NC)"
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

linter: ## Запустить собственный multichecker (cmd/linter)
	@echo "$(GREEN)Running linter...$(NC)"
	@if [ ! -f "./$(LINTER_BINARY)" ]; then \
		echo "$(YELLOW)Building linter...$(NC)"; \
		go build -o ./$(LINTER_BINARY) ./cmd/linter; \
	fi
	./$(LINTER_BINARY) ./...
	@echo "$(GREEN)✅ linter passed!$(NC)"

staticcheck: ## Запустить staticcheck (honnef.co/go/tools)
	@echo "$(GREEN)Running staticcheck...$(NC)"
	@if ! command -v staticcheck >/dev/null 2>&1; then \
		echo "$(YELLOW)Installing staticcheck...$(NC)"; \
		go install honnef.co/go/tools/cmd/staticcheck@latest; \
	fi
	staticcheck ./...
	@echo "$(GREEN)✅ staticcheck passed!$(NC)"

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
	rm -f ./$(LINTER_BINARY)
	@echo "$(GREEN)✅ Cleaned!$(NC)"
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

iter5: build check-metricstest ## Автотесты итерации 5
	@echo "$(GREEN)Running iteration 5 tests...$(NC)"
	@SERVER_PORT=$$(python3 -c 'import socket; s=socket.socket(); s.bind(("", 0)); print(s.getsockname()[1]); s.close()'); \
	ADDRESS="localhost:$$SERVER_PORT"; \
	TEMP_FILE=$$(mktemp); \
	echo "$(YELLOW)Using random port: $$SERVER_PORT$(NC)"; \
	$(METRICSTEST) -test.v -test.run='^TestIteration5$$' \
		-agent-binary-path=$(AGENT_BINARY) \
		-binary-path=$(SERVER_BINARY) \
		-server-port=$$SERVER_PORT \
		-source-path=.; \
	rm -f $$TEMP_FILE
	@echo "$(GREEN)✅ Iteration 5 passed!$(NC)"

iter6: build check-metricstest ## Автотесты итерации 6
	@echo "$(GREEN)Running iteration 6 tests...$(NC)"
	$(METRICSTEST) -test.v -test.run='^TestIteration6$$' \
		-agent-binary-path=$(AGENT_BINARY) \
		-binary-path=$(SERVER_BINARY) \
		-server-port=8080 \
		-source-path=.
	@echo "$(GREEN)✅ Iteration 6 passed!$(NC)"

iter7: build check-metricstest ## Автотесты итерации 7
	@echo "$(GREEN)Running iteration 7 tests...$(NC)"
	@SERVER_PORT=$$(python3 -c 'import socket; s=socket.socket(); s.bind(("", 0)); print(s.getsockname()[1]); s.close()'); \
	ADDRESS="localhost:$$SERVER_PORT"; \
	TEMP_FILE=$$(mktemp); \
	echo "$(YELLOW)Using random port: $$SERVER_PORT$(NC)"; \
	$(METRICSTEST) -test.v -test.run='^TestIteration7$$' \
		-agent-binary-path=$(AGENT_BINARY) \
		-binary-path=$(SERVER_BINARY) \
		-server-port=$$SERVER_PORT \
		-source-path=.; \
	rm -f $$TEMP_FILE
	@echo "$(GREEN)✅ Iteration 7 passed!$(NC)"
iter8: build check-metricstest ## Автотесты итерации 8
	@echo "$(GREEN)Running iteration 8 tests...$(NC)"
	@SERVER_PORT=$$(python3 -c 'import socket; s=socket.socket(); s.bind(("", 0)); print(s.getsockname()[1]); s.close()'); \
	ADDRESS="localhost:$$SERVER_PORT"; \
	TEMP_FILE=$$(mktemp); \
	echo "$(YELLOW)Using random port: $$SERVER_PORT$(NC)"; \
	$(METRICSTEST) -test.v -test.run='^TestIteration8$$' \
		-agent-binary-path=$(AGENT_BINARY) \
		-binary-path=$(SERVER_BINARY) \
		-server-port=$$SERVER_PORT \
		-source-path=.; \
	rm -f $$TEMP_FILE
	@echo "$(GREEN)✅ Iteration 8 passed!$(NC)"
iter9: build check-metricstest ## Автотесты итерации 9
	@echo "$(GREEN)Running iteration 9 tests...$(NC)"
	@SERVER_PORT=$$(python3 -c 'import socket; s=socket.socket(); s.bind(("", 0)); print(s.getsockname()[1]); s.close()'); \
	ADDRESS="localhost:$$SERVER_PORT"; \
	TEMP_FILE=$$(mktemp); \
	echo "$(YELLOW)Using random port: $$SERVER_PORT$(NC)"; \
	$(METRICSTEST) -test.v -test.run='^TestIteration9$$' \
		-agent-binary-path=$(AGENT_BINARY) \
		-binary-path=$(SERVER_BINARY) \
		-server-port=$$SERVER_PORT \
		-file-storage-path=./backup \
		-source-path=.; \
	rm -f $$TEMP_FILE
	@echo "$(GREEN)✅ Iteration 9 passed!$(NC)"
iter10: build check-metricstest ## Автотесты итерации 10
	@echo "$(GREEN)Running iteration 10 tests...$(NC)"
	@SERVER_PORT=$$(python3 -c 'import socket; s=socket.socket(); s.bind(("", 0)); print(s.getsockname()[1]); s.close()'); \
	ADDRESS="localhost:$$SERVER_PORT"; \
	TEMP_FILE=$$(mktemp); \
	echo "$(YELLOW)Using random port: $$SERVER_PORT$(NC)"; \
	$(METRICSTEST) -test.v -test.run='^TestIteration10$$' \
		-agent-binary-path=$(AGENT_BINARY) \
		-binary-path=$(SERVER_BINARY) \
		-server-port=$$SERVER_PORT \
		-database-dsn='$(DATABASE_DSN)' \
		-file-storage-path=./backup \
		-source-path=.; \
	rm -f $$TEMP_FILE
	@echo "$(GREEN)✅ Iteration 10 passed!$(NC)"
iter11: build check-metricstest ## Автотесты итерации 11
	@echo "$(GREEN)Running iteration 11 tests...$(NC)"
	@SERVER_PORT=$$(python3 -c 'import socket; s=socket.socket(); s.bind(("", 0)); print(s.getsockname()[1]); s.close()'); \
	ADDRESS="localhost:$$SERVER_PORT"; \
	TEMP_FILE=$$(mktemp); \
	echo "$(YELLOW)Using random port: $$SERVER_PORT$(NC)"; \
	$(METRICSTEST) -test.v -test.run='^TestIteration11$$' \
		-agent-binary-path=$(AGENT_BINARY) \
		-binary-path=$(SERVER_BINARY) \
		-server-port=$$SERVER_PORT \
		-database-dsn='$(DATABASE_DSN)' \
		-source-path=. ; \
	rm -f $$TEMP_FILE
	@echo "$(GREEN)✅ Iteration 11 passed!$(NC)"
iter12: build check-metricstest ## Автотесты итерации 12
	@echo "$(GREEN)Running iteration 12 tests...$(NC)"
	@SERVER_PORT=$$(python3 -c 'import socket; s=socket.socket(); s.bind(("", 0)); print(s.getsockname()[1]); s.close()'); \
	ADDRESS="localhost:$$SERVER_PORT"; \
	TEMP_FILE=$$(mktemp); \
	echo "$(YELLOW)Using random port: $$SERVER_PORT$(NC)"; \
	$(METRICSTEST) -test.v -test.run='^TestIteration12$$' \
		-agent-binary-path=$(AGENT_BINARY) \
		-binary-path=$(SERVER_BINARY) \
		-server-port=$$SERVER_PORT \
		-database-dsn='$(DATABASE_DSN)' \
		-file-storage-path=./backup \
		-source-path=.; \
	rm -f $$TEMP_FILE
	@echo "$(GREEN)✅ Iteration 12 passed!$(NC)"
iter13: build check-metricstest ## Автотесты итерации 13
	@echo "$(GREEN)Running iteration 13 tests...$(NC)"
	@SERVER_PORT=$$(python3 -c 'import socket; s=socket.socket(); s.bind(("", 0)); print(s.getsockname()[1]); s.close()'); \
	ADDRESS="localhost:$$SERVER_PORT"; \
	TEMP_FILE=$$(mktemp); \
	echo "$(YELLOW)Using random port: $$SERVER_PORT$(NC)"; \
	$(METRICSTEST) -test.v -test.run='^TestIteration13$$' \
		-agent-binary-path=$(AGENT_BINARY) \
		-binary-path=$(SERVER_BINARY) \
		-server-port=$$SERVER_PORT \
		-database-dsn='$(DATABASE_DSN)' \
		-file-storage-path=./backup \
		-source-path=.; \
	rm -f $$TEMP_FILE
	@echo "$(GREEN)✅ Iteration 13 passed!$(NC)"
iter14: build check-metricstest ## Автотесты итерации 14
	@echo "$(GREEN)Running iteration 14 tests...$(NC)"
	@SERVER_PORT=$$(python3 -c 'import socket; s=socket.socket(); s.bind(("", 0)); print(s.getsockname()[1]); s.close()'); \
	ADDRESS="localhost:$$SERVER_PORT"; \
	TEMP_FILE=$$(mktemp); \
	echo "$(YELLOW)Using random port: $$SERVER_PORT$(NC)"; \
	$(METRICSTEST) -test.v -test.run='^TestIteration14$$' \
		-agent-binary-path=$(AGENT_BINARY) \
		-binary-path=$(SERVER_BINARY) \
		-server-port=$$SERVER_PORT \
		-database-dsn='$(DATABASE_DSN)' \
		-file-storage-path=./backup \
		-key='test'\
		-source-path=.; \
	rm -f $$TEMP_FILE
	@echo "$(GREEN)✅ Iteration 14 passed!$(NC)"
.DEFAULT_GOAL := help

