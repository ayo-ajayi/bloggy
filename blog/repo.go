package blog

import (
	"errors"
	"time"

	"github.com/ayo-ajayi/bloggy/db"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type BlogPost struct {
	Id          primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Title       string             `json:"title,omitempty" bson:"title,omitempty"`
	Slug        string             `json:"slug,omitempty" bson:"slug,omitempty"`
	Description string             `json:"description,omitempty" bson:"description,omitempty"`
	Content     string             `json:"content,omitempty" bson:"content,omitempty"`
	Likes       []Like             `json:"likes,omitempty" bson:"likes,omitempty"`
	CreatedAt   time.Time          `json:"created_at,omitempty" bson:"created_at,omitempty"`
	UpdatedAt   time.Time          `json:"updated_at,omitempty" bson:"updated_at,omitempty"`
}

type Comment struct {
	Id         primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	AuthorId   string             `json:"author_id,omitempty" bson:"author_id,omitempty"`
	BlogPostId primitive.ObjectID `json:"blog_post_id,omitempty" bson:"blog_post_id,omitempty"`
	ParentId   primitive.ObjectID `json:"parent_id,omitempty" bson:"parent_id,omitempty"`
	Content    string             `json:"content,omitempty" bson:"content,omitempty"`
	Likes      []Like             `json:"likes,omitempty" bson:"likes,omitempty"`
	CreatedAt  time.Time          `json:"created_at,omitempty" bson:"created_at,omitempty"`
	UpdatedAt  time.Time          `json:"updated_at,omitempty" bson:"updated_at,omitempty"`
}
type Like struct {
	UserId string `json:"user_id,omitempty" bson:"user_id,omitempty"`
}

type BlogRepo struct {
	blogCollection    *mongo.Collection
	commentCollection *mongo.Collection
}

func NewBlogRepo(blogCollection, commentCollection *mongo.Collection) *BlogRepo {
	return &BlogRepo{blogCollection, commentCollection}
}

func (repo *BlogRepo) CreateBlogPost(blogPost *BlogPost) (*mongo.InsertOneResult, error) {
	ctx, cancel := db.DBReqContext(20)
	defer cancel()
	return repo.blogCollection.InsertOne(ctx, blogPost)
}
func (rep *BlogRepo) GetBlogPosts(filter interface{}, opts ...*options.FindOptions) ([]*BlogPost, error) {
	ctx, cancel := db.DBReqContext(20)
	defer cancel()
	cur, err := rep.blogCollection.Find(ctx, filter, opts...)
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
	err := repo.blogCollection.FindOne(ctx, filter, opts...).Decode(&blogPost)
	if err != nil {
		return nil, err
	}
	return &blogPost, nil
}

func (repo *BlogRepo) UpdateBlogPost(filter interface{}, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	ctx, cancel := db.DBReqContext(20)
	defer cancel()
	return repo.blogCollection.UpdateOne(ctx, filter, update, opts...)
}

func (repo *BlogRepo) DeleteBlogPost(filter interface{}, opts ...*options.DeleteOptions) (*mongo.DeleteResult, error) {
	ctx, cancel := db.DBReqContext(20)
	defer cancel()
	return repo.blogCollection.DeleteOne(ctx, filter, opts...)
}

func InitSearchIndex(blogCollection *mongo.Collection) error {
	ctx, cancel := db.DBReqContext(20)
	defer cancel()
	_, err := blogCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "title", Value: "text"}, {Key: "content", Value: "text"}, {Key: "description", Value: "text"}},
		Options: options.Index().SetName("text_index"),
	})
	if err != nil {
		return errors.New("Error creating search index for blog collection: " + err.Error())
	}
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
			{"description": primitive.Regex{
				Pattern: query,
				Options: "i",
			},
			},
		}}
	return repo.GetBlogPosts(filter)
}

func (repo *BlogRepo) PostComment(comment *Comment) (*mongo.InsertOneResult, error) {
	ctx, cancel := db.DBReqContext(20)
	defer cancel()
	return repo.commentCollection.InsertOne(ctx, comment)
}
func (rep *BlogRepo) GetComments(filter interface{}, opts ...*options.FindOptions) ([]*Comment, error) {
	ctx, cancel := db.DBReqContext(20)
	defer cancel()
	cur, err := rep.commentCollection.Find(ctx, filter, opts...)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var comments []*Comment
	for cur.Next(ctx) {
		var comment Comment
		err := cur.Decode(&comment)
		if err != nil {
			return nil, err
		}
		comments = append(comments, &comment)
	}
	if err := cur.Err(); err != nil {
		return nil, err
	}
	return comments, nil
}

func (repo *BlogRepo) GetComment(filter interface{}, opts ...*options.FindOneOptions) (*Comment, error) {
	ctx, cancel := db.DBReqContext(20)
	defer cancel()
	var comment Comment
	err := repo.commentCollection.FindOne(ctx, filter, opts...).Decode(&comment)
	if err != nil {
		return nil, err
	}
	return &comment, nil
}

func (repo *BlogRepo) UpdateComment(filter interface{}, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	ctx, cancel := db.DBReqContext(20)
	defer cancel()
	return repo.commentCollection.UpdateOne(ctx, filter, update, opts...)
}

func (repo *BlogRepo) DeleteComment(filter interface{}, opts ...*options.DeleteOptions) (*mongo.DeleteResult, error) {
	ctx, cancel := db.DBReqContext(20)
	defer cancel()
	return repo.commentCollection.DeleteOne(ctx, filter, opts...)
}
