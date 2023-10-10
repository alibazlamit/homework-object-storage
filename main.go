package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/alibazlamit/homework-object-storage/loadbalancer"
	"github.com/alibazlamit/homework-object-storage/storage_client"
	"github.com/gorilla/mux"
)

var storageClientInstance storage_client.StorageClient
var lb loadbalancer.Loadbalancer

func main() {
	//handle routes here so far we only have two
	router := mux.NewRouter()
	router.HandleFunc("/object/{id:[a-zA-Z0-9]+}", getObject).Methods("GET")
	router.HandleFunc("/object/{id:[a-zA-Z0-9]+}", putObject).Methods("PUT")

	//define the instances of loadbalancer and storage client services
	lb = loadbalancer.NewMinioLoadBalancer()
	storageClientInstance = storage_client.NewMinioStorageClient()

	// run this in another routine to not block API
	go lb.WatchContainerChanges()
	//run the server
	http.Handle("/", router)
	http.ListenAndServe(":3000", nil)
}

func getObject(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	params := mux.Vars(r)
	objectID := params["id"]
	//based on id get the lb'd instance
	storageInstance := lb.SelectStorageInstance(objectID)
	storageClientInstance = storageClientInstance.GetStorageClient(storageInstance)

	body, err := storageClientInstance.GetObject(ctx, objectID)
	if err != nil {
		switch {
		case ctx.Err() == context.DeadlineExceeded:
			// Handle timeout error
			handleError(w, err, http.StatusGatewayTimeout, "Request timed out")
		case strings.Contains(err.Error(), "The specified key does not exist"):
			// Handle not found error
			handleError(w, err, http.StatusNotFound, "Object not found")
		default:
			// Handle other errors
			handleError(w, err, http.StatusInternalServerError, "Internal Server Error")
		}
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(body)

}

func putObject(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	params := mux.Vars(r)
	objectID := params["id"]

	body, err := io.ReadAll(r.Body)
	if err != nil {
		handleError(w, err, http.StatusBadRequest, "Error reading request body")
		return
	}
	// based on id get the lb'd instance
	storageInstance := lb.SelectStorageInstance(objectID)
	storageClientInstance = storageClientInstance.GetStorageClient(storageInstance)

	err = storageClientInstance.UpdateObject(ctx, objectID, body)
	if err != nil {
		switch {
		case ctx.Err() == context.DeadlineExceeded:
			handleError(w, err, http.StatusGatewayTimeout, "Request timed out")
		default:
			handleError(w, err, http.StatusInternalServerError, "Error updating object")
		}
		return
	}
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintln(w, "Object created or updated successfully")
}

func handleError(w http.ResponseWriter, err error, statusCode int, message string) {
	w.WriteHeader(statusCode)
	fmt.Fprintf(w, "%s: %v", message, err)
}
