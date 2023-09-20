package user

import (
	"github.com/ayo-ajayi/bloggy/db"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type UserRepo struct {
	collection *mongo.Collection
}

func NewUserRepo(collection *mongo.Collection) *UserRepo {
	return &UserRepo{collection}
}

func (repo *UserRepo)IsExists(filter interface{}, opts ...*options.FindOneOptions) (bool, error) {
	ctx, cancel := db.DBReqContext(5)
	defer cancel()
	err := repo.collection.FindOne(ctx, filter, opts...).Err()
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (repo *UserRepo) CreateUser(user *User) (*mongo.InsertOneResult, error) {
	ctx, cancel := db.DBReqContext(20)
	defer cancel()
	return repo.collection.InsertOne(ctx, user)
}
func(repo *UserRepo)GetUser(filter interface{}, opts ...*options.FindOneOptions)(*User, error){
	ctx, cancel := db.DBReqContext(20)
	defer cancel()
	var user User
	err := repo.collection.FindOne(ctx, filter, opts...).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func(repo *UserRepo)UpdateUser(filter interface{}, update interface{}, opts ...*options.UpdateOptions)(*mongo.UpdateResult, error){
	ctx, cancel := db.DBReqContext(20)
	defer cancel()
	return repo.collection.UpdateOne(ctx, filter, update, opts...)
}