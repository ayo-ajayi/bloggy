package user

import (
	"context"
	"mime/multipart"
	"net/http"

	"github.com/gin-gonic/gin"
)

type UserController struct {
	service  UserServices
	uploader Uploader
}

func NewUserController(service UserServices, uploader Uploader) *UserController {
	return &UserController{service, uploader}
}

type UserServices interface {
	Login(ctx *gin.Context) (string, error)
	Callback(ctx *gin.Context) (*GoogleLoginResponse, error)
	SaveAccessToken(ctx context.Context,userId string, td *TokenDetails) error
	GenerateAccessToken(userId string) (*TokenDetails, error)
	SaveUser(ctx context.Context,googleLoginResponse *GoogleLoginResponse) error
	Logout(ctx context.Context,accessUuid string) error
	Profile(ctx context.Context,userId string) (*User, error)
	UpdateAboutMe(ctx context.Context,userId, aboutMe, profilePicture string) error
	GetAboutMe(ctx context.Context) (*AboutMe, error)
	SubscribeToMailingList(ctx context.Context,id string) error
	UnSubscribeFromMailingList(ctx context.Context,id string) error
	GetMailingList(ctx context.Context) (*MailingList, error)
	GetUsers(ctx context.Context) ([]*User, error)
}

type Uploader interface {
	UploadImage(ctx context.Context, file *multipart.FileHeader, collection string) (string, error)
}

func (uc *UserController) Login(c *gin.Context) {
	url, err := uc.service.Login(c)
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	c.Redirect(http.StatusTemporaryRedirect, url)
}

func (uc *UserController) Callback(c *gin.Context) {
	content, err := uc.service.Callback(c)
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	if err := uc.service.SaveUser(c, content); err != nil {
		c.JSON(500, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	td, err := uc.service.GenerateAccessToken(content.ID)
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	if err := uc.service.SaveAccessToken(c, content.ID, td); err != nil {
		c.JSON(500, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	lr := &LoginResponse{
		AccessToken:         td.AccessToken,
		AtExpires:           td.AtExpires,
		GoogleLoginResponse: *content,
	}
	c.JSON(http.StatusOK, gin.H{"data": lr, "message": "Successfully logged in"})
}

type LoginResponse struct {
	AccessToken string `json:"access_token"`
	AtExpires   int64  `json:"at_expires"`
	GoogleLoginResponse
}

type AboutMe struct {
	Id             string `json:"id" bson:"_id"`
	AboutMe        string `json:"about_me" bson:"about_me"`
	ProfilePicture string `json:"profile_picture" bson:"profile_picture"`
}

func (uc *UserController) Logout(c *gin.Context) {
	accessUuid := c.GetString("access_uuid")
	if err := uc.service.Logout(c, accessUuid); err != nil {
		c.JSON(500, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	c.JSON(200, gin.H{"message": "Successfully logged out"})
}

func (uc *UserController) Profile(c *gin.Context) {
	userid, exists := c.Get("user_id")
	if !exists {
		c.JSON(500, gin.H{"error": gin.H{"message": "user not found"}})	
		return
	}

	user, err := uc.service.Profile(c, userid.(string))
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	c.JSON(200, gin.H{"data": user})
}

func (uc *UserController) UpdateAboutMe(c *gin.Context) {
	userid,exists := c.Get("user_id")
	if !exists {
		c.JSON(500, gin.H{"error": gin.H{"message": "user not found"}})
		return
	}	
	err := c.Request.ParseMultipartForm(32 << 20)
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	req := struct {
		AboutMe        string `json:"about_me"`
		ProfilePicture string `json:"profile_picture"`
	}{}
	req.AboutMe = c.PostForm("about_me")
	files := c.Request.MultipartForm.File["profile_picture"]
	if len(files) != 0 {
		image, err := uc.uploader.UploadImage(c, files[0], "profile_picture")
		if err != nil {
			c.JSON(500, gin.H{"error": gin.H{"message": err.Error()}})
			return
		}
		req.ProfilePicture = image
	} else {
		req.ProfilePicture = ""
	}
	if err := uc.service.UpdateAboutMe(c,userid.(string) , req.AboutMe, req.ProfilePicture); err != nil {
		c.JSON(500, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	c.JSON(200, gin.H{"message": "Successfully updated about me"})
}

func (uc *UserController) GetAboutMe(c *gin.Context) {
	aboutMe, err := uc.service.GetAboutMe(c)
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	c.JSON(200, gin.H{"data": gin.H{
		"about_me":        aboutMe.AboutMe,
		"profile_picture": aboutMe.ProfilePicture,
	}})
}

func (uc *UserController) SubscribeToMailingList(c *gin.Context) {
	id, exists := c.Get("user_id")
	if !exists {
		c.JSON(500, gin.H{"error": gin.H{"message": "user not found"}})
		return
	}
	if err := uc.service.SubscribeToMailingList(c, id.(string)); err != nil {
		c.JSON(500, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	c.JSON(200, gin.H{"message": "Successfully subscribed to mailing list"})
}

func (uc *UserController) UnSubscribeFromMailingList(c *gin.Context) {
	id, exists := c.Get("user_id")
	if !exists {
		c.JSON(500, gin.H{"error": gin.H{"message": "user not found"}})
		return
	}
	if err := uc.service.UnSubscribeFromMailingList(c, id.(string)); err != nil {
		c.JSON(500, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	c.JSON(200, gin.H{"message": "Successfully unsubscribed from mailing list"})
}

func (uc *UserController) GetMailingList(c *gin.Context) {
	mailingList, err := uc.service.GetMailingList(c)
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	c.JSON(200, gin.H{"data": mailingList})
}

func (uc *UserController) GetUsers(c *gin.Context) {
	u, err := uc.service.GetUsers(c)
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	c.JSON(200, gin.H{"data": u})
}
