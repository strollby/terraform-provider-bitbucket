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

func dataPipelineOidcConfig() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataReadPipelineOidcConfig,

		Schema: map[string]*schema.Schema{
			"workspace": {
				Type:     schema.TypeString,
				Required: true,
			},
			"oidc_config": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataReadPipelineOidcConfig(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(Clients).httpClient

	workspace := d.Get("workspace").(string)
	req, err := c.Get(fmt.Sprintf("2.0/workspaces/%s/pipelines-config/identity/oidc/.well-known/openid-configuration", workspace))
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

	log.Printf("[DEBUG] Pipeline Oidc Config Response JSON: %v", string(body))

	d.SetId(workspace)
	d.Set("workspace", workspace)
	d.Set("oidc_config", string(body))

	return nil
}
