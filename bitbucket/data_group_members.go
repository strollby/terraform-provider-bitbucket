package bitbucket

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataGroupMembers() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataReadGroupMembers,

		Schema: map[string]*schema.Schema{
			"workspace": {
				Type:     schema.TypeString,
				Required: true,
			},
			"slug": {
				Type:     schema.TypeString,
				Required: true,
			},
			"members": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},
		},
	}
}

func dataReadGroupMembers(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(Clients).httpClient

	workspace := d.Get("workspace").(string)
	slug := d.Get("slug").(string)

	groupsReq, _ := client.Get(fmt.Sprintf("1.0/groups/%s/%s/members", workspace, slug))

	if groupsReq.Body == nil {
		return diag.Errorf("error reading Group (%s): empty response", d.Id())
	}

	var members []*UserGroupMembership

	body, readerr := io.ReadAll(groupsReq.Body)
	if readerr != nil {
		return diag.FromErr(readerr)
	}

	log.Printf("[DEBUG] Group Membership Response JSON: %v", string(body))

	decodeerr := json.Unmarshal(body, &members)
	if decodeerr != nil {
		return diag.FromErr(decodeerr)
	}

	log.Printf("[DEBUG] Group Membership Response Decoded: %#v", members)

	var mems []string
	for _, mbr := range members {
		mems = append(mems, mbr.UUID)
	}

	d.SetId(fmt.Sprintf("%s/%s", workspace, slug))
	d.Set("workspace", workspace)
	d.Set("slug", slug)
	d.Set("members", mems)

	return nil
}
