package utils

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/fatih/color"
)

func CheckDependencies(dependencies []string) error {
	missing := []string{}
	for _, dep := range dependencies {
		_, err := exec.LookPath(dep)
		if err != nil {
			missing = append(missing, dep)
		}
	}

	if len(missing) > 0 {
		warn := color.New(color.FgYellow).SprintFunc()
		fmt.Printf("%s: The following dependencies are missing: %v\n", warn("Warning"), missing)
		fmt.Println("Please install them before proceeding.")
		return fmt.Errorf("missing dependencies: %v", missing)
	}

	return nil
}

func PromptInstall(dependency string) error {
	// Implement logic to install the dependency
	// For security reasons, you might want to avoid auto-installing
	// Instead, guide the user on how to install
	fmt.Printf("To install %s, please follow the instructions at:\n", dependency)
	switch dependency {
	case "ansible":
		fmt.Println("https://docs.ansible.com/ansible/latest/installation_guide/intro_installation.html")
	case "terraform":
		fmt.Println("https://learn.hashicorp.com/terraform/getting-started/install.html")
	default:
		fmt.Printf("Please refer to the official documentation for %s.\n", dependency)
	}
	return nil
}

func CheckDependencyWithRetry(dependency string) error {
	for {
		_, err := exec.LookPath(dependency)
		if err == nil {
			// Dependency is installed
			return nil
		}

		warn := color.New(color.FgYellow).SprintFunc()
		fmt.Printf("%s: '%s' is not installed and is required to continue.\n", warn("Warning"), dependency)
		fmt.Printf("Please install '%s' and then press Enter to retry.\n", dependency)
		fmt.Println("Alternatively, you can type 'exit' to cancel.")

		var input string
		fmt.Scanln(&input)
		if strings.ToLower(input) == "exit" {
			return fmt.Errorf("missing dependency: %s", dependency)
		}
		// Retry the loop
	}
}
