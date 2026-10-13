package validate

import (
	"fmt"

	keyVaultParse "github.com/terraform-providers/terraform-provider-azurestack/azurestack/helpers/keyvault/parse"
)

// stringIsNotEmpty mirrors the SDK's validation.StringIsNotEmpty, which is not
// available in the terraform-plugin-sdk version this provider is pinned to.
func stringIsNotEmpty(i interface{}, k string) (warnings []string, errors []error) {
	v, ok := i.(string)
	if !ok {
		errors = append(errors, fmt.Errorf("expected type of %q to be string", k))
		return
	}

	if v == "" {
		errors = append(errors, fmt.Errorf("%q must not be empty", k))
	}

	return warnings, errors
}

func NestedItemId(i interface{}, k string) (warnings []string, errors []error) {
	if warnings, errors = stringIsNotEmpty(i, k); len(errors) > 0 {
		return warnings, errors
	}

	v, ok := i.(string)
	if !ok {
		errors = append(errors, fmt.Errorf("Expected %s to be a string!", k))
		return warnings, errors
	}

	if _, err := keyVaultParse.ParseNestedItemID(v); err != nil {
		errors = append(errors, fmt.Errorf("parsing %q: %s", v, err))
		return warnings, errors
	}

	return warnings, errors
}

func VersionlessNestedItemId(i interface{}, k string) (warnings []string, errors []error) {
	if warnings, errors = stringIsNotEmpty(i, k); len(errors) > 0 {
		return warnings, errors
	}

	v, ok := i.(string)
	if !ok {
		errors = append(errors, fmt.Errorf("Expected %s to be a string!", k))
		return warnings, errors
	}

	id, err := keyVaultParse.ParseOptionallyVersionedNestedItemID(v)
	if err != nil {
		errors = append(errors, fmt.Errorf("parsing %q: %s", v, err))
		return warnings, errors
	}

	if id.Version != "" {
		errors = append(errors, fmt.Errorf("expected %s to not have a version", k))
	}

	return warnings, errors
}

func NestedItemIdWithOptionalVersion(i interface{}, k string) (warnings []string, errors []error) {
	if warnings, errors = stringIsNotEmpty(i, k); len(errors) > 0 {
		return warnings, errors
	}

	v, ok := i.(string)
	if !ok {
		errors = append(errors, fmt.Errorf("Expected %s to be a string!", k))
		return warnings, errors
	}

	if _, err := keyVaultParse.ParseOptionallyVersionedNestedItemID(v); err != nil {
		errors = append(errors, fmt.Errorf("parsing %q: %s", v, err))
		return warnings, errors
	}

	return warnings, errors
}
