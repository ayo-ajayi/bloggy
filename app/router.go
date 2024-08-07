package app

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/ayo-ajayi/bloggy/blog"
	"github.com/ayo-ajayi/bloggy/db"
	"github.com/ayo-ajayi/bloggy/user"
	"github.com/gin-gonic/gin"
	cors "github.com/rs/cors/wrapper/gin"
)

func BlogRouter() *gin.Engine {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := db.MongoClient(ctx, os.Getenv("MONGODB_URI"))
	if err != nil {
		log.Fatal(err.Error())
	}
	postCollection := client.Database("bloggy").Collection("posts")
	commentCollection := client.Database("bloggy").Collection("comments")
	tokenCollection := client.Database("bloggy").Collection("tokens")
	accessTokenSecret := os.Getenv("ACCESS_TOKEN_SECRET")
	accessTokenValidaityInHours := int64(24)
	tokenManager := user.NewTokenManager(accessTokenSecret, accessTokenValidaityInHours, tokenCollection)
	blogController := blog.NewBlogController(blog.NewBlogService(blog.NewBlogRepo(postCollection, commentCollection)))
	userRepo := user.NewUserRepo(client.Database("bloggy").Collection("users"))
	cloudinary, err := user.NewMediaCloudManager(os.Getenv("CLOUDINARY_URI"), "bloggy")
	if err != nil {
		log.Fatal(err.Error())
	}
	userController := user.NewUserController(user.NewUserService(userRepo, tokenManager), cloudinary)
	middleware := user.NewMiddleware(accessTokenSecret, userRepo, tokenManager)
	if err := blog.InitSearchIndex(ctx, postCollection); err != nil {
		log.Fatal(err.Error())
	}
	if err := user.InitTokenExpiryIndex(ctx, tokenCollection); err != nil {
		log.Fatal(err.Error())
	}
	r := gin.Default()
	r.Use(func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		c.Next()
	}, cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowCredentials: true,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
	}))

	r.NoRoute(func(ctx *gin.Context) { ctx.JSON(404, gin.H{"error": "endpoint not found"}) })
	r.GET("/", func(ctx *gin.Context) { ctx.JSON(200, gin.H{"message": "welcome to bloggy"}) })
	r.POST("/blog", middleware.Authentication(), middleware.Authorization([]user.Role{user.Admin}), blogController.CreateBlogPost)
	r.GET("/blog", blogController.GetBlogPosts)
	r.GET("/blog/:id", blogController.GetBlogPostByID)
	r.GET("/blog/slug/:slug", blogController.GetBlogPostBySlug)
	r.PUT("/blog/:id", middleware.Authentication(), middleware.Authorization([]user.Role{user.Admin}), blogController.UpdateBlogPost)
	r.DELETE("/blog/:id", middleware.Authentication(), middleware.Authorization([]user.Role{user.Admin}), blogController.DeleteBlogPost)
	r.GET("/search", blogController.Search)
	r.GET("/login", userController.Login)
	r.GET("/callback", userController.Callback)
	r.GET("/profile", middleware.Authentication(), userController.Profile)
	r.GET("/users", middleware.Authentication(), middleware.Authorization([]user.Role{user.Admin}), userController.GetUsers)
	r.DELETE("/logout", middleware.Authentication(), userController.Logout)
	r.POST("/like-unlike-post", middleware.Authentication(), blogController.LikeOrUnlikePost)
	r.POST("/like-unlike-comment", middleware.Authentication(), blogController.LikeOrUnlikeComment)
	r.POST("/comment", middleware.Authentication(), blogController.PostComment)
	r.PUT("/comment/:id", middleware.Authentication(), blogController.UpdateComment)
	r.GET("/comments/:postId", blogController.GetComments)
	r.GET("/comment/:id", blogController.GetComment)
	r.PUT("/about", middleware.Authentication(), middleware.Authorization([]user.Role{user.Admin}), userController.UpdateAboutMe)
	r.GET("/about", userController.GetAboutMe)
	r.POST("/subscribe", middleware.Authentication(), userController.SubscribeToMailingList)
	r.DELETE("/unsubscribe", middleware.Authentication(), userController.UnSubscribeFromMailingList)
	r.GET("/mailing-list", middleware.Authentication(), middleware.Authorization([]user.Role{user.Admin}), userController.GetMailingList)
	return r
}
