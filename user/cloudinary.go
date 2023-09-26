package user

import (
	"context"
	"errors"
	"mime/multipart"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

type MediaCloudManager struct {
	folder     string
	cloudinary *cloudinary.Cloudinary
}

func NewMediaCloudManager(api_uri, folder string) (*MediaCloudManager, error) {
	cld, err := cloudinary.NewFromURL(api_uri)
	if err != nil {
		return nil, errors.New("failed to init cloudinary: " + err.Error())
	}
	return &MediaCloudManager{
		folder:     folder,
		cloudinary: cld,
	}, nil
}

func (mcm *MediaCloudManager) UploadImage(ctx context.Context, file *multipart.FileHeader, collection string) (string, error) {
	f, err := file.Open()
	if err != nil {
		return "", errors.New("failed to open file: " + err.Error())
	}
	defer f.Close()
	res, err := mcm.cloudinary.Upload.Upload(ctx, f, uploader.UploadParams{Folder: mcm.folder + "/" + collection, UniqueFilename: api.Bool(false), ResourceType: "image", PublicID: "profile-picture"})
	if err != nil {
		return "", errors.New("failed to upload file: " + err.Error())
	}
	return res.SecureURL, nil

}
