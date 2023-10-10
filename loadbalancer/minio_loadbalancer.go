package loadbalancer

import (
	"context"
	"fmt"
	"hash/fnv"
	"sort"
	"strings"

	"github.com/alibazlamit/homework-object-storage/config"
	"github.com/alibazlamit/homework-object-storage/logger"
	"github.com/alibazlamit/homework-object-storage/models"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
)

var minioInstances = map[string]models.StorageInstanceInfo{}
var keyName string = "amazin-object-storage"

type MinioLoadbalancer struct {
}

func NewMinioLoadBalancer() *MinioLoadbalancer {
	return &MinioLoadbalancer{}
}

func (lb MinioLoadbalancer) DiscoverStorageInstances() map[string]models.StorageInstanceInfo {
	minioInstances = map[string]models.StorageInstanceInfo{}
	dc, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		logger.Log.Errorf("Error initializing Minio client: %v", err)
	}
	containers, err := dc.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		logger.Log.Errorf("Failed to list containers: %v", err)
	}

	// Iterate through containers and check if name has our prefix its a minio instance. Get IP user and pass
	for _, cnt := range containers {
		for _, cntName := range cnt.Names {
			if strings.Contains(cntName, keyName) {
				inspect, err := dc.ContainerInspect(context.Background(), cnt.ID)
				if err != nil {
					logger.Log.Errorf("Failed to inspect container %s: %v", cnt.ID, err)
					break
				}

				if inspect.State.Running {
					user, pass := getUserAndPassFromEnv(inspect.Config.Env)
					host := inspect.NetworkSettings.Networks[string(inspect.HostConfig.NetworkMode)].IPAddress
					hostWithPort := fmt.Sprintf("%s:%s", host, config.AppConfig.StoragePort)
					minioInstances[cnt.ID] = models.StorageInstanceInfo{
						Host:     hostWithPort,
						User:     user,
						Password: pass,
					}
					break
				}
			}

		}
	}
	logger.Log.Infof("Discovered %d instances", len(minioInstances))
	return minioInstances
}

func (lb MinioLoadbalancer) SelectStorageInstance(objectID string) *models.StorageInstanceInfo {
	hashValue := hash(objectID)

	instanceIDs := make([]string, 0, len(minioInstances))

	for k := range minioInstances {
		instanceIDs = append(instanceIDs, k)
	}

	sort.Slice(instanceIDs, func(i, j int) bool {
		return hash(instanceIDs[i]) < hash(instanceIDs[j])
	})

	selectedInstanceID := ""
	for _, id := range instanceIDs {
		if hash(id) > hashValue {
			selectedInstanceID = id
			break
		}
	}

	if selectedInstanceID == "" {
		selectedInstanceID = instanceIDs[0]
		logger.Log.Warnf("consistent hashing could not find instance, selected instance %v", selectedInstanceID)
	}
	selectedInstance := minioInstances[selectedInstanceID]
	return &selectedInstance
}

func (lb MinioLoadbalancer) WatchContainerChanges() {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	options := types.EventsOptions{
		Filters: filters.NewArgs(),
	}

	events, errCh := cli.Events(context.Background(), options)

	for {
		select {
		case event := <-events:
			if event.Action == "start" || event.Action == "die" {
				lb.DiscoverStorageInstances()
			}
		case err := <-errCh:
			if err != nil {
				logger.Log.Error(err)
			}
		}
	}
}

func getUserAndPassFromEnv(envVars []string) (user string, pass string) {
	for _, envVar := range envVars {
		parts := strings.SplitN(envVar, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key, value := parts[0], parts[1]
		if key == "MINIO_ROOT_USER" {
			user = value
		} else if key == "MINIO_ROOT_PASSWORD" {
			pass = value
		}
	}
	return user, pass
}

func hash(s string) int {
	h := fnv.New32a()
	h.Write([]byte(s))
	return int(h.Sum32())
}
