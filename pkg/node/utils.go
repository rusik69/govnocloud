package node

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/rusik69/govnocloud/pkg/types"
	"github.com/sirupsen/logrus"
	"libvirt.org/go/libvirt"
)

// resizeImage resizes the image.
func resizeImage(image string, flavor types.VMFlavor) error {
	imgInfo, err := exec.Command("qemu-img", "info", image).CombinedOutput()
	if err != nil {
		return err
	}
	lines := strings.Split(string(imgInfo), "\n")
	re := regexp.MustCompile(`^virtual size: (\d+)`)
	var virtualSize string
	for _, line := range lines {
		if re.MatchString(line) {
			matches := re.FindStringSubmatch(line)
			virtualSize = matches[1]
			break
		}
	}
	size, err := strconv.Atoi(virtualSize)
	if err != nil {
		return err
	}
	var cmdStrings []string
	if size < int(flavor.Disk) {
		cmdStrings = []string{"resize", image, strconv.Itoa(int(flavor.Disk)) + "G"}
	} else {
		cmdStrings = []string{"resize", "--shrink", image, strconv.Itoa(int(flavor.Disk)) + "G"}
	}
	cmd := exec.Command("qemu-img", cmdStrings...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		logrus.Println(string(output))
		return err
	}
	return nil
}

// ParseState parses the state of the vm.
func ParseState(state libvirt.DomainState) string {
	switch state {
	case libvirt.DOMAIN_NOSTATE:
		return "NOSTATE"
	case libvirt.DOMAIN_RUNNING:
		return "RUNNING"
	case libvirt.DOMAIN_BLOCKED:
		return "BLOCKED"
	case libvirt.DOMAIN_PAUSED:
		return "PAUSED"
	case libvirt.DOMAIN_SHUTDOWN:
		return "SHUTDOWN"
	case libvirt.DOMAIN_SHUTOFF:
		return "SHUTOFF"
	case libvirt.DOMAIN_CRASHED:
		return "CRASHED"
	case libvirt.DOMAIN_PMSUSPENDED:
		return "PMSUSPENDED"
	}
	return ""
}

// DownloadFile downloads the file.
func DownloadFile(url string, dir string) error {
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return err
	}
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	fileName := path.Base(url)
	filePath := path.Join(dir, fileName)
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()
	bufferSize := 4096
	buffer := make([]byte, bufferSize)
	_, err = io.CopyBuffer(file, resp.Body, buffer)
	if err != nil {
		return err
	}
	return err
}

// CreateSSHKey creates the ssh key.
func CreateSSHKey() error {
	// check if file exists
	fileName := "/root/.ssh/id_rsa"
	if _, err := os.Stat(fileName); err == nil {
		return nil
	}
	logrus.Println("Creating ssh key", fileName)
	cmd := exec.Command("ssh-keygen", "-f", fileName, "-t", "rsa", "-N", "")
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

// AddSSHPublicKey adds the ssh public key to image.
func AddSSHPublicKey(image, hostName string) error {
	logrus.Println("Adding ssh public key to", image)
	cmdSlice := []string{"--no-selinux-relabel", "-a", image, "--mkdir", "/root/.ssh", "--root-password", "password:password", "--password", "password:password", "--network", "--ssh-inject", "root:file:/root/.ssh/id_rsa.pub", "--firstboot-command", "dhclient;dpkg-reconfigure openssh-server;systemctl restart sshd; growpart /dev/vda 1; resize2fs /dev/vda1; hostnamectl set-hostname " + hostName}
	mkdirCmd := exec.Command("virt-customize", cmdSlice...)
	output, err := mkdirCmd.CombinedOutput()
	logrus.Println(string(output))
	if err != nil {
		return err
	}
	return nil
}

// wait for the vm to be up
func waitForVMUp(domain *libvirt.Domain) (string, error) {
	logrus.Println("Waiting for VM to be up")
	count := 0
	for {
		if count == 120 {
			return "", errors.New("timeout")
		}
		ifaces, err := domain.ListAllInterfaceAddresses(libvirt.DOMAIN_INTERFACE_ADDRESSES_SRC_LEASE)
		if err != nil {
			logrus.Println("Failed to get interface addresses")
			return "", err
		}

		for _, iface := range ifaces {
			for _, addr := range iface.Addrs {
				if addr.Addr != "" {
					fmt.Println("IP address:", addr.Addr)
					return addr.Addr, nil
				}
			}
		}

		// Wait before checking again
		count++
		time.Sleep(1 * time.Second)
	}
}

// wait for ssh connection
func waitForSSH(ip string) error {
	logrus.Println("Waiting for SSH on", ip)
	count := 0
	for {
		if count == 120 {
			return errors.New("timeout")
		}
		conn, err := net.DialTimeout("tcp", net.JoinHostPort(ip, "22"), time.Second)
		if err != nil {
			count++
			time.Sleep(1 * time.Second)
			continue
		} else {
			if conn != nil {
				conn.Close()
				logrus.Println("SSH is up")
				return nil
			}
		}
	}
}

// apply ansible to vm
func applyAnsible(ip, playbook, hostname string) error {
	logrus.Println("Applying ansible to", ip)
	cmd := exec.Command("ansible-playbook", "-u", "root", "-i", ip+",", playbook)
	cmd.Env = append(cmd.Env, "ANSIBLE_HOST_KEY_CHECKING=False")
	cmd.Env = append(cmd.Env, "HOSTNAME="+hostname)
	output, err := cmd.CombinedOutput()
	logrus.Println(string(output))
	if err != nil {
		return err
	}
	return nil

}

// copyFile copies the file.
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()
	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()
	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return err
	}
	return nil
}

// tailscaleRemove removes the vm from tailscale.
func tailscaleRemove(deviceID string) error {
	url := fmt.Sprintf("https://api.tailscale.com/api/v2/device/%s", deviceID)
	req, _ := http.NewRequest("DELETE", url, nil)
	req.Header.Add("Authorization", "Bearer "+types.NodeEnvInstance.TailscaleAccessToken)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	if res.StatusCode != http.StatusOK {
		return errors.New("Failed to remove device " + deviceID + " " + string(bodyBytes))
	}
	return nil
}

type tailscaleDevice struct {
	ID   string `json:"id"`
	IP   string `json:"ip"`
	Name string `json:"name"`
}

type tailscaleDevices struct {
	Devices []tailscaleDevice `json:"devices"`
}

// tailscaleGetDeviceInfo gets the device ip and id from tailscale.
func tailscaleGetDeviceInfo(deviceName string) (string, string, error) {
	url := "https://api.tailscale.com/api/v2/tailnet/-/devices"
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Authorization", "Bearer "+types.NodeEnvInstance.TailscaleAccessToken)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", "", err
	}
	if res.StatusCode != http.StatusOK {
		return "", "", errors.New("Failed to get devices")
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)
	var devices tailscaleDevices
	json.Unmarshal(body, &devices)
	logrus.Println("Devices")
	for _, device := range devices.Devices {
		deviceHostName := strings.Split(device.Name, ".")[0]
		if deviceHostName == deviceName {
			logrus.Println("Device found", device)
			return device.IP, device.ID, nil
		}
	}
	return "", "", errors.New("Device not found")
}
