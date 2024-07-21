package user

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type UserRepo struct {
	collection *mongo.Collection
}

func NewUserRepo(collection *mongo.Collection) *UserRepo {
	return &UserRepo{collection}
}

func (repo *UserRepo) IsExists(ctx context.Context, filter interface{}, opts ...*options.FindOneOptions) (bool, error) {
	err := repo.collection.FindOne(ctx, filter, opts...).Err()
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (repo *UserRepo) CreateUser(ctx context.Context, user *User) (*mongo.InsertOneResult, error) {
	return repo.collection.InsertOne(ctx, user)
}

func (repo *UserRepo) GetUser(ctx context.Context, filter interface{}, opts ...*options.FindOneOptions) (*User, error) {
	var user User
	err := repo.collection.FindOne(ctx, filter, opts...).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (repo *UserRepo) UpdateUser(ctx context.Context, filter interface{}, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	return repo.collection.UpdateOne(ctx, filter, update, opts...)
}

func (repo *UserRepo) CreateAboutMe(ctx context.Context, filter interface{}, opts ...*options.InsertOneOptions) (*mongo.InsertOneResult, error) {
	return repo.collection.InsertOne(ctx, filter, opts...)
}
func (repo *UserRepo) UpdateAboutMe(ctx context.Context, filter interface{}, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	return repo.UpdateUser(ctx, filter, update, opts...)
}
func (repo *UserRepo) GetAboutMe(ctx context.Context, filter interface{}, opts ...*options.FindOneOptions) (*AboutMe, error) {
	var aboutMe AboutMe
	err := repo.collection.FindOne(ctx, filter, opts...).Decode(&aboutMe)
	if err != nil {
		return nil, err
	}
	return &aboutMe, nil
}

func (repo *UserRepo) CreateMailingList(ctx context.Context, filter interface{}, opts ...*options.InsertOneOptions) (*mongo.InsertOneResult, error) {
	return repo.collection.InsertOne(ctx, filter, opts...)
}

func (repo *UserRepo) GetMailingList(ctx context.Context, filter interface{}, opts ...*options.FindOneOptions) (*MailingList, error) {
	var mailingList MailingList
	err := repo.collection.FindOne(ctx, filter, opts...).Decode(&mailingList)
	if err != nil {
		return nil, err
	}
	return &mailingList, nil
}

func (repo *UserRepo) UpdateMailingList(ctx context.Context, filter interface{}, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	return repo.UpdateUser(ctx, filter, update, opts...)
}

func (repo *UserRepo) GetUsers(ctx context.Context, filter interface{}, opts ...*options.FindOptions) ([]*User, error) {
	var users []*User
	cursor, err := repo.collection.Find(ctx, filter, opts...)
	if err != nil {
		return nil, err
	}
	if err = cursor.All(ctx, &users); err != nil {
		return nil, err
	}
	return users, nil
}
