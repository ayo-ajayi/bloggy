package comment
import "time"
import "go.mongodb.org/mongo-driver/bson/primitive"


type Comment struct {
	Id         primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	AuthorId   string             `json:"authorId,omitempty" bson:"authorId,omitempty"`
	BlogPostId string             `json:"blog_post_id,omitempty" bson:"blog_post_id,omitempty"`
	ParentId   string             `json:"parent_id,omitempty" bson:"parent_id,omitempty"`
	Content    string             `json:"content,omitempty" bson:"content,omitempty"`
	Likes      []Like             `json:"likes,omitempty" bson:"likes,omitempty"`
	CreatedAt  time.Time          `json:"created_at,omitempty" bson:"created_at,omitempty"`
	UpdatedAt  time.Time          `json:"updated_at,omitempty" bson:"updated_at,omitempty"`
}

type Like struct {
	UserId string `json:"user_id,omitempty" bson:"user_id,omitempty"`
}

