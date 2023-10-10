package storage_client

import (
	"context"

	"github.com/alibazlamit/homework-object-storage/models"
)

type StorageClient interface {
	GetObject(ctx context.Context, objectId string) (body []byte, err error)
	UpdateObject(ctx context.Context, objId string, body []byte) error
	GetStorageClient(instance *models.StorageInstanceInfo) StorageClient
}
