# Refatoração do Rate Limiter - Documentação

## 📋 Resumo das Mudanças

O `RateLimiter` foi refatorado para ser acoplável aos **Strategy Usecases** (IP e Token), permitindo validações customizadas além da verificação local de limite.

---

## 🔄 Mudanças Implementadas

### 1. **Interface Strategy** (`internal/ratelimiter/local_rate_limiter.go`)
```go
type Strategy interface {
	Validate(ctx context.Context, key string) *internal_error.InternalError
}
```

Todos os strategies devem implementar este método para validação customizada.

---

### 2. **RateLimitMessage Atualizado**
```go
type RateLimitMessage struct {
	Key       string
	ReplyChan chan error
	Ctx       context.Context  // ← NOVO: contexto para validações
}
```

---

### 3. **Constructor do RateLimiter**
**Antes:**
```go
NewRateLimiter(workers int, ttl time.Duration, limit int64)
```

**Depois:**
```go
NewRateLimiter(workers int, ttl time.Duration, limit int64, strategy Strategy)
```

---

### 4. **Worker - Fluxo de Execução**

Agora o worker executa em 3 fases:

```
┌─────────────────────────────────────────────────────────┐
│ 1. VERIFICAÇÃO LOCAL                                     │
│    Verifica se a chave excedeu o limite de requisições  │
│    em memória                                            │
└─────────────────────────────────────────────────────────┘
                           ↓
            ┌──────────────────────────────┐
            │ Limite excedido?             │
            └──────────────────────────────┘
            ↙ SIM                      NÃO ↘
    RETORNA ERRO              CONTINUA
                                   ↓
            ┌──────────────────────────────┐
            │ 2. VALIDAÇÃO STRATEGY        │
            │    Executa validações via    │
            │    Strategy específica       │
            └──────────────────────────────┘
                           ↓
        ┌──────────────────────────────┐
        │ Strategy retornou erro?      │
        └──────────────────────────────┘
        ↙ SIM                      NÃO ↘
    RETORNA ERRO              CONTINUA
                                   ↓
    ┌──────────────────────────────────┐
    │ 3. SUCESSO                       │
    │    Requisição autorizada         │
    └──────────────────────────────────┘
```

---

### 5. **Strategy Usecases Implementados**

#### **IPStrategyUsecase** (`internal/usecase/strategy_usecase/ip_strategy_usecase.go`)
```go
func (ts *IPStrategyUsecase) Validate(ctx context.Context, key string) *internal_error.InternalError {
	// 1. Verifica se IP foi bloqueado
	blocked := ts.Expirer.IsExpired(ctx, key)
	if blocked {
		return internal_error.NewInternalServerError("IP address is rate limited")
	}

	// 2. Salva informação da requisição
	if err := ts.RequestInfo.Save(ctx, key, request_info_entity.Active, request_info_entity.FONTE_IP); err != nil {
		return err
	}

	return nil
}
```

#### **TolkenStrategyUsecase** (`internal/usecase/strategy_usecase/tolken_strategy_usecase.go`)
```go
func (ts *TolkenStrategyUsecase) Validate(ctx context.Context, key string) *internal_error.InternalError {
	// 1. Valida se o token é válido
	isValid := ts.TolkenRepository.ValidateTolken(ctx, key)
	if !isValid {
		return internal_error.NewInternalServerError("Invalid or expired token")
	}

	// 2. Verifica se token foi bloqueado
	blocked := ts.Expirer.IsExpired(ctx, key)
	if blocked {
		return internal_error.NewInternalServerError("Token is rate limited")
	}

	// 3. Salva informação da requisição
	if err := ts.RequestInfo.Save(ctx, key, request_info_entity.Active, request_info_entity.FONTE_TOLKEN); err != nil {
		return err
	}

	return nil
}
```

---

### 6. **ExpirerInterface Expandida** (`internal/usecase/expire_usecase/expire_usecase.go`)

**Antes:**
```go
type ExpirerInterface interface {
	SetExpiration(Key string, duration time.Duration, callback func())
}
```

**Depois:**
```go
type ExpirerInterface interface {
	SetExpiration(Key string, duration time.Duration, callback func())
	IsExpired(key string) bool        // ← NOVO: Verifica se chave expirou
	ExpireKey(key string)              // ← NOVO: Marca chave como expirada
}
```

---

## 🚀 Como Usar

### Instanciar com IP Strategy:
```go
expirer := expire_usecase.NewDefaultExpirer()
requestInfoRepo := repository.NewRequestInfoRepository(db)

ipStrategy := strategy_usecase.NewIPStrategyUsecase(expirer, requestInfoRepo)
rl := ratelimiter.NewRateLimiter(4, 1*time.Minute, 100, ipStrategy)
```

### Instanciar com Token Strategy:
```go
tokenRepo := repository.NewTolkenRepository(redis)

tokenStrategy := strategy_usecase.NewTolkenStrategyUsecase(expirer, tokenRepo, requestInfoRepo)
rl := ratelimiter.NewRateLimiter(4, 1*time.Minute, 100, tokenStrategy)
```

### No Middleware:
```go
replyChan := make(chan error)
rl.InputChan <- ratelimiter.RateLimitMessage{
	Key:       ipOrToken,
	Ctx:       r.Context(),
	ReplyChan: replyChan,
}

if err := <-replyChan; err != nil {
	http.Error(w, err.Error(), http.StatusTooManyRequests)
	return
}

next.ServeHTTP(w, r)
```

---

## ✅ Benefícios

| Benefício | Descrição |
|-----------|-----------|
| **Desacoplamento** | RateLimiter não conhece detalhes de validação |
| **Extensibilidade** | Novas strategies podem ser criadas sem modificar RateLimiter |
| **Reutilização** | Strategies podem ser compartilhadas entre componentes |
| **Testabilidade** | Strategies podem ser mockadas facilmente |
| **Separação de Concerns** | Lógica de rate limit separada de validações específicas |

---

## 📁 Arquivos Modificados

- ✅ `internal/ratelimiter/local_rate_limiter.go` - Refatorado com Strategy pattern
- ✅ `internal/usecase/strategy_usecase/ip_strategy_usecase.go` - Implementa validação por IP
- ✅ `internal/usecase/strategy_usecase/tolken_strategy_usecase.go` - Implementa validação por Token
- ✅ `internal/usecase/expire_usecase/expire_usecase.go` - Expandida com métodos de expiração
- ✨ `internal/ratelimiter/example_usage.go` - Exemplos de uso
