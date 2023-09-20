package blog

import (
	"github.com/gin-gonic/gin"
)

type BlogController struct {
	service BlogServices
}

type BlogServices interface {
	CreateBlogPost(blogPost *BlogPost) error
	GetBlogPosts() ([]*BlogPost, error)
	GetBlogPost(idStr string) (*BlogPost, error)
	UpdateBlogPost(blogPost *BlogPost) error
	DeleteBlogPost(idStr string) error
	SearchBlogPosts(query string) ([]*BlogPost, error)
}

func NewBlogController(service BlogServices) *BlogController {
	return &BlogController{service}
}

func (controller *BlogController) CreateBlogPost(c *gin.Context) {
	req := struct {
		Title   string `json:"title" binding:"required"`
		Content string `json:"content" binding:"required"`
	}{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	bp := &BlogPost{
		Title:   req.Title,
		Content: req.Content,
	}

	if err := controller.service.CreateBlogPost(bp); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "Blog post created successfully"})
}

func (controller *BlogController) GetBlogPosts(c *gin.Context) {
	posts, err := controller.service.GetBlogPosts()
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"data": posts})
}

func (controller *BlogController) Search(c *gin.Context) {
	q := c.Query("query")
	if q == "" {
		c.JSON(400, gin.H{"error": "query param is required"})
		return
	}
	bp, err := controller.service.SearchBlogPosts(q)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"data": bp})
}

func (controller *BlogController) GetBlogPost(c *gin.Context) {
	id := c.Param("id")
	post, err := controller.service.GetBlogPost(id)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"data": post})
}

func (controller *BlogController) UpdateBlogPost(c *gin.Context) {
	id := c.Param("id")
	post, err := controller.service.GetBlogPost(id)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	req := struct {
		Title   string `json:"title" binding:"required"`
		Content string `json:"content" binding:"required"`
	}{
		Title:   "",
		Content: "",
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	post.Title = req.Title
	post.Content = req.Content
	if err := controller.service.UpdateBlogPost(post); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "Blog post updated successfully"})
}

func (controller *BlogController) DeleteBlogPost(c *gin.Context) {
	id := c.Param("id")
	if err := controller.service.DeleteBlogPost(id); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "Blog post deleted successfully"})
}
