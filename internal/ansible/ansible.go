package ansible

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"os/exec"
)

// Embed all templates into the binary using go:embed.
//
//go:embed templates/*
var ansibleTemplates embed.FS

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

	// Change to the temporary directory
	err = os.Chdir(tempDir)
	if err != nil {
		return fmt.Errorf("failed to change to temporary directory: %v", err)
	}

	// Generate inventory.ini
	inventoryContent, err := parseTemplate("templates/inventory.tpl", config)
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
	ansibleCfgContent, err := parseTemplate("templates/ansible.cfg.tpl", config)
	if err != nil {
		return err
	}
	err = os.WriteFile("ansible.cfg", []byte(ansibleCfgContent), 0644)
	if err != nil {
		return err
	}

	// Generate requirements.yml
	requirementsContent, err := parseTemplate("templates/requirements.yml.tpl", config)
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
	playbookContent, err := parseTemplate("templates/playbook.yml.tpl", config)
	if err != nil {
		return err
	}
	err = os.WriteFile("playbook.yml", []byte(playbookContent), 0644)
	if err != nil {
		return err
	}

	// Run ansible-playbook
	varsMap := map[string]interface{}{
		"jenkins_admin_user":          config.JenkinsAdminUser,
		"jenkins_admin_password":      config.JenkinsAdminPassword,
		"jenkins_http_port":           config.JenkinsHTTPPort,
		"jenkins_docker_image":        config.JenkinsDockerImage,
		"jenkins_container_name":      config.JenkinsContainerName,
		"jenkins_plugin_list":         config.JenkinsPluginList,
		"jenkins_job_dsl_repo":        config.JenkinsJobDSLRepo,
		"jenkins_shared_library_repo": config.JenkinsSharedLibraryRepo,
	}
	extraVarsJSON, err := json.Marshal(varsMap)
	if err != nil {
		return fmt.Errorf("failed to marshal extraVars: %v", err)
	}
	ansibleCmd := exec.Command("ansible-playbook", "playbook.yml", "-e", string(extraVarsJSON))
	ansibleCmd.Stdout = os.Stdout
	ansibleCmd.Stderr = os.Stderr
	err = ansibleCmd.Run()
	if err != nil {
		return fmt.Errorf("ansible deployment failed: %v", err)
	}

	return nil
}

// parseTemplate reads a file from the embedded templates and executes it with data
func parseTemplate(templateFile string, data interface{}) (string, error) {
	tmpl, err := template.ParseFS(ansibleTemplates, templateFile)
	if err != nil {
		return "", fmt.Errorf("error parsing template: %v", err)
	}
	var tpl bytes.Buffer
	if err := tmpl.Execute(&tpl, data); err != nil {
		return "", fmt.Errorf("error executing template: %v", err)
	}
	return tpl.String(), nil
}
