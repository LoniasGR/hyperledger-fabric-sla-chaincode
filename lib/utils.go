package lib

import (
	"fmt"
	"os"
	"strconv"
)



// setDiscoveryAsLocalhost sets the environmental variable DISCOVERY_AS_LOCALHOST
func SetDiscoveryAsLocalhost(value bool) error {
	err := os.Setenv("DISCOVERY_AS_LOCALHOST", strconv.FormatBool(value))
	if err != nil {
		return fmt.Errorf("failed to set DISCOVERY_AS_LOCALHOST environment variable: %v", err)
	}
	return nil
}
