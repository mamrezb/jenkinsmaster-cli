package utils

import (
	"fmt"
	"os/exec"
	"time"
)

func WaitForSSH(host, port, user, privateKey string, timeout time.Duration) error {
	endTime := time.Now().Add(timeout)
	for {
		// Attempt to SSH into the host and run a simple command
		sshCmd := exec.Command("ssh",
			"-o", "BatchMode=yes",
			"-o", "StrictHostKeyChecking=no",
			"-i", privateKey,
			"-p", port,
			fmt.Sprintf("%s@%s", user, host),
			"echo SSH connection successful")

		output, err := sshCmd.CombinedOutput()
		if err == nil {
			// SSH command succeeded
			fmt.Println(string(output))
			return nil
		}

		if time.Now().After(endTime) {
			return fmt.Errorf("SSH connection to %s:%s timed out", host, port)
		}

		fmt.Println("Waiting for SSH to become available...")
		time.Sleep(10 * time.Second)
	}
}
