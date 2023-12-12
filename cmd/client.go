/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"strconv"

	"github.com/rusik69/govnocloud/pkg/client"
	"github.com/spf13/cobra"
)

// clientCmd represents the client command
var clientCmd = &cobra.Command{
	Use:   "client",
	Short: "govnocloud client",
	Long:  `govnocloud client`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("client called")
	},
}

// vm represents the vm commands
var vmCmd = &cobra.Command{
	Use:   "vm",
	Short: "vm commands",
	Long:  `vm commands`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("usage: vm [create|delete|list]")
	},
}

// vmCreateCmd represents the vm create command
var vmCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "create vm",
	Long:  `create vm`,
	Run: func(cmd *cobra.Command, args []string) {
		vmName := cmd.PersistentFlags().Lookup("name").Value.String()
		if vmName == "" {
			panic("vm name is required")
		}
		vmImage := cmd.PersistentFlags().Lookup("image").Value.String()
		if vmImage == "" {
			panic("vm image is required")
		}
		vmFlavor := cmd.PersistentFlags().Lookup("flavor").Value.String()
		if vmFlavor == "" {
			panic("vm flavor is required")
		}
		host := cmd.PersistentFlags().Lookup("host").Value.String()
		if host == "" {
			panic("host is required")
		}
		port := cmd.PersistentFlags().Lookup("port").Value.String()
		if port == "" {
			panic("port is required")
		}
		id, err := client.CreateVM(host, port, vmName, vmImage, vmFlavor)
		if err != nil {
			panic(err)
		}
		fmt.Println("VM created with id " + string(id))
	},
}

// vmDeleteCmd represents the vm delete command
var vmDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "delete vm",
	Long:  `delete vm`,
	Run: func(cmd *cobra.Command, args []string) {
		host := cmd.PersistentFlags().Lookup("host").Value.String()
		if host == "" {
			panic("host is required")
		}
		port := cmd.PersistentFlags().Lookup("port").Value.String()
		if port == "" {
			panic("port is required")
		}
		idString := cmd.PersistentFlags().Lookup("id").Value.String()
		if idString == "" {
			panic("id is required")
		}
		id, err := strconv.Atoi(idString)
		if err != nil {
			panic(err)
		}
		err = client.DeleteVM(id, host, port)
		if err != nil {
			panic(err)
		}
	},
}

var vmGetCmd = &cobra.Command{
	Use:   "get",
	Short: "get vm",
	Long:  `get vm`,
	Run: func(cmd *cobra.Command, args []string) {
		host := cmd.PersistentFlags().Lookup("host").Value.String()
		if host == "" {
			panic("host is required")
		}
		port := cmd.PersistentFlags().Lookup("port").Value.String()
		if port == "" {
			panic("port is required")
		}
		idString := cmd.PersistentFlags().Lookup("id").Value.String()
		if idString == "" {
			panic("id is required")
		}
		id, err := strconv.Atoi(idString)
		if err != nil {
			panic(err)
		}
		vm, err := client.GetVM(id, host, port)
		if err != nil {
			panic(err)
		}
		fmt.Println(
			"ID: " + string(vm.ID) + "\n" +
				"Name: " + vm.Name + "\n" +
				"IP: " + vm.IP + "\n" +
				"Host: " + vm.Host + "\n" +
				"State: " + vm.State + "\n" +
				"Image: " + vm.Image + "\n" +
				"Flavor: " + vm.Flavor + "\n",
		)
	},
}

// vmListCmd represents the vm list command
var vmListCmd = &cobra.Command{
	Use:   "list",
	Short: "list vm",
	Long:  `list vm`,
	Run: func(cmd *cobra.Command, args []string) {
		host := cmd.PersistentFlags().Lookup("host").Value.String()
		if host == "" {
			panic("host is required")
		}
		port := cmd.PersistentFlags().Lookup("port").Value.String()
		if port == "" {
			panic("port is required")
		}
		vms, err := client.ListVMs(host, port)
		if err != nil {
			panic(err)
		}
		fmt.Printf("| %-10s | %-10s | %-16s | %-5s | %-7s | %-20s | %-5s | %-10s |\n", "ID", "NAME", "IP", "Host", "Status", "Image", "Flavor", "Volumes")
		fmt.Println("------------------------------------------------------------------------------------------------------------------------")
		for _, vm := range vms {
			fmt.Println("| %-10s | %-10s | %-16s | %-5s | %-7s | %-20s | %-5s | %-10s |\n", vm.ID, vm.Name, vm.IP, vm.Host, vm.State, vm.Image, vm.Flavor, vm.Volumes)
		}
	},
}

func init() {
	rootCmd.AddCommand(clientCmd)
	clientCmd.AddCommand(vmCmd)
	vmCmd.AddCommand(vmCreateCmd)
	vmCmd.AddCommand(vmDeleteCmd)
	vmCmd.AddCommand(vmListCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	clientCmd.PersistentFlags().String("host", "localhost", "host to connect to")
	clientCmd.PersistentFlags().String("port", "6969", "port to connect to")
	vmCmd.PersistentFlags().String("name", "", "name of the vm")
	vmCmd.PersistentFlags().String("image", "", "image of the vm")
	vmCmd.PersistentFlags().String("flavor", "", "flavor of the vm")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// clientCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
