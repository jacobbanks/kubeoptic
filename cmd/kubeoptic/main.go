package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"kubeoptic/internal/models"
	"kubeoptic/internal/services"
	"kubeoptic/internal/tui"
	"kubeoptic/internal/tui/components"
	"kubeoptic/internal/tui/styles"
)

func main() {
	// Parse command line flags
	configPath := flag.String("config", "", "path to kubeconfig file")
	debug := flag.Bool("debug", false, "enable debug mode (skip TUI)")
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

	// Debug mode - print information and exit
	if *debug {
		fmt.Printf("kubeoptic initialized successfully!\n")
		fmt.Printf("Current Context: %s\n", kubeoptic.GetSelectedContext())
		fmt.Printf("Available Contexts: ")
		for i, ctx := range kubeoptic.GetContexts() {
			if i > 0 {
				fmt.Print(", ")
			}
			fmt.Print(ctx.Name)
		}
		fmt.Printf("\n")
		fmt.Printf("Available Namespaces: ")
		for i, ns := range kubeoptic.GetNamespaces() {
			if i > 0 {
				fmt.Print(", ")
			}
			fmt.Print(ns.Name)
		}
		fmt.Printf("\n")

		// Test TUI component rendering
		fmt.Printf("\n=== TUI Component Test ===\n")
		contexts := kubeoptic.GetContexts()
		contextList := components.NewContextList(contexts, kubeoptic.GetSelectedContext())
		fmt.Printf("ContextList View:\n%s\n", contextList.View())
		return
	}

	// Launch TUI
	app := tui.NewApp(kubeoptic)

	// Create TUI components
	theme := styles.DefaultTheme()

	// Create context list with available contexts
	contexts := kubeoptic.GetContexts()
	contextList := components.NewContextList(contexts, kubeoptic.GetSelectedContext())

	// Create namespace list
	namespaceList := components.NewNamespaceList(kubeoptic)

	// Create pod list (initially empty, will be populated when namespace is selected)
	podList := components.NewPodList([]services.Pod{}, 0, 0)

	// Create log view
	logView := components.NewLogViewer(kubeoptic, 0, 0)

	// Create status bar
	statusBar := components.NewStatusBar(theme, kubeoptic)

	// Connect components to the app
	app.SetComponents(contextList, namespaceList, podList, logView, statusBar)

	// Create and run the Bubble Tea program
	program := tea.NewProgram(
		app,
		tea.WithAltScreen(),       // Use alternate screen buffer
		tea.WithMouseCellMotion(), // Enable mouse support
	)

	// Run the program
	if _, err := program.Run(); err != nil {
		log.Fatalf("Error running TUI: %v", err)
	}
}
