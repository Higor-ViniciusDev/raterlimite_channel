# Makefile para executar testes de carga com Vegeta
# Uso padrão: 10 segundos, 500 requisições por segundo, targets em `targets.txt`

RATE ?= 500
DURATION ?= 10s
TARGETS ?= targets.txt
OUTPUT ?= results.bin
REPORT_TXT ?= report.txt

.PHONY: help run report plot attack live

help:
	@echo "Usage: make attack [RATE=500] [DURATION=10s] [TARGETS=targets.txt]"
	@echo "Targets file must be a Vegeta-formatted targets file (one request per line)."

# Executa o ataque e grava os resultados binários em $(OUTPUT)
run:
	vegeta attack -targets=$(TARGETS) -rate=$(RATE) -duration=$(DURATION) -output=$(OUTPUT)

# Gera um relatório texto simples em $(REPORT_TXT) e JSON em report.json
report: run
	vegeta report -inputs=$(OUTPUT) > $(REPORT_TXT)
	@echo "Report written to $(REPORT_TXT)"
	vegeta report -inputs=$(OUTPUT) -type=json > report.json || true
	@echo "JSON report written to report.json"

# Gera um gráfico HTML (plot) a partir do arquivo de resultados
plot: run
	vegeta plot -inputs=$(OUTPUT) -output=plot.html
	@echo "Plot saved to plot.html"

# Ataque padrão + relatório (alias)
attack: report

# Modo 'live' para ver relatório em stdout enquanto grava resultados em arquivo
live:
	vegeta attack -targets=$(TARGETS) -rate=$(RATE) -duration=$(DURATION) | tee $(OUTPUT) | vegeta report
