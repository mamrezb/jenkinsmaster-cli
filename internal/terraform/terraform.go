package terraform

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
)

func Apply(tfVars map[string]interface{}, moduleSource string) (string, error) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "terraform")
	if err != nil {
		return "", fmt.Errorf("failed to create temporary directory: %v", err)
	}

	// Change to the temporary directory
	originalDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current working directory: %v", err)
	}
	defer os.Chdir(originalDir) // Change back after we're done

	err = os.Chdir(tempDir)
	if err != nil {
		return "", fmt.Errorf("failed to change to temporary directory: %v", err)
	}

	// print the current working directory
	fmt.Println("Current working directory: ", tempDir)
	// print ls -tlrha
	cmd := exec.Command("ls", "-tlrha")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return "", fmt.Errorf("failed to list files in the directory: %v", err)
	}

	// Prepare -var arguments
	varArgs := []string{}
	for key, value := range tfVars {
		var varString string

		switch v := value.(type) {
		case int, int32, int64, float32, float64:
			varString = fmt.Sprintf("%s=%v", key, v)
		case string:
			varString = fmt.Sprintf("%s=%s", key, v)
		default:
			varString = fmt.Sprintf("%s=%v", key, v)
		}
		varArgs = append(varArgs, "-var", varString)
	}

	// Run terraform init with the module source
	cmdInit := exec.Command("terraform", "init", "-from-module="+moduleSource)
	cmdInit.Stdout = os.Stdout
	cmdInit.Stderr = os.Stderr
	err = cmdInit.Run()
	if err != nil {
		return "", fmt.Errorf("terraform init failed: %v", err)
	}

	// Run terraform apply with variables
	args := append([]string{"apply", "-auto-approve"}, varArgs...)
	cmdApply := exec.Command("terraform", args...)
	cmdApply.Stdout = os.Stdout
	cmdApply.Stderr = os.Stderr
	err = cmdApply.Run()
	if err != nil {
		return "", fmt.Errorf("terraform apply failed: %v", err)
	}

	return tempDir, nil
}

func GetOutput(tempDir, outputName string) (string, error) {
	// Change to the temporary directory
	originalDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current working directory: %v", err)
	}
	defer os.Chdir(originalDir)

	err = os.Chdir(tempDir)
	if err != nil {
		return "", fmt.Errorf("failed to change to temporary directory: %v", err)
	}

	cmd := exec.Command("terraform", "output", "-json")
	var out bytes.Buffer
	cmd.Stdout = &out
	err = cmd.Run()
	if err != nil {
		return "", fmt.Errorf("failed to get Terraform outputs: %v", err)
	}

	var outputs map[string]struct {
		Value interface{} `json:"value"`
	}
	err = json.Unmarshal(out.Bytes(), &outputs)
	if err != nil {
		return "", fmt.Errorf("failed to parse Terraform outputs: %v", err)
	}

	value, ok := outputs[outputName]
	if !ok {
		return "", fmt.Errorf("output %s not found", outputName)
	}

	// Convert value to string
	valueStr := fmt.Sprintf("%v", value.Value)
	return valueStr, nil
}
