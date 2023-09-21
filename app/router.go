package app

import (
	"log"
	"os"

	"github.com/ayo-ajayi/bloggy/blog"
	"github.com/ayo-ajayi/bloggy/db"
	"github.com/ayo-ajayi/bloggy/user"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func BlogRouter() *gin.Engine {
	r := gin.Default()
	cfg := cors.DefaultConfig()
	cfg.AllowAllOrigins = true
	cfg.AllowHeaders = []string{
		"Origin",
		"Content-Length",
		"Content-Type",
		"Authorization",
		"X-Requested-With",
		"Accept",
		"Accept-Language",
		"Host",
		"Referer",
		"User-Agent",
		"Access-Control-Allow-Origin",
		"Access-Control-Allow-Headers",
		"Access-Control-Allow-Methods",
		"Access-Control-Allow-Credentials",
		"Access-Control-Max-Age",
		"Access-Control-Expose-Headers",
		"Access-Control-Request-Headers",
		"Access-Control-Request-Method",
		"Connection",
		"Accept-Encoding",
	}

	r.Use(cors.New(cfg))

	client, err := db.MongoClient(os.Getenv("MONGODB_URI"))
	if err != nil {
		panic(err)
	}
	postCollection := client.Database("bloggy").Collection("posts")
	commentCollection := client.Database("bloggy").Collection("comments")
	tokenCollection := client.Database("bloggy").Collection("tokens")
	accessTokenSecret := os.Getenv("ACCESS_TOKEN_SECRET")
	accessTokenValidaityInHours := int64(24)
	tokenManager := user.NewTokenManager(accessTokenSecret, accessTokenValidaityInHours, tokenCollection)
	blogController := blog.NewBlogController(blog.NewBlogService(blog.NewBlogRepo(postCollection, commentCollection)))
	userRepo := user.NewUserRepo(client.Database("bloggy").Collection("users"))
	userController := user.NewUserController(user.NewUserService(userRepo, tokenManager))
	middleware := user.NewMiddleware(accessTokenSecret, userRepo, tokenManager)
	if err := blog.InitSearchIndex(postCollection); err != nil {
		log.Fatal(err.Error())
	}
	if err := user.InitTokenExpiryIndex(tokenCollection); err != nil {
		log.Fatal(err.Error())
	}

	r.NoRoute(func(ctx *gin.Context) { ctx.JSON(404, gin.H{"error": "endpoint not found"}) })
	r.GET("/", func(ctx *gin.Context) { ctx.JSON(200, gin.H{"message": "welcome to bloggy"}) })
	r.POST("/blog", middleware.Authentication(), middleware.Authorization([]user.Role{user.Admin}), blogController.CreateBlogPost)
	r.GET("/blog", blogController.GetBlogPosts)
	r.GET("/blog/:id", blogController.GetBlogPost)
	r.PUT("/blog/:id", middleware.Authentication(), middleware.Authorization([]user.Role{user.Admin}), blogController.UpdateBlogPost)
	r.DELETE("/blog/:id", middleware.Authentication(), middleware.Authorization([]user.Role{user.Admin}), blogController.DeleteBlogPost)
	r.GET("/search", blogController.Search)
	r.GET("/login", userController.Login)
	r.GET("/callback", userController.Callback)
	r.GET("/profile", middleware.Authentication(), userController.Profile)
	r.DELETE("/logout", middleware.Authentication(), userController.Logout)
	r.POST("/like-unlike-post", middleware.Authentication(), blogController.LikeOrUnlikePost)
	r.POST("/like-unlike-comment", middleware.Authentication(), blogController.LikeOrUnlikeComment)
	r.POST("/comment", middleware.Authentication(), blogController.PostComment)
	r.PUT("/comment/:id", middleware.Authentication(), blogController.UpdateComment)
	r.GET("/comments/:postId", blogController.GetComments)
	r.GET("/comment/:id", blogController.GetComment)

	return r
}
