package ansible

import (
	"fmt"
	"math/rand"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/manifoldco/promptui"
)

func CollectAnsibleVariables() (Config, error) {
	var config Config

	// Set default values
	config.JenkinsAdminUser = "admin"
	config.JenkinsHTTPPort = 8080
	config.JenkinsDockerImage = "jenkins/jenkins:lts"
	config.JenkinsContainerName = "jenkinsmaster"
	config.JenkinsPluginList = []string{
		"configuration-as-code",
		"job-dsl",
		"pipeline-groovy-lib",
		"git",
		"ldap",
		"sonar",
		"jira",
		"github",
		"bitbucket",
		"gitlab-plugin",
	}
	config.JenkinsJobDSLRepo = "https://github.com/mamrezb/jenkinsmaster-job-dsl.git"
	config.JenkinsSharedLibraryRepo = "https://github.com/mamrezb/jenkinsmaster-shared-library.git"

	// Prompt for Jenkins admin user
	for {
		promptAdminUser := promptui.Prompt{
			Label:   "Jenkins Admin Username",
			Default: config.JenkinsAdminUser,
			Validate: func(input string) error {
				if strings.TrimSpace(input) == "" {
					return fmt.Errorf("Username cannot be empty")
				}
				return nil
			},
		}
		result, err := promptAdminUser.Run()
		if err != nil {
			if err == promptui.ErrInterrupt || err == promptui.ErrEOF {
				fmt.Println("\nInput cancelled by user.")
				return config, fmt.Errorf("input cancelled by user")
			}
			fmt.Println(err)
			continue
		}
		config.JenkinsAdminUser = result
		break
	}

	// Prompt for Jenkins admin password
	for {
		promptAdminPassword := promptui.Prompt{
			Label: "Jenkins Admin Password (or type 'generate' to generate a strong password)",
			Mask:  '*',
			Validate: func(input string) error {
				if strings.TrimSpace(input) == "" {
					return fmt.Errorf("Password cannot be empty")
				}
				return nil
			},
		}
		result, err := promptAdminPassword.Run()
		if err != nil {
			if err == promptui.ErrInterrupt || err == promptui.ErrEOF {
				fmt.Println("\nInput cancelled by user.")
				return config, fmt.Errorf("input cancelled by user")
			}
			fmt.Println(err)
			continue
		}

		if result == "generate" {
			// Generate a random strong password
			generatedPassword := generateStrongPassword()
			config.JenkinsAdminPassword = generatedPassword
			fmt.Printf("Generated strong password: %s\n", generatedPassword)
			break
		} else {
			if isStrongPassword(result) {
				config.JenkinsAdminPassword = result
				break
			} else {
				fmt.Println("Password is not strong enough. It should be at least 8 characters long, and include uppercase, lowercase, numbers, and special characters.")
			}
		}
	}

	// Prompt for Jenkins HTTP port
	for {
		promptHTTPPort := promptui.Prompt{
			Label:    "Jenkins HTTP Port",
			Default:  strconv.Itoa(config.JenkinsHTTPPort),
			Validate: validatePort,
		}
		result, err := promptHTTPPort.Run()
		if err != nil {
			if err == promptui.ErrInterrupt || err == promptui.ErrEOF {
				fmt.Println("\nInput cancelled by user.")
				return config, fmt.Errorf("input cancelled by user")
			}
			fmt.Println(err)
			continue
		}
		port, err := strconv.Atoi(result)
		if err != nil {
			fmt.Println("Invalid port number")
			continue
		}
		config.JenkinsHTTPPort = port
		break
	}

	// Prompt for Jenkins Docker image
	for {
		promptDockerImage := promptui.Prompt{
			Label:   "Jenkins Docker Image",
			Default: config.JenkinsDockerImage,
			Validate: func(input string) error {
				if strings.TrimSpace(input) == "" {
					return fmt.Errorf("Docker image cannot be empty")
				}
				return nil
			},
		}
		result, err := promptDockerImage.Run()
		if err != nil {
			if err == promptui.ErrInterrupt || err == promptui.ErrEOF {
				fmt.Println("\nInput cancelled by user.")
				return config, fmt.Errorf("input cancelled by user")
			}
			fmt.Println(err)
			continue
		}
		config.JenkinsDockerImage = result
		// Validate Docker image
		if validateDockerImage(config.JenkinsDockerImage) {
			break
		} else {
			fmt.Println("Docker image not found on Docker Hub. Please enter a valid image.")
		}
	}

	// Prompt for Jenkins container name
	for {
		promptContainerName := promptui.Prompt{
			Label:   "Jenkins Container Name",
			Default: config.JenkinsContainerName,
			Validate: func(input string) error {
				if strings.TrimSpace(input) == "" {
					return fmt.Errorf("Container name cannot be empty")
				}
				return nil
			},
		}
		result, err := promptContainerName.Run()
		if err != nil {
			if err == promptui.ErrInterrupt || err == promptui.ErrEOF {
				fmt.Println("\nInput cancelled by user.")
				return config, fmt.Errorf("input cancelled by user")
			}
			fmt.Println(err)
			continue
		}
		config.JenkinsContainerName = result
		break
	}

	// Prompt for Jenkins plugin list
	plugins, err := promptPluginList(config.JenkinsPluginList)
	if err != nil {
		if err == promptui.ErrInterrupt || err == promptui.ErrEOF {
			fmt.Println("\nInput cancelled by user.")
			return config, fmt.Errorf("input cancelled by user")
		}
		return config, err
	}
	config.JenkinsPluginList = plugins

	// Prompt for Jenkins Job DSL Repo
	for {
		promptJobDSLRepo := promptui.Prompt{
			Label:   "Jenkins Job DSL Repository",
			Default: config.JenkinsJobDSLRepo,
			Validate: func(input string) error {
				if strings.TrimSpace(input) == "" {
					return fmt.Errorf("Repository URL cannot be empty")
				}
				return nil
			},
		}
		result, err := promptJobDSLRepo.Run()
		if err != nil {
			if err == promptui.ErrInterrupt || err == promptui.ErrEOF {
				fmt.Println("\nInput cancelled by user.")
				return config, fmt.Errorf("input cancelled by user")
			}
			fmt.Println(err)
			continue
		}
		config.JenkinsJobDSLRepo = result
		if validateGitRepo(config.JenkinsJobDSLRepo) {
			break
		} else {
			// Ask if user wants to bypass validation
			if promptBypassValidation("Jenkins Job DSL Repository") {
				break
			} else {
				fmt.Println("Please enter a valid repository URL.")
			}
		}
	}

	// Prompt for Jenkins Shared Library Repo
	for {
		promptSharedLibRepo := promptui.Prompt{
			Label:   "Jenkins Shared Library Repository",
			Default: config.JenkinsSharedLibraryRepo,
			Validate: func(input string) error {
				if strings.TrimSpace(input) == "" {
					return fmt.Errorf("Repository URL cannot be empty")
				}
				return nil
			},
		}
		result, err := promptSharedLibRepo.Run()
		if err != nil {
			if err == promptui.ErrInterrupt || err == promptui.ErrEOF {
				fmt.Println("\nInput cancelled by user.")
				return config, fmt.Errorf("input cancelled by user")
			}
			fmt.Println(err)
			continue
		}
		config.JenkinsSharedLibraryRepo = result
		if validateGitRepo(config.JenkinsSharedLibraryRepo) {
			break
		} else {
			// Ask if user wants to bypass validation
			if promptBypassValidation("Jenkins Shared Library Repository") {
				break
			} else {
				fmt.Println("Please enter a valid repository URL.")
			}
		}
	}

	return config, nil
}

