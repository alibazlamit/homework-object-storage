package loadbalancer

import "github.com/alibazlamit/homework-object-storage/models"

type MockLoadBalancer struct{}

func (m *MockLoadBalancer) SelectStorageInstance(objectID string) *models.StorageInstanceInfo {
	return &models.StorageInstanceInfo{
		Host:     "localhost",
		User:     "test",
		Password: "test",
	}
}

func (m *MockLoadBalancer) DiscoverStorageInstances() map[string]models.StorageInstanceInfo {
	return nil
}

func (m *MockLoadBalancer) WatchContainerChanges() {}
