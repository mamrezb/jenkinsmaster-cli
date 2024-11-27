package hetzner

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/mamrezb/jenkinsmaster-cli/internal/ansible"
	"github.com/mamrezb/jenkinsmaster-cli/internal/terraform"
	"github.com/mamrezb/jenkinsmaster-cli/internal/utils"
	"github.com/manifoldco/promptui"
)

type HetznerProvider struct {
	Token          string
	ServerType     string
	ServerLocation string
	ServerImage    string
	SSHKeyPath     string
	SSHKeyName     string
	ServerName     string
	Client         *hcloud.Client
}

func (h *HetznerProvider) GetName() string {
	return "Hetzner Cloud"
}

func (h *HetznerProvider) RequiresTerraform() bool {
	return true
}

func (h *HetznerProvider) Deploy() error {
	// Collect user inputs
	err := h.collectToken()
	if err != nil {
		return err
	}

	err = h.selectServerLocation()
	if err != nil {
		return err
	}

	err = h.selectServerType()
	if err != nil {
		return err
	}

	err = h.selectServerImage()
	if err != nil {
		return err
	}

	err = h.collectSSHKeyPath()
	if err != nil {
		return err
	}

	err = h.collectSSHKeyName()
	if err != nil {
		return err
	}

	err = h.collectServerName()
	if err != nil {
		return err
	}

	ansibleConfig, err := ansible.CollectAnsibleVariables()
	if err != nil {
		return err
	}

	// Prepare Terraform variables
	tfVars := map[string]interface{}{
		"hcloud_token":        h.Token,
		"server_name":         h.ServerName,
		"server_type":         h.ServerType,
		"server_image":        h.ServerImage,
		"ssh_public_key_path": h.SSHKeyPath,
		"ssh_key_name":        h.SSHKeyName,
		"server_location":     h.ServerLocation,
		"ssh_port":            22,
		"jenkins_http_port":   ansibleConfig.JenkinsHTTPPort,
	}

	// Display a summary and prompt for confirmation
	err = h.confirmInputs(ansibleConfig)
	if err != nil {
		return err
	}

	// Check for Terraform installation
	err = utils.CheckDependencyWithRetry("terraform")
	if err != nil {
		return err
	}

	// Apply Terraform
	fmt.Println("\nProvisioning server with Terraform...")
	tempDir, err := terraform.Apply(tfVars, "registry.terraform.io/mamrezb/jenkinsmaster/hcloud")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempDir) // Clean up tempDir after we're done

	// Get server IP from Terraform outputs
	serverIP, err := terraform.GetOutput(tempDir, "server_ip")
	if err != nil {
		return err
	}

	// Wait for SSH to become available
	fmt.Println("\nWaiting for the server to be ready for SSH connections...")
	err = utils.WaitForSSH(serverIP, "22", "root", h.SSHKeyPath, 5*time.Minute)
	if err != nil {
		return err
	}

	// Check for Ansible installation
	err = utils.CheckDependencyWithRetry("ansible")
	if err != nil {
		return err
	}

	// wait for 60 seconds for the server to be ready
	time.Sleep(60 * time.Second)

	// Deploy with Ansible
	fmt.Println("\nDeploying JenkinsMaster with Ansible...")
	err = h.deployAnsible(serverIP, ansibleConfig)
	if err != nil {
		return err
	}

	fmt.Println("\nDeployment completed successfully!")
	return nil
}

func (h *HetznerProvider) collectToken() error {
	prompt := promptui.Prompt{
		Label:    "Enter your Hetzner API Token",
		Validate: validateNonEmpty,
		Mask:     '*',
	}

	for {
		result, err := prompt.Run()
		if err != nil {
			if err == promptui.ErrInterrupt || err == promptui.ErrEOF {
				fmt.Println("\nInput cancelled by user.")
				return fmt.Errorf("input cancelled by user")
			}
			fmt.Println(err)
			continue
		}
		h.Token = strings.TrimSpace(result)

		// Initialize Hetzner client
		h.Client = hcloud.NewClient(hcloud.WithToken(h.Token))

		// Validate token
		err = h.validateToken()
		if err == nil {
			break
		} else {
			fmt.Println("Invalid token, please try again.")
		}
	}

	return nil
}

func (h *HetznerProvider) validateToken() error {
	// Try to list server types as a simple validation
	_, err := h.Client.ServerType.All(context.Background())
	if err != nil {
		return err
	}
	return nil
}

// Helper function
func validateNonEmpty(input string) error {
	if len(strings.TrimSpace(input)) == 0 {
		return fmt.Errorf("input cannot be empty")
	}
	return nil
}

func (h *HetznerProvider) selectServerLocation() error {
	locations, err := h.fetchServerLocations()
	if err != nil {
		return err
	}

	prompt := promptui.Select{
		Label: "Select Server Location",
		Items: locations,
	}

	index, _, err := prompt.Run()
	if err != nil {
		return err
	}

	h.ServerLocation = locations[index]
	return nil
}

func (h *HetznerProvider) fetchServerLocations() ([]string, error) {
	locations, err := h.Client.Location.All(context.Background())
	if err != nil {
		return nil, err
	}

	var locationNames []string
	for _, loc := range locations {
		locationNames = append(locationNames, loc.Name)
	}

	sort.Strings(locationNames)
	return locationNames, nil
}