// Helper functions used in CollectAnsibleVariables
func validatePort(input string) error {
	port, err := strconv.Atoi(input)
	if err != nil || port < 1 || port > 65535 {
		return fmt.Errorf("Invalid port number")
	}
	return nil
}

func isStrongPassword(password string) bool {
	var (
		hasMinLen  = false
		hasUpper   = false
		hasLower   = false
		hasNumber  = false
		hasSpecial = false
	)

	if len(password) >= 8 {
		hasMinLen = true
	}

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	return hasMinLen && hasUpper && hasLower && hasNumber && hasSpecial
}

func generateStrongPassword() string {
	const passwordLength = 12
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()-_=+[]{}<>?,./"

	rand.Seed(time.Now().UnixNano())
	password := make([]byte, passwordLength)
	for i := 0; i < passwordLength; i++ {
		password[i] = letters[rand.Intn(len(letters))]
	}
	return string(password)
}

func validateDockerImage(image string) bool {
	// Split image into name and tag
	imageParts := strings.Split(image, ":")
	imageName := imageParts[0]
	var imageTag string
	if len(imageParts) > 1 {
		imageTag = imageParts[1]
	} else {
		imageTag = "latest"
	}

	// Replace "/" with "%2F" to properly encode the URL
	imageNameEncoded := strings.ReplaceAll(imageName, "/", "%2F")
	url := fmt.Sprintf("https://hub.docker.com/v2/repositories/%s/tags/%s", imageNameEncoded, imageTag)
	resp, err := http.Get(url)
	if err != nil || resp.StatusCode != 200 {
		return false
	}
	return true
}

var fixedPlugins = []string{
	"configuration-as-code",
	"job-dsl",
	"pipeline-groovy-lib",
	"git",
}

