package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/spf13/cobra"
)

var logsCmd = &cobra.Command{
	Use:   "logs [function]",
	Short: "Stream logs for a function",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		funcName := args[0]
		
		// 1. Find the container ID
		// We need to look for containers with name prefix "nanolambda-<funcName>-"
		cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
		if err != nil {
			fmt.Printf("Error connecting to Docker: %v\n", err)
			return
		}

		containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{})
		if err != nil {
			fmt.Printf("Error listing containers: %v\n", err)
			return
		}

		var targetID string
		prefix := "nanolambda-" + funcName + "-"
		
		// Find the most recent one
		for _, container := range containers {
			for _, name := range container.Names {
				// name comes with slash e.g. /nanolambda-my-func-123...
				if strings.HasPrefix(name, "/"+prefix) {
					targetID = container.ID
					break
				}
			}
			if targetID != "" {
				break
			}
		}

		if targetID == "" {
			fmt.Printf("No running container found for function '%s'.\n(Note: Containers scale to zero when idle. Invoke the function first!)\n", funcName)
			return
		}

		fmt.Printf("Streaming logs for container %s...\n", targetID[:12])
		
		// 2. Stream logs using docker cli (easier for streaming to stdout)
		c := exec.Command("docker", "logs", "-f", targetID)
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		
		if err := c.Run(); err != nil {
			// Ignore error if user Ctrl+C
			return
		}
	},
}

func init() {
	rootCmd.AddCommand(logsCmd)
}
