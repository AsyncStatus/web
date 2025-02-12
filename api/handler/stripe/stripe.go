package stripe

import (
	"api/handler"
	emailtemplate "api/internal/email/templates"
	"api/internal/http/aserr"
	"api/internal/http/asres"
	"api/internal/utils/authutils"
	"api/repository"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/guregu/null/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/resend/resend-go/v2"
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/customer"
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

		session, err := session.Get(cs.ID, nil)
		if err != nil {
			return nil, aserr.ErrBadRequest.NewWithError(err)
		}

		if session.Subscription == nil {
			return nil, aserr.ErrBadRequest.NewWithError(errors.New("no subscription associated with the checkout session"))
		}

		subscriptionID := session.Subscription.ID
		name := session.CustomerDetails.Name
		customerID := session.Customer.ID
		email := session.CustomerDetails.Email

		subscriptionObj, err := subscription.Get(subscriptionID, nil)
		if err != nil {
			return nil, aserr.ErrBadRequest.NewWithError(err)
		}
		h.Logger().Debug("Fetched subscription details", zap.Any("subscription", subscriptionObj))

		now := pgtype.Timestamptz{Time: time.Now(), Valid: true}
		user, err := h.Repository().CreateUser(r.Context(), h.Pg().Pool, &repository.CreateUserBody{
			Email:                email,
			Name:                 name,
			StripeCustomerID:     null.StringFrom(customerID),
			StripeSubscriptionID: null.StringFrom(subscriptionID),
			VerifiedAt:           now,
			ApprovedAt:           now,
		})
		if err != nil && !errors.Is(err, aserr.ErrUserAlreadyExists) {
			return nil, aserr.ErrInternalServerError.NewWithError(err)
		}
		if errors.Is(err, aserr.ErrUserAlreadyExists) {
			customerIDNullable := null.StringFrom(customerID)
			subscriptionIDNullable := null.StringFrom(subscriptionID)
			if _, err = h.Repository().UpdateUser(r.Context(), h.Pg().Pool, &repository.UpdateUserBodyFilter{
				Email: email,
			}, &repository.UpdateUserBodyData{
				Email:                &email,
				Name:                 &name,
				StripeCustomerID:     &customerIDNullable,
				StripeSubscriptionID: &subscriptionIDNullable,
				VerifiedAt:           &now,
				ApprovedAt:           &now,
			}); err != nil {
				return nil, aserr.ErrInternalServerError.NewWithError(err)
			}
		}

		params := &stripe.CustomerParams{}
		params.AddMetadata("user_id", user.ID.String())
		if _, err := customer.Update(customerID, params); err != nil {
			return nil, aserr.ErrInternalServerError.NewWithError(err)
		}

		now = pgtype.Timestamptz{Time: time.Now(), Valid: true}
		code := authutils.GenerateCode(h.Config().SecretAuthCodeValue, string(repository.AuthCodeTypeCreateAccount), user.ID.String())
		valueHash := authutils.GenerateCodeHash(h.Config().SecretAuthCodeValueHash, code)
		if _, err := h.Repository().CreateAuthCode(r.Context(), h.Pg().Pool, &repository.CreateAuthCodeInput{
			UserID:    user.ID,
			Type:      repository.AuthCodeTypeCreateAccount,
			Value:     code,
			ValueHash: valueHash,
			CreatedAt: now,
			ExpiresAt: pgtype.Timestamptz{Time: now.Time.Add(24 * time.Hour), Valid: true},
		}); err != nil {
			return nil, aserr.ErrInternalServerError.NewWithError(err)
		}

		emailData := emailtemplate.NewCreateAccountTemplateData(h.Config(), strings.Split(name, " ")[0], email, valueHash)
		if err := h.EmailClient().Send(&resend.SendEmailRequest{
			From:    "AsyncStatus <accounts@a.asyncstatus.com>",
			ReplyTo: "support@asyncstatus.com",
			To:      []string{email},
			Subject: "Create account",
			Html:    emailtemplate.CreateAccountTemplate.Render(emailData),
			Headers: map[string]string{"X-Entity-Ref-ID": valueHash},
			Text:    fmt.Sprintf("Create account: %s", emailData.CreateAccountURL),
		}); err != nil {
			return nil, aserr.ErrInternalServerError.NewWithError(err)
		}

	}

	return asres.NewResponse("ok", http.StatusOK), nil
}

func (h *Handler) PaymentLinkSuccess(w http.ResponseWriter, r *http.Request) (*asres.Response[string], error) {
	getRedirectURL := func(err string) string {
		return fmt.Sprintf("%s://%s/confirmation-error?error=%s", h.Config().AppProtocol, h.Config().AppHost, url.QueryEscape(err))
	}

	csid := r.URL.Query().Get("csid")
	if csid == "" {
		return asres.NewRedirect(getRedirectURL("Missing checkout session ID"), http.StatusTemporaryRedirect), nil
	}

	session, err := session.Get(csid, nil)
	if err != nil {
		return asres.NewRedirect(getRedirectURL("No checkout session found"), http.StatusTemporaryRedirect), nil
	}

	if session.Status != stripe.CheckoutSessionStatusComplete {
		return asres.NewRedirect(getRedirectURL("No subscription associated with the checkout session."), http.StatusTemporaryRedirect), nil
	}

	customer, err := customer.Get(session.Customer.ID, nil)
	if err != nil {
		return asres.NewRedirect(getRedirectURL("No customer associated with the checkout session."), http.StatusTemporaryRedirect), nil
	}

	userID, err := uuid.Parse(customer.Metadata["user_id"])
	if err != nil {
		return asres.NewRedirect(getRedirectURL("No user associated with the checkout session."), http.StatusTemporaryRedirect), nil
	}

	authCode, err := h.Repository().GetAuthCode(
		r.Context(), h.Pg().Pool,
		repository.WithGetAuthCodeInputType(repository.AuthCodeTypeCreateAccount),
		repository.WithGetAuthCodeInputUserID(userID),
	)
	if err != nil {
		return asres.NewRedirect(getRedirectURL("No auth code associated with the checkout session."), http.StatusTemporaryRedirect), nil
	}

	user, err := h.Repository().GetUser(r.Context(), h.Pg().Pool, &repository.GetUserBody{ID: &userID})
	if err != nil {
		return asres.NewRedirect(getRedirectURL("No user associated with the checkout session."), http.StatusTemporaryRedirect), nil
	}

	return asres.NewRedirect(fmt.Sprintf("%s://%s/sign-up?token=%s&email=%s", h.Config().AppProtocol, h.Config().AppHost, url.QueryEscape(authCode.ValueHash), url.QueryEscape(user.Email)), http.StatusTemporaryRedirect), nil
}

func (h *Handler) Router() chi.Router {
	r := chi.NewRouter()

	r.Post("/webhook", asres.ToHandlerFunc(h.Webhook))
	r.Get("/payment-link-success", asres.ToHandlerFunc(h.PaymentLinkSuccess))

	return r
}
