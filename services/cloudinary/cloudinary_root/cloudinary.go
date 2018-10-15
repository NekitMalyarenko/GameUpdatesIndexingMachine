package cloudinary_root

import (
	"io"
	"golang.org/x/net/context"
)

func UploadStaticImage(ctx context.Context, fileName string, data io.Reader) error {
	c, _ := FromContext(ctx)
	_, err := c.UploadStaticImage(fileName, data, "")
	return err
}

func UploadStaticRaw(ctx context.Context, fileName string, data io.Reader) error {
	c, _ := FromContext(ctx)
	_, err := c.UploadStaticRaw(fileName, data, "")
	return err
}

func Resources(ctx context.Context) ([]*Resource, error) {
	c, _ := FromContext(ctx)
	return c.Resources(ImageType)
}

func ResourceURL(ctx context.Context, fileName string) string {
	c, _ := FromContext(ctx)
	return c.Url(fileName, ImageType)
}

func DeleteStaticImage(ctx context.Context, fileName string) error {
	c, _ := FromContext(ctx)
	return c.Delete(fileName, "", ImageType)
}
