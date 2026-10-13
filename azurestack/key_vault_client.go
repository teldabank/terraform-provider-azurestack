package azurestack

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"sync"

	"github.com/Azure/azure-sdk-for-go/profiles/2017-03-09/resources/mgmt/resources"
	keyvaultdata "github.com/Azure/azure-sdk-for-go/services/keyvault/2016-10-01/keyvault"
	keyvaultmgmt "github.com/Azure/azure-sdk-for-go/services/keyvault/mgmt/2019-09-01/keyvault"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/helpers/keyvault/parse"
)

// keyVaultClient holds the management-plane client used to manage vaults and the
// data-plane client used to manage the items inside them.
type keyVaultClient struct {
	ManagementClient *keyvaultdata.BaseClient
	VaultsClient     *keyvaultmgmt.VaultsClient
}

var (
	keyVaultsCache = map[string]keyVaultDetails{}
	keysmith       = &sync.RWMutex{}
	keyVaultLocks  = map[string]*sync.RWMutex{}
)

type keyVaultDetails struct {
	keyVaultId       string
	dataPlaneBaseUri string
	resourceGroup    string
}

func (c *keyVaultClient) AddToCache(keyVaultId parse.VaultId, dataPlaneUri string) {
	cacheKey := c.cacheKeyForKeyVault(keyVaultId.Name)
	keysmith.Lock()
	keyVaultsCache[cacheKey] = keyVaultDetails{
		keyVaultId:       keyVaultId.ID(),
		dataPlaneBaseUri: dataPlaneUri,
		resourceGroup:    keyVaultId.ResourceGroup,
	}
	keysmith.Unlock()
}

func (c *keyVaultClient) BaseUriForKeyVault(ctx context.Context, keyVaultId parse.VaultId) (*string, error) {
	cacheKey := c.cacheKeyForKeyVault(keyVaultId.Name)
	keysmith.Lock()
	if keyVaultLocks[cacheKey] == nil {
		keyVaultLocks[cacheKey] = &sync.RWMutex{}
	}
	keysmith.Unlock()
	keyVaultLocks[cacheKey].Lock()
	defer keyVaultLocks[cacheKey].Unlock()

	resp, err := c.VaultsClient.Get(ctx, keyVaultId.ResourceGroup, keyVaultId.Name)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			return nil, fmt.Errorf("%s was not found", keyVaultId)
		}
		return nil, fmt.Errorf("retrieving %s: %+v", keyVaultId, err)
	}

	if resp.Properties == nil || resp.Properties.VaultURI == nil {
		return nil, fmt.Errorf("`properties` was nil for %s", keyVaultId)
	}

	return resp.Properties.VaultURI, nil
}

func (c *keyVaultClient) Exists(ctx context.Context, keyVaultId parse.VaultId) (bool, error) {
	cacheKey := c.cacheKeyForKeyVault(keyVaultId.Name)
	keysmith.Lock()
	if keyVaultLocks[cacheKey] == nil {
		keyVaultLocks[cacheKey] = &sync.RWMutex{}
	}
	keysmith.Unlock()
	keyVaultLocks[cacheKey].Lock()
	defer keyVaultLocks[cacheKey].Unlock()

	if _, ok := keyVaultsCache[cacheKey]; ok {
		return true, nil
	}

	resp, err := c.VaultsClient.Get(ctx, keyVaultId.ResourceGroup, keyVaultId.Name)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			return false, nil
		}
		return false, fmt.Errorf("retrieving %s: %+v", keyVaultId, err)
	}

	if resp.Properties == nil || resp.Properties.VaultURI == nil {
		return false, fmt.Errorf("`properties` was nil for %s", keyVaultId)
	}

	c.AddToCache(keyVaultId, *resp.Properties.VaultURI)

	return true, nil
}

func (c *keyVaultClient) KeyVaultIDFromBaseUrl(ctx context.Context, resourcesClient resources.Client, keyVaultBaseUrl string) (*string, error) {
	keyVaultName, err := c.parseNameFromBaseUrl(keyVaultBaseUrl)
	if err != nil {
		return nil, err
	}

	cacheKey := c.cacheKeyForKeyVault(*keyVaultName)
	keysmith.Lock()
	if keyVaultLocks[cacheKey] == nil {
		keyVaultLocks[cacheKey] = &sync.RWMutex{}
	}
	keysmith.Unlock()
	keyVaultLocks[cacheKey].Lock()
	defer keyVaultLocks[cacheKey].Unlock()

	if v, ok := keyVaultsCache[cacheKey]; ok {
		return &v.keyVaultId, nil
	}

	filter := fmt.Sprintf("resourceType eq 'Microsoft.KeyVault/vaults' and name eq '%s'", *keyVaultName)
	result, err := resourcesClient.List(ctx, filter, "", utils.Int32(5))
	if err != nil {
		return nil, fmt.Errorf("listing resources matching %q: %+v", filter, err)
	}

	for result.NotDone() {
		for _, v := range result.Values() {
			if v.ID == nil {
				continue
			}

			id, err := parse.VaultID(*v.ID)
			if err != nil {
				return nil, fmt.Errorf("parsing %q: %+v", *v.ID, err)
			}
			if !strings.EqualFold(id.Name, *keyVaultName) {
				continue
			}

			props, err := c.VaultsClient.Get(ctx, id.ResourceGroup, id.Name)
			if err != nil {
				return nil, fmt.Errorf("retrieving %s: %+v", *id, err)
			}
			if props.Properties == nil || props.Properties.VaultURI == nil {
				return nil, fmt.Errorf("retrieving %s: `properties.VaultUri` was nil", *id)
			}

			c.AddToCache(*id, *props.Properties.VaultURI)
			return utils.String(id.ID()), nil
		}

		if err := result.NextWithContext(ctx); err != nil {
			return nil, fmt.Errorf("iterating over results: %+v", err)
		}
	}

	// we haven't found it, but Data Sources and Resources need to handle this error separately
	return nil, nil // nolint:nilnil
}

func (c *keyVaultClient) Purge(keyVaultId parse.VaultId) {
	cacheKey := c.cacheKeyForKeyVault(keyVaultId.Name)
	keysmith.Lock()
	if keyVaultLocks[cacheKey] == nil {
		keyVaultLocks[cacheKey] = &sync.RWMutex{}
	}
	keysmith.Unlock()
	keyVaultLocks[cacheKey].Lock()
	delete(keyVaultsCache, cacheKey)
	keyVaultLocks[cacheKey].Unlock()
}

func (c *keyVaultClient) cacheKeyForKeyVault(name string) string {
	return strings.ToLower(name)
}

func (c *keyVaultClient) parseNameFromBaseUrl(input string) (*string, error) {
	uri, err := url.Parse(input)
	if err != nil {
		return nil, err
	}
	// https://tharvey-keyvault.vault.azure.net/
	segments := strings.Split(uri.Host, ".")
	return &segments[0], nil
}
