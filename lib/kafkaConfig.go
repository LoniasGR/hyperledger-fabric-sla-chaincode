package lib

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
)

// ParseArgs parses the command line arguments and
// returns the config file on success, or exits on error
func ParseArgs() *string {
	configFile := flag.String("f", "", "Path to Kafka configuration file")
	if *configFile == "" {
		flag.Usage()
		os.Exit(2) // the same exit code flag.Parse uses
	}

	return configFile
}

// ReadConfig reads the file specified by configFile and
// creates a map of key-value pairs that correspond to each
// line of the file. ReadConfig returns the map on success,
// or nil and an error
func ReadConfig(configFile string) (map[string]string, error) {
	m := make(map[string]string)

	file, err := os.Open(configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "#") && len(line) != 0 {
			kv := strings.Split(line, "=")
			parameter := strings.TrimSpace(kv[0])
			value := strings.TrimSpace(kv[1])
			m[parameter] = value
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read file: %v", err)
	}

	return m, nil
}