func (h *HetznerProvider) selectServerType() error {
	serverTypes, err := h.fetchServerTypes()
	if err != nil {
		return err
	}

	if len(serverTypes) == 0 {
		return fmt.Errorf("no available server types for location %s", h.ServerLocation)
	}

	prompt := promptui.Select{
		Label: "Select Server Type",
		Items: serverTypes,
	}

	index, _, err := prompt.Run()
	if err != nil {
		return err
	}

	// Extract the server type name from the selected string
	selected := serverTypes[index]
	h.ServerType = strings.Split(selected, ":")[0]
	return nil
}

func (h *HetznerProvider) fetchServerTypes() ([]string, error) {
	// Retrieve all server types
	serverTypes, err := h.Client.ServerType.All(context.Background())
	if err != nil {
		return nil, err
	}

	// Create a map of server type IDs to server types
	serverTypeMap := make(map[int64]*hcloud.ServerType)
	for _, st := range serverTypes {
		// Skip deprecated server types
		if st.Deprecation != nil {
			continue
		}
		serverTypeMap[st.ID] = st
	}

	// Retrieve all datacenters
	datacenters, err := h.Client.Datacenter.All(context.Background())
	if err != nil {
		return nil, err
	}

	var availableServerTypes []string
	for _, dc := range datacenters {
		if dc.Location.Name == h.ServerLocation {
			for _, stRef := range dc.ServerTypes.Available {
				st := serverTypeMap[stRef.ID]
				if st == nil {
					continue
				}

				// Retrieve pricing information
				var monthlyPrice float64
				for _, pricing := range st.Pricings {
					if pricing.Location != nil && pricing.Location.Name == h.ServerLocation {
						monthlyPrice, _ = strconv.ParseFloat(pricing.Monthly.Gross, 64)
						break
					}
				}

				// Format the server type details
				formatted := fmt.Sprintf(
					"%s: %d vCPUs, %.2f GB RAM, %d GB Disk, %s, €%.2f/month",
					st.Name,
					st.Cores,
					st.Memory,
					st.Disk,
					st.Architecture,
					monthlyPrice,
				)
				availableServerTypes = append(availableServerTypes, formatted)
			}
			break
		}
	}

	sort.Strings(availableServerTypes)
	return availableServerTypes, nil
}

func (h *HetznerProvider) selectServerImage() error {
	images, err := h.fetchServerImages()
	if err != nil {
		return err
	}

	if len(images) == 0 {
		return fmt.Errorf("no available images")
	}

	prompt := promptui.Select{
		Label: "Select Server Image",
		Items: images,
	}

	index, _, err := prompt.Run()
	if err != nil {
		return err
	}

	h.ServerImage = images[index]
	return nil
}

func (h *HetznerProvider) fetchServerImages() ([]string, error) {
	images, err := h.Client.Image.All(context.Background())
	if err != nil {
		return nil, err
	}

	imageSet := make(map[string]struct{})
	var imageList []string

	for _, img := range images {
		if img.Type != hcloud.ImageTypeSystem || !img.Deprecated.IsZero() {
			continue
		}
		if img.Architecture != hcloud.ArchitectureX86 {
			continue // Skip non-x86 architectures
		}
		if _, exists := imageSet[img.Name]; !exists {
			imageSet[img.Name] = struct{}{}
			imageList = append(imageList, img.Name)
		}
	}

	sort.Strings(imageList)
	return imageList, nil
}

func (h *HetznerProvider) collectSSHKeyPath() error {
	prompt := promptui.Prompt{
		Label:    "Enter path to your SSH public key",
		Default:  "~/.ssh/id_rsa.pub",
		Validate: validateFilePath,
	}

	result, err := prompt.Run()
	if err != nil {
		return err
	}

	h.SSHKeyPath = expandPath(result)
	return nil
}

func (h *HetznerProvider) collectSSHKeyName() error {
	prompt := promptui.Prompt{
		Label:   "Enter a name for the SSH key in Hetzner Cloud",
		Default: "jenkinsmaster-key",
	}

	result, err := prompt.Run()
	if err != nil {
		return err
	}

	h.SSHKeyName = result
	return nil
}

func (h *HetznerProvider) collectServerName() error {
	prompt := promptui.Prompt{
		Label:   "Enter a name for the JenkinsMaster server",
		Default: "jenkinsmaster-server",
	}

	result, err := prompt.Run()
	if err != nil {
		return err
	}

	h.ServerName = result
	return nil
}

func (h *HetznerProvider) confirmInputs(ansibleConfig ansible.Config) error {
	fmt.Println("\nPlease review the following settings:")
	// Provider settings
	fmt.Printf("Server Name: %s\n", h.ServerName)
	fmt.Printf("Server Type: %s\n", h.ServerType)
	fmt.Printf("Server Image: %s\n", h.ServerImage)
	fmt.Printf("Server Location: %s\n", h.ServerLocation)
	fmt.Printf("SSH Key Path: %s\n", h.SSHKeyPath)
	fmt.Printf("SSH Key Name: %s\n", h.SSHKeyName)
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
				return fmt.Errorf("please enter 'yes' or 'no'")
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

func (h *HetznerProvider) deployAnsible(serverIP string, ansibleConfig ansible.Config) error {
	ansibleConfig.Host = serverIP
	ansibleConfig.User = "root"
	ansibleConfig.Port = "22"
	ansibleConfig.PrivateKey = h.SSHKeyPath
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

func validateFilePath(input string) error {
	if len(strings.TrimSpace(input)) == 0 {
		return fmt.Errorf("path cannot be empty")
	}
	expandedPath := expandPath(input)
	if _, err := os.Stat(expandedPath); os.IsNotExist(err) {
		return fmt.Errorf("file does not exist")
	}
	return nil
}
