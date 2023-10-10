package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
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
	//TODO:: figure out how often we need to update/discover
	lb.DiscoverStorageInstances()

	//run the server
	http.Handle("/", router)
	http.ListenAndServe(":3000", nil)
}

func getObject(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	params := mux.Vars(r)
	objectID := params["id"]
	storageClientInstance = storageClientInstance.GetStorageClient(lb.SelectStorageInstance(objectID))

	body, err := storageClientInstance.GetObject(ctx, objectID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintln(w, "Object not found")
		return
		//TODO: handle different error scenarios to distinguish between 404 and 500
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
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error reading request body: %v", err)
		return
	}

	storageClientInstance = storageClientInstance.GetStorageClient(lb.SelectStorageInstance(objectID))

	err = storageClientInstance.UpdateObject(ctx, objectID, body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error updating object: %v", err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintln(w, "Object created or updated successfully")
}
