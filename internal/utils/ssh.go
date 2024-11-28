package utils

import (
	"fmt"
	"os/exec"
	"time"
)

func ValidateSSHConnection(host, port, user, privateKey string) error {
	sshCmd := fmt.Sprintf("ssh -o BatchMode=yes -o StrictHostKeyChecking=no -i %s -p %s %s@%s echo Connection successful", privateKey, port, user, host)
	cmd := exec.Command("sh", "-c", sshCmd)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("SSH connection failed: %v", err)
	}
	return nil
}

func WaitForSSH(host, port, user, privateKey string, timeout time.Duration) error {
	endTime := time.Now().Add(timeout)
	for {
		// Attempt to SSH into the host and run a simple command
		err := ValidateSSHConnection(host, port, user, privateKey)
		if err == nil {
			return nil
		}

		if time.Now().After(endTime) {
			return fmt.Errorf("SSH connection to %s:%s timed out", host, port)
		}

		fmt.Println("Waiting for SSH to become available...")
		time.Sleep(10 * time.Second)
	}
}
