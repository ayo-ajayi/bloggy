package user

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Middleware struct {
	accessTokenSecret string
	userRepo          MiddlewareUserRepo
	tokenManager      MiddlewareTokenManager
}

type MiddlewareTokenManager interface {
	FindToken(filter interface{}, opts ...*options.FindOneOptions) (*AccessDetails, error)
	ExtractTokenMetadata(token *jwt.Token) (*AccessDetails, error)
}
type MiddlewareUserRepo interface {
	GetUser(filter interface{}, opts ...*options.FindOneOptions) (*User, error)
}

func NewMiddleware(accessTokenSecret string, userRepo MiddlewareUserRepo, tokenManager MiddlewareTokenManager) *Middleware {
	return &Middleware{accessTokenSecret, userRepo, tokenManager}
}

func (m *Middleware) Authentication() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractToken(c.Request)
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": gin.H{"message": "unauthorized: token is required"}})
			c.Abort()
			return
		}
		jwtToken, err := ValidateToken(token, m.accessTokenSecret)
		if err != nil {
			if errors.Is(err, jwt.ErrSignatureInvalid) {
				c.JSON(http.StatusUnauthorized, gin.H{"error": gin.H{"message": "unauthorized: "+err.Error()}})
				c.Abort()
				return
			}
			if errors.Is(err, jwt.ErrTokenExpired) {
				c.JSON(http.StatusUnauthorized, gin.H{"error": gin.H{"message": "unauthorized: "+err.Error()}})
				c.Abort()
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": err.Error()}})
			c.Abort()
			return
		}
		td, err := m.tokenManager.ExtractTokenMetadata(jwtToken)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"message": "unauthorized: "+err.Error()}})
			c.Abort()
			return
		}
		_, err = m.tokenManager.FindToken(bson.M{"access_uuid": td.AccessUuid})
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"message": "unauthorized: "+err.Error()}})
			c.Abort()
			return
		}
		c.Set("access_uuid", td.AccessUuid)
		c.Set("user_id", td.UserId)
		c.Next()
	}
}

func extractToken(r *http.Request) string {
	token := r.Header.Get("Authorization")
	ttoken := strings.Split(token, " ")
	if len(ttoken) != 2 {
		return ""
	}
	return ttoken[1]
}

func (m *Middleware) Authorization(roles []Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		userId := c.MustGet("user_id").(string)
		user, err := m.userRepo.GetUser(bson.M{
			"_id": userId,
		})
		if err != nil {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": gin.H{"message": err.Error() + ": you are not authorized to acess this resource"}})
			return
		}
		allowed := false
		for _, role := range roles {
			if role == user.Role {
				allowed = true
				break
			}
		}
		if !allowed {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": gin.H{"message": err.Error() + ": you are not authorized to acess this resource"}})
			return
		}
		c.Next()
	}
}
