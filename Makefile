# Makefile para testes de carga com Vegeta

# Configurações padrão
RATE ?= 500
DURATION ?= 10s
TARGETS ?= targets.txt
OUTPUT ?= results.bin
REPORT_TXT ?= report.txt

.PHONY: help test-ip test-token test-ip-stress test-token-stress analyze clean

help:
	@echo "Testes disponíveis:"
	@echo "  make test-ip          - Testa rate limit por IP (10 req/s por 2s)"
	@echo "  make test-token       - Testa rate limit por Token (15 req/s por 10s)"
	@echo "  make test-ip-stress   - Teste de stress IP (50 req/s por 5s)"
	@echo "  make test-token-stress- Teste de stress Token (30 req/s por 10s)"
	@echo "  make analyze          - Analisa resultados dos testes"
	@echo "  make clean            - Remove arquivos de resultado"

# Teste IP: 20 requisições em 2 segundos (10 req/s)
# Esperado: 10 sucesso, 10 bloqueadas (limite: 5 req/s × 2s = 10 total)
test-ip:
	@echo "🧪 Testando Rate Limit por IP..."
	@echo "Configuração: 10 req/s por 2 segundos"
	@echo "Esperado: ~10 sucesso (200), ~10 bloqueadas (429)"
	@vegeta attack -targets=targets-ip.txt -rate=10 -duration=3s -output=results-ip.bin
	@echo "\n📊 Relatório:"
	@vegeta report results-ip.bin
	@vegeta report results-ip.bin -type=json | jq '{success: .status_codes."200", blocked: .status_codes."429", total: .requests}'

# Teste Token: 150 requisições em 10 segundos (15 req/s)
# Esperado: 100 sucesso, 50 bloqueadas (limite: 10 req/s × 10s = 100 total)
test-token:
	@echo "🧪 Testando Rate Limit por Token..."
	@echo "Configuração: 15 req/s por 10 segundos"
	@echo "Esperado: ~100 sucesso (200), ~50 bloqueadas (429)"
	@echo "⚠️  Primeiro, crie um token:"
	@curl -s -X POST http://localhost:8080/tolken | jq -r '.tolken' > .token
	@echo "Token criado: $$(cat .token)"
	@vegeta attack -targets=targets-token.txt -rate=15 -duration=10s -output=results-token.bin
	@echo "\n📊 Relatório:"
	@vegeta report results-token.bin
	@vegeta report results-token.bin -type=json | jq '{success: .status_codes."200", blocked: .status_codes."429", total: .requests}'

# Teste de Stress IP: 250 requisições em 5 segundos (50 req/s)
test-ip-stress:
	@echo "🔥 Teste de STRESS por IP..."
	@echo "Configuração: 50 req/s por 5 segundos"
	@vegeta attack -targets=targets-ip.txt -rate=50 -duration=5s -output=results-ip-stress.bin
	@vegeta report results-ip-stress.bin
	@vegeta plot results-ip-stress.bin > plot-ip-stress.html
	@echo "Gráfico salvo em: plot-ip-stress.html"

# Teste de Stress Token: 300 requisições em 10 segundos (30 req/s)
test-token-stress:
	@echo "🔥 Teste de STRESS por Token..."
	@curl -s -X POST http://localhost:8080/tolken | jq -r '.tolken' > .token
	@vegeta attack -targets=targets-token.txt -rate=30 -duration=10s -output=results-token-stress.bin
	@vegeta report results-token-stress.bin
	@vegeta plot results-token-stress.bin > plot-token-stress.html
	@echo "Gráfico salvo em: plot-token-stress.html"

# Analisa todos os resultados
analyze:
	@echo "📊 Análise Completa dos Testes\n"
	@if [ -f results-ip.bin ]; then \
		echo "=== Teste IP Normal ==="; \
		vegeta report results-ip.bin -type=json | jq '{success: .status_codes."200", blocked: .status_codes."429", success_rate: .success_ratio, latency_p99: .latencies.p99}'; \
	fi
	@if [ -f results-token.bin ]; then \
		echo "\n=== Teste Token Normal ==="; \
		vegeta report results-token.bin -type=json | jq '{success: .status_codes."200", blocked: .status_codes."429", success_rate: .success_ratio, latency_p99: .latencies.p99}'; \
	fi

# Limpa arquivos de resultado
clean:
	rm -f results*.bin plot*.html .token report.json

# Teste completo: IP + Token
test-all: test-ip test-token analyze
	@echo "✅ Todos os testes concluídos!"