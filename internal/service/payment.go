package service

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/imperatorofdwelling/payment-svc/internal/domain/model"
	"github.com/imperatorofdwelling/payment-svc/internal/storage/postgres"
	"go.uber.org/zap"
)

type IPaymentSvc interface {
	CreatePayment(context.Context, *model.Payment) error
}

type PaymentSvc struct {
	repo    postgres.IPaymentRepo
	log     *zap.SugaredLogger
	logsSvc ILogsSvc
}

func NewPaymentSvc(repo postgres.IPaymentRepo, logsSvc ILogsSvc, log *zap.SugaredLogger) *PaymentSvc {
	return &PaymentSvc{
		repo,
		log,
		logsSvc,
	}
}

func (s *PaymentSvc) CreatePayment(ctx context.Context, payment *model.Payment) error {
	const op = "service.payments.CreatePayment"

	idUUID, err := uuid.Parse(payment.ID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	newLog := &model.Log{
		TransactionID:   idUUID,
		MethodType:      payment.PaymentMethodData.Type,
		TransactionType: model.PaymentType,
		Status:          payment.Status,
		Value:           payment.Amount.Value,
		Currency:        payment.Amount.Currency,
	}

	err = s.logsSvc.InsertLog(ctx, newLog)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
