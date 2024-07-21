package blog

import (
	"context"
	"errors"
	"time"

	"github.com/gosimple/slug"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type BlogService struct {
	repo BlogRepository
}

type BlogRepository interface {
	CreateBlogPost(ctx context.Context, blogPost *BlogPost) (*mongo.InsertOneResult, error)
	GetBlogPosts(ctx context.Context, filter interface{}, opts ...*options.FindOptions) ([]*BlogPost, error)
	GetBlogPost(ctx context.Context, filter interface{}, opts ...*options.FindOneOptions) (*BlogPost, error)
	UpdateBlogPost(ctx context.Context, filter interface{}, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error)
	DeleteBlogPost(ctx context.Context, filter interface{}, opts ...*options.DeleteOptions) (*mongo.DeleteResult, error)
	SearchBlogPosts(ctx context.Context, query string) ([]*BlogPost, error)
	PostComment(ctx context.Context, comment *Comment) (*mongo.InsertOneResult, error)
	GetComments(ctx context.Context, filter interface{}, opts ...*options.FindOptions) ([]*Comment, error)
	GetComment(ctx context.Context, filter interface{}, opts ...*options.FindOneOptions) (*Comment, error)
	UpdateComment(ctx context.Context, filter interface{}, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error)
	DeleteComment(ctx context.Context, filter interface{}, opts ...*options.DeleteOptions) (*mongo.DeleteResult, error)
}

func NewBlogService(repo BlogRepository) *BlogService {
	return &BlogService{repo}
}

func (service *BlogService) CreateBlogPost(ctx context.Context, blogPost *BlogPost) error {
	blogPost.Slug = slug.Make(blogPost.Title)
	p, err := service.repo.GetBlogPost(ctx, bson.M{"slug": blogPost.Slug})
	if err != nil {
		if err == mongo.ErrNoDocuments {
		} else {
			return err
		}
	}
	if p != nil {
		blogPost.Slug = blogPost.Slug + "-" + primitive.NewObjectID().Hex()
	}
	blogPost.CreatedAt = time.Now()
	blogPost.UpdatedAt = time.Now()
	_, err = service.repo.CreateBlogPost(ctx, blogPost)
	return err
}

func (service *BlogService) GetBlogPosts(ctx context.Context) ([]*BlogPost, error) {
	posts, err := service.repo.GetBlogPosts(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	return posts, nil
}

func (service *BlogService) GetBlogPostByID(ctx context.Context, idStr string) (*BlogPost, error) {
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		return nil, err
	}
	post, err := service.repo.GetBlogPost(ctx, bson.M{"_id": id})
	if err != nil {
		return nil, err
	}
	return post, nil
}

func (service *BlogService) GetBlogPostBySlug(ctx context.Context, slug string) (*BlogPost, error) {
	post, err := service.repo.GetBlogPost(ctx, bson.M{"slug": slug})
	if err != nil {
		return nil, err
	}
	return post, nil
}

func (service *BlogService) UpdateBlogPost(ctx context.Context, blogPost *BlogPost) error {
	oldPost, err := service.repo.GetBlogPost(ctx, bson.M{"_id": blogPost.Id})
	if err != nil {
		return err
	}
	blogPost.CreatedAt = oldPost.CreatedAt
	blogPost.Likes = oldPost.Likes
	blogPost.UpdatedAt = time.Now()
	blogPost.Slug = slug.Make(blogPost.Title)
	_, err = service.repo.UpdateBlogPost(ctx, bson.M{"_id": blogPost.Id}, bson.M{"$set": blogPost})
	return err
}

func (service *BlogService) DeleteBlogPost(ctx context.Context, idStr string) error {
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		return err
	}
	_, err = service.repo.DeleteBlogPost(ctx, bson.M{"_id": id})
	return err
}

func (service *BlogService) SearchBlogPosts(ctx context.Context, query string) ([]*BlogPost, error) {
	return service.repo.SearchBlogPosts(ctx, query)
}

func (service *BlogService) PostComment(ctx context.Context, comment *Comment) error {
	comment.CreatedAt = time.Now()
	comment.UpdatedAt = time.Now()
	_, err := service.repo.PostComment(ctx, comment)
	return err
}

func (service *BlogService) GetComments(ctx context.Context, postIdStr string) ([]*Comment, error) {
	postId, err := primitive.ObjectIDFromHex(postIdStr)
	if err != nil {
		return nil, err
	}
	comments, err := service.repo.GetComments(ctx, bson.M{"blog_post_id": postId})
	if err != nil {
		return nil, err
	}
	return comments, nil
}

