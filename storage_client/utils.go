package storage_client

import (
	"bytes"
	"io"
	"strings"

	"github.com/alibazlamit/homework-object-storage/logger"
	"github.com/minio/minio-go/v7"
)

// converts bytes to io.Reader
func getFileContent(content []byte) io.Reader {
	reader := bytes.NewReader(content)
	return reader
}

// coverts a minio object to data
func convertObjToData(obj *minio.Object) ([]byte, error) {
	var buffer bytes.Buffer
	_, err := io.Copy(&buffer, obj)
	if err != nil {
		if strings.Contains(err.Error(), "The specified key does not exist") {
			return nil, err
		}

		logger.Log.Errorf("Error reading object data: %v", err)
		return nil, err
	}

	return buffer.Bytes(), nil
}
