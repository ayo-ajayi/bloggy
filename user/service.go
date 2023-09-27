package user

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/sessions"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type UserService struct {
	repo              UserRepository
	store             *sessions.CookieStore
	googleOauthConfig *oauth2.Config
	tokenMgr          TokenMgr
}

type TokenMgr interface {
	SaveToken(userId string, td *TokenDetails) error
	GenerateToken(userId string) (*TokenDetails, error)
	FindToken(filter interface{}, opts ...*options.FindOneOptions) (*AccessDetails, error)
	IsExists(filter interface{}, opts ...*options.FindOneOptions) (bool, error)
	DeleteToken(filter interface{}, opts ...*options.DeleteOptions) error
}

type UserRepository interface {
	CreateUser(user *User) (*mongo.InsertOneResult, error)
	GetUser(filter interface{}, opts ...*options.FindOneOptions) (*User, error)
	GetUsers(filter interface{}, opts ...*options.FindOptions) ([]*User, error)
	UpdateUser(filter interface{}, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error)
	IsExists(filter interface{}, opts ...*options.FindOneOptions) (bool, error)
	CreateAboutMe(filter interface{}, opts ...*options.InsertOneOptions) (*mongo.InsertOneResult, error)
	GetAboutMe(filter interface{}, opts ...*options.FindOneOptions) (*AboutMe, error)
	UpdateAboutMe(filter interface{}, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error)

	CreateMailingList(filter interface{}, opts ...*options.InsertOneOptions) (*mongo.InsertOneResult, error)
	GetMailingList(filter interface{}, opts ...*options.FindOneOptions) (*MailingList, error)
	UpdateMailingList(filter interface{}, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error)
}

func NewUserService(repo UserRepository, tokenMgr TokenMgr) *UserService {
	googleOauthConfig := &oauth2.Config{
		RedirectURL:  os.Getenv("REDIRECT_URL"),
		ClientID:     os.Getenv("CLIENT_ID"),
		ClientSecret: os.Getenv("CLIENT_SECRET"),
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
		Endpoint:     google.Endpoint,
	}
	store := sessions.NewCookieStore([]byte(os.Getenv("SESSION_SECRET")))

	return &UserService{
		repo:              repo,
		store:             store,
		googleOauthConfig: googleOauthConfig,
		tokenMgr:          tokenMgr,
	}
}
func generateRandomState() string {
	return uuid.New().String()
}
func (us *UserService) Login(ctx *gin.Context) (string, error) {
	state := generateRandomState()
	session, err := us.store.New(ctx.Request, "session-name")
	if err != nil {
		return "", err
	}
	session.Values["state"] = state
	if err = session.Save(ctx.Request, ctx.Writer); err != nil {
		return "", err
	}
	return us.googleOauthConfig.AuthCodeURL(state), nil
}

func (us *UserService) Callback(ctx *gin.Context) (*GoogleLoginResponse, error) {
	session, err := us.getSession(ctx)
	if err != nil {
		return nil, err
	}
	retrievedState, ok := session.Values["state"].(string)
	if !ok || retrievedState != ctx.Request.URL.Query().Get("state") {
		return nil, errors.New("unable to retrieve state")
	}
	session.Options.MaxAge = -1
	if err = session.Save(ctx.Request, ctx.Writer); err != nil {
		return nil, err
	}
	token, err := us.googleOauthConfig.Exchange(ctx, ctx.Request.URL.Query().Get("code"))
	if err != nil {
		return nil, err
	}
	client := us.googleOauthConfig.Client(ctx, token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token.AccessToken)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("unable to retrieve user info")
	}

	googleResponse := &GoogleLoginResponse{}
	if err := json.NewDecoder(resp.Body).Decode(&googleResponse); err != nil {
		return nil, err
	}
	return googleResponse, nil
}

func (us *UserService) getSession(ctx *gin.Context) (*sessions.Session, error) {
	return us.store.Get(ctx.Request, "session-name")
}

type GoogleLoginResponse struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
	Locale        string `json:"locale"`
}

func (us *UserService) GenerateAccessToken(userId string) (*TokenDetails, error) {
	td, err := us.tokenMgr.GenerateToken(userId)
	if err != nil {
		return nil, err
	}
	return td, nil
}

func (us *UserService) SaveAccessToken(userId string, td *TokenDetails) error {
	exists, err := us.tokenMgr.IsExists(bson.M{"user_id": userId})
	if err != nil {
		return err
	}
	if exists {
		if err := us.tokenMgr.DeleteToken(bson.M{"user_id": userId}); err != nil {
			return err
		}
	}
	return us.tokenMgr.SaveToken(userId, td)
}

