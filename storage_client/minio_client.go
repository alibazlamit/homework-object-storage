package storage_client

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/alibazlamit/homework-object-storage/config"
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

func (msc *MinioStorageClient) GetStorageClient(instance *models.StorageInstanceInfo) StorageClient {
	msc = loadClient(*instance)
	createBucket(msc)
	return msc
}

func loadClient(inst models.StorageInstanceInfo) *MinioStorageClient {
	client, err := minio.New(inst.Host, &minio.Options{
		Creds:  credentials.NewStaticV4(inst.User, inst.Password, ""),
		Secure: false,
	})
	if err != nil {
		log.Println(err)
	}

	return &MinioStorageClient{minioClient: client, instance: &inst}
}

func createBucket(msc *MinioStorageClient) {
	exists, err := msc.minioClient.BucketExists(context.Background(), config.AppConfig.MainBucket)
	if err != nil {
		log.Println(err)
	}

	if !exists {
		err = msc.minioClient.MakeBucket(context.Background(), config.AppConfig.MainBucket, minio.MakeBucketOptions{})
		if err != nil {
			log.Println(err)
			return
		}
		fmt.Printf("Created bucket: %s\n", config.AppConfig.MainBucket)
	} else {
		fmt.Printf("Bucket already exists: %s\n", config.AppConfig.MainBucket)
	}
}

func (msc *MinioStorageClient) GetObject(ctx context.Context, objectId string) (body []byte, err error) {

	obj, err := msc.minioClient.GetObject(ctx, config.AppConfig.MainBucket, objectId, minio.GetObjectOptions{})
	if err != nil {
		log.Printf("Error getting the object %v", err)
		return nil, err
	}

	return convertObjToData(obj)
}

func (msc *MinioStorageClient) UpdateObject(ctx context.Context, objId string, body []byte) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := msc.minioClient.PutObject(ctx, config.AppConfig.MainBucket, objId, getFileContent(body),
		int64(len(body)), minio.PutObjectOptions{})
	if err != nil {
		log.Printf("Error updating the object %v", err)
		return err
	}
	return nil
}

func getFileContent(content []byte) io.Reader {
	reader := bytes.NewReader(content)
	return reader

}

func convertObjToData(obj *minio.Object) ([]byte, error) {
	var buffer bytes.Buffer
	_, err := io.Copy(&buffer, obj)
	if err != nil {
		log.Printf("Error reading object data: %v", err)
		return nil, err
	}

	return buffer.Bytes(), nil
}
