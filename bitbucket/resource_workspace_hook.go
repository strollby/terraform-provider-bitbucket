package bitbucket

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceWorkspaceHook() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceWorkspaceHookCreate,
		ReadWithoutTimeout:   resourceWorkspaceHookRead,
		UpdateWithoutTimeout: resourceWorkspaceHookUpdate,
		DeleteWithoutTimeout: resourceWorkspaceHookDelete,
		Importer: &schema.ResourceImporter{
			State: func(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				idParts := strings.Split(d.Id(), "/")
				if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
					return nil, fmt.Errorf("unexpected format of ID (%q), expected workspace/REPO/HOOK-ID", d.Id())
				}
				d.SetId(idParts[1])
				d.Set("workspace", idParts[0])
				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"workspace": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"active": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"url": {
				Type:     schema.TypeString,
				Required: true,
			},
			"uuid": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Required: true,
			},
			"events": {
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
					ValidateFunc: validation.StringInSlice([]string{
						"pullrequest:unapproved",
						"issue:comment_created",
						"repo:imported",
						"repo:created",
						"repo:commit_comment_created",
						"pullrequest:approved",
						"pullrequest:comment_updated",
						"issue:updated",
						"project:updated",
						"repo:deleted",
						"pullrequest:changes_request_created",
						"pullrequest:comment_created",
						"repo:commit_status_updated",
						"pullrequest:updated",
						"issue:created",
						"repo:fork",
						"pullrequest:comment_deleted",
						"repo:commit_status_created",
						"repo:updated",
						"pullrequest:rejected",
						"pullrequest:fulfilled",
						"pullrequest:created",
						"pullrequest:changes_request_removed",
						"repo:transfer",
						"repo:push",
					}, false),
				},
			},
			"skip_cert_verification": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
		},
	}
}

func resourceWorkspaceHookCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(Clients).httpClient
	hook := createHook(d)

	payload, err := json.Marshal(hook)
	if err != nil {
		return diag.FromErr(err)
	}

	hookReq, err := client.Post(fmt.Sprintf("2.0/workspaces/%s/hooks",
		d.Get("workspace").(string),
	), bytes.NewBuffer(payload))

	if err != nil {
		return diag.FromErr(err)
	}

	body, readerr := io.ReadAll(hookReq.Body)
	if readerr != nil {
		return diag.FromErr(readerr)
	}

	decodeerr := json.Unmarshal(body, &hook)
	if decodeerr != nil {
		return diag.FromErr(decodeerr)
	}

	d.SetId(hook.UUID)

	return resourceWorkspaceHookRead(ctx, d, m)
}
func resourceWorkspaceHookRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(Clients).httpClient

	hookReq, err := client.Get(fmt.Sprintf("2.0/workspaces/%s/hooks/%s",
		d.Get("workspace").(string),
		url.PathEscape(d.Id()),
	))

	if hookReq.StatusCode == http.StatusNotFound {
		log.Printf("[WARN] Repository Hook (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("ID: %s", url.PathEscape(d.Id()))

	if hookReq.StatusCode == 200 {
		var hook Hook

		body, readerr := io.ReadAll(hookReq.Body)
		if readerr != nil {
			return diag.FromErr(readerr)
		}

		decodeerr := json.Unmarshal(body, &hook)
		if decodeerr != nil {
			return diag.FromErr(decodeerr)
		}

		d.Set("uuid", hook.UUID)
		d.Set("description", hook.Description)
		d.Set("active", hook.Active)
		d.Set("url", hook.URL)
		d.Set("skip_cert_verification", hook.SkipCertVerification)
		d.Set("events", hook.Events)
	}

	return nil
}

func resourceWorkspaceHookUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(Clients).httpClient
	hook := createHook(d)
	payload, err := json.Marshal(hook)
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = client.Put(fmt.Sprintf("2.0/workspaces/%s/hooks/%s",
		d.Get("workspace").(string),
		url.PathEscape(d.Id()),
	), bytes.NewBuffer(payload))

	if err != nil {
		return diag.FromErr(err)
	}

	return resourceWorkspaceHookRead(ctx, d, m)
}

func resourceWorkspaceHookDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(Clients).httpClient
	_, err := client.Delete(fmt.Sprintf("2.0/workspaces/%s/hooks/%s",
		d.Get("workspace").(string),
		url.PathEscape(d.Id()),
	))

	return diag.FromErr(err)

}
