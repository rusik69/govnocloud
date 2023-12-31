package master

import (
	"encoding/json"

	"github.com/gin-gonic/gin"
	"github.com/rusik69/govnocloud/pkg/client"
	"github.com/rusik69/govnocloud/pkg/types"
	"github.com/sirupsen/logrus"
)

// CreateContainerHandler handles the create container request.
func CreateContainerHandler(c *gin.Context) {
	body := c.Request.Body
	defer body.Close()
	var tempContainer types.Container
	if err := c.ShouldBindJSON(&tempContainer); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		logrus.Error(err.Error())
		return
	}
	if tempContainer.Name == "" || tempContainer.Image == "" {
		c.JSON(400, gin.H{"error": "name or image is empty"})
		logrus.Error("name or image is empty")
		return
	}
	logrus.Println("Creating container", tempContainer)
	containerInfoString, err := ETCDGet("/containers/" + tempContainer.Name)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		logrus.Error(err.Error())
		return
	}
	if containerInfoString != "" {
		c.JSON(400, gin.H{"error": "container with this id already exists"})
		logrus.Error("container with this id already exists")
		return
	}
	var newContainerID string
	created := false
	var newContainer types.Container
	for _, node := range types.MasterEnvInstance.Nodes {
		newContainerID, err = client.CreateContainer(node.Host, node.Port, tempContainer.Name, tempContainer.Image)
		if err != nil {
			logrus.Error(err.Error())
			continue
		}
		newContainer.ID = newContainerID
		newContainer.Host = node.Host
		created = true
		break
	}
	if !created {
		c.JSON(500, gin.H{"error": "can't create container"})
		logrus.Error("can't create container")
		return
	}
	newContainer.Committed = true
	newContainerString, err := json.Marshal(newContainer)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		logrus.Error(err.Error())
		return
	}
	err = ETCDPut("/containers/"+tempContainer.Name, string(newContainerString))
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		logrus.Error(err.Error())
		return
	}
	c.JSON(200, newContainer)
}

// DeleteContainerHandler handles the delete container request.
func DeleteContainerHandler(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(400, gin.H{"error": "id is empty"})
		return
	}
	logrus.Printf("Deleting container with id %s\n", id)
	vmInfoString, err := ETCDGet("/containers/" + id)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		logrus.Error(err.Error())
		return
	}
	if vmInfoString == "" {
		c.JSON(400, gin.H{"error": "container with this id doesn't exist"})
		logrus.Error("container with this id doesn't exist")
		return
	}
	var tempContainer types.Container
	err = json.Unmarshal([]byte(vmInfoString), &tempContainer)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		logrus.Error(err.Error())
		return
	}
	deleted := false
	for _, node := range types.MasterEnvInstance.Nodes {
		if node.Host == tempContainer.Host {
			err = client.DeleteContainer(node.Host, node.Port, tempContainer.ID)
			if err != nil {
				logrus.Error(err.Error())
				break
			}
			deleted = true
		}
	}
	if !deleted {
		c.JSON(500, gin.H{"error": "can't delete container"})
		logrus.Error("can't delete container")
		return
	}
	err = ETCDDelete("/containers/" + id)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		logrus.Error(err.Error())
		return
	}
	c.JSON(200, gin.H{"message": "container deleted"})
}

// ListContainerHandler handles the list container request.
func ListContainerHandler(c *gin.Context) {
	containers := []types.Container{}
	for _, node := range types.MasterEnvInstance.Nodes {
		tempContainers, err := client.ListContainers(node.Host, node.Port)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			logrus.Error(err.Error())
			return
		}
		for _, container := range tempContainers {
			containers = append(containers, container)
		}
	}
	c.JSON(200, containers)
}

// GetContainerHandler handles the get container request.
func GetContainerHandler(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(400, gin.H{"error": "id is empty"})
		return
	}
	logrus.Printf("Getting container with id %s\n", id)
	containerInfoString, err := ETCDGet("/containers/" + id)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		logrus.Error(err.Error())
		return
	}
	if containerInfoString == "" {
		c.JSON(400, gin.H{"error": "container with this id doesn't exist"})
		logrus.Error("container with this id doesn't exist")
		return
	}
	var container types.Container
	err = json.Unmarshal([]byte(containerInfoString), &container)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		logrus.Error(err.Error())
		return
	}
	c.JSON(200, container)
}

// StartContainerHandler handles the start container request.
func StartContainerHandler(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(400, gin.H{"error": "id is empty"})
		return
	}
	logrus.Printf("Starting container with id %s\n", id)
	containerInfoString, err := ETCDGet("/containers/" + id)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		logrus.Error(err.Error())
		return
	}
	if containerInfoString == "" {
		c.JSON(400, gin.H{"error": "container with this id doesn't exist"})
		logrus.Error("container with this id doesn't exist")
		return
	}
	var container types.Container
	err = json.Unmarshal([]byte(containerInfoString), &container)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		logrus.Error(err.Error())
		return
	}
	started := false
	for _, node := range types.MasterEnvInstance.Nodes {
		if node.Host == container.Host {
			err = client.StartContainer(node.Host, node.Port, container.ID)
			if err != nil {
				logrus.Error(err.Error())
				break
			}
			started = true
		}
	}
	if !started {
		c.JSON(500, gin.H{"error": "can't start container"})
		logrus.Error("can't start container")
		return
	}
	container.State = "running"
	containerString, err := json.Marshal(container)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		logrus.Error(err.Error())
		return
	}
	err = ETCDPut("/containers/"+id, string(containerString))
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		logrus.Error(err.Error())
		return
	}
	c.JSON(200, container)
}

// StopContainerHandler handles the stop container request.
func StopContainerHandler(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(400, gin.H{"error": "id is empty"})
		return
	}
	logrus.Printf("Stopping container with id %s\n", id)
	containerInfoString, err := ETCDGet("/containers/" + id)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		logrus.Error(err.Error())
		return
	}
	if containerInfoString == "" {
		c.JSON(400, gin.H{"error": "container with this id doesn't exist"})
		logrus.Error("container with this id doesn't exist")
		return
	}
	var container types.Container
	err = json.Unmarshal([]byte(containerInfoString), &container)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		logrus.Error(err.Error())
		return
	}
	stopped := false
	for _, node := range types.MasterEnvInstance.Nodes {
		if node.Host == container.Host {
			err = client.StopContainer(node.Host, node.Port, container.ID)
			if err != nil {
				logrus.Error(err.Error())
				break
			}
			stopped = true
		}
	}
	if !stopped {
		c.JSON(500, gin.H{"error": "can't stop container"})
		logrus.Error("can't stop container")
		return
	}
	container.State = "stopped"
	containerString, err := json.Marshal(container)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		logrus.Error(err.Error())
		return
	}
	err = ETCDPut("/containers/"+id, string(containerString))
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		logrus.Error(err.Error())
		return
	}
	c.JSON(200, container)
}
