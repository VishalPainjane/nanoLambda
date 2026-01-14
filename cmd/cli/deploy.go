package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/nikhi/nanolambda/pkg/registry"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

type FunctionConfig struct {
	Name    string `yaml:"name"`
	Runtime string `yaml:"runtime"`
	Timeout int    `yaml:"timeout"`
}

var deployCmd = &cobra.Command{
	Use:   "deploy [path]",
	Short: "Deploy a function",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		path := "."
		if len(args) > 0 {
			path = args[0]
		}

		// 1. Read config
		configFile := filepath.Join(path, "nanolambda.yaml")
		data, err := os.ReadFile(configFile)
		if err != nil {
			fmt.Printf("Error reading nanolambda.yaml: %v\n", err)
			return
		}

		var config FunctionConfig
		if err := yaml.Unmarshal(data, &config); err != nil {
			fmt.Printf("Error parsing nanolambda.yaml: %v\n", err)
			return
		}

		fmt.Printf("Deploying function '%s'...\n", config.Name)

		// 2. Build Docker Image
		imageTag := fmt.Sprintf("nanolambda/%s:latest", config.Name)
		fmt.Printf("Building image %s...\n", imageTag)
		
		buildCmd := exec.Command("docker", "build", "-t", imageTag, path)
		buildCmd.Stdout = os.Stdout
		buildCmd.Stderr = os.Stderr
		if err := buildCmd.Run(); err != nil {
			fmt.Printf("Docker build failed: %v\n", err)
			return
		}

		// 3. Register Function
		// Assuming we share the DB file with Gateway (single machine)
		// We need to know where the DB is. Let's assume ./data/nanolambda.db relative to where CLI is run?
		// Or strictly use an absolute path or flag. For demo: ./data/nanolambda.db
		os.MkdirAll("./data", 0755)
		reg, err := registry.NewManager("./data/nanolambda.db")
		if err != nil {
			fmt.Printf("Error connecting to registry: %v\n", err)
			return
		}
		defer reg.Close()

		fn := registry.Function{
			Name:        config.Name,
			Runtime:     config.Runtime,
			ImageTag:    imageTag,
			CreatedAt:   time.Now(),
			MemoryLimit: 128, // Default
			Timeout:     config.Timeout,
		}

		if err := reg.RegisterFunction(fn); err != nil {
			fmt.Printf("Error registering function: %v\n", err)
			return
		}

		fmt.Println("Function deployed successfully!")
	},
}

func init() {
	rootCmd.AddCommand(deployCmd)
}
