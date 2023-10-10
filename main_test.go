package main

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alibazlamit/homework-object-storage/config"
	"github.com/alibazlamit/homework-object-storage/loadbalancer"
	"github.com/alibazlamit/homework-object-storage/storage_client"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func init() {
	config.AppConfig.SetMockEnv(config.Config{MainBucket: "test", StoragePort: "9000", GeneralTimeout: 1})
}

func TestGetObject(t *testing.T) {
	lb = &loadbalancer.MockLoadBalancer{}
	storageClientInstance = &storage_client.MockStorageClient{}

	req, err := http.NewRequest("GET", "/object/abcdefg12345", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	router := mux.NewRouter()
	router.HandleFunc("/object/{id:[a-zA-Z0-9]+}", getObject).Methods("GET")

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, bytes.NewBuffer(storage_client.MockObjects[0].Data), rr.Body)
}

func TestGetObjectNotFound(t *testing.T) {
	lb = &loadbalancer.MockLoadBalancer{}
	storageClientInstance = &storage_client.MockStorageClient{}

	req, err := http.NewRequest("GET", "/object/test", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	router := mux.NewRouter()
	router.HandleFunc("/object/{id:[a-zA-Z0-9]+}", getObject).Methods("GET")

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
	assert.Equal(t, "Object not found: The specified key does not exist", rr.Body.String())
}

func TestGetObjectTimeOut(t *testing.T) {
	lb = &loadbalancer.MockLoadBalancer{}
	storageClientInstance = &storage_client.MockStorageClient{}

	req, err := http.NewRequest("GET", "/object/timeout", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	router := mux.NewRouter()
	router.HandleFunc("/object/{id:[a-zA-Z0-9]+}", getObject).Methods("GET")

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusRequestTimeout, rr.Code)
	assert.Equal(t, "Request timed out: service not up", rr.Body.String())
}

func TestPutObject(t *testing.T) {
	lb = &loadbalancer.MockLoadBalancer{}
	storageClientInstance = &storage_client.MockStorageClient{}

	req, err := http.NewRequest("PUT", "/object/test", bytes.NewBuffer([]byte("test")))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	router := mux.NewRouter()
	router.HandleFunc("/object/{id:[a-zA-Z0-9]+}", putObject).Methods("PUT")

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)
	assert.Equal(t, "Object created or updated successfully\n", rr.Body.String())
}

func TestPutObjectBadBody(t *testing.T) {
	lb = &loadbalancer.MockLoadBalancer{}
	storageClientInstance = &storage_client.MockStorageClient{}
	req, err := http.NewRequest("PUT", "/object/test", &ErrorReader{})
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	router := mux.NewRouter()
	router.HandleFunc("/object/{id:[a-zA-Z0-9]+}", putObject).Methods("PUT")

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Equal(t, "Error reading request body: Custom error: Unable to read", rr.Body.String())
}

func TestPutObjectTimeout(t *testing.T) {

	lb = &loadbalancer.MockLoadBalancer{}
	storageClientInstance = &storage_client.MockStorageClient{}
	req, err := http.NewRequest("PUT", "/object/timeout", bytes.NewBuffer([]byte("test")))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	router := mux.NewRouter()
	router.HandleFunc("/object/{id:[a-zA-Z0-9]+}", putObject).Methods("PUT")

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusRequestTimeout, rr.Code)
	assert.Equal(t, "Request timed out: service not up", rr.Body.String())
}

type ErrorReader struct{}

func (er *ErrorReader) Read(p []byte) (n int, err error) {
	return 0, fmt.Errorf("Custom error: Unable to read")
}
