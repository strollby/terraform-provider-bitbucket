package bitbucket

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/DrFaust92/bitbucket-go-client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceRepositoryVariable() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceRepositoryVariableCreate,
		UpdateWithoutTimeout: resourceRepositoryVariableUpdate,
		ReadWithoutTimeout:   resourceRepositoryVariableRead,
		DeleteWithoutTimeout: resourceRepositoryVariableDelete,

		Schema: map[string]*schema.Schema{
			"uuid": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"key": {
				Type:     schema.TypeString,
				Required: true,
			},
			"value": {
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: true,
			},
			"secured": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"repository": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func newRepositoryVariableFromResource(d *schema.ResourceData) bitbucket.PipelineVariable {
	dk := bitbucket.PipelineVariable{
		Key:     d.Get("key").(string),
		Value:   d.Get("value").(string),
		Secured: d.Get("secured").(bool),
	}
	return dk
}

func resourceRepositoryVariableCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(Clients).genClient
	pipeApi := c.ApiClient.PipelinesApi
	rvcr := newRepositoryVariableFromResource(d)

	repo := d.Get("repository").(string)
	workspace, repoSlug, err := repoVarId(repo)
	if err != nil {
		return diag.FromErr(err)
	}

	rvRes, _, err := pipeApi.CreateRepositoryPipelineVariable(c.AuthContext, rvcr, workspace, repoSlug)
	if err := handleClientError(err); err != nil {
		return diag.FromErr(err)
	}

	d.Set("uuid", rvRes.Uuid)
	d.SetId(rvRes.Key)

	return resourceRepositoryVariableRead(ctx, d, m)
}

func resourceRepositoryVariableRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(Clients).genClient
	pipeApi := c.ApiClient.PipelinesApi

	repo := d.Get("repository").(string)
	workspace, repoSlug, err := repoVarId(repo)
	if err != nil {
		return diag.FromErr(err)
	}

	rvRes, res, err := pipeApi.GetRepositoryPipelineVariable(c.AuthContext, workspace, repoSlug, d.Get("uuid").(string))

	if res.StatusCode == http.StatusNotFound {
		log.Printf("[WARN] Repository Variable (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err := handleClientError(err); err != nil {
		return diag.FromErr(err)
	}

	d.Set("uuid", rvRes.Uuid)
	d.Set("key", rvRes.Key)
	d.Set("secured", rvRes.Secured)

	if !rvRes.Secured {
		d.Set("value", rvRes.Value)
	} else {
		d.Set("value", d.Get("value").(string))
	}

	return nil
}

func resourceRepositoryVariableUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(Clients).genClient
	pipeApi := c.ApiClient.PipelinesApi

	repo := d.Get("repository").(string)
	workspace, repoSlug, err := repoVarId(repo)
	if err != nil {
		return diag.FromErr(err)
	}

	rvcr := newRepositoryVariableFromResource(d)

	_, _, err = pipeApi.UpdateRepositoryPipelineVariable(c.AuthContext, rvcr, workspace, repoSlug, d.Get("uuid").(string))
	if err := handleClientError(err); err != nil {
		return diag.FromErr(err)
	}

	return resourceRepositoryVariableRead(ctx, d, m)
}

func resourceRepositoryVariableDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(Clients).genClient
	pipeApi := c.ApiClient.PipelinesApi

	repo := d.Get("repository").(string)
	workspace, repoSlug, err := repoVarId(repo)
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = pipeApi.DeleteRepositoryPipelineVariable(c.AuthContext, workspace, repoSlug, d.Get("uuid").(string))
	if err := handleClientError(err); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func repoVarId(repo string) (string, string, error) {
	idparts := strings.Split(repo, "/")
	if len(idparts) == 2 {
		return idparts[0], idparts[1], nil
	} else {
		return "", "", fmt.Errorf("incorrect ID format, should match `owner/key`")
	}
}
