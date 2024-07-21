package user

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type TokenDetails struct {
	AccessToken string `json:"access_token"`
	AcessUuid   string `json:"-"`
	AtExpires   int64  `json:"at_expires"`
}

type AccessDetails struct {
	ID         primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	AccessUuid string             `json:"access_uuid" bson:"access_uuid"`
	UserId     string             `json:"user_id" bson:"user_id"`
	ExpiresAt  time.Time          `json:"expires_at" bson:"expires_at"`
}

type TokenManager struct {
	accessTokenSecret           string
	accessTokenValidaityInHours int64
	collection                  *mongo.Collection
}

func NewTokenManager(accessTokenSecret string, accessTokenValidaityInHours int64, collection *mongo.Collection) *TokenManager {
	tm := &TokenManager{
		accessTokenSecret:           accessTokenSecret,
		accessTokenValidaityInHours: accessTokenValidaityInHours,
		collection:                  collection,
	}

	return tm
}

func InitTokenExpiryIndex(ctx context.Context, collection *mongo.Collection) error {
	indexModel := mongo.IndexModel{
		Keys: bson.M{
			"expires_at": 1}, Options: options.Index().SetExpireAfterSeconds(0),
	}
	_, err := collection.Indexes().CreateOne(ctx, indexModel)
	if err != nil {
		return errors.New("Error creating TTL index for token collection: " + err.Error())
	}
	return nil
}

func (tm *TokenManager) GenerateToken(userId string) (*TokenDetails, error) {
	td := &TokenDetails{}
	td.AtExpires = time.Now().Add(time.Hour * time.Duration(tm.accessTokenValidaityInHours)).Unix()
	td.AcessUuid = uuid.New().String()

	var err error
	td.AccessToken, err = createToken(userId, td.AcessUuid, td.AtExpires, tm.accessTokenSecret)
	if err != nil {
		return nil, err
	}
	if td.AccessToken == "" {
		return nil, errors.New("access token is empty")
	}
	return td, nil
}

func createToken(userId string, uuid string, expires int64, secret string) (string, error) {
	claims := jwt.MapClaims{}
	claims["authorized"] = true
	claims["user_id"] = userId
	claims["access_uuid"] = uuid
	claims["exp"] = expires
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return at.SignedString([]byte(secret))
}

func (tm *TokenManager) SaveToken(ctx context.Context, userId string, td *TokenDetails) error {
	_, err := tm.collection.InsertOne(ctx, &AccessDetails{
		AccessUuid: td.AcessUuid,
		UserId:     userId,
		ExpiresAt:  time.Unix(td.AtExpires, 0),
	})
	return err
}

func (tm *TokenManager) DeleteToken(ctx context.Context, filter interface{}, opts ...*options.DeleteOptions) error {
	_, err := tm.collection.DeleteOne(ctx, filter, opts...)
	return err
}
func (tm *TokenManager) IsExists(ctx context.Context, filter interface{}, opts ...*options.FindOneOptions) (bool, error) {
	err := tm.collection.FindOne(ctx, filter, opts...).Err()
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (tm *TokenManager) FindToken(ctx context.Context, filter interface{}, opts ...*options.FindOneOptions) (*AccessDetails, error) {
	var accessDetails AccessDetails
	err := tm.collection.FindOne(ctx, filter, opts...).Decode(&accessDetails)
	if err != nil {
		return nil, err
	}
	return &accessDetails, nil
}

func ValidateToken(token string, secret string) (*jwt.Token, error) {
	return jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(secret), nil
	})
}

func (tm *TokenManager) ExtractTokenMetadata(token *jwt.Token) (*AccessDetails, error) {
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, errors.New("unauthorized")
	}
	accessUuid, ok := claims["access_uuid"].(string)
	if !ok || accessUuid == "" {
		return nil, errors.New("unauthorized")
	}
	userId, ok := claims["user_id"].(string)
	if !ok || userId == "" {
		return nil, errors.New("unauthorized")
	}
	return &AccessDetails{
		AccessUuid: accessUuid,
		UserId:     userId,
	}, nil
}
