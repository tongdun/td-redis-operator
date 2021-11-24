package web

import "github.com/gin-gonic/gin"

func Result(msg string, issuccess bool) gin.H {
	return gin.H{
		"message": msg,
		"success": issuccess,
	}
}
