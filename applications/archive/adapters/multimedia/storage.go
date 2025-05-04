package multimedia

import (
	"bytes"
	"context"
	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"sherry.archive.com/shared/logger"
)

type StorageClient interface {
	Save(ctx context.Context, data []byte) (string, error)
}

func NewCloudinaryClient(cfg CloudinaryConfig) StorageClient {
	cld, err := cloudinary.NewFromParams(cfg.Cloud, cfg.ApiKey, cfg.ApiSecret)
	if err != nil {
		logger.Errorf("error init cloudinary client: %s", err.Error())
	}
	return &cloudinaryImpl{
		cld: cld,
	}
}

type CloudinaryConfig struct {
	Cloud     string `env:"CLOUDINARY_CLOUD"`
	ApiKey    string `env:"CLOUDINARY_API_KEY"`
	ApiSecret string `env:"CLOUDINARY_API_SECRET"`
}

type cloudinaryImpl struct {
	cld *cloudinary.Cloudinary
}

func (i *cloudinaryImpl) Save(ctx context.Context, data []byte) (string, error) {
	uploadData := bytes.NewReader(data)
	resp, err := i.cld.Upload.Upload(ctx, uploadData, uploader.UploadParams{})
	if err != nil {
		return "", err
	}

	return resp.URL, nil
}
