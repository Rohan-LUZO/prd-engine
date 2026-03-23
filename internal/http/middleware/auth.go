package middleware

import (
	"net/http"
	"strings"

	"prd-engine/internal/auth"
	"prd-engine/internal/http/response"

	"github.com/gin-gonic/gin"
)

// ContextUserKey is the key used to store the authenticated username in gin.Context (e.g. for SaveModule "updatedBy").
const ContextUserKey = "user"

// Auth is a middleware that authenticates requests using a bearer token looked up in the given UserStore.
// If valid, the user's Username is set in context under ContextUserKey and the next handler runs.
// If missing/invalid token, it responds with 401 and aborts the chain.
// Clients must send: Authorization: Bearer <token>
func Auth(store auth.UserStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			unauthorized(c, "Missing Authorization header")
			return
		}
		const prefix = "Bearer "
		if !strings.HasPrefix(authHeader, prefix) {
			unauthorized(c, "Invalid Authorization scheme")
			return
		}
		token := strings.TrimSpace(strings.TrimPrefix(authHeader, prefix))
		if token == "" {
			unauthorized(c, "Empty bearer token")
			return
		}

		user, err := store.FindByToken(token)
		if err != nil {
			unauthorized(c, "Authentication error")
			return
		}
		if user == nil {
			unauthorized(c, "Invalid token")
			return
		}

		// Attach username to context for downstream handlers
		c.Set(ContextUserKey, user.Username)
		c.Next()
	}
}

func unauthorized(c *gin.Context, msg string) {
	response.Send(
		c,
		response.New(http.StatusUnauthorized).
			WithMessage(msg).
			WithUserMessage(
				"Unauthorized",
				"Your session is not valid or has expired. Please log in again.",
			),
	)
	c.Abort()
}

