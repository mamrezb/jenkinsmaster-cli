package vm

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/mamrezb/jenkinsmaster-cli/internal/ansible"
	"github.com/mamrezb/jenkinsmaster-cli/internal/utils"
	"github.com/manifoldco/promptui"
)

type VMProvider struct {
	IPAddress  string
	Port       string
	Username   string
	PrivateKey string
}

func (vm *VMProvider) GetName() string {
	return "SSH to Existing VM"
}

func (vm *VMProvider) RequiresTerraform() bool {
	return false
}
func (vm *VMProvider) Deploy() error {
	// Collect SSH details
	err := vm.collectSSHDetails()
	if err != nil {
		return err
	}

	// Collect Ansible variables
	ansibleConfig, err := ansible.CollectAnsibleVariables()
	if err != nil {
		return err
	}

	// Display a summary and prompt for confirmation
	err = vm.confirmInputs(ansibleConfig)
	if err != nil {
		return err
	}

	// Check for Ansible installation
	err = utils.CheckDependencyWithRetry("ansible")
	if err != nil {
		return err
	}

	// Validate SSH connection
	fmt.Println("\nValidating SSH connection...")
	err = vm.validateSSHConnection()
	if err != nil {
		return err
	}

	// Deploy with Ansible
	fmt.Println("\nDeploying JenkinsMaster with Ansible...")
	err = vm.deployAnsible(ansibleConfig)
	if err != nil {
		return err
	}

	fmt.Println("\nDeployment completed successfully!")
	return nil
}

func (vm *VMProvider) collectSSHDetails() error {
	promptIP := promptui.Prompt{
		Label:    "Enter the IP address",
		Validate: validateIPAddress,
	}

	ip, err := promptIP.Run()
	if err != nil {
		return err
	}
	vm.IPAddress = ip

	promptPort := promptui.Prompt{
		Label:    "Enter the SSH port",
		Default:  "22",
		Validate: validatePort,
	}

	port, err := promptPort.Run()
	if err != nil {
		return err
	}
	vm.Port = port

	promptUser := promptui.Prompt{
		Label:   "Enter the SSH username",
		Default: "root",
	}

	user, err := promptUser.Run()
	if err != nil {
		return err
	}
	vm.Username = user

	promptKey := promptui.Prompt{
		Label:    "Enter path to your SSH private key",
		Default:  "~/.ssh/id_rsa",
		Validate: validateFilePath,
	}

	keyPath, err := promptKey.Run()
	if err != nil {
		return err
	}
	vm.PrivateKey = expandPath(keyPath)

	return nil
}

func validateFilePath(input string) error {
	expandedPath := expandPath(input)
	fileInfo, err := os.Stat(expandedPath)
	if os.IsNotExist(err) {
		return fmt.Errorf("file does not exist")
	}
	if fileInfo.IsDir() {
		return fmt.Errorf("path is a directory, not a file")
	}
	return nil
}

func (vm *VMProvider) validateSSHConnection() error {
	sshCmd := fmt.Sprintf("ssh -o BatchMode=yes -o StrictHostKeyChecking=no -i %s -p %s %s@%s echo Connection successful", vm.PrivateKey, vm.Port, vm.Username, vm.IPAddress)
	cmd := exec.Command("sh", "-c", sshCmd)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("SSH connection failed: %v", err)
	}
	return nil
}

func validateIPAddress(input string) error {
	if net.ParseIP(input) == nil {
		return fmt.Errorf("invalid IP address")
	}
	return nil
}

func (vm *VMProvider) confirmInputs(ansibleConfig ansible.Config) error {
	fmt.Println("\nPlease review the following settings:")
	// SSH Provider settings
	fmt.Printf("IP Address: %s\n", vm.IPAddress)
	fmt.Printf("SSH Port: %s\n", vm.Port)
	fmt.Printf("SSH Username: %s\n", vm.Username)
	fmt.Printf("SSH Private Key: %s\n", vm.PrivateKey)
	// Ansible variables
	fmt.Printf("Jenkins Admin User: %s\n", ansibleConfig.JenkinsAdminUser)
	fmt.Printf("Jenkins HTTP Port: %d\n", ansibleConfig.JenkinsHTTPPort)
	fmt.Printf("Jenkins Docker Image: %s\n", ansibleConfig.JenkinsDockerImage)
	fmt.Printf("Jenkins Container Name: %s\n", ansibleConfig.JenkinsContainerName)
	fmt.Printf("Jenkins Plugin List: %v\n", ansibleConfig.JenkinsPluginList)
	fmt.Printf("Jenkins Job DSL Repo: %s\n", ansibleConfig.JenkinsJobDSLRepo)
	fmt.Printf("Jenkins Shared Library Repo: %s\n", ansibleConfig.JenkinsSharedLibraryRepo)

	for {
		prompt := promptui.Prompt{
			Label: "Do you want to proceed with these settings? (yes/no)",
			Validate: func(input string) error {
				lowerInput := strings.ToLower(strings.TrimSpace(input))
				if lowerInput == "yes" || lowerInput == "no" || lowerInput == "y" || lowerInput == "n" {
					return nil
				}
				return fmt.Errorf("Please enter 'yes' or 'no'")
			},
		}
		result, err := prompt.Run()
		if err != nil {
			if err == promptui.ErrInterrupt || err == promptui.ErrEOF {
				fmt.Println("\nInput cancelled by user.")
				return fmt.Errorf("input cancelled by user")
			}
			fmt.Println(err)
			continue
		}
		lowerResult := strings.ToLower(strings.TrimSpace(result))
		if lowerResult == "yes" || lowerResult == "y" {
			return nil
		} else if lowerResult == "no" || lowerResult == "n" {
			fmt.Println("Deployment cancelled.")
			return fmt.Errorf("deployment cancelled by user")
		}
	}
}

func validatePort(input string) error {
	port, err := strconv.Atoi(input)
	if err != nil || port < 1 || port > 65535 {
		return fmt.Errorf("invalid port number")
	}
	return nil
}

func (vm *VMProvider) deployAnsible(ansibleConfig ansible.Config) error {
	ansibleConfig.Host = vm.IPAddress
	ansibleConfig.User = vm.Username
	ansibleConfig.Port = vm.Port
	ansibleConfig.PrivateKey = vm.PrivateKey
	ansibleConfig.Forks = 10

	err := ansible.DeployAnsible(ansibleConfig)
	if err != nil {
		return err
	}

	return nil
}

// Helper function to expand ~ in file paths
func expandPath(path string) string {
	if strings.HasPrefix(path, "~") {
		homeDir, _ := os.UserHomeDir()
		return filepath.Join(homeDir, path[1:])
	}
	return path
}
