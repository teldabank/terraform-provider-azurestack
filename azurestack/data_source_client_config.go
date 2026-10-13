package azurestack

import (
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func dataSourceArmClientConfig() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceArmClientConfigRead,

		Schema: map[string]*schema.Schema{
			"client_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tenant_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"subscription_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"service_principal_application_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"service_principal_object_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceArmClientConfigRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient)

	d.SetId(time.Now().UTC().String())
	d.Set("client_id", client.clientId)
	d.Set("tenant_id", client.tenantId)
	d.Set("subscription_id", client.subscriptionId)

	// The object ID is read from the ARM token's `oid` claim at provider
	// configuration time rather than by querying the Graph API: the service
	// principal used to run Terraform does not have permission to list service
	// principals on this Azure Stack Hub stamp (Graph returns 403).
	if client.usingServicePrincipal && client.authenticatedObjectId != "" {
		d.Set("service_principal_application_id", client.clientId)
		d.Set("service_principal_object_id", client.authenticatedObjectId)
	}

	return nil
}
