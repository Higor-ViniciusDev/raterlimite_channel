package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"

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

	wokers := os.Getenv("WORKERS_RATE_LIMITER")
	bufferSize := os.Getenv("BUFFER_SIZE_RATE_LIMITER")

	wokerNumber, _ := strconv.Atoi(wokers)
	bufSizeNumber, _ := strconv.Atoi(bufferSize)

	// RateLimiter
	raterLimite := ratelimiter.NewRateLimiter(wokerNumber, bufSizeNumber)

	webServerPort := os.Getenv("WEB_SERVER_PORTA")
	webServer := web.NovoWebServer(fmt.Sprintf(":%v", webServerPort))

	webServer.RegistrarRota("/tolken", tolkenController.CreateTolken, "POST")
	webServer.RegistrarRota("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}, "GET", middleware.RateLimiterMiddleware(&policyUsecase, raterLimite))
	webServer.IniciarWebServer()
}

func initDependeces(redisCli *redis.Client) (controller.TolkenController, policy_usecase.PolicyUsecase) {
	var tolkenController controller.TolkenController

	expirerUsecase := expire_usecase.NewDefaultExpirer()

	requestInfoRepository := repository.NewRequestInfoRepository(redisCli)
	//Tolken dependeces
	tolkeRepository := repository.NewTolkenDB(redisCli)
	tolkenUsecase := tolken_usecase.NewTolkenUsecase(tolkeRepository, expirerUsecase)
	tolkenController = *controller.NewTolkenController(tolkenUsecase)

	ipStrategy := strategy_usecase.NewIPStrategyUsecase(requestInfoRepository)
	tokenStrategy := strategy_usecase.NewTokenStrategyUsecase(tolkeRepository, requestInfoRepository)

	policyUsecase := *policy_usecase.NewPolicyUsecase(ipStrategy, tokenStrategy)

	return tolkenController, policyUsecase
}