func (us *UserService) SaveUser(googleLoginResponse *GoogleLoginResponse) error {
	role := Admin
	if googleLoginResponse.Email != os.Getenv("ADMIN_EMAIL") {
		role = Reader
	}
	user := &User{
		ID:         googleLoginResponse.ID,
		Name:       googleLoginResponse.Name,
		Email:      googleLoginResponse.Email,
		IsVerified: googleLoginResponse.VerifiedEmail,
		Role:       role,
		Picture:    googleLoginResponse.Picture,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	exists, err := us.repo.IsExists(bson.M{"email": user.Email})
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	_, err = us.repo.CreateUser(user)
	if err != nil {
		return err
	}
	return nil
}

func (us *UserService) GetUsers() ([]*User, error) {
	users, err := us.repo.GetUsers(bson.M{
		"role": bson.M{
			"$ne": Admin,
		},
		"name": bson.M{
			"$ne": "mailing_list",
		},
		"_id": bson.M{
			"$not": bson.M{
				"$regex": "profile_picture.*",
			},
		},
	})
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (us *UserService) Logout(accessUuid string) error {
	return us.tokenMgr.DeleteToken(bson.M{"access_uuid": accessUuid})
}

func (us *UserService) Profile(userId string) (*User, error) {
	return us.repo.GetUser(bson.M{"_id": userId})
}

func (us *UserService) UpdateAboutMe(userId, aboutMe, profilePicture string) error {
	exists, err := us.repo.IsExists(bson.M{"_id": "profile_picture" + userId})
	if err != nil {
		return err
	}
	if !exists {
		_, err := us.repo.CreateAboutMe(bson.M{"_id": "profile_picture" + userId, "about_me": aboutMe, "profile_picture": profilePicture, "updated_at": time.Now()})
		if err != nil {
			return err
		}
	}
	_, err = us.repo.UpdateUser(bson.M{"_id": "profile_picture" + userId}, bson.M{"$set": bson.M{"about_me": aboutMe, "profile_picture": profilePicture, "updated_at": time.Now()}})
	return err
}

func (us *UserService) GetAboutMe() (*AboutMe, error) {
	admin, err := us.repo.GetUser(bson.M{"role": Admin})
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	aboutMe, err := us.repo.GetAboutMe(bson.M{"_id": "profile_picture" + admin.ID})
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return aboutMe, nil
}

func (us *UserService) SubscribeToMailingList(id string) error {
	user, err := us.Profile(id)
	if err != nil {
		return err
	}
	mailingList, err := us.GetMailingList()
	if err != nil {
		return err
	}
	if mailingList == nil {
		_, err := us.repo.CreateMailingList(bson.M{"name": "mailing_list", "subscribers": []Subscriber{
			{
				Email:     user.Email,
				Name:      user.Name,
				CreatedAt: time.Now(),
			},
		}})
		if err != nil {
			return err
		}
		return nil
	}
	for _, subscriber := range mailingList.Subscribers {
		if subscriber.Email == user.Email {
			return errors.New("user is already subscribed to mailing list")
		}
	}
	_, err = us.repo.UpdateMailingList(bson.M{"name": "mailing_list"}, bson.M{"$push": bson.M{"subscribers": bson.M{"email": user.Email, "name": user.Name, "created_at": time.Now()}}})
	return err
}

func (us *UserService) GetMailingList() (*MailingList, error) {
	mailingList, err := us.repo.GetMailingList(bson.M{"name": "mailing_list"})
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return mailingList, nil
}

func (us *UserService) UnSubscribeFromMailingList(id string) error {
	user, err := us.Profile(id)
	if err != nil {
		return err
	}
	mailingList, err := us.GetMailingList()
	if err != nil {
		return err
	}
	if mailingList == nil {
		return errors.New("user is not subscribed to mailing list")
	}

	removeSubscriber := false
	for _, subscriber := range mailingList.Subscribers {
		if subscriber.Email == user.Email {
			removeSubscriber = true
			break
		}
	}
	if !removeSubscriber {
		return errors.New("user is not subscribed to mailing list")
	}
	_, err = us.repo.UpdateMailingList(bson.M{"name": "mailing_list"}, bson.M{"$pull": bson.M{"subscribers": bson.M{"email": user.Email}}})
	return err
}
