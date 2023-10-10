package loadbalancer

import (
	"github.com/alibazlamit/homework-object-storage/models"
)

type Loadbalancer interface {
	SelectStorageInstance(objectID string) *models.StorageInstanceInfo
	DiscoverStorageInstances() map[string]models.StorageInstanceInfo
	WatchContainerChanges()
}
