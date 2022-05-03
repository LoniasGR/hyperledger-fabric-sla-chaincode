package lib

import (
	"fmt"
	"os"
	"strconv"
)

const ColorReset = "\033[0m"
const ColorBlue = "\033[34m"
const ColorRed = "\033[31m"
const ColorGreen = "\033[32m"
const ColorCyan = "\033[36m"

// setDiscoveryAsLocalhost sets the environmental variable DISCOVERY_AS_LOCALHOST
func SetDiscoveryAsLocalhost(value bool) error {
	err := os.Setenv("DISCOVERY_AS_LOCALHOST", strconv.FormatBool(value))
	if err != nil {
		return fmt.Errorf("failed to set DISCOVERY_AS_LOCALHOST environment variable: %v", err)
	}
	return nil
}
