package router

import (
	"github.com/cdfmlr/crud/log"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// NewRouter creates a new router (a gin.New() router)
// with gin.Recovery() middleware, the log.Logger4Gin middleware
// and addon middlewares indicated by the options parameters.
func NewRouter(options ...RouterOption) *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery(), log.Logger4Gin)

	for _, option := range options {
		router = option(router).(*gin.Engine)
	}

	return router
}

type RouterOption func(router gin.IRouter) gin.IRouter

func AllowAllCors() RouterOption {
	return func(router gin.IRouter) gin.IRouter {
		router.Use(cors.Default())
		return router
	}
}
