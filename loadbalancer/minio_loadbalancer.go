package loadbalancer

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/alibazlamit/homework-object-storage/config"
	"github.com/alibazlamit/homework-object-storage/models"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

var minioInstances = map[string]models.StorageInstanceInfo{}
var name string = "amazin-object-storage"
var hackID string
var loadbalancer *MinioLoadbalancer

type MinioLoadbalancer struct {
}

func NewMinioLoadBalancer() *MinioLoadbalancer {
	return &MinioLoadbalancer{}
}

func (lb MinioLoadbalancer) DiscoverStorageInstances() map[string]models.StorageInstanceInfo {
	dc, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Printf("Error initializing Minio client: %v", err)
	}
	containers, err := dc.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		log.Fatalf("Failed to list containers: %v", err)
	}

	// Iterate through containers and check if name has our prefix its a minio instance. Get IP user and pass
	for _, cnt := range containers {
		for _, cntName := range cnt.Names {
			if strings.Contains(cntName, name) {
				inspect, err := dc.ContainerInspect(context.Background(), cnt.ID)
				if err != nil {
					log.Printf("Failed to inspect container %s: %v", cnt.ID, err)
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
					hackID = cnt.ID
					break
				}
			}

		}
	}
	return minioInstances
}

func (lb MinioLoadbalancer) SelectStorageInstance(objectID string) *models.StorageInstanceInfo {
	// lastDigit, err := strconv.Atoi(string(objectID[len(objectID)-1]))
	// if err != nil {
	// 	log.Printf("Error parsing objectID: %v", err)
	// }
	// instanceIndex := lastDigit % len(minioInstances)
	selectedInstance := minioInstances[hackID]
	return &selectedInstance
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
