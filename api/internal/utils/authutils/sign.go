package authutils

import (
	"api/config"
	"api/internal/http/aserr"
	"api/repository"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

func SignCode(secret string, userID string, value string, _type string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(fmt.Sprintf("%s:%s:%s", userID, value, _type)))
	return hex.EncodeToString(h.Sum(nil))
}

const codeCharset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
const codeLength = 8

func GenerateCode(secret string, _type string, userID string) string {
	randomBytes := make([]byte, 16)
	_, err := rand.Read(randomBytes)
	if err != nil {
		panic("Failed to generate random bytes")
	}

	h := hmac.New(sha512.New, []byte(secret))
	data := fmt.Sprintf("%s-%s-%s-%x", _type, userID, secret, randomBytes)
	h.Write([]byte(data))
	hash := h.Sum(nil)

	code := make([]byte, codeLength)
	for i := 0; i < codeLength; i++ {
		index := hash[i%len(hash)] % byte(len(codeCharset))
		code[i] = codeCharset[index]
	}
	return string(code)
}

func GenerateCodeHash(secret string, code string) string {
	h := hmac.New(sha512.New, []byte(secret))
	h.Write([]byte(code))
	return hex.EncodeToString(h.Sum(nil))
}

func VerifyCodeHash(secret, code, providedHash string) bool {
	expectedHash := GenerateCodeHash(secret, code)
	return hmac.Equal([]byte(expectedHash), []byte(providedHash))
}

func sessionCookieSignature(secret string, userID string, sessionID string, timestamp time.Time) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(fmt.Sprintf("%s:%s:%d", userID, sessionID, timestamp.Unix())))
	return hex.EncodeToString(h.Sum(nil))
}

func SessionCookie(secret string, userID string, sessionID string, timestamp time.Time) string {
	signature := sessionCookieSignature(secret, userID, sessionID, timestamp)
	return base64.URLEncoding.EncodeToString([]byte(
		fmt.Sprintf("%s:%s:%d:%s", userID, sessionID, timestamp.Unix(), signature),
	))
}

const SessionCookieName = "as-session"

func NewSessionCookie(cfg *config.Config, session *repository.Session, timestamp time.Time) *http.Cookie {
	return &http.Cookie{
		Name:     SessionCookieName,
		Value:    SessionCookie(cfg.SecretSession, session.UserID.String(), session.ID.String(), timestamp),
		Path:     "/",
		Domain:   cfg.SessionCookieDomain,
		MaxAge:   int(cfg.SessionTTL.Seconds()),
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
}

func ClearSessionCookie(cfg *config.Config) *http.Cookie {
	return &http.Cookie{Name: SessionCookieName, MaxAge: -1}
}

const (
	sessionCookieValuesLen = 4
)

type SessionCookieValue struct {
	ID     string
	UserID string
}

func GetSessionCookieValue(secret string, signedValue string) (*SessionCookieValue, error) {
	var userID, sessionID, signature string
	var timestamp int64

	decodedValue, err := url.QueryUnescape(signedValue)
	if err != nil {
		return nil, err
	}
	decoded, err := base64.URLEncoding.DecodeString(decodedValue)
	if err != nil {
		return nil, err
	}

	parts := strings.SplitN(string(decoded), ":", sessionCookieValuesLen)
	if len(parts) != sessionCookieValuesLen {
		return nil, aserr.ErrInvalidSignedValueFormat
	}

	userID, sessionID = parts[0], parts[1]
	timestamp, err = strconv.ParseInt(parts[2], 10, 64)
	if err != nil {
		return nil, aserr.ErrInvalidSignedValueTimestamp
	}
	signature = parts[3]

	expectedSignature := sessionCookieSignature(secret, userID, sessionID, time.Unix(timestamp, 0))
	if hmac.Equal([]byte(expectedSignature), []byte(signature)) {
		return &SessionCookieValue{UserID: userID, ID: sessionID}, nil
	}

	return nil, aserr.ErrInvalidSignedValueSignature
}
