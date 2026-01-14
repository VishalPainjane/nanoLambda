package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init [name]",
	Short: "Initialize a new function",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		runtime, _ := cmd.Flags().GetString("runtime")

		if runtime != "python" {
			fmt.Println("only python runtime is supported in this demo")
			return
		}

		// create directory
		if err := os.Mkdir(name, 0755); err != nil {
			fmt.Printf("error creating directory: %v\n", err)
			return
		}

		// create handler.py
		handlerCode := `def handle(request):
    return {"message": "Hello from NanoLambda!", "input": request}
`
		if err := os.WriteFile(filepath.Join(name, "handler.py"), []byte(handlerCode), 0644); err != nil {
			fmt.Printf("error creating handler.py: %v\n", err)
			return
		}

		// create dockerfile (dynamic generation based on runtime)
		dockerfile := `FROM nanolambda/base-python:3.11

COPY handler.py /function/handler.py
# If you have requirements.txt, uncomment:
# COPY requirements.txt /function/requirements.txt
# RUN pip install -r /function/requirements.txt
`
		if err := os.WriteFile(filepath.Join(name, "Dockerfile"), []byte(dockerfile), 0644); err != nil {
			fmt.Printf("error creating dockerfile: %v\n", err)
			return
		}

		// create nanolambda.yaml (metadata)
		yamlContent := fmt.Sprintf("name: %s\nruntime: %s\n", name, runtime)
		if err := os.WriteFile(filepath.Join(name, "nanolambda.yaml"), []byte(yamlContent), 0644); err != nil {
			fmt.Printf("error creating nanolambda.yaml: %v\n", err)
			return
		}

		fmt.Printf("initialized function '%s' in ./%s\n", name, name)
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().String("runtime", "python", "Runtime environment (python, nodejs)")
}