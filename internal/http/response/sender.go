package response

import "github.com/gin-gonic/gin"

func Send(c *gin.Context, resp *APIResponse) {
	c.JSON(resp.StatusCode, resp)
}
