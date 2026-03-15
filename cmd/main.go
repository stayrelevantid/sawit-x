package main

import (
	"log"
	"os"

	// Import the root package to trigger its init() and Functions Framework registration
	_ "github.com/indragiri/sawit-x"
	"github.com/GoogleCloudPlatform/functions-framework-go/funcframework"
)

func main() {
	// Use the standard Functions Framework local development server
	port := "8080"
	if p := os.Getenv("PORT"); p != "" {
		port = p
	}

	log.Printf("Starting SAWIT-X local development server on port %s", port)
	if err := funcframework.Start(port); err != nil {
		log.Fatalf("funcframework.Start: %v", err)
	}
}
