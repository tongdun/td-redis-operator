package common

import "github.com/gin-gonic/gin"

func Common(g *gin.Context) {
	g.Set("dba", true)
	g.Set("realname", "administrator")
}
