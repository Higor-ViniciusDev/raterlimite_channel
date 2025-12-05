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
	@echo "- Testando Rate Limit por IP..."
	@echo "Configuração: 10 req/s por 2 segundos"
	@echo "Esperado: ~10 sucesso (200), ~10 bloqueadas (429)"
	@vegeta attack -targets=targets-ip.txt -rate=10 -duration=2s -output=results-ip.bin
	@echo "\n- Relatório:"
	@vegeta report results-ip.bin

# Teste 4: Token com taxa alta
# COMPORTAMENTO CORRETO:
# - Limite: 10 req/s
# - TTL da key: 1s (contador reseta a cada segundo)
# - Taxa de envio: 15 req/s por 10 segundos = 150 requisições total
# - A cada segundo: primeiras 10 req passam, próximas 5 bloqueadas
# - 6ª requisição de cada segundo ATIVA penalidade de 10s
# - Durante penalidade: TUDO bloqueado
# 
# Esperado (SEM considerar penalidade acumulativa):
# - Por segundo: 10 sucesso + 5 bloqueadas
# - Em 10 segundos: ~100 sucesso, ~50 bloqueadas
# 
# Esperado REAL (COM penalidade):
# - 1º segundo: 10 sucesso + 5 bloqueadas (ativa penalidade 10s)
# - 2º-10º segundo: TUDO bloqueado (penalidade ativa)
# - Resultado: ~10-15 sucesso, ~135-140 bloqueadas
test-token:
	@echo "-Testando Rate Limit por Token..."
	@echo "   - Limite: 10 req/s"
	@echo "   - TTL key: 1s (sliding window)"
	@echo "   - Penalidade: 10s (ao exceder)"
	@echo "   - Taxa envio: 15 req/s"
	@echo "   - Duracao: 10 segundos"
	@echo "   - Total: 150 requisicoes"
	@echo ""
	@echo "- Criando token..."
	@powershell -Command "$$response = Invoke-RestMethod -Method Post 'http://localhost:8080/tolken'; $$response.tolken" > .token
	@powershell -Command "$$token = Get-Content .token -Raw; $$token = $$token.Trim(); Write-Host \"   Token criado: $$token\""
	@echo ""
	@echo "- Criando arquivo de targets com token..."
	@powershell -Command "$$token = (Get-Content .token -Raw).Trim(); $$content = \"GET http://localhost:8080/`nAPI-KEY: $$token`n\"; Set-Content -Path targets-token.txt -Value $$content -NoNewline"
	@echo "   Arquivo targets-token.txt criado"
	@echo ""
	@echo "- Executando teste de carga..."
	@echo "   COMPORTAMENTO ESPERADO:"
	@echo "   1. Primeiro segundo: 10 OK + 5 bloqueadas"
	@echo "   2. 6a requisicao ATIVA penalidade de 10s"
	@echo "   3. Proximos 9s: TUDO bloqueado"
	@echo "   Resultado: ~10-15 sucesso, ~135-140 bloqueadas"
	@echo ""
	@vegeta attack -targets=targets-token.txt -rate=15 -duration=10s -output=results-token.bin
	@echo ""
	@echo "- Relatorio:"
	@vegeta report results-token.bin
	@echo ""
	
# Limpa arquivos de resultado
clean:
	rm -f results*.bin plot*.html .token report.json