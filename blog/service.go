package blog

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

type BlogService struct {
	repo BlogRepository
}

type BlogRepository interface {
	CreateBlogPost(blogPost *BlogPost) (*mongo.InsertOneResult, error)
	GetBlogPosts(filter interface{}, opts ...*options.FindOptions) ([]*BlogPost, error)
	GetBlogPost(filter interface{}, opts ...*options.FindOneOptions) (*BlogPost, error)
	UpdateBlogPost(filter interface{}, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error)
	DeleteBlogPost(filter interface{}, opts ...*options.DeleteOptions) (*mongo.DeleteResult, error)
	SearchBlogPosts(query string) ([]*BlogPost, error)
}

func NewBlogService(repo BlogRepository) *BlogService {
	return &BlogService{repo}
}

func (service *BlogService) CreateBlogPost(blogPost *BlogPost) error {
	blogPost.CreatedAt = time.Now()
	blogPost.UpdatedAt = time.Now()
	_, err := service.repo.CreateBlogPost(blogPost)
	return err
}

func (service *BlogService) GetBlogPosts() ([]*BlogPost, error) {
	posts, err := service.repo.GetBlogPosts(bson.M{})
	if err != nil {
		return nil, err
	}
	return posts, nil
}

func (service *BlogService) GetBlogPost(idStr string) (*BlogPost, error) {
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		return nil, err
	}
	post, err := service.repo.GetBlogPost(bson.M{"_id": id})
	if err != nil {
		return nil, err
	}
	return post, nil
}

func (service *BlogService) UpdateBlogPost(blogPost *BlogPost) error {
	blogPost.UpdatedAt = time.Now()
	_, err := service.repo.UpdateBlogPost(bson.M{"_id": blogPost.Id}, bson.M{"$set": blogPost})
	return err
}

func (service *BlogService) DeleteBlogPost(idStr string) error {
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		return err
	}
	_, err = service.repo.DeleteBlogPost(bson.M{"_id": id})
	return err
}

func (service *BlogService) SearchBlogPosts(query string) ([]*BlogPost, error) {
	return service.repo.SearchBlogPosts(query)
}
