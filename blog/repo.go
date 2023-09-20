package blog

import (
	"time"

	"github.com/ayo-ajayi/bloggy/db"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)



type BlogPost struct {
	Id primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	//AuthorId string `json:"authorId,omitempty" bson:"authorId,omitempty"`
	Title     string    `json:"title,omitempty" bson:"title,omitempty"`
	Content   string    `json:"content,omitempty" bson:"content,omitempty"`
	CreatedAt time.Time `json:"created_at,omitempty" bson:"created_at,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty" bson:"updated_at,omitempty"`
}

type Comment struct {
	Id primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	// AuthorId string `json:"authorId,omitempty" bson:"authorId,omitempty"`
	BlogPostId string    `json:"blog_post_id,omitempty" bson:"blog_post_id,omitempty"`
	Content    string    `json:"content,omitempty" bson:"content,omitempty"`
	CreatedAt  time.Time `json:"created_at,omitempty" bson:"created_at,omitempty"`
}

type BlogRepo struct {
	collection *mongo.Collection
}

func NewBlogRepo(collection *mongo.Collection) *BlogRepo {
	return &BlogRepo{collection}
}
func (repo *BlogRepo) CreateBlogPost(blogPost *BlogPost) (*mongo.InsertOneResult, error) {
	ctx, cancel := db.DBReqContext(20)
	defer cancel()
	return repo.collection.InsertOne(ctx, blogPost)
}
func (rep *BlogRepo) GetBlogPosts(filter interface{}, opts ...*options.FindOptions) ([]*BlogPost, error) {
	ctx, cancel := db.DBReqContext(20)
	defer cancel()
	cur, err := rep.collection.Find(ctx, filter, opts...)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var blogPosts []*BlogPost
	for cur.Next(ctx) {
		var blogPost BlogPost
		err := cur.Decode(&blogPost)
		if err != nil {
			return nil, err
		}
		blogPosts = append(blogPosts, &blogPost)
	}
	if err := cur.Err(); err != nil {
		return nil, err
	}
	return blogPosts, nil
}

func (repo *BlogRepo) GetBlogPost(filter interface{}, opts ...*options.FindOneOptions) (*BlogPost, error) {
	ctx, cancel := db.DBReqContext(20)
	defer cancel()
	var blogPost BlogPost
	err := repo.collection.FindOne(ctx, filter, opts...).Decode(&blogPost)
	if err != nil {
		return nil, err
	}
	return &blogPost, nil
}

func (repo *BlogRepo) UpdateBlogPost(filter interface{}, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	ctx, cancel := db.DBReqContext(20)
	defer cancel()
	return repo.collection.UpdateOne(ctx, filter, update, opts...)
}

func (repo *BlogRepo) DeleteBlogPost(filter interface{}, opts ...*options.DeleteOptions) (*mongo.DeleteResult, error) {
	ctx, cancel := db.DBReqContext(20)
	defer cancel()
	return repo.collection.DeleteOne(ctx, filter, opts...)
}

func InitSearchIndex(collection *mongo.Collection) error {
	ctx, cancel := db.DBReqContext(5)
	defer cancel()
	_, err := collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "title", Value: "text"}, {Key: "content", Value: "text"}},
		Options: options.Index().SetName("text_index"),
	})
	return err
}

func (repo *BlogRepo) SearchBlogPosts(query string) ([]*BlogPost, error) {

	filter := bson.M{
		"$or": []bson.M{
			{"title": primitive.Regex{
				Pattern: query,
				Options: "i",
			}},
			{"content": primitive.Regex{
				Pattern: query,
				Options: "i",
			},
			},
		}}
	return repo.GetBlogPosts(filter)
}
