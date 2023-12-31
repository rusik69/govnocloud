package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/rusik69/govnocloud/pkg/types"
)

// CreateContainer creates a container.
func CreateContainer(host, port, name, image string) (string, error) {
	container := types.Container{
		Name:  name,
		Image: image,
	}
	url := "http://" + host + ":" + port + "/api/v1/container/create"
	body, err := json.Marshal(container)
	if err != nil {
		return "", err
	}
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	bodyText, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != 200 {
		return "", errors.New(string(bodyText))
	}
	err = json.Unmarshal(bodyText, &container)
	if err != nil {
		return "", err
	}
	return container.ID, nil
}

// StartContainer starts a container.
func StartContainer(host, port, id string) error {
	url := "http://" + host + ":" + port + "/api/v1/container/start/" + id
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		bodyText, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		return errors.New(string(bodyText))
	}
	return nil
}

// StopContainer stops a container.
func StopContainer(host, port, id string) error {
	url := "http://" + host + ":" + port + "/api/v1/container/stop/" + id
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		bodyText, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		return errors.New(string(bodyText))
	}
	return nil
}

// ListContainers lists containers.
func ListContainers(host, port string) ([]types.Container, error) {
	url := "http://" + host + ":" + port + "/api/v1/container/list"
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var containers []types.Container
	bodyText, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, errors.New(string(bodyText))
	}
	err = json.Unmarshal(bodyText, &containers)
	if err != nil {
		return nil, err
	}
	return containers, nil
}

// GetContainer gets a container.
func GetContainer(host, port, id string) (types.Container, error) {
	container := types.Container{
		ID: id,
	}
	url := "http://" + host + ":" + port + "/api/v1/container/" + id
	resp, err := http.Get(url)
	if err != nil {
		return container, err
	}
	defer resp.Body.Close()
	bodyText, err := io.ReadAll(resp.Body)
	if err != nil {
		return container, err
	}
	if resp.StatusCode != 200 {
		return container, errors.New(string(bodyText))
	}
	err = json.Unmarshal(bodyText, &container)
	if err != nil {
		return container, err
	}
	return container, nil
}

// DeleteContainer deletes a container.
func DeleteContainer(host, port, id string) error {
	url := "http://" + host + ":" + port + "/api/v1/container/" + id
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		bodyText, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		return errors.New(string(bodyText))
	}
	return nil
}
