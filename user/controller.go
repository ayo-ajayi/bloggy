package user

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type UserController struct {
	service UserServices
}

func NewUserController(service UserServices) *UserController {
	return &UserController{service}
}

type UserServices interface {
	Login(ctx *gin.Context) (string, error)
	Callback(ctx *gin.Context) (*GoogleLoginResponse, error)
	SaveAccessToken(userId string, td *TokenDetails) error
	GenerateAccessToken(userId string) (*TokenDetails, error)
	SaveUser(googleLoginResponse *GoogleLoginResponse) error 
	Logout(accessUuid string) error 
	Profile(userId string) (*User, error)
}

func (uc *UserController) Login(c *gin.Context) {
	url, err := uc.service.Login(c)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.Redirect(http.StatusTemporaryRedirect, url)
}

func (uc *UserController) Callback(c *gin.Context) {
	content, err := uc.service.Callback(c)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	if err:=uc.service.SaveUser(content);err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	td, err := uc.service.GenerateAccessToken(content.ID)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	if err := uc.service.SaveAccessToken(content.ID, td); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	lr := &LoginResponse{
		AccessToken:         td.AccessToken,
		AtExpires:           td.AtExpires,
		GoogleLoginResponse: *content,
	}
	c.JSON(http.StatusOK, gin.H{"data": lr})
}

type LoginResponse struct {
	AccessToken string `json:"access_token"`
	AtExpires   int64  `json:"at_expires"`
	GoogleLoginResponse
}

func(uc *UserController)Logout(c *gin.Context) {
	accessUuid := c.GetString("access_uuid")
	if err := uc.service.Logout(accessUuid); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "Successfully logged out"})
}

func (uc *UserController)Profile(c *gin.Context) {
	userid:=c.MustGet("user_id").(string)
	user, err := uc.service.Profile(userid)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, user)
}