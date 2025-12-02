package main

import (
	"fmt"
	"os"

	"github.com/Higor-ViniciusDev/posgo_raterlimite/configuration/logger"
	"github.com/Higor-ViniciusDev/posgo_raterlimite/internal/infra/api/web"
	"github.com/joho/godotenv"
)

func main() {
	defer logger.GetLogger().Sync()

	if err := godotenv.Load("cmd/raterlimite/.env"); err != nil {
		logger.Error("Erro ao carregar variaveis de ambiente", err)
		return
	}

	// _ := database.NewConnectionRedis()

	webServerPort := os.Getenv("WEB_SERVER_PORT")
	webServer := web.NovoWebServer(fmt.Sprintf(":%v", webServerPort))

	webServer.IniciarWebServer()
	// webServer.RegistrarRota("/tolken", tolkenController.CreateTolken, "POST")
}
