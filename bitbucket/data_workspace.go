package bitbucket

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataWorkspace() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataReadWorkspace,

		Schema: map[string]*schema.Schema{
			"workspace": {
				Type:     schema.TypeString,
				Required: true,
			},
			"slug": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"is_private": {
				Type:     schema.TypeBool,
				Computed: true,
			},
		},
	}
}

func dataReadWorkspace(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(Clients).genClient

	workspaceApi := c.ApiClient.WorkspacesApi

	workspace := d.Get("workspace").(string)
	workspaceReq, res, err := workspaceApi.WorkspacesWorkspaceGet(c.AuthContext, workspace)
	if err := handleClientError(res, err); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(workspaceReq.Uuid)
	d.Set("workspace", workspace)
	d.Set("name", workspaceReq.Name)
	d.Set("slug", workspaceReq.Slug)
	d.Set("is_private", workspaceReq.IsPrivate)

	return nil
}
