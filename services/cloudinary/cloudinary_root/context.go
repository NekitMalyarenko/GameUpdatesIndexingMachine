package cloudinary_root

import (
	"net/url"
	"golang.org/x/net/context"
)

type key int

const cloudinaryKey key = 0

func NewContext(ctx context.Context, uri string) context.Context {
	cURI, err := url.Parse(uri)
	if err != nil {
		return ctx
	}

	service, err := Dial(cURI.String())
	if err != nil {
		return ctx
	}
	return WithCloudinary(ctx, service)
}

func WithCloudinary(ctx context.Context, service *Service) context.Context {
	return context.WithValue(ctx, cloudinaryKey, service)
}

func FromContext(ctx context.Context) (*Service, bool) {
	c, ok := ctx.Value(cloudinaryKey).(*Service)
	return c, ok
}
