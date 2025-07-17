package api

import "github.com/gin-gonic/gin"

func SetupRouter(r *gin.Engine) {

	apiGroup := r.Group("/api/v1")
	{
		apiGroup.GET("/hello", helloWorld)
	}
}
