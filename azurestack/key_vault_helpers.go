package azurestack

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/helpers/azure"
)

// wasNotFound mirrors the upstream keyvault helper: the data-plane client
// returns a bare *http.Response rather than an autorest.Response.
func wasNotFound(resp *http.Response) bool {
	return resp != nil && resp.StatusCode == http.StatusNotFound
}

// hashIPv4AddressOrCIDR is used as a Set hash for CIDR blocks, where the same
// value should hash the same regardless of case.
func hashIPv4AddressOrCIDR(v interface{}) int {
	return schema.HashString(strings.ToLower(v.(string)))
}

type subnetId struct {
	SubscriptionId     string
	ResourceGroup      string
	VirtualNetworkName string
	Name               string
}

func (id subnetId) ID() string {
	return fmt.Sprintf("/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Network/virtualNetworks/%s/subnets/%s", id.SubscriptionId, id.ResourceGroup, id.VirtualNetworkName, id.Name)
}

// subnetIDInsensitively parses a subnet resource ID and pulls out the virtual
// network and subnet names, matching the segments case-insensitively (ARM is
// not consistent about the casing it returns).
func subnetIDInsensitively(input string) (*subnetId, error) {
	id, err := azure.ParseAzureResourceID(input)
	if err != nil {
		return nil, err
	}

	resourceId := subnetId{
		SubscriptionId: id.SubscriptionID,
		ResourceGroup:  id.ResourceGroup,
	}

	if resourceId.SubscriptionId == "" {
		return nil, fmt.Errorf("ID was missing the 'subscriptions' element")
	}

	if resourceId.ResourceGroup == "" {
		return nil, fmt.Errorf("ID was missing the 'resourceGroups' element")
	}

	resourceId.VirtualNetworkName = pathValueInsensitively(id.Path, "virtualNetworks")
	if resourceId.VirtualNetworkName == "" {
		return nil, fmt.Errorf("ID was missing the 'virtualNetworks' element")
	}

	resourceId.Name = pathValueInsensitively(id.Path, "subnets")
	if resourceId.Name == "" {
		return nil, fmt.Errorf("ID was missing the 'subnets' element")
	}

	return &resourceId, nil
}

func pathValueInsensitively(path map[string]string, key string) string {
	for k, v := range path {
		if strings.EqualFold(k, key) {
			return v
		}
	}
	return ""
}

func azureStackNormalizeLocationNilable(input *string) string {
	if input == nil {
		return ""
	}
	return azureStackNormalizeLocation(*input)
}
