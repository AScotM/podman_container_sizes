package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"sort"
	"strings" // This was missing in the improved version
)

type PodmanContainer struct {
	ID     string   `json:"Id"`
	Names  []string `json:"Names"`
	Image  string   `json:"Image"`
	Status string   `json:"Status"`
	Size   struct {
		RootFsSize int64 `json:"RootFsSize"`
		RwSize     int64 `json:"RwSize"`
	} `json:"Size"`
}

type Config struct {
	showAll    bool
	sortBy     string
	outputJSON bool
}

func main() {
	cfg := parseFlags()
	containers, err := getContainers(cfg.showAll)
	if err != nil {
		log.Fatalf("Error getting containers: %v", err)
	}

	if len(containers) == 0 {
		fmt.Println("No containers found.")
		return
	}

	sortContainers(containers, cfg.sortBy)

	if cfg.outputJSON {
		printJSON(containers)
	} else {
		printTable(containers)
	}
}

func parseFlags() Config {
	var cfg Config
	flag.BoolVar(&cfg.showAll, "a", false, "Show all containers (default shows just running)")
	flag.StringVar(&cfg.sortBy, "sort", "name", "Sort by (name, size, rwsize)")
	flag.BoolVar(&cfg.outputJSON, "json", false, "Output in JSON format")
	flag.Parse()
	return cfg
}

func getContainers(showAll bool) ([]PodmanContainer, error) {
	args := []string{"ps", "--size", "--format", "json"}
	if showAll {
		args = append(args, "-a")
	}

	cmd := exec.Command("podman", args...)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("podman command failed: %w", err)
	}

	var containers []PodmanContainer
	if err := json.Unmarshal(output, &containers); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return containers, nil
}

func sortContainers(containers []PodmanContainer, sortBy string) {
	switch sortBy {
	case "size":
		sort.Slice(containers, func(i, j int) bool {
			return (containers[i].Size.RootFsSize + containers[i].Size.RwSize) > 
			       (containers[j].Size.RootFsSize + containers[j].Size.RwSize)
		})
	case "rwsize":
		sort.Slice(containers, func(i, j int) bool {
			return containers[i].Size.RwSize > containers[j].Size.RwSize
		})
	default: // name
		sort.Slice(containers, func(i, j int) bool {
			name1 := getContainerName(containers[i])
			name2 := getContainerName(containers[j])
			return name1 < name2
		})
	}
}

func getContainerName(c PodmanContainer) string {
	if len(c.Names) > 0 {
		return c.Names[0]
	}
	return "<unnamed>"
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

func printTable(containers []PodmanContainer) {
	fmt.Printf("%-20s %-30s %-12s %-12s %s\n", 
		"NAME", "IMAGE", "RW SIZE", "ROOT SIZE", "TOTAL SIZE")
	fmt.Println(strings.Repeat("-", 90))

	var totalRw, totalRoot int64

	for _, c := range containers {
		name := getContainerName(c)
		rwSize := formatSize(c.Size.RwSize)
		rootSize := formatSize(c.Size.RootFsSize)
		totalSize := formatSize(c.Size.RwSize + c.Size.RootFsSize)

		fmt.Printf("%-20s %-30s %-12s %-12s %s\n", 
			name, c.Image, rwSize, rootSize, totalSize)

		totalRw += c.Size.RwSize
		totalRoot += c.Size.RootFsSize
	}

	fmt.Println("\nTOTAL:")
	fmt.Printf("Read/Write: %s\n", formatSize(totalRw))
	fmt.Printf("Root FS:    %s\n", formatSize(totalRoot))
	fmt.Printf("Combined:   %s\n", formatSize(totalRw + totalRoot))
}

func printJSON(containers []PodmanContainer) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(containers); err != nil {
		log.Fatalf("Error encoding JSON: %v", err)
	}
}
