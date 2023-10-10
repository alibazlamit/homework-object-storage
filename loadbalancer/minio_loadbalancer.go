package loadbalancer

import (
	"context"
	"errors"
	"fmt"
	"hash/fnv"
	"sort"
	"strings"

	"github.com/alibazlamit/homework-object-storage/config"
	"github.com/alibazlamit/homework-object-storage/logger"
	"github.com/alibazlamit/homework-object-storage/models"
	"github.com/docker/docker/api/types"
	dockerEvents "github.com/docker/docker/api/types/events"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
)

var minioInstances = map[string]models.StorageInstanceInfo{}

const keyName string = "amazin-object-storage"

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
	// list all containers
	containers, err := dc.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		logger.Log.Errorf("Failed to list containers: %v", err)
	}

	// iterate through containers and check if name has our prefix its a minio instance. Get IP user and pass
	for _, cnt := range containers {
		for _, cntName := range cnt.Names {
			// check if containers name contains keyName value
			if strings.Contains(cntName, keyName) {
				inspect, err := dc.ContainerInspect(context.Background(), cnt.ID)
				if err != nil {
					logger.Log.Errorf("Failed to inspect container %s: %v", cnt.ID, err)
					break
				}

				// if instance is not running no need to add it
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
	if len(minioInstances) == 0 {
		logger.Log.Warnf("Discovered %d instances, please try to bring some instances up", len(minioInstances))
	}
	logger.Log.Infof("Discovered %d instances", len(minioInstances))
	return minioInstances
}

func (lb MinioLoadbalancer) SelectStorageInstance(objectID string) (*models.StorageInstanceInfo, error) {
	// check if no instances available handle that case
	if len(minioInstances) > 0 {
		//consistent hashing used below to load balance
		hashValue := hash(objectID)

		instanceIDs := make([]string, 0, len(minioInstances))

		// collect instanceIds
		for k := range minioInstances {
			instanceIDs = append(instanceIDs, k)
		}

		//sort to be used in an order for consistent hashing
		sort.Slice(instanceIDs, func(i, j int) bool {
			return hash(instanceIDs[i]) < hash(instanceIDs[j])
		})

		selectedInstanceID := ""
		// go through sorted ids and find a one instance who has greater hash value than hashed id
		for _, id := range instanceIDs {
			if hash(id) > hashValue {
				selectedInstanceID = id
				break
			}
		}

		// default case if no hashed ID is bigger than objectIds hash
		if selectedInstanceID == "" {
			selectedInstanceID = instanceIDs[0]
			logger.Log.Warnf("consistent hashing could not find instance, selected instance %v", selectedInstanceID)
		}
		selectedInstance := minioInstances[selectedInstanceID]
		return &selectedInstance, nil
	}
	return nil, errors.New("no minio instances available")
}

func (lb MinioLoadbalancer) WatchContainerChanges() {
	dc, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}
	defer dc.Close()
	options := types.EventsOptions{
		Filters: filters.NewArgs(),
	}

	// using docker get the events stream and look for any actions that might need to trigger the rediscovery of instances
	events, errCh := dc.Events(context.Background(), options)
	defer func() {
		if err := dc.Close(); err != nil {
			logger.Log.Error(err)
		}
	}()

	triggerActions := []string{"start", "oom", "stop"}
	for {
		select {
		//events channel check for any action of type container below
		case event := <-events:
			if event.Type == dockerEvents.ContainerEventType {
				logger.Log.Infof("Event in container %v, of actor id %v", event.Action, event.Actor.ID)
				for _, action := range triggerActions {
					if event.Action == action {
						lb.DiscoverStorageInstances()
						break
					}
				}
			}
		//error channel
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
