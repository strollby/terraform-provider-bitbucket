package bitbucket

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/strollby/bitbucket-go-client"
)

func resourcePipelineSchedule() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourcePipelineScheduleCreate,
		ReadWithoutTimeout:   resourcePipelineScheduleRead,
		UpdateWithoutTimeout: resourcePipelineScheduleUpdate,
		DeleteWithoutTimeout: resourcePipelineScheduleDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"workspace": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"repository": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"enabled": {
				Type:     schema.TypeBool,
				Required: true,
			},
			"cron_pattern": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"target": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ref_name": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"ref_type": {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringInSlice([]string{"branch", "tag"}, false),
						},
						"selector": {
							Type:     schema.TypeList,
							Required: true,
							ForceNew: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"type": {
										Type:     schema.TypeString,
										Optional: true,
										ForceNew: true,
										Default:  "branches",
									},
									"pattern": {
										Type:     schema.TypeString,
										Required: true,
										ForceNew: true,
									},
								},
							},
						},
					},
				},
			},
			"uuid": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourcePipelineScheduleCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(Clients).genClient
	pipeApi := c.ApiClient.PipelinesApi

	pipeSchedule := expandCreatePipelineSchedule(d)
	log.Printf("[DEBUG] Pipeline Schedule Request: %#v", pipeSchedule)

	repo := d.Get("repository").(string)
	workspace := d.Get("workspace").(string)
	schedule, _, err := pipeApi.CreateRepositoryPipelineSchedule(c.AuthContext, *pipeSchedule, workspace, repo)
	if err := handleClientError(err); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(string(fmt.Sprintf("%s/%s/%s", workspace, repo, schedule.Uuid)))

	if !d.Get("enabled").(bool) {
		pipeScheduleUpdate := expandUpdatePipelineSchedule(d)
		_, _, err = pipeApi.UpdateRepositoryPipelineSchedule(c.AuthContext, *pipeScheduleUpdate, workspace, repo, schedule.Uuid)
		if err := handleClientError(err); err != nil {
			return diag.FromErr(err)
		}
	}

	return resourcePipelineScheduleRead(ctx, d, m)
}

func resourcePipelineScheduleUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(Clients).genClient
	pipeApi := c.ApiClient.PipelinesApi

	workspace, repo, uuid, err := pipeScheduleId(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	pipeScheduleUpdate := expandUpdatePipelineSchedule(d)
	log.Printf("[DEBUG] Pipeline Schedule Request: %#v", pipeScheduleUpdate)
	_, _, err = pipeApi.UpdateRepositoryPipelineSchedule(c.AuthContext, *pipeScheduleUpdate, workspace, repo, uuid)
	if err := handleClientError(err); err != nil {
		return diag.FromErr(err)
	}

	return resourcePipelineScheduleRead(ctx, d, m)
}

func resourcePipelineScheduleRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(Clients).genClient
	pipeApi := c.ApiClient.PipelinesApi

	workspace, repo, uuid, err := pipeScheduleId(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	schedule, res, err := pipeApi.GetRepositoryPipelineSchedule(c.AuthContext, workspace, repo, uuid)

	if res.StatusCode == http.StatusNotFound {
		log.Printf("[WARN] Pipeline Schedule (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err := handleClientError(err); err != nil {
		return diag.FromErr(err)
	}

	d.Set("repository", repo)
	d.Set("workspace", workspace)
	d.Set("uuid", schedule.Uuid)
	d.Set("enabled", schedule.Enabled)
	d.Set("cron_pattern", schedule.CronPattern)

	d.Set("target", flattenPipelineRefTarget(schedule.Target))

	return nil
}

func resourcePipelineScheduleDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(Clients).genClient
	pipeApi := c.ApiClient.PipelinesApi

	workspace, repo, uuid, err := pipeScheduleId(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	_, err = pipeApi.DeleteRepositoryPipelineSchedule(c.AuthContext, workspace, repo, uuid)
	if err := handleClientError(err); err != nil {
		return diag.FromErr(err)
	}

	return diag.FromErr(err)
}

func expandUpdatePipelineSchedule(d *schema.ResourceData) *bitbucket.PipelineSchedulePutRequestBody {
	schedule := &bitbucket.PipelineSchedulePutRequestBody{
		Enabled: d.Get("enabled").(bool),
	}

	return schedule
}

func expandCreatePipelineSchedule(d *schema.ResourceData) *bitbucket.PipelineSchedulePostRequestBody {
	schedule := &bitbucket.PipelineSchedulePostRequestBody{
		Enabled:     d.Get("enabled").(bool),
		CronPattern: d.Get("cron_pattern").(string),
		Target:      expandPipelineRefTarget(d.Get("target").([]interface{})),
	}

	return schedule
}

func expandPipelineRefTarget(conf []interface{}) *bitbucket.PipelineSchedulePostRequestBodyTarget {
	tfMap, _ := conf[0].(map[string]interface{})

	target := &bitbucket.PipelineSchedulePostRequestBodyTarget{
		RefName:  tfMap["ref_name"].(string),
		RefType:  tfMap["ref_type"].(string),
		Selector: expandPipelineRefTargetSelector(tfMap["selector"].([]interface{})),
		Type_:    "pipeline_ref_target",
	}

	return target
}

func expandPipelineRefTargetSelector(conf []interface{}) *bitbucket.PipelineSelector {
	tfMap, _ := conf[0].(map[string]interface{})

	selector := &bitbucket.PipelineSelector{
		Pattern: tfMap["pattern"].(string),
		Type_:   tfMap["type"].(string),
	}

	return selector
}

func flattenPipelineRefTarget(rp *bitbucket.PipelineRefTarget) []interface{} {
	if rp == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"ref_name": rp.RefName,
		"ref_type": rp.RefType,
		"selector": flattenPipelineSelector(rp.Selector),
	}

	return []interface{}{m}
}

func flattenPipelineSelector(rp *bitbucket.PipelineSelector) []interface{} {
	if rp == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"pattern": rp.Pattern,
		"type":    rp.Type_,
	}

	return []interface{}{m}
}

func pipeScheduleId(id string) (string, string, string, error) {
	parts := strings.Split(id, "/")

	if len(parts) != 3 {
		return "", "", "", fmt.Errorf("unexpected format of ID (%q), expected WORKSPACE-ID/REPO-ID/UUID", id)
	}

	return parts[0], parts[1], parts[2], nil
}
