.PHONY: help build test clean up down logs shell db-shell test-db prod-db analyze verify generate

# Цвета для вывода
GREEN  := $(shell tput -Txterm setaf 2)
YELLOW := $(shell tput -Txterm setaf 3)
WHITE  := $(shell tput -Txterm setaf 7)
RESET  := $(shell tput -Txterm sgr0)

help: ## Показать эту справку
	@echo ''
	@echo 'Использование:'
	@echo '  ${YELLOW}make${RESET} ${GREEN}<target>${RESET}'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  ${YELLOW}%-15s${RESET} %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Собрать Docker образ
	@echo "${GREEN}Сборка Docker образа...${RESET}"
	docker-compose build

test: ## Запустить все тесты
	@echo "${GREEN}Запуск тестов...${RESET}"
	docker-compose up --build --abort-on-container-exit pg-sec-lab
	@echo "${GREEN}Результаты тестов:${RESET}"
	@cat test-results/summary.txt 2>/dev/null || echo "Результаты не найдены"

test-quick: ## Быстрый запуск тестов (без пересборки)
	@echo "${GREEN}Быстрый запуск тестов...${RESET}"
	docker-compose up --abort-on-container-exit pg-sec-lab

up: ## Запустить все сервисы
	@echo "${GREEN}Запуск всех сервисов...${RESET}"
	docker-compose up -d postgres postgres-prod
	@echo "${GREEN}Ожидание готовности PostgreSQL...${RESET}"
	@sleep 5
	@echo "${GREEN}Сервисы запущены!${RESET}"

down: ## Остановить все сервисы
	@echo "${YELLOW}Остановка сервисов...${RESET}"
	docker-compose down

clean: ## Полная очистка (volumes, images, результаты)
	@echo "${YELLOW}Полная очистка...${RESET}"
	docker-compose down -v --rmi all
	rm -rf output/ test-results/ || true
	@echo "${GREEN}Очистка завершена${RESET}"

logs: ## Показать логи
	docker-compose logs -f

shell: ## Войти в shell контейнера pg-sec-lab
	docker-compose run --rm pg-sec-lab sh

db-shell: ## Войти в psql тестовой БД
	docker-compose exec postgres psql -U postgres -d testdb

prod-db-shell: ## Войти в psql продакшн БД
	docker-compose exec postgres-prod psql -U produser -d proddb

generate: ## Запустить команду generate
	@echo "${GREEN}Генерация SQL...${RESET}"
	docker-compose run --rm pg-sec-lab generate --policy /app/policy.yaml --out /app/output/generated.sql
	@echo "${GREEN}SQL сгенерирован: output/generated.sql${RESET}"

verify: up ## Запустить команду verify
	@echo "${GREEN}Проверка политик...${RESET}"
	docker-compose run --rm pg-sec-lab verify \
		--policy /app/policy.yaml \
		--dsn "postgres://postgres:postgres@postgres:5432/testdb?sslmode=disable"

analyze: up ## Запустить команду analyze
	@echo "${GREEN}Анализ конфигурации...${RESET}"
	docker-compose run --rm pg-sec-lab analyze \
		--dsn "postgres://postgres:postgres@postgres:5432/testdb?sslmode=disable" \
		--out /app/output/report.json
	@echo "${GREEN}Отчёт создан: output/report.json${RESET}"
	@cat output/report.json | jq '.' 2>/dev/null || cat output/report.json

analyze-prod: up ## Анализ продакшн БД
	@echo "${GREEN}Анализ продакшн конфигурации...${RESET}"
	docker-compose run --rm pg-sec-lab analyze \
		--dsn "postgres://produser:prodpass@postgres-prod:5432/proddb?sslmode=disable" \
		--out /app/output/report-prod.json
	@echo "${GREEN}Отчёт создан: output/report-prod.json${RESET}"
	@cat output/report-prod.json | jq '.' 2>/dev/null || cat output/report-prod.json

results: ## Показать результаты тестов
	@echo "${GREEN}=== Результаты тестов ===${RESET}"
	@cat test-results/summary.txt 2>/dev/null || echo "Результаты не найдены. Запустите 'make test'"
	@echo ""
	@echo "${GREEN}=== Сгенерированный SQL (первые 30 строк) ===${RESET}"
	@head -30 output/generated.sql 2>/dev/null || echo "SQL не найден"
	@echo ""
	@echo "${GREEN}=== JSON отчёт (findings) ===${RESET}"
	@cat output/report.json 2>/dev/null | jq '.findings' || echo "Отчёт не найден"

install-local: ## Локальная установка (без Docker)
	@echo "${GREEN}Установка зависимостей...${RESET}"
	go mod download
	@echo "${GREEN}Сборка приложения...${RESET}"
	go build -o pg-sec-lab.exe
	@echo "${GREEN}Установка завершена! Запустите: ./pg-sec-lab.exe${RESET}"

test-local: install-local ## Запуск локальных Go тестов
	@echo "${GREEN}Запуск Go тестов...${RESET}"
	go test -v ./...

benchmark: ## Бенчмарк производительности
	@echo "${GREEN}Запуск бенчмарка...${RESET}"
	@echo "Тест 1: Генерация SQL"
	@time docker-compose run --rm pg-sec-lab generate --policy /app/policy.yaml > /dev/null
	@echo ""
	@echo "Тест 2: Анализ БД"
	@time docker-compose run --rm pg-sec-lab analyze \
		--dsn "postgres://postgres:postgres@postgres:5432/testdb?sslmode=disable" \
		--out /dev/null

ci: build test ## CI pipeline (build + test)
	@echo "${GREEN}CI pipeline завершён${RESET}"

watch: ## Следить за изменениями и перезапускать тесты
	@echo "${YELLOW}Наблюдение за изменениями (Ctrl+C для остановки)...${RESET}"
	@while true; do \
		inotifywait -r -e modify,create,delete --exclude '(output|test-results)' . 2>/dev/null || sleep 2; \
		echo "${GREEN}Изменения обнаружены, перезапуск тестов...${RESET}"; \
		make test-quick; \
		sleep 1; \
	done

status: ## Показать статус сервисов
	@echo "${GREEN}=== Статус Docker сервисов ===${RESET}"
	@docker-compose ps
	@echo ""
	@echo "${GREEN}=== Сети ===${RESET}"
	@docker network ls | grep pg-sec-lab || echo "Сети не найдены"
	@echo ""
	@echo "${GREEN}=== Volumes ===${RESET}"
	@docker volume ls | grep course-4 || echo "Volumes не найдены"

prune: ## Очистка неиспользуемых Docker ресурсов
	@echo "${YELLOW}Очистка Docker...${RESET}"
	docker system prune -f
	@echo "${GREEN}Очистка завершена${RESET}"
