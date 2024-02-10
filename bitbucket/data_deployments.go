package bitbucket

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataDeployments() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataReadDeployments,

		Schema: map[string]*schema.Schema{
			"uuids": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},
			"workspace": {
				Type:     schema.TypeString,
				Required: true,
			},
			"repository": {
				Type:     schema.TypeString,
				Required: true,
			},
			"names": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},
		},
	}
}

func dataReadDeployments(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(Clients).genClient
	deployApi := c.ApiClient.DeploymentsApi

	workspace := d.Get("workspace").(string)
	repoId := d.Get("repository").(string)

	deploymentsResp, _, err := deployApi.GetEnvironmentsForRepository(c.AuthContext, workspace, repoId, nil)
	if err := handleClientError(err); err != nil {
		return diag.FromErr(err)
	}

	deployments := deploymentsResp.Values

	log.Printf("haha %#v", deployments)

	var uuids []string
	for _, deployment := range deployments {
		uuids = append(uuids, deployment.Uuid)
	}

	var names []string
	for _, deployment := range deployments {
		names = append(names, deployment.Name)
	}

	d.SetId(fmt.Sprintf("%s/%s", workspace, repoId))
	d.Set("uuids", uuids)
	d.Set("names", names)

	return nil
}
