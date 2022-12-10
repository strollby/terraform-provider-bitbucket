package bitbucket

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

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

func resourceDeployment() *schema.Resource {
	return &schema.Resource{
		Create: resourceDeploymentCreate,
		Update: resourceDeploymentUpdate,
		Read:   resourceDeploymentRead,
		Delete: resourceDeploymentDelete,
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

func resourceDeploymentCreate(d *schema.ResourceData, m interface{}) error {

	client := m.(Clients).httpClient
	rvcr := newDeploymentFromResource(d)
	bytedata, err := json.Marshal(rvcr)

	if err != nil {
		return err
	}
	req, err := client.Post(fmt.Sprintf("2.0/repositories/%s/environments/",
		d.Get("repository").(string),
	), bytes.NewBuffer(bytedata))

	if err != nil {
		return err
	}

	var deployment Deployment

	body, readerr := io.ReadAll(req.Body)
	if readerr != nil {
		return readerr
	}

	decodeerr := json.Unmarshal(body, &deployment)
	if decodeerr != nil {
		return decodeerr
	}
	d.Set("uuid", deployment.UUID)
	d.SetId(fmt.Sprintf("%s:%s", d.Get("repository"), deployment.UUID))

	return resourceDeploymentRead(d, m)
}

func resourceDeploymentRead(d *schema.ResourceData, m interface{}) error {

	repoId, deployId, err := deploymentId(d.Id())
	if err != nil {
		return err
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
		return err
	}

	var deploy Deployment
	body, readerr := io.ReadAll(res.Body)
	if readerr != nil {
		return readerr
	}

	log.Printf("[DEBUG] deployment response: %s", string(body))

	decodeerr := json.Unmarshal(body, &deploy)
	if decodeerr != nil {
		return decodeerr
	}

	d.Set("uuid", deploy.UUID)
	d.Set("name", deploy.Name)
	d.Set("stage", deploy.Stage.Name)
	d.Set("repository", repoId)

	return nil
}

func resourceDeploymentUpdate(d *schema.ResourceData, m interface{}) error {
	client := m.(Clients).httpClient
	rvcr := newDeploymentFromResource(d)
	bytedata, err := json.Marshal(rvcr)

	if err != nil {
		return err
	}
	req, err := client.Put(fmt.Sprintf("2.0/repositories/%s/environments/%s",
		d.Get("repository").(string),
		d.Get("uuid").(string),
	), bytes.NewBuffer(bytedata))

	if err != nil {
		return err
	}

	if req.StatusCode != 200 {
		return nil
	}

	return resourceDeploymentRead(d, m)
}

func resourceDeploymentDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(Clients).httpClient
	_, err := client.Delete(fmt.Sprintf("2.0/repositories/%s/environments/%s",
		d.Get("repository").(string),
		d.Get("uuid").(string),
	))
	return err
}

func deploymentId(id string) (string, string, error) {
	parts := strings.Split(id, ":")

	if len(parts) != 2 {
		return "", "", fmt.Errorf("unexpected format of ID (%q), expected REPO-ID/DEPLOYMENT-UUID", id)
	}

	return parts[0], parts[1], nil
}
