package middleware

import (
	"net/http"

	"prd-engine/internal/http/response"

	"github.com/gin-gonic/gin"
)

func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				resp := response.New(http.StatusInternalServerError).
					WithMessage("Internal server error").
					WithUserMessage(
						"Something went wrong",
						"Our team has been notified. Please try again later.",
					).
					WithTag("PANIC_RECOVERY").
					WithCaller(3)

				response.Send(c, resp)
				c.Abort()
			}
		}()

		c.Next()
	}
}
