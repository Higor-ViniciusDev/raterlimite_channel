package expire_usecase

import (
	"testing"
	"time"
)

func TestDefaultExpirer_CallbackFires(t *testing.T) {
	e := NewDefaultExpirer()

	ch := make(chan struct{}, 1)

	// callback dispara um sinal no channel
	e.SetExpiration("key1", 20*time.Millisecond, func() {
		ch <- struct{}{}
	})

	select {
	case <-ch:
		//Se preencher valor, tudo ok
	case <-time.After(200 * time.Millisecond):
		t.Fatal("callback não foi chamado dentro do tempo esperado")
	}
}
