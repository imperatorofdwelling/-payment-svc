package http

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/imperatorofdwelling/payment-svc/internal/config"
	v1 "github.com/imperatorofdwelling/payment-svc/internal/handler/http/api/v1"
	"github.com/imperatorofdwelling/payment-svc/internal/service"
	"github.com/imperatorofdwelling/payment-svc/internal/storage"
	"github.com/imperatorofdwelling/payment-svc/internal/storage/postgres"
	"github.com/imperatorofdwelling/payment-svc/internal/storage/redis"
	"github.com/rvinnie/yookassa-sdk-go/yookassa"
	"go.uber.org/zap"
	"strconv"
	"time"
)

type Router struct {
	Handler *chi.Mux
}

func NewRouter(s *storage.Storage, log *zap.SugaredLogger, cfg *config.Config) *Router {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.URLFormat)
	r.Use(middleware.Timeout(60 * time.Second))

	r.Route("/api/v1", func(r chi.Router) {
		yooclient := yookassa.NewClient(strconv.Itoa(cfg.ShopID), cfg.SecretKey)

		yookassaHdl := yookassa.NewPaymentHandler(yooclient)

		paymentRepo := postgres.NewPaymentRepo(s.Psql, log.Named("payment_repo"))
		paymentSvc := service.NewPaymentSvc(paymentRepo, log.Named("payment_service"))
		v1.NewPaymentsHandler(r, paymentSvc, log.Named("payment_handler"))

		_ = postgres.NewLogsRepo(s.Psql)

		_ = postgres.NewCardRepo(s.Psql)

		_ = redis.NewTransactionRepo(s.Redis)
	})

	return &Router{
		Handler: r,
	}
}