func (service *BlogService) GetComment(ctx context.Context, idStr string) (*Comment, error) {
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		return nil, err
	}
	comment, err := service.repo.GetComment(ctx, bson.M{"_id": id})
	if err != nil {
		return nil, err
	}
	return comment, nil
}

func (service *BlogService) UpdateComment(ctx context.Context, comment *Comment) error {
	old, err := service.repo.GetComment(ctx, bson.M{"_id": comment.Id, "author_id": comment.AuthorId})
	if err != nil {
		return err
	}
	comment.CreatedAt = old.CreatedAt
	comment.BlogPostId = old.BlogPostId
	comment.ParentId = old.ParentId
	comment.Likes = old.Likes
	comment.UpdatedAt = time.Now()

	_, err = service.repo.UpdateComment(ctx, bson.M{"_id": comment.Id}, bson.M{"$set": comment})
	return err
}

func (service *BlogService) DeleteComment(ctx context.Context, idStr string) error {
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		return err
	}
	_, err = service.repo.DeleteComment(ctx, bson.M{"_id": id})
	return err
}

func (service *BlogService) LikeOrUnlikePost(ctx context.Context, postIdStr, userId string, opt PostOption) error {
	postId, err := primitive.ObjectIDFromHex(postIdStr)
	if err != nil {
		return err
	}
	if opt == LikePost {
		return service.likePost(ctx, postId, userId)
	} else if opt == UnlikePost {
		return service.unlikePost(ctx, postId, userId)
	}
	return nil
}

func (service *BlogService) likePost(ctx context.Context, postId primitive.ObjectID, userId string) error {
	post, err := service.repo.GetBlogPost(ctx, bson.M{"_id": postId})
	if err != nil {
		return err
	}
	for _, like := range post.Likes {
		if like.UserId == userId {
			return errors.New("already liked post")
		}
	}
	post.Likes = append(post.Likes, Like{UserId: userId})
	_, err = service.repo.UpdateBlogPost(ctx, bson.M{"_id": postId}, bson.M{"$set": post})
	return err
}

func (service *BlogService) unlikePost(ctx context.Context, postId primitive.ObjectID, userId string) error {
	post, err := service.repo.GetBlogPost(ctx, bson.M{"_id": postId})
	if err != nil {
		return err
	}
	currentlylikesPost := false
	var updatedLikes []Like
	for _, like := range post.Likes {
		if like.UserId == userId {
			currentlylikesPost = true
		} else {
			updatedLikes = append(updatedLikes, like)
		}
	}
	if !currentlylikesPost {
		return errors.New("post is not currently liked")
	}
	post.Likes = updatedLikes
	_, err = service.repo.UpdateBlogPost(ctx, bson.M{"_id": postId}, bson.M{"$set": bson.M{"likes": post.Likes}})
	return err
}

type PostOption string

const LikePost PostOption = "like"
const UnlikePost PostOption = "unlike"

type CommnentOption string

const LikeComment CommnentOption = "like"
const UnlikeComment CommnentOption = "unlike"

func (service *BlogService) LikeOrUnlikeComment(ctx context.Context, commentIdStr, userId string, opt CommnentOption) error {
	postId, err := primitive.ObjectIDFromHex(commentIdStr)
	if err != nil {
		return err
	}
	if opt == LikeComment {
		return service.likeComment(ctx, postId, userId)
	} else if opt == UnlikeComment {
		return service.unlikeComment(ctx, postId, userId)
	}
	return nil
}

func (service *BlogService) likeComment(ctx context.Context, commentId primitive.ObjectID, userId string) error {
	comment, err := service.repo.GetComment(ctx, bson.M{"_id": commentId})
	if err != nil {
		return err
	}
	for _, like := range comment.Likes {
		if like.UserId == userId {
			return errors.New("already liked comment")
		}
	}
	comment.Likes = append(comment.Likes, Like{UserId: userId})
	_, err = service.repo.UpdateComment(ctx, bson.M{"_id": commentId}, bson.M{"$set": comment})
	return err
}

func (service *BlogService) unlikeComment(ctx context.Context, commentId primitive.ObjectID, userId string) error {
	comment, err := service.repo.GetComment(ctx, bson.M{"_id": commentId})
	if err != nil {
		return err
	}
	currentlylikesComment := false
	var updatedLikes []Like
	for _, like := range comment.Likes {
		if like.UserId == userId {
			currentlylikesComment = true
		} else {
			updatedLikes = append(updatedLikes, like)
		}
	}
	if !currentlylikesComment {
		return errors.New("comment is not currently liked")
	}
	comment.Likes = updatedLikes
	_, err = service.repo.UpdateComment(ctx, bson.M{"_id": commentId}, bson.M{"$set": bson.M{"likes": comment.Likes}})
	return err
}
