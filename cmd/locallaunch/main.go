package main

import (
	"fmt"
	"log"
	"os"

	"locallaunch/internal/api"
	"locallaunch/internal/config"
)

const version = "0.1.0"

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "version":
			fmt.Printf("locallaunch %s\n", version)
			return
		case "token":
			cfg, err := config.Load()
			if err != nil {
				log.Fatalf("Error loading config: %v\n", err)
			}
			fmt.Println(cfg.Token)
			return
		default:
			fmt.Fprintf(os.Stderr, "Unknown command: %s\n", os.Args[1])
			fmt.Fprintf(os.Stderr, "Usage: locallaunch [version|token]\n")
			os.Exit(1)
		}
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Error loading config: %v\n", err)
	}

	cfgPath, _ := config.Dir()
	log.Printf("LocalLaunch started\n\nConfig:\n%s/config.json\n\n", cfgPath)

	srv := api.New(cfg)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("Server error: %v\n", err)
	}
}
