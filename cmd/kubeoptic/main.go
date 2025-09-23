package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"kubeoptic/internal/models"
	"kubeoptic/internal/services"
)

func main() {
	// Parse command line flags
	configPath := flag.String("config", "", "path to kubeconfig file")
	flag.Parse()

	// Initialize services
	configSvc := services.NewConfigService()

	// Auto-discover config if not provided
	var kubeConfigPath string
	if *configPath != "" {
		kubeConfigPath = *configPath
	} else {
		discoveredPath, err := configSvc.DiscoverConfig()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: No kubeconfig found. Please specify:\n")
			fmt.Fprintf(os.Stderr, "  kubeoptic --config /path/to/kubeconfig\n\n")
			fmt.Fprintf(os.Stderr, "Searched:\n")
			fmt.Fprintf(os.Stderr, "  - $KUBECONFIG environment variable\n")
			fmt.Fprintf(os.Stderr, "  - ~/.kube/config\n")
			fmt.Fprintf(os.Stderr, "  - In-cluster config\n")
			os.Exit(1)
		}
		kubeConfigPath = discoveredPath
	}

	// Create kubeoptic coordinator
	kubeoptic := models.NewKubeoptic(
		configSvc,
		nil, // Will be set after loading config
		nil, // Will be set after loading config
	)

	// Load kubernetes configuration
	err := kubeoptic.LoadContexts(kubeConfigPath)
	if err != nil {
		log.Fatalf("Failed to load kubeconfig: %v", err)
	}

	// Print current state (temporary - will be replaced by TUI)
	fmt.Printf("kubeoptic initialized successfully!\n")
	fmt.Printf("Current Context: %s\n", kubeoptic.GetSelectedContext())
	fmt.Printf("Available Contexts: ")
	for i, ctx := range kubeoptic.GetContexts() {
		if i > 0 {
			fmt.Print(", ")
		}
		fmt.Print(ctx.Name)
	}
	fmt.Printf("\nAvailable Namespaces: ")
	for i, ns := range kubeoptic.GetNamespaces() {
		if i > 0 {
			fmt.Print(", ")
		}
		fmt.Print(ns)
	}
	fmt.Printf("\n")

	// TODO: Launch TUI here
	fmt.Println("\nTUI not implemented yet - this is just the architecture demo!")
}
