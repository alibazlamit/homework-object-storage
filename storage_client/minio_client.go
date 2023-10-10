package storage_client

import (
	"context"

	"github.com/alibazlamit/homework-object-storage/config"
	"github.com/alibazlamit/homework-object-storage/logger"
	"github.com/alibazlamit/homework-object-storage/models"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinioStorageClient struct {
	minioClient *minio.Client
	instance    *models.StorageInstanceInfo
}

func NewMinioStorageClient() *MinioStorageClient {
	return &MinioStorageClient{}
}

// build a minio client based on the provided instance
func (msc *MinioStorageClient) GetStorageClient(instance *models.StorageInstanceInfo) StorageClient {
	return msc.loadClient(*instance)
}

func (msc *MinioStorageClient) loadClient(inst models.StorageInstanceInfo) *MinioStorageClient {
	client, err := minio.New(inst.Host, &minio.Options{
		Creds:  credentials.NewStaticV4(inst.User, inst.Password, ""),
		Secure: false,
	})
	if err != nil {
		logger.Log.Error(err)
	}

	// a check to keep the last instance if its the same provider
	if msc == nil || msc.minioClient == nil {
		msc = &MinioStorageClient{minioClient: client, instance: &inst}
		createBucket(msc)
	} else if msc.instance.Host != inst.Host {
		msc = &MinioStorageClient{minioClient: client, instance: &inst}
		createBucket(msc)
	}

	return msc
}

// create the main bucket if doesn't exist
func createBucket(msc *MinioStorageClient) {
	exists, err := msc.minioClient.BucketExists(context.Background(), config.AppConfig.MainBucket)
	if err != nil {
		logger.Log.Error(err)
	}

	if !exists {
		err = msc.minioClient.MakeBucket(context.Background(), config.AppConfig.MainBucket, minio.MakeBucketOptions{})
		if err != nil {
			logger.Log.Error(err)
			return
		}
		logger.Log.Infof("Created bucket: %s\n", config.AppConfig.MainBucket)
	}
}

// get object providing an object id
func (msc *MinioStorageClient) GetObject(ctx context.Context, objectId string) (body []byte, err error) {

	obj, err := msc.minioClient.GetObject(ctx, config.AppConfig.MainBucket, objectId, minio.GetObjectOptions{})

	if err != nil {
		logger.Log.Errorf("Error getting the object %v", err)
		return nil, err
	}
	defer obj.Close()

	return convertObjToData(obj)
}

// put object will create or update the given object id with body
func (msc *MinioStorageClient) UpdateObject(ctx context.Context, objId string, body []byte) error {

	_, err := msc.minioClient.PutObject(ctx, config.AppConfig.MainBucket, objId, getFileContent(body),
		int64(len(body)), minio.PutObjectOptions{})
	if err != nil {
		logger.Log.Errorf("Error updating the object %v", err)
		return err
	}
	return nil
}
