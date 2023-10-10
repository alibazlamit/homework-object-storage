package storage_client

import (
	"context"
	"errors"
	"time"

	"github.com/alibazlamit/homework-object-storage/models"
)

var MockObjects = []models.MockObject{
	{ObjectId: "abcdefg12345", Data: []byte("find me please")},
}

type MockStorageClient struct{}

func (m *MockStorageClient) GetObject(ctx context.Context, objectId string) (body []byte, err error) {
	if objectId == "timeout" {
		time.Sleep(100 * time.Millisecond)
		return nil, errors.New("service not up")
	}
	for _, obj := range MockObjects {
		if obj.ObjectId == objectId {
			return obj.Data, nil
		}
	}
	return nil, errors.New("The specified key does not exist")
}

func (m *MockStorageClient) UpdateObject(ctx context.Context, objId string, body []byte) error {
	if objId == "timeout" {
		time.Sleep(100 * time.Millisecond)
		return errors.New("service not up")
	}
	return nil
}

func (m *MockStorageClient) GetStorageClient(instance *models.StorageInstanceInfo) StorageClient {
	return m
}
