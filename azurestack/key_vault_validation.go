package azurestack

import (
	"fmt"
	"net"
	"time"

	uuid "github.com/satori/go.uuid"
)

// The terraform-plugin-sdk version this provider is pinned to predates these
// validators, so they are provided locally with the same semantics as the SDK's.

func isUUID(i interface{}, k string) (warnings []string, errors []error) {
	v, ok := i.(string)
	if !ok {
		errors = append(errors, fmt.Errorf("expected type of %q to be string", k))
		return
	}

	if _, err := uuid.FromString(v); err != nil {
		errors = append(errors, fmt.Errorf("expected %q to be a valid UUID, got %v", k, v))
	}

	return warnings, errors
}

func isRFC3339Time(i interface{}, k string) (warnings []string, errors []error) {
	v, ok := i.(string)
	if !ok {
		errors = append(errors, fmt.Errorf("expected type of %q to be string", k))
		return
	}

	if _, err := time.Parse(time.RFC3339, v); err != nil {
		errors = append(errors, fmt.Errorf("expected %q to be a valid RFC3339 date/time, got %v", k, v))
	}

	return warnings, errors
}

func isIPv4Address(i interface{}, k string) (warnings []string, errors []error) {
	v, ok := i.(string)
	if !ok {
		errors = append(errors, fmt.Errorf("expected type of %q to be string", k))
		return
	}

	ip := net.ParseIP(v)
	if ip == nil || ip.To4() == nil {
		errors = append(errors, fmt.Errorf("expected %q to be a valid IPv4 address, got %v", k, v))
	}

	return warnings, errors
}

func isCIDR(i interface{}, k string) (warnings []string, errors []error) {
	v, ok := i.(string)
	if !ok {
		errors = append(errors, fmt.Errorf("expected type of %q to be string", k))
		return
	}

	if _, _, err := net.ParseCIDR(v); err != nil {
		errors = append(errors, fmt.Errorf("expected %q to be a valid CIDR, got %v", k, v))
	}

	return warnings, errors
}
