package gcs

import (
	"context"
	"errors"
	"fmt"

	"cloud.google.com/go/storage"
)

func NewGCSWriter(ctx context.Context, bucket, key string) (*storage.Writer, error) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("storage.NewClient: %v", err)
	}

	wc := client.Bucket(bucket).Object(key).NewWriter(ctx)

	return wc, nil
}

func NewGCSReader(ctx context.Context, bucket, key string) (*storage.Reader, error) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("storage.NewClient: %v", err)
	}

	wc, err := client.Bucket(bucket).Object(key).NewReader(ctx)
	if err != nil {
		return nil, err
	}

	return wc, nil

}

func KeyExists(ctx context.Context, bucket, key string) (bool, error) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return false, fmt.Errorf("storage.NewClient: %v", err)
	}

	_, err = client.Bucket(bucket).Object(key).Attrs(ctx)
	if err != nil {
		if errors.Is(err, storage.ErrObjectNotExist) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}
