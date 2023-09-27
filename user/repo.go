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

func (repo *UserRepo) IsExists(filter interface{}, opts ...*options.FindOneOptions) (bool, error) {
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

func (repo *UserRepo) GetUser(filter interface{}, opts ...*options.FindOneOptions) (*User, error) {
	ctx, cancel := db.DBReqContext(20)
	defer cancel()
	var user User
	err := repo.collection.FindOne(ctx, filter, opts...).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (repo *UserRepo) UpdateUser(filter interface{}, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	ctx, cancel := db.DBReqContext(20)
	defer cancel()
	return repo.collection.UpdateOne(ctx, filter, update, opts...)
}

func (repo *UserRepo) CreateAboutMe(filter interface{}, opts ...*options.InsertOneOptions) (*mongo.InsertOneResult, error) {
	ctx, cancel := db.DBReqContext(20)
	defer cancel()
	return repo.collection.InsertOne(ctx, filter, opts...)
}
func (repo *UserRepo) UpdateAboutMe(filter interface{}, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	return repo.UpdateUser(filter, update, opts...)
}
func (repo *UserRepo) GetAboutMe(filter interface{}, opts ...*options.FindOneOptions) (*AboutMe, error) {
	ctx, cancel := db.DBReqContext(20)
	defer cancel()
	var aboutMe AboutMe
	err := repo.collection.FindOne(ctx, filter, opts...).Decode(&aboutMe)
	if err != nil {
		return nil, err
	}
	return &aboutMe, nil
}

func (repo *UserRepo) CreateMailingList(filter interface{}, opts ...*options.InsertOneOptions) (*mongo.InsertOneResult, error) {
	ctx, cancel := db.DBReqContext(20)
	defer cancel()
	return repo.collection.InsertOne(ctx, filter, opts...)
}

func (repo *UserRepo) GetMailingList(filter interface{}, opts ...*options.FindOneOptions) (*MailingList, error) {
	ctx, cancel := db.DBReqContext(20)
	defer cancel()
	var mailingList MailingList
	err := repo.collection.FindOne(ctx, filter, opts...).Decode(&mailingList)
	if err != nil {
		return nil, err
	}
	return &mailingList, nil
}

func (repo *UserRepo) UpdateMailingList(filter interface{}, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	return repo.UpdateUser(filter, update, opts...)
}

func (repo *UserRepo) GetUsers(filter interface{}, opts ...*options.FindOptions) ([]*User, error) {
	ctx, cancel := db.DBReqContext(20)
	defer cancel()
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
