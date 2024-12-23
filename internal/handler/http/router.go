package http

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/imperatorofdwelling/payment-svc/internal/config"
	v1 "github.com/imperatorofdwelling/payment-svc/internal/handler/http/api/v1"
	"github.com/imperatorofdwelling/payment-svc/internal/handler/http/htmx"
	"github.com/imperatorofdwelling/payment-svc/internal/service"
	"github.com/imperatorofdwelling/payment-svc/internal/storage"
	"github.com/imperatorofdwelling/payment-svc/internal/storage/postgres"
	"github.com/imperatorofdwelling/payment-svc/internal/storage/redis"
	"github.com/imperatorofdwelling/payment-svc/pkg/yookassa"
	"go.uber.org/zap"
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

		yooClient := yookassa.NewYookassaClient(cfg.PayApi)
		yookassaHdl := yookassa.NewPaymentsHandler(yooClient, log.Named("yookassa_handler"))

		htmx.NewHTMXHandler(r, log.Named("htmx_handler"))

		cardsRepo := postgres.NewCardsRepo(s.Psql, log.Named("cards_repo"))
		cardsSvc := service.NewCardsService(cardsRepo, log.Named("cards_service"))

		logsRepo := postgres.NewLogsRepo(s.Psql, log.Named("logs_repo"))
		logsSvc := service.NewLogsService(logsRepo, log.Named("logs_service"))
		v1.NewLogsHandler(r, logsSvc, log.Named("logs_handler"))

		paymentRepo := postgres.NewPaymentRepo(s.Psql, log.Named("payment_repo"))
		paymentSvc := service.NewPaymentSvc(paymentRepo, logsSvc, log.Named("payment_service"))
		v1.NewPaymentsHandler(r, paymentSvc, yookassaHdl, log.Named("payment_handler"))

		v1.NewPayoutsHandler(r, cardsSvc, log.Named("payout_handler"))

		_ = redis.NewTransactionRepo(s.Redis)
	})

	return &Router{
		Handler: r,
	}
}
