package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/thakurfurqaan/excinity-posts/internal/auth"
	"github.com/thakurfurqaan/excinity-posts/internal/handlers"
	"gorm.io/gorm"
)

func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			c.JSON(401, gin.H{"error": "Request does not contain an access token"})
			c.Abort()
			return
		}

		token, err := auth.ValidateToken(tokenString)
		if err != nil {
			c.JSON(401, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok || !token.Valid {
			c.JSON(401, gin.H{"error": "Invalid token claims"})
			c.Abort()
			return
		}

		c.Set("userID", uint(claims["id"].(float64)))
		c.Next()
	}
}

func SetupRoutes(r *gin.Engine, db *gorm.DB) {
	r.POST("/register", handlers.Register(db))
	r.POST("/login", handlers.Login(db))

	authorized := r.Group("/")
	authorized.Use(authMiddleware())
	{
		authorized.POST("/posts", handlers.CreatePost(db))
		authorized.GET("/posts", handlers.GetPosts(db))
	}
}
