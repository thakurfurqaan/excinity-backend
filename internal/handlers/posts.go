package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/thakurfurqaan/excinity-posts/internal/models"
	"gorm.io/gorm"
)

func CreatePost(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var post models.Post
		if err := c.ShouldBindJSON(&post); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		userID, _ := c.Get("userID")
		post.UserID = userID.(uint)

		if err := db.Create(&post).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create post"})
			return
		}

		c.JSON(http.StatusCreated, post)
	}
}

func GetPosts(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var posts []models.Post
		if err := db.Preload("User").Find(&posts).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch posts"})
			return
		}

		c.JSON(http.StatusOK, posts)
	}
}
