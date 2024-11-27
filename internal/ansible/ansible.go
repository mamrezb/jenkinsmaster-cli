package ansible

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"os/exec"
	"path/filepath"
)

type Config struct {
	Host                     string
	User                     string
	Port                     string
	PrivateKey               string
	Forks                    int
	InventoryFile            string
	JenkinsAdminUser         string
	JenkinsAdminPassword     string
	JenkinsHTTPPort          int
	JenkinsDockerImage       string
	JenkinsContainerName     string
	JenkinsPluginList        []string
	JenkinsJobDSLRepo        string
	JenkinsSharedLibraryRepo string
}

func DeployAnsible(config Config) error {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "ansible")
	if err != nil {
		return fmt.Errorf("failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir) // Clean up tempDir after we're done

	// Save the original working directory
	originalDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current working directory: %v", err)
	}
	defer os.Chdir(originalDir) // Change back after we're done

	// Change to the temporary directory
	err = os.Chdir(tempDir)
	if err != nil {
		return fmt.Errorf("failed to change to temporary directory: %v", err)
	}

	// Ensure templates directory exists
	exePath, err := os.Executable()
	if err != nil {
		return err
	}
	exeDir := filepath.Dir(exePath)
	templatesDir := filepath.Join(exeDir, "templates")

	// Generate inventory.ini
	inventoryContent, err := parseTemplate(filepath.Join(templatesDir, "inventory.tpl"), config)
	if err != nil {
		return err
	}
	inventoryFile := "inventory.ini"
	err = os.WriteFile(inventoryFile, []byte(inventoryContent), 0644)
	if err != nil {
		return err
	}

	// Update config with inventory file path
	config.InventoryFile = inventoryFile

	// Generate ansible.cfg
	ansibleCfgContent, err := parseTemplate(filepath.Join(templatesDir, "ansible.cfg.tpl"), config)
	if err != nil {
		return err
	}
	err = os.WriteFile("ansible.cfg", []byte(ansibleCfgContent), 0644)
	if err != nil {
		return err
	}

	// Generate requirements.yml
	requirementsContent, err := parseTemplate(filepath.Join(templatesDir, "requirements.yml.tpl"), config)
	if err != nil {
		return err
	}
	err = os.WriteFile("requirements.yml", []byte(requirementsContent), 0644)
	if err != nil {
		return err
	}

	// Install Ansible Galaxy roles
	fmt.Println("Installing Ansible Galaxy roles...")
	galaxyCmd := exec.Command("ansible-galaxy", "install", "-r", "requirements.yml", "--force")
	galaxyCmd.Stdout = os.Stdout
	galaxyCmd.Stderr = os.Stderr
	err = galaxyCmd.Run()
	if err != nil {
		return fmt.Errorf("failed to install Ansible Galaxy roles: %v", err)
	}

	// Generate playbook.yml
	playbookContent, err := parseTemplate(filepath.Join(templatesDir, "playbook.yml.tpl"), config)
	if err != nil {
		return err
	}
	err = os.WriteFile("playbook.yml", []byte(playbookContent), 0644)
	if err != nil {
		return err
	}

	// Run ansible-playbook
	ansibleCmd := exec.Command("ansible-playbook", "playbook.yml")
	ansibleCmd.Stdout = os.Stdout
	ansibleCmd.Stderr = os.Stderr
	err = ansibleCmd.Run()
	if err != nil {
		return fmt.Errorf("ansible deployment failed: %v", err)
	}

	return nil
}

func parseTemplate(templateFile string, data interface{}) (string, error) {
	tmpl, err := template.ParseFiles(templateFile)
	if err != nil {
		return "", err
	}
	var tpl bytes.Buffer
	if err := tmpl.Execute(&tpl, data); err != nil {
		return "", err
	}
	return tpl.String(), nil
}
