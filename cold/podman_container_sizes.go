package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
)

type PodmanContainer struct {
	ID         string `json:"Id"`
	Names      []string `json:"Names"`
	Image      string `json:"Image"`
	Size       struct {
		RootFsSize int64 `json:"RootFsSize"`
		RwSize     int64 `json:"RwSize"`
	} `json:"Size"`
}

func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func main() {
	cmd := exec.Command("podman", "ps", "--size", "--format", "json")
	output, err := cmd.Output()
	if err != nil {
		log.Fatalf("Failed to run podman: %v", err)
	}

	var containers []PodmanContainer
	if err := json.Unmarshal(output, &containers); err != nil {
		log.Fatalf("Failed to parse JSON: %v", err)
	}

	if len(containers) == 0 {
		fmt.Println("No running containers found.")
		return
	}

	for _, c := range containers {
		fmt.Printf("Container: %s\n", c.Names[0])
		fmt.Printf("  Image: %s\n", c.Image)
		fmt.Printf("  Read/Write Layer Size: %s\n", formatSize(c.Size.RwSize))
		fmt.Printf("  Root Filesystem Size:  %s\n", formatSize(c.Size.RootFsSize))
		fmt.Println()
	}
}
