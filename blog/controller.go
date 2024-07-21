package blog

import (
	"context"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type BlogController struct {
	service BlogServices
}

type BlogServices interface {
	CreateBlogPost(ctx context.Context, blogPost *BlogPost) error
	GetBlogPosts(ctx context.Context) ([]*BlogPost, error)
	GetBlogPostByID(ctx context.Context, idStr string) (*BlogPost, error)
	GetBlogPostBySlug(ctx context.Context, slug string) (*BlogPost, error)
	UpdateBlogPost(ctx context.Context, blogPost *BlogPost) error
	DeleteBlogPost(ctx context.Context, idStr string) error
	SearchBlogPosts(ctx context.Context, query string) ([]*BlogPost, error)

	PostComment(ctx context.Context, comment *Comment) error
	GetComments(ctx context.Context, postIdStr string) ([]*Comment, error)
	GetComment(ctx context.Context, idStr string) (*Comment, error)
	UpdateComment(ctx context.Context, comment *Comment) error
	DeleteComment(ctx context.Context, idStr string) error
	LikeOrUnlikePost(ctx context.Context, postIdStr, userId string, opt PostOption) error
	LikeOrUnlikeComment(ctx context.Context, commentIdStr, userId string, opt CommnentOption) error
}

func NewBlogController(service BlogServices) *BlogController {
	return &BlogController{service}
}

func (controller *BlogController) CreateBlogPost(c *gin.Context) {
	req := struct {
		Title       string `json:"title" binding:"required"`
		Content     string `json:"content" binding:"required"`
		Description string `json:"description" binding:"required"`
	}{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	bp := &BlogPost{
		Title:       req.Title,
		Content:     req.Content,
		Description: req.Description,
	}

	if err := controller.service.CreateBlogPost(c, bp); err != nil {
		c.JSON(500, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	c.JSON(200, gin.H{"message": "Blog post created successfully"})
}

func (controller *BlogController) GetBlogPosts(c *gin.Context) {
	posts, err := controller.service.GetBlogPosts(c)
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	c.JSON(200, gin.H{"data": posts})
}

func (controller *BlogController) Search(c *gin.Context) {
	q := c.Query("q")
	if q == "" {
		c.JSON(400, gin.H{"error": gin.H{"message": "query param is required"}})
		return
	}
	bp, err := controller.service.SearchBlogPosts(c, q)
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	c.JSON(200, gin.H{"data": bp})
}

func (controller *BlogController) GetBlogPostByID(c *gin.Context) {
	id := c.Param("id")
	post, err := controller.service.GetBlogPostByID(c, id)
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	c.JSON(200, gin.H{"data": post})
}
func (controller *BlogController) GetBlogPostBySlug(c *gin.Context) {
	slug := c.Param("slug")
	post, err := controller.service.GetBlogPostBySlug(c, slug)
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	c.JSON(200, gin.H{"data": post})
}

func (controller *BlogController) UpdateBlogPost(c *gin.Context) {
	id := c.Param("id")
	post, err := controller.service.GetBlogPostByID(c, id)
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	req := struct {
		Title       string `json:"title" binding:"required"`
		Content     string `json:"content" binding:"required"`
		Description string `json:"description" binding:"required"`
	}{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	post.Title = req.Title
	post.Content = req.Content
	post.Description = req.Description
	if err := controller.service.UpdateBlogPost(c, post); err != nil {
		c.JSON(500, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	c.JSON(200, gin.H{"message": "Blog post updated successfully"})
}

func (controller *BlogController) DeleteBlogPost(c *gin.Context) {
	id := c.Param("id")
	if err := controller.service.DeleteBlogPost(c, id); err != nil {
		c.JSON(500, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	c.JSON(200, gin.H{"message": "Blog post deleted successfully"})
}

func (controller *BlogController) LikeOrUnlikePost(c *gin.Context) {
	userid, exists := c.Get("user_id")
	if !exists {
		c.JSON(500, gin.H{"error": gin.H{"message": "user not found"}})
		return
	}
	req := struct {
		ID     string     `json:"id" binding:"required"`
		Option PostOption `json:"option" binding:"required"`
	}{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	if err := controller.service.LikeOrUnlikePost(c, req.ID, userid.(string), req.Option); err != nil {
		c.JSON(500, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	if req.Option == LikePost {
		c.JSON(200, gin.H{"message": "Blog post liked successfully"})
	} else {
		c.JSON(200, gin.H{"message": "Blog post unliked successfully"})
	}
}

func (controller *BlogController) LikeOrUnlikeComment(c *gin.Context) {
	userid, exists := c.Get("user_id")
	if !exists {
		c.JSON(500, gin.H{"error": gin.H{"message": "user not found"}})
		return
	}
	req := struct {
		ID     string         `json:"id" binding:"required"`
		Option CommnentOption `json:"option" binding:"required"`
	}{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	if err := controller.service.LikeOrUnlikeComment(c, req.ID, userid.(string), req.Option); err != nil {
		c.JSON(500, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	if req.Option == LikeComment {
		c.JSON(200, gin.H{"message": "Comment liked successfully"})
	} else {
		c.JSON(200, gin.H{"message": "Comment unliked successfully"})
	}
}

func (controller *BlogController) PostComment(c *gin.Context) {
	userid, exists := c.Get("user_id")
	if !exists {
		c.JSON(500, gin.H{"error": gin.H{"message": "user not found"}})
		return
	}
	req := struct {
		BlogPostId string `json:"blog_post_id" binding:"required"`
		ParentId   string `json:"parent_id"`
		Content    string `json:"content" binding:"required"`
	}{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	blogPostId, err := primitive.ObjectIDFromHex(req.BlogPostId)
	if err != nil {
		c.JSON(400, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	comment := &Comment{
		AuthorId:   userid.(string),
		BlogPostId: blogPostId,
		Content:    req.Content,
	}
	if req.ParentId != "" {
		parentId, err := primitive.ObjectIDFromHex(req.ParentId)
		if err != nil {
			c.JSON(400, gin.H{"error": gin.H{"message": err.Error()}})
			return
		}
		comment.ParentId = parentId
	}

	if err := controller.service.PostComment(c, comment); err != nil {
		c.JSON(500, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	c.JSON(200, gin.H{"message": "Comment posted successfully"})
}

func (controller *BlogController) GetComments(c *gin.Context) {
	postId := c.Param("postId")
	comments, err := controller.service.GetComments(c, postId)
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	c.JSON(200, gin.H{"data": comments})
}

func (controller *BlogController) GetComment(c *gin.Context) {
	id := c.Param("id")
	comment, err := controller.service.GetComment(c, id)
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	c.JSON(200, gin.H{"data": comment})
}

func (controller *BlogController) UpdateComment(c *gin.Context) {
	userid, exists := c.Get("user_id")
	if !exists {
		c.JSON(500, gin.H{"error": gin.H{"message": "user not found"}})
		return
	}
	idStr := c.Param("id")
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		c.JSON(400, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	req := struct {
		Content string `json:"content" binding:"required"`
	}{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	comment := &Comment{
		AuthorId: userid.(string),
		Id:       id,
		Content:  req.Content,
	}
	if err := controller.service.UpdateComment(c, comment); err != nil {
		c.JSON(500, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	c.JSON(200, gin.H{"message": "Comment updated successfully"})
}
