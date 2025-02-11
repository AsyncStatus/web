package stripe

import (
	"api/handler"
	"api/internal/http/aserr"
	"api/internal/http/asres"
	"api/repository"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/guregu/null/v5"
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/webhook"
	"go.uber.org/zap"

	"github.com/stripe/stripe-go/v81/checkout/session"
	"github.com/stripe/stripe-go/v81/subscription"
)

type Handler struct {
	handler.BaseHandlerDeps
}

func NewHandler(deps handler.BaseHandlerDeps) *Handler {
	return &Handler{deps}
}

func (h *Handler) Webhook(w http.ResponseWriter, r *http.Request) (*asres.Response[string], error) {
	const MaxBodyBytes = int64(65536)
	r.Body = http.MaxBytesReader(w, r.Body, MaxBodyBytes)
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, aserr.ErrServiceUnavailable
	}

	sigHeader := r.Header.Get("Stripe-Signature")
	event, err := webhook.ConstructEvent(payload, sigHeader, h.Config().StripeSigningSecretWebhook)
	if err != nil {
		return nil, aserr.ErrBadRequest
	}

	h.Logger().Debug("Received event", zap.Any("event", event))

	switch event.Type {
	case "checkout.session.completed":
		var cs stripe.CheckoutSession
		if err := json.Unmarshal(event.Data.Raw, &cs); err != nil {
			return nil, aserr.ErrBadRequest.NewWithError(err)
		}

		updatedSession, err := session.Get(cs.ID, nil)
		if err != nil {
			return nil, aserr.ErrBadRequest.NewWithError(err)
		}

		if updatedSession.Subscription == nil {
			return nil, aserr.ErrBadRequest.NewWithError(errors.New("no subscription associated with the checkout session"))
		}

		subscriptionID := updatedSession.Subscription.ID
		name := updatedSession.CustomerDetails.Name
		customerID := updatedSession.Customer.ID
		email := updatedSession.CustomerDetails.Email

		subscriptionObj, err := subscription.Get(subscriptionID, nil)
		if err != nil {
			return nil, aserr.ErrBadRequest.NewWithError(err)
		}
		h.Logger().Debug("Fetched subscription details", zap.Any("subscription", subscriptionObj))

		if _, err := h.Repository().GetUser(r.Context(), h.Pg().Pool, &repository.GetUserBody{
			Email: &email,
		}); err != nil && !errors.Is(err, aserr.ErrUserNotFound) {
			return nil, aserr.ErrInternalServerError.NewWithError(err)
		} else if err != nil && errors.Is(err, aserr.ErrUserNotFound) {
			if _, err = h.Repository().CreateUser(r.Context(), h.Pg().Pool, &repository.CreateUserBody{
				Email:                email,
				Name:                 name,
				StripeCustomerID:     null.StringFrom(customerID),
				StripeSubscriptionID: null.StringFrom(subscriptionID),
			}); err != nil {
				return nil, aserr.ErrInternalServerError.NewWithError(err)
			}

			return asres.NewResponse("ok", http.StatusOK), nil
		}

		customerIDNullable := null.StringFrom(customerID)
		subscriptionIDNullable := null.StringFrom(subscriptionID)
		if _, err = h.Repository().UpdateUser(r.Context(), h.Pg().Pool, &repository.UpdateUserBodyFilter{
			Email: email,
		}, &repository.UpdateUserBodyData{
			Email:                &email,
			Name:                 &name,
			StripeCustomerID:     &customerIDNullable,
			StripeSubscriptionID: &subscriptionIDNullable,
		}); err != nil {
			return nil, aserr.ErrInternalServerError.NewWithError(err)
		}

	default:
		h.Logger().Error("Unhandled event type", zap.String("type", string(event.Type)))
	}

	return asres.NewResponse("ok", http.StatusOK), nil
}

func (h *Handler) Router() chi.Router {
	r := chi.NewRouter()

	r.Post("/webhook", asres.ToHandlerFunc(h.Webhook))

	return r
}
