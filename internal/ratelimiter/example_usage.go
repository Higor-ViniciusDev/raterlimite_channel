package ratelimiter

/*
EXEMPLO DE USO - RATE LIMITER COM STRATEGIES

1. Para usar com IP Strategy:
   -------
   ipStrategy := strategy_usecase.NewIPStrategyUsecase(expirer, requestInfoRepo)
   rl := NewRateLimiter(workers, ttl, limit, ipStrategy)

2. Para usar com Token Strategy:
   -------
   tokenStrategy := strategy_usecase.NewTolkenStrategyUsecase(expirer, tolkenRepo, requestInfoRepo)
   rl := NewRateLimiter(workers, ttl, limit, tokenStrategy)

3. No middleware, enviar mensagem com contexto:
   -------
   replyChan := make(chan error)
   rl.InputChan <- RateLimitMessage{
       Key:       ipOrToken,
       Ctx:       r.Context(),
       ReplyChan: replyChan,
   }

   if err := <-replyChan; err != nil {
       http.Error(w, err.Error(), http.StatusTooManyRequests)
       return
   }

FLUXO:
1. RateLimiter recebe mensagem com chave (IP ou Token)
2. Verifica o counter local (em memória)
3. Se não excedeu o limite, executa a estratégia de validação
4. A estratégia pode fazer validações adicionais (Redis, BD, etc)
5. Retorna sucesso ou erro pelo canal
*/
