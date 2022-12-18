package bitbucket

import (
	"context"
	"log"
	"net/http"

	"github.com/DrFaust92/bitbucket-go-client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataHookTypes() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataReadHookTypes,

		Schema: map[string]*schema.Schema{
			"subject_type": {
				Type:     schema.TypeString,
				Required: true,
			},
			"hook_types": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"event": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"category": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"label": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"description": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataReadHookTypes(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(Clients).genClient
	webhooksApi := c.ApiClient.WebhooksApi

	subjectType := d.Get("subject_type").(string)
	hookTypes, res, err := webhooksApi.HookEventsSubjectTypeGet(c.AuthContext, subjectType)
	if err != nil {
		return diag.FromErr(err)
	}

	if res.StatusCode == http.StatusNotFound {
		return diag.Errorf("user not found")
	}

	if res.StatusCode >= http.StatusInternalServerError {
		return diag.Errorf("internal server error fetching hook types")
	}

	d.SetId(subjectType)
	d.Set("hook_types", flattenHookTypes(hookTypes.Values))

	return nil
}

func flattenHookTypes(hookTypes []bitbucket.HookEvent) []interface{} {
	if len(hookTypes) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, btRaw := range hookTypes {
		log.Printf("[DEBUG] HookType Response Decoded: %#v", btRaw)

		hookType := map[string]interface{}{
			"event":       btRaw.Event,
			"category":    btRaw.Category,
			"label":       btRaw.Label,
			"description": btRaw.Description,
		}

		tfList = append(tfList, hookType)
	}

	return tfList
}
