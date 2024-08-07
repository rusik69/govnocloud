package node

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
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

	"github.com/pkg/sftp"
	"github.com/rusik69/govnocloud/pkg/types"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
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
		return "nostate"
	case libvirt.DOMAIN_RUNNING:
		return "running"
	case libvirt.DOMAIN_BLOCKED:
		return "blocked"
	case libvirt.DOMAIN_PAUSED:
		return "paused"
	case libvirt.DOMAIN_SHUTDOWN:
		return "shutdown"
	case libvirt.DOMAIN_SHUTOFF:
		return "shutoff"
	case libvirt.DOMAIN_CRASHED:
		return "crashed"
	case libvirt.DOMAIN_PMSUSPENDED:
		return "pmsuspended"
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

// PrepareImage adds the ssh public key to image.
func PrepareImage(image, hostName string) error {
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
	logrus.Println("Waiting for VM to get up")
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

// getKubeConfig gets the kubeconfig.
func getKubeConfig(ip string) (string, error) {
	logrus.Println("Getting kubeconfig")
	key, err := os.ReadFile("/root/.ssh/id_rsa")
	if err != nil {
		return "", err
	}
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return "", err
	}
	sshConfig := &ssh.ClientConfig{
		User: "root",
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	conn, err := ssh.Dial("tcp", net.JoinHostPort(ip, "22"), sshConfig)
	if err != nil {
		return "", err
	}
	client, err := sftp.NewClient(conn)
	if err != nil {
		return "", err
	}
	defer client.Close()
	srcFile, err := client.Open("/etc/rancher/k3s/k3s.yaml")
	if err != nil {
		return "", err
	}
	defer srcFile.Close()
	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(srcFile)
	if err != nil {
		return "", err
	}
	encoded := base64.StdEncoding.EncodeToString(buf.Bytes())
	return encoded, nil
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

// HealthResponse represents the JSON response from the /health endpoint
type HealthResponse struct {
	Status string `json:"status"`
}

// waitForLLM waits for the llm to be up.
func waitForLLM(ip string) error {
	logrus.Println("Waiting for LLM on", ip)
	count := 0
	var healthResponse HealthResponse
	for {
		if count == 600 {
			logrus.Println("LLM wait timeout")
			return errors.New("llm wait timeout")
		}
		resp, err := http.Get("http://" + ip + "/health")
		if err != nil {
			logrus.Println("Error getting response:", err)
			count++
			time.Sleep(1 * time.Second)
			continue
		} else {
			err = json.NewDecoder(resp.Body).Decode(&healthResponse)
			if err != nil {
				logrus.Println("Error decoding response:", err)
				time.Sleep(1 * time.Second)
				count++
				continue
			}
			if healthResponse.Status == "healthy" {
				logrus.Println("LLM is ready.")
				break
			} else {
				logrus.Println("Waiting for LLM...")
				time.Sleep(1 * time.Second)
				count++
			}
		}
	}
	return nil
}
