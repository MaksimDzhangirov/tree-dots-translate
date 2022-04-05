package auth

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"firebase.google.com/go/auth"
	"github.com/MaksimDzhangirov/three-dots/part13/internal/common/server/httperr"
)

type FirebaseHttpMiddleware struct {
	AuthClient *auth.Client
}

func (a FirebaseHttpMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		bearerToken := a.tokenFromHeader(r)
		if bearerToken == "" {
			httperr.Unauthorised("empty-bearer-token", nil, w, r)
			return
		}

		token, err := a.AuthClient.VerifyIDToken(ctx, bearerToken)
		if err != nil {
			httperr.Unauthorised("unable-to-verity-jwt", err, w, r)
			return
		}

		// всегда рекомендуется использовать пользовательский тип в качестве значения контекста (в данном случае ctxKey)
		// потому что никто из вне пакета не сможет переопределить/прочитать это значение
		ctx = context.WithValue(ctx, userContentKey, User{
			UUID:        token.UID,
			Email:       token.Claims["email"].(string),
			Role:        token.Claims["role"].(string),
			DisplayName: token.Claims["name"].(string),
		})
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

func (a FirebaseHttpMiddleware) tokenFromHeader(r *http.Request) string {
	headerValue := r.Header.Get("Authorization")
	if len(headerValue) > 7 && strings.ToLower(headerValue[0:6]) == "bearer" {
		return headerValue[7:]
	}

	return ""
}

type User struct {
	UUID  string
	Email string
	Role  string

	DisplayName string
}

type ctxKey int

const (
	userContentKey ctxKey = iota
)

var (
	// если мы ожидаем, что пользователя функции может заинтересовать конкретная ошибка,
	// хорошей практикой является создать переменную с этой ошибкой
	NoUserInContextError = errors.New("no user in context")
)

func UserFromCtx(ctx context.Context) (User, error) {
	u, ok := ctx.Value(userContentKey).(User)
	if ok {
		return u, nil
	}

	return User{}, NoUserInContextError
}
