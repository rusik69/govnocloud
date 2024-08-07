package main

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/rusik69/govnocloud/pkg/deploy"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var nodes, osds []string
var master, ansibleInventoryFile string
var key, user, nodesString, osdsString string

var rootCmd = &cobra.Command{
	Use:   "deploy",
	Short: "deploy a shitty cloud",
	Long:  `deploy a shitty cloud`,
	Run: func(cmd *cobra.Command, args []string) {
		nodes = strings.Split(nodesString, ",")
		osds = strings.Split(osdsString, ",")
		if len(nodes) == 0 || len(osds) == 0 || master == "" {
			logrus.Println("Nodes, osds and master must be specified")
			os.Exit(1)
		}
		if nodes[0] == "" {
			logrus.Println("Nodes must be specified")
			os.Exit(1)
		}
		if osds[0] == "" {
			logrus.Println("OSDs must be specified")
			os.Exit(1)
		}
		logrus.Println("Deploying govnocloud on nodes", nodesString, "osds", osdsString, "and master", master)
		logrus.Println("Generating Ansible inventory file", ansibleInventoryFile)
		err := deploy.GenerateAnsibleConfig(nodes, osds, master, ansibleInventoryFile)
		if err != nil {
			logrus.Panic(err)
		}
		logrus.Println("Running Ansible on inventory file", ansibleInventoryFile)
		err = deploy.RunAnsible(ansibleInventoryFile, user, key)
		if err != nil {
			logrus.Panic(err)
		}
		for _, node := range nodes {
			logrus.Println("Stopping govnocloud on node", node)
			err := deploy.RunSSHCommand(node, key, user, "sudo systemctl stop govnocloud-node; cleanup.sh")
			if err != nil {
				logrus.Panic(err)
			}
		}
		logrus.Println("Stopping govnocloud on master", master)
		err = deploy.RunSSHCommand(master, key, user, "sudo systemctl stop govnocloud-master; cleanup.sh")
		if err != nil {
			logrus.Panic(err)
		}
		logrus.Println("Running cleanup.sh on master", master)
		err = deploy.RunSSHCommand(master, key, user, "/usr/local/bin/cleanup.sh")
		if err != nil {
			logrus.Panic(err)
		}
		for _, node := range nodes {
			logrus.Println("Running cleanup.sh on node", node)
			err := deploy.RunSSHCommand(node, key, user, "/usr/local/bin/cleanup.sh")
			if err != nil {
				logrus.Panic(err)
			}
		}
		logrus.Println("Copying govnocloud-master-linux-amd64 to master", master)
		err = deploy.CopyFile(master, key, user, "bin/govnocloud-master-linux-amd64", "/usr/local/bin/govnocloud-master")
		if err != nil {
			logrus.Panic(err)
		}
		for _, node := range nodes {
			logrus.Println("Copying govnocloud-node-linux-amd64 to node", node)
			err := deploy.CopyFile(node, key, user, "bin/govnocloud-node-linux-amd64", "/usr/local/bin/govnocloud-node")
			if err != nil {
				logrus.Panic(err)
			}
			err = deploy.SyncDir(node, user, "deployments/ansible", "/var/lib/libvirt/")
			if err != nil {
				logrus.Panic(err)
			}
		}
		logrus.Println("Starting govnocloud on master", master)
		err = deploy.RunSSHCommand(master, key, user, "chmod +x /usr/local/bin/govnocloud-master")
		if err != nil {
			logrus.Panic(err)
		}
		err = deploy.RunSSHCommand(master, key, user, "sudo systemctl start govnocloud-master")
		if err != nil {
			logrus.Panic(err)
		}
		for _, node := range nodes {
			logrus.Println("Starting govnocloud on node", node)
			err := deploy.RunSSHCommand(node, key, user, "chmod +x /usr/local/bin/govnocloud-node")
			if err != nil {
				logrus.Panic(err)
			}
			err = deploy.RunSSHCommand(node, key, user, "sudo systemctl start govnocloud-node")
			if err != nil {
				logrus.Panic(err)
			}
		}
		logrus.Println("Starting govnocloud front on master", master)
		err = deploy.RunSSHCommand(master, key, user, "docker stop govnocloud-front; docker rm govnocloud-front;docker pull loqutus/govnocloud-front:latest; docker run --name govnocloud-front -d -p 8080:80 loqutus/govnocloud-front:latest")
		if err != nil {
			logrus.Panic(err)
		}
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	currentUserHomeDir, err := os.UserHomeDir()
	if err != nil {
		logrus.Panic(err)
	}
	rootCmd.PersistentFlags().StringVar(&nodesString, "nodes", "", "nodes to deploy")
	rootCmd.PersistentFlags().StringVar(&osdsString, "osds", "", "osds to deploy")
	rootCmd.PersistentFlags().StringVar(&master, "master", "", "master to deploy")
	rootCmd.PersistentFlags().StringVar(&ansibleInventoryFile, "inv", "./deployments/ansible/inventories/deploy_hosts", "ansible inventory file")
	rootCmd.PersistentFlags().StringVar(&key, "key", filepath.Join(currentUserHomeDir, ".ssh/id_rsa"), "private ssh key path")
	rootCmd.PersistentFlags().StringVar(&user, "user", "root", "ssh user")
}

func main() {
	Execute()
}
