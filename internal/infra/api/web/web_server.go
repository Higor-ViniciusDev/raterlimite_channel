package web

import (
	"fmt"
	"net/http"

	"github.com/Higor-ViniciusDev/posgo_raterlimite/configuration/logger"
	"github.com/go-chi/chi/v5"
)

type HandlerInfo struct {
	Metodo      string
	Handler     http.HandlerFunc
	Middlewares []func(http.Handler) http.Handler
}

type WebServer struct {
	Porta    string
	Handlers map[string][]HandlerInfo // path -> HandlerInfo (method + handler)
	Rotas    chi.Router
}

func NovoWebServer(porta string) *WebServer {
	return &WebServer{
		Porta:    porta,
		Handlers: make(map[string][]HandlerInfo),
		Rotas:    chi.NewRouter(),
	}
}

func (w WebServer) RegistrarRota(caminho string, handlerFunc http.HandlerFunc, metodo string, middlewares ...func(http.Handler) http.Handler,
) {
	w.Handlers[caminho] = append(w.Handlers[caminho], HandlerInfo{
		Metodo:      metodo,
		Handler:     handlerFunc,
		Middlewares: middlewares,
	})
}

func (w WebServer) IniciarWebServer() {
	for rota, handlers := range w.Handlers {
		for _, infoHandle := range handlers {
			w.Rotas.Method(infoHandle.Metodo, rota, infoHandle.Handler)
			logger.Info(fmt.Sprintf("Registrando na rota %v com o metodo %v", rota, infoHandle.Metodo))
		}
	}

	err := http.ListenAndServe(w.Porta, w.Rotas)

	if err != nil {
		logger.Error("Error ao iniciar webserver", err)
	}
}
