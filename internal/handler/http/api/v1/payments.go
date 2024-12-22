package v1

import (
	"github.com/go-chi/chi/v5"
	"github.com/imperatorofdwelling/payment-svc/internal/domain/model"
	"github.com/imperatorofdwelling/payment-svc/internal/service"
	"github.com/imperatorofdwelling/payment-svc/pkg/json"
	"github.com/imperatorofdwelling/payment-svc/pkg/yookassa"
	"go.uber.org/zap"
	"net/http"
)

type paymentsHandler struct {
	svc         service.IPaymentSvc
	log         *zap.SugaredLogger
	yookassaHdl *yookassa.PaymentsHandler
}

func NewPaymentsHandler(r chi.Router, svc service.IPaymentSvc, yookassaHdl *yookassa.PaymentsHandler, log *zap.SugaredLogger) {
	handler := &paymentsHandler{
		svc:         svc,
		log:         log,
		yookassaHdl: yookassaHdl,
	}

	r.Route("/payments", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Post("/", handler.createPayment)
		})
	})
}

func (h *paymentsHandler) createPayment(w http.ResponseWriter, r *http.Request) {
	const op = "handler.v1.payments.createPayment"
	var payment model.Payment

	idempotenceKey := r.Header.Get("Idempotence-Key")
	if idempotenceKey == "" {
		h.log.Errorf("%s: %v", op, ErrGettingIdempotenceKey)
		json.WriteError(w, http.StatusBadRequest, ErrGettingIdempotenceKey.Error(), json.GettingHeaderDataError)
		return
	}

	err := json.Read(r, &payment)
	if err != nil {
		json.WriteError(w, http.StatusBadRequest, err.Error(), json.DecodeBodyError)
		return
	}

	newPayment, err := h.yookassaHdl.CreatePayment(&payment, idempotenceKey)
	if err != nil {
		h.log.Errorf("%s: %v", op, err.Error())
		json.WriteError(w, http.StatusBadRequest, err.Error(), json.ExternalApiError)
		return
	}

	if newPayment.Status == "" {
		h.log.Error("invalid response from API", zap.String("op", op), zap.String("description", newPayment.Description))
		json.WriteError(w, http.StatusBadRequest, newPayment.Description, json.ExternalApiError)
		return
	}

	err = h.svc.CreatePayment(r.Context(), newPayment)
	if err != nil {
		h.log.Errorf("%s: %v", op, zap.Error(err))
		json.WriteError(w, http.StatusInternalServerError, err.Error(), json.InternalApiError)
		return
	}

	json.Write(w, http.StatusOK, newPayment)
}

func (h *paymentsHandler) ChangeStatus(w http.ResponseWriter, r *http.Request) {
	const op = "handler.v1.payments.ChangeStatus"
}

//if err := v10.Validate.Struct(tt); err != nil {
//	validationErr := err.(validator.ValidationErrors)
//	json.WriteError(w, http.StatusBadRequest, validationErr.Error(), json.ValidationError)
//	return
//}

//json.Write(w, http.StatusOK, tt)
