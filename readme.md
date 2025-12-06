# 🚀 Rate Limiter - Sliding Window com Penalidade
---

## 📋 O que faz

Sistema de limitação de requisições que:

- **Controla taxa de requisições** por IP ou Token (API-KEY)
- **Sliding Window**: Contador reseta a cada 1 segundo (TTL da key)
- **Sistema de Penalidade**: Ao exceder o limite, bloqueia TODAS as requisições por X segundos
- **Priorização**: Token tem prioridade sobre IP

### 🔄 Como Funciona

```
Timeline Exemplo (Token):
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
t=0.0s   → Req #1-10: ✅ 200 OK (dentro do limite)
t=0.0s   → Req #11: 🚫 429 (excedeu) + ATIVA PENALIDADE (10s)
t=0.1s   → Req #12-50: 🚫 429 (penalidade ativa)
t=1.0s   → TTL expira (key resetaria), MAS penalidade continua
t=2-9s   → TODAS req: 🚫 429 (penalidade ativa)
t=10s+   → Penalidade expira → Volta ao normal
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

---

## 🛠️ Como Executar

### 1️⃣ Pré-requisitos

- **Docker** (para Redis)
- **Go 1.24+**

### 2️⃣ Subir o Redis

```bash
# Baixar imagem e iniciar container
docker-compose up -d

```

### 4️⃣ Iniciar a Aplicação

```bash
go run cmd/ratelimite/main.go
```

A aplicação estará rodando em `http://localhost:8080`

---

## 🧪 Testes de Carga

### Instalar Vegeta (Ferramenta de Load Testing)

**Windows (PowerShell como Admin):**
```powershell
# Usando Chocolatey
choco install vegeta

# Ou baixar direto do GitHub
# https://github.com/tsenart/vegeta/releases
# Adicionar vegeta.exe ao PATH
```
---
Só funciona no windows, e se não tiver o make instalado apenas rode 'vegeta attack -targets=targets-token.txt -rate=15 -duration=10s -output=results-token.bin'

### 📊 Executar Testes

#### Teste 1: Rate Limit por IP
### Se tiver o vegeta instalado, rode:

```bash
make test-ip
```

**Configuração:**
- Taxa: 10 req/s
- Duração: 2 segundos
- Total: 20 requisições
- **Esperado**: ~10 sucesso (200), ~10 bloqueadas (429)

---

#### Teste 2: Rate Limit por Token
```bash
make test-token
```

**Configuração:**
- Limite: 10 req/s
- TTL key: 1s (sliding window)
- Penalidade: 10s (ao exceder)
- Taxa envio: 15 req/s
- Duração: 10 segundos
- Total: 150 requisições

**Esperado:**
```
1º segundo: 10 OK + 5 bloqueadas
11ª requisição ATIVA penalidade de 10s
Próximos 9s: TUDO bloqueado

Resultado: ~10-15 sucesso (200), ~60-65 bloqueadas (429), ~74 inválidas (400)

Status Codes:
  200: ~10-15   (6-10%)   ✅ Requisições permitidas
  429: ~60-65   (40-43%)  🚫 Bloqueadas pelo rate limiter
  400: ~74      (49%)     ⚠️ Token expirou (TOLKEN_EXPIRATION=8s)
```

> **Nota**: Os ~74 erros `400 Bad Request` são esperados porque o token expira após 8 segundos (configurado no `.env`), e o teste dura 10 segundos.


---

## 📂 Estrutura do Projeto

```
.
├── cmd/ratelimite/          # Aplicação principal
│   ├── main.go
│   └── .env                 # Configurações
├── internal/
│   ├── entity/              # Entidades de domínio
│   ├── infra/               # Infraestrutura (controllers, repositories)
│   ├── middleware/          # Rate Limiter middleware
│   ├── ratelimiter/         # Lógica do rate limiter
│   └── usecase/             # Casos de uso (strategies, policies)
├── configuration/           # Configurações (logger, database)
├── Makefile                 # Comandos de teste
└── docker-compose.yaml      # Redis
```

---

## ⚙️ Configuração (.env)

```env
# Redis
REDIS_URL=localhost
REDIS_PORT=6379

# Token
TOLKEN_EXPIRATION=8          # Token expira em 8 segundos
JWT_SECRET=secret

# Rate Limits
REQUEST_PER_SECOND_IP=5      # Limite IP: 5 req/s
REQUEST_PER_SECOND_TOLKEN=10 # Limite Token: 10 req/s

# Penalidades
TIME_UNLOCKED_NEW_REQUEST_IP=1       # 1 segundo de bloqueio
TIME_UNLOCKED_NEW_REQUEST_TOLKEN=10  # 10 segundos de bloqueio

# TTL das Keys (Sliding Window)
TLL_KEY_IP=1                 # Key IP expira em 1s
TLL_KEY_TOLKEN=1             # Key Token expira em 1s

# Workers
WORKER_POOL_SIZE=5           # 5 workers para processar
SIZE_BUFFER_CHANNEL=1000     # Buffer do canal
```

---

## 🔑 Endpoints

### 1. Criar Token
```bash
POST http://localhost:8080/tolken
```

**Response:**
```json
{
  "tolken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

### 2. Endpoint com Rate Limiter
```bash
# Por IP
GET http://localhost:8080/

# Por Token
GET http://localhost:8080/
API-KEY: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**Respostas:**
- `200 OK` - Requisição permitida
- `429 Too Many Requests` - Rate limit excedido
- `400 Bad Request` - Token inválido/expirado
---