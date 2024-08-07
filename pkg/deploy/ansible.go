package deploy

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"

	"github.com/sirupsen/logrus"
)

// GenerateAnsibleConfig generates ansible config
func GenerateAnsibleConfig(nodes, osds []string, master, invFile string) error {
	file, err := os.Create(invFile)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.WriteString("[all]\n")
	if err != nil {
		return err
	}
	for _, node := range nodes {
		_, err = file.WriteString(node + " ansible_become=true\n")
		if err != nil {
			return err
		}
	}
	_, err = file.WriteString(master + " ansible_become=true\n")
	if err != nil {
		return err
	}
	_, err = file.WriteString("[nodes]\n")
	if err != nil {
		return err
	}
	for _, node := range nodes {
		_, err = file.WriteString(node + " ansible_become=true\n")
		if err != nil {
			return err
		}
	}
	_, err = file.WriteString("[osds]\n")
	if err != nil {
		return err
	}
	for _, osd := range osds {
		_, err = file.WriteString(osd + " ansible_become=true\n")
		if err != nil {
			return err
		}
	}
	_, err = file.WriteString("[masters]\n")
	if err != nil {
		return err
	}
	_, err = file.WriteString(master + " ansible_become=true\n")
	if err != nil {
		return err
	}
	file.Sync()
	fileContent, err := os.ReadFile(invFile)
	if err != nil {
		return err
	}
	logrus.Println(string(fileContent))
	return nil
}

// RunAnsible runs ansible
func RunAnsible(invFile, user, key string) error {
	cmd := exec.Command("ansible-playbook", "-i", invFile, "-u", user, "--private-key="+key, "deployments/ansible/main.yml")
	cmd.Env = append(cmd.Env, "ANSIBLE_HOST_KEY_CHECKING=False")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		logrus.Panic(err)
	}
	cmd.Stderr = cmd.Stdout
	if err := cmd.Start(); err != nil {
		return err
	}
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}
	err = cmd.Wait()
	return err
}
