package bitbucket

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// Deployment structure for handling key info
type Deployment struct {
	Name  string `json:"name"`
	Stage *Stage `json:"environment_type"`
	UUID  string `json:"uuid,omitempty"`
}

type Stage struct {
	Name string `json:"name"`
}

type Changes struct {
	Change *Change `json:"change"`
}

type Change struct {
	Name string `json:"name"`
}

func resourceDeployment() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceDeploymentCreate,
		UpdateWithoutTimeout: resourceDeploymentUpdate,
		ReadWithoutTimeout:   resourceDeploymentRead,
		DeleteWithoutTimeout: resourceDeploymentDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"uuid": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"stage": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					"Test",
					"Staging",
					"Production",
				},
					false),
			},
			"repository": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func newDeploymentFromResource(d *schema.ResourceData) *Deployment {
	dk := &Deployment{
		Name: d.Get("name").(string),
		Stage: &Stage{
			Name: d.Get("stage").(string),
		},
	}
	return dk
}

func resourceDeploymentCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	client := m.(Clients).httpClient
	rvcr := newDeploymentFromResource(d)
	bytedata, err := json.Marshal(rvcr)

	if err != nil {
		return diag.FromErr(err)
	}
	req, err := client.Post(fmt.Sprintf("2.0/repositories/%s/environments/",
		d.Get("repository").(string),
	), bytes.NewBuffer(bytedata))

	if err != nil {
		return diag.FromErr(err)
	}

	var deployment Deployment

	body, readerr := io.ReadAll(req.Body)
	if readerr != nil {
		return diag.FromErr(readerr)
	}

	decodeerr := json.Unmarshal(body, &deployment)
	if decodeerr != nil {
		return diag.FromErr(decodeerr)
	}
	d.Set("uuid", deployment.UUID)
	d.SetId(fmt.Sprintf("%s:%s", d.Get("repository"), deployment.UUID))

	return resourceDeploymentRead(ctx, d, m)
}

func resourceDeploymentRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	repoId, deployId, err := deploymentId(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	client := m.(Clients).httpClient
	res, err := client.Get(fmt.Sprintf("2.0/repositories/%s/environments/%s",
		repoId,
		deployId,
	))

	if res != nil && res.StatusCode == http.StatusNotFound {
		log.Printf("[WARN] Deployment (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.FromErr(err)
	}

	var deploy Deployment
	body, readerr := io.ReadAll(res.Body)
	if readerr != nil {
		return diag.FromErr(readerr)
	}

	log.Printf("[DEBUG] deployment response: %s", string(body))

	decodeerr := json.Unmarshal(body, &deploy)
	if decodeerr != nil {
		return diag.FromErr(decodeerr)
	}

	d.Set("uuid", deploy.UUID)
	d.Set("name", deploy.Name)
	d.Set("stage", deploy.Stage.Name)
	d.Set("repository", repoId)

	return nil
}

func resourceDeploymentUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(Clients).httpClient
	rvcr := &Changes{
		Change: &Change{
			Name: d.Get("name").(string),
		},
	}
	bytedata, err := json.Marshal(rvcr)

	if err != nil {
		return diag.FromErr(err)
	}

	req, err := client.Post(fmt.Sprintf("2.0/repositories/%s/environments/%s/changes/",
		d.Get("repository").(string),
		d.Get("uuid").(string),
	), bytes.NewBuffer(bytedata))

	if err != nil {
		return diag.FromErr(err)
	}

	if req.StatusCode != 200 {
		return nil
	}

	return resourceDeploymentRead(ctx, d, m)
}

func resourceDeploymentDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(Clients).httpClient
	_, err := client.Delete(fmt.Sprintf("2.0/repositories/%s/environments/%s",
		d.Get("repository").(string),
		d.Get("uuid").(string),
	))
	return diag.FromErr(err)
}

func deploymentId(id string) (string, string, error) {
	parts := strings.Split(id, ":")

	if len(parts) != 2 {
		return "", "", fmt.Errorf("unexpected format of ID (%q), expected REPO-ID/DEPLOYMENT-UUID", id)
	}

	return parts[0], parts[1], nil
}
