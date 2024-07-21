package sites

import (
	"fmt"
)

// GetRedirect returns the redirect URL for the given hostname
func GetRedirect(hostname string) (string, error) {
	redirectURL, ok := sites[hostname]
	if !ok {
		return "", fmt.Errorf("No redirect found for %s", hostname)
	}

	return redirectURL, nil
}