func promptPluginList(defaultPlugins []string) ([]string, error) {
	plugins := defaultPlugins

	// Ensure fixed plugins are included and not removable
	plugins = mergePlugins(fixedPlugins, plugins)

	for {
		fmt.Println("\nCurrent Jenkins Plugin List:")
		for i, plugin := range plugins {
			if isFixedPlugin(plugin) {
				fmt.Printf("%d. %s (fixed)\n", i+1, plugin)
			} else {
				fmt.Printf("%d. %s\n", i+1, plugin)
			}
		}
		prompt := promptui.Select{
			Label: "Select an action",
			Items: []string{
				"Add a plugin",
				"Remove a plugin",
				"Continue",
			},
		}
		index, _, err := prompt.Run()
		if err != nil {
			if err == promptui.ErrInterrupt || err == promptui.ErrEOF {
				fmt.Println("\nInput cancelled by user.")
				return nil, err
			}
			fmt.Println(err)
			continue
		}
		switch index {
		case 0:
			// Add a plugin
			promptAdd := promptui.Prompt{
				Label: "Enter plugin ID to add",
				Validate: func(input string) error {
					if strings.TrimSpace(input) == "" {
						return fmt.Errorf("plugin ID cannot be empty")
					}
					return nil
				},
			}
			pluginID, err := promptAdd.Run()
			if err != nil {
				if err == promptui.ErrInterrupt || err == promptui.ErrEOF {
					fmt.Println("\nInput cancelled by user.")
					return nil, err
				}
				fmt.Println(err)
				continue
			}
			pluginID = strings.TrimSpace(pluginID)
			if !validateJenkinsPlugin(pluginID) {
				fmt.Println("Invalid plugin ID. Plugin not found.")
			} else if contains(plugins, pluginID) {
				fmt.Println("Plugin already in the list.")
			} else {
				plugins = append(plugins, pluginID)
				fmt.Println("Plugin added.")
			}
		case 1:
			// Remove a plugin
			removablePlugins := []string{"Cancel"}
			for _, plugin := range plugins {
				if !isFixedPlugin(plugin) {
					removablePlugins = append(removablePlugins, plugin)
				}
			}
			if len(removablePlugins) == 1 {
				fmt.Println("No plugins to remove.")
				continue
			}
			promptRemove := promptui.Select{
				Label: "Select a plugin to remove",
				Items: removablePlugins,
			}
			idx, _, err := promptRemove.Run()
			if err != nil {
				if err == promptui.ErrInterrupt || err == promptui.ErrEOF {
					fmt.Println("\nInput cancelled by user.")
					return nil, err
				}
				fmt.Println(err)
				continue
			}
			if idx == 0 {
				// User selected "Cancel"
				continue
			}
			pluginToRemove := removablePlugins[idx]
			// Remove from plugins
			plugins = removePlugin(plugins, pluginToRemove)
			fmt.Println("Plugin removed.")
		case 2:
			// Continue
			return plugins, nil
		}
	}
}

func isFixedPlugin(plugin string) bool {
	for _, fixedPlugin := range fixedPlugins {
		if plugin == fixedPlugin {
			return true
		}
	}
	return false
}

func contains(plugins []string, plugin string) bool {
	for _, p := range plugins {
		if p == plugin {
			return true
		}
	}
	return false
}

func removePlugin(plugins []string, plugin string) []string {
	newPlugins := []string{}
	for _, p := range plugins {
		if p != plugin {
			newPlugins = append(newPlugins, p)
		}
	}
	return newPlugins
}

func mergePlugins(fixed, current []string) []string {
	pluginSet := make(map[string]bool)
	for _, plugin := range fixed {
		pluginSet[plugin] = true
	}
	for _, plugin := range current {
		pluginSet[plugin] = true
	}
	merged := []string{}
	for plugin := range pluginSet {
		merged = append(merged, plugin)
	}
	return merged
}

func validateJenkinsPlugin(pluginID string) bool {
	url := fmt.Sprintf("https://plugins.jenkins.io/api/plugin/%s", pluginID)
	resp, err := http.Get(url)
	if err != nil || resp.StatusCode != 200 {
		return false
	}
	return true
}

func validateGitRepo(repoURL string) bool {
	cmd := exec.Command("git", "ls-remote", repoURL)
	err := cmd.Run()
	return err == nil
}

func promptBypassValidation(item string) bool {
	for {
		prompt := promptui.Prompt{
			Label: fmt.Sprintf("Unable to validate %s. Do you want to proceed anyway? (yes/no)", item),
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
				return false
			}
			fmt.Println(err)
			continue
		}
		lowerResult := strings.ToLower(strings.TrimSpace(result))
		if lowerResult == "yes" || lowerResult == "y" {
			return true
		} else if lowerResult == "no" || lowerResult == "n" {
			return false
		}
	}
}
