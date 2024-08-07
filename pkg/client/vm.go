package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/rusik69/govnocloud/pkg/types"
)

// CreateVM creates a vm.
func CreateVM(host, port, name, image, flavor string) (types.VM, error) {
	vm := types.VM{
		Name:   name,
		Image:  image,
		Flavor: flavor,
	}
	url := "http://" + host + ":" + port + "/api/v1/vms"
	body, err := json.Marshal(vm)
	if err != nil {
		return types.VM{}, err
	}
	client := &http.Client{
		Timeout: 300 * time.Second,
	}
	req, err := http.NewRequestWithContext(context.Background(), "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return types.VM{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return types.VM{}, err
	}
	defer resp.Body.Close()
	bodyText, err := io.ReadAll(resp.Body)
	if err != nil {
		return types.VM{}, err
	}
	if resp.StatusCode != 200 {
		return types.VM{}, errors.New(string(bodyText))
	}
	var newVM types.VM
	err = json.Unmarshal(bodyText, &newVM)
	if err != nil {
		return types.VM{}, err
	}
	return newVM, nil
}

// DeleteVM deletes a vm.
func DeleteVM(host, port, name string) error {
	url := "http://" + host + ":" + port + "/api/v1/vm/" + name
	client := &http.Client{
		Timeout: 60 * time.Second,
	}
	req, err := http.NewRequestWithContext(context.Background(), "DELETE", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
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
	return err
}

// StartVM starts a vm.
func StartVM(host, port, name string) error {
	url := "http://" + host + ":" + port + "/api/v1/vmstart/" + name
	client := &http.Client{
		Timeout: 60 * time.Second,
	}
	req, err := http.NewRequestWithContext(context.Background(), "GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
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
	return err
}

// StopVM stops a vm.
func StopVM(host, port, name string) error {
	url := "http://" + host + ":" + port + "/api/v1/vmstop/" + name
	client := &http.Client{
		Timeout: 60 * time.Second,
	}
	req, err := http.NewRequestWithContext(context.Background(), "GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
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
	return err
}

// ListVMs lists vms.
func ListVMs(host, port string) ([]types.VM, error) {
	url := "http://" + host + ":" + port + "/api/v1/vms"
	client := &http.Client{
		Timeout: 60 * time.Second,
	}
	req, err := http.NewRequestWithContext(context.Background(), "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var vms []types.VM
	err = json.NewDecoder(resp.Body).Decode(&vms)
	return vms, err
}

// GetVM gets a vm.
func GetVM(host, port, name string) (types.VM, error) {
	vm := types.VM{}
	url := "http://" + host + ":" + port + "/api/v1/vm/" + name
	client := &http.Client{
		Timeout: 60 * time.Second,
	}
	req, err := http.NewRequestWithContext(context.Background(), "GET", url, nil)
	if err != nil {
		return vm, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return vm, err
	}
	defer resp.Body.Close()
	err = json.NewDecoder(resp.Body).Decode(&vm)
	return vm, err
}
