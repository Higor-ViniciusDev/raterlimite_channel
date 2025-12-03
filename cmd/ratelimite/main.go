package main

import (
	"fmt"
	"os"

	"github.com/Higor-ViniciusDev/posgo_raterlimite/configuration/database"
	"github.com/Higor-ViniciusDev/posgo_raterlimite/configuration/logger"
	"github.com/Higor-ViniciusDev/posgo_raterlimite/internal/infra/api/controller"
	"github.com/Higor-ViniciusDev/posgo_raterlimite/internal/infra/api/web"
	"github.com/Higor-ViniciusDev/posgo_raterlimite/internal/infra/repository"
	"github.com/Higor-ViniciusDev/posgo_raterlimite/internal/middleware"
	"github.com/Higor-ViniciusDev/posgo_raterlimite/internal/ratelimiter"
	"github.com/Higor-ViniciusDev/posgo_raterlimite/internal/usecase/expire_usecase"
	"github.com/Higor-ViniciusDev/posgo_raterlimite/internal/usecase/policy_usecase"
	"github.com/Higor-ViniciusDev/posgo_raterlimite/internal/usecase/strategy_usecase"
	"github.com/Higor-ViniciusDev/posgo_raterlimite/internal/usecase/tolken_usecase"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

func main() {
	defer logger.GetLogger().Sync()

	if err := godotenv.Load("cmd/ratelimite/.env"); err != nil {
		logger.Error("Erro ao carregar variaveis de ambiente", err)
		return
	}

	redis := database.NewConnectionRedis()

	tolkenController, policyUsecase := initDependeces(redis)

	//rater limite 	middleware
	raterLimite := ratelimiter.NewRateLimiter(5, 1000)

	webServerPort := os.Getenv("WEB_SERVER_PORT")
	webServer := web.NovoWebServer(fmt.Sprintf(":%v", webServerPort))

	webServer.RegistrarRota("/tolken", tolkenController.CreateTolken, "POST")
	webServer.RegistrarRota("/", nil, "GET", middleware.RateLimiterMiddleware(&policyUsecase, raterLimite))
	webServer.IniciarWebServer()
}

func initDependeces(redisCli *redis.Client) (controller.TolkenController, policy_usecase.PolicyUsecase) {
	var tolkenController controller.TolkenController

	expirerUsecase := expire_usecase.NewDefaultExpirer()

	//Tolken dependeces
	tolkeRepository := repository.NewTolkenDB(redisCli)
	tolkenUsecase := tolken_usecase.NewTolkenUsecase(tolkeRepository, expirerUsecase)
	tolkenController = *controller.NewTolkenController(tolkenUsecase)

	ipStrategy := strategy_usecase.NewIPStrategyUsecase()
	tokenStrategy := strategy_usecase.NewTokenStrategyUsecase(tolkeRepository)

	policyUsecase := *policy_usecase.NewPolicyUsecase(ipStrategy, tokenStrategy)

	return tolkenController, policyUsecase
}
