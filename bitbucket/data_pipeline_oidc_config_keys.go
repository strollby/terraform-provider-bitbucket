package bitbucket

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataPipelineOidcConfigKeys() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataReadPipelineOidcConfigKeys,

		Schema: map[string]*schema.Schema{
			"workspace": {
				Type:     schema.TypeString,
				Required: true,
			},
			"keys": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
		},
	}
}

func dataReadPipelineOidcConfigKeys(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(Clients).httpClient

	workspace := d.Get("workspace").(string)
	req, err := c.Get(fmt.Sprintf("2.0/workspaces/%s/pipelines-config/identity/oidc/keys.json", workspace))
	if err != nil {
		return diag.FromErr(err)
	}

	if req.StatusCode == http.StatusNotFound {
		return diag.Errorf("user not found")
	}

	if req.StatusCode >= http.StatusInternalServerError {
		return diag.Errorf("internal server error fetching user")
	}

	body, readerr := io.ReadAll(req.Body)
	if readerr != nil {
		return diag.FromErr(readerr)
	}

	log.Printf("[DEBUG] Pipeline Oidc Config Keys Response JSON: %v", string(body))

	d.SetId(workspace)
	d.Set("workspace", workspace)
	d.Set("keys", string(body))

	return nil
}
