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

	client, err := db.MongoClient("mongodb://localhost:27017")
	if err != nil {
		panic(err)
	}
	postsCollection := client.Database("bloggy").Collection("posts")
	accessTokenSecret := os.Getenv("ACCESS_TOKEN_SECRET")
	accessTokenValidaityInHours := int64(24)
	tokenManager := user.NewTokenManager(accessTokenSecret, accessTokenValidaityInHours, client.Database("bloggy").Collection("tokens"))
	blogController := blog.NewBlogController(blog.NewBlogService(blog.NewBlogRepo(postsCollection)))
	userRepo := user.NewUserRepo(client.Database("bloggy").Collection("users"))
	userController := user.NewUserController(user.NewUserService(userRepo, tokenManager))
	middleware := user.NewMiddleware(accessTokenSecret, userRepo, tokenManager)
	if err := blog.InitSearchIndex(postsCollection); err != nil {
		log.Println(err.Error())
	}
	
	r.NoRoute(func(ctx *gin.Context) { ctx.JSON(404, gin.H{"error": "endpoint not found"}) })
	r.GET("/", func(ctx *gin.Context) { ctx.JSON(200, gin.H{"message": "welcome to bloggy"}) })
	r.POST("/blog", blogController.CreateBlogPost)
	r.GET("/blog", blogController.GetBlogPosts)
	r.GET("/blog/:id", blogController.GetBlogPost)
	r.PUT("/blog/:id", blogController.UpdateBlogPost)
	r.DELETE("/blog/:id", blogController.DeleteBlogPost)
	r.GET("/search", blogController.Search)
	r.GET("/login", userController.Login)
	r.GET("/callback", userController.Callback)
	r.GET("/profile", middleware.Authentication(), userController.Profile)
	r.DELETE("/logout", middleware.Authentication(), userController.Logout)
	return r
}
