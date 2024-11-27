package cmd

import (
	"fmt"

	"github.com/mamrezb/jenkinsmaster-cli/internal/providers"
	"github.com/mamrezb/jenkinsmaster-cli/internal/providers/hetzner"
	"github.com/mamrezb/jenkinsmaster-cli/internal/providers/vm"
	"github.com/mamrezb/jenkinsmaster-cli/internal/utils"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy JenkinsMaster",
	Run: func(cmd *cobra.Command, args []string) {
		startDeployment()
	},
}

func init() {
	rootCmd.AddCommand(deployCmd)
}

func startDeployment() {
	// Check for Ansible installation
	err := utils.CheckDependencies([]string{"ansible"})
	if err != nil {
	}

	provider, err := selectProvider()
	if err != nil {
		fmt.Println("Error selecting provider:", err)
		return
	}

	// if requires terraform, check for terraform installation
	if provider.RequiresTerraform() {
		err = utils.CheckDependencies([]string{"terraform"})
		if err != nil {
		}
	}

	err = provider.Deploy()
	if err != nil {
		fmt.Println("Deployment failed:", err)
	} else {
		fmt.Println("Deployment successful!")
	}
}

func selectProvider() (providers.Provider, error) {
	providerOptions := []providers.Provider{
		&hetzner.HetznerProvider{},
		&vm.VMProvider{},
	}

	prompt := promptui.Select{
		Label: "Select Deployment Provider",
		Items: []string{
			providerOptions[0].GetName(),
			providerOptions[1].GetName(),
		},
	}

	index, _, err := prompt.Run()
	if err != nil {
		return nil, err
	}

	return providerOptions[index], nil
}
