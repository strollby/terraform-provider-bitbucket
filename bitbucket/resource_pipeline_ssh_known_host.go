package bitbucket

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/strollby/bitbucket-go-client"
)

func resourcePipelineSshKnownHost() *schema.Resource {
	return &schema.Resource{
		Create: resourcePipelineSshKnownHostsCreate,
		Read:   resourcePipelineSshKnownHostsRead,
		Update: resourcePipelineSshKnownHostsUpdate,
		Delete: resourcePipelineSshKnownHostsDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
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
			"hostname": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"public_key": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"key_type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice([]string{"ssh-ed25519", "ecdsa-sha2-nistp256", "ssh-rsa", "ssh-dss"}, false),
						},
						"key": {
							Type:     schema.TypeString,
							Required: true,
						},
						"md5_fingerprint": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"sha256_fingerprint": {
							Type:     schema.TypeString,
							Computed: true,
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

func resourcePipelineSshKnownHostsCreate(d *schema.ResourceData, m interface{}) error {
	c := m.(Clients).genClient
	pipeApi := c.ApiClient.PipelinesApi

	pipeSshKnownHost := expandPipelineSshKnownHost(d)
	log.Printf("[DEBUG] Pipeline Ssh Key Request: %#v", pipeSshKnownHost)

	repo := d.Get("repository").(string)
	workspace := d.Get("workspace").(string)
	host, _, err := pipeApi.CreateRepositoryPipelineKnownHost(c.AuthContext, *pipeSshKnownHost, workspace, repo)

	if err != nil {
		return fmt.Errorf("error creating pipeline ssh known host: %w", err)
	}

	d.SetId(string(fmt.Sprintf("%s/%s/%s", workspace, repo, host.Uuid)))

	return resourcePipelineSshKnownHostsRead(d, m)
}

func resourcePipelineSshKnownHostsUpdate(d *schema.ResourceData, m interface{}) error {
	c := m.(Clients).genClient
	pipeApi := c.ApiClient.PipelinesApi

	workspace, repo, uuid, err := pipeSshKnownHostId(d.Id())
	if err != nil {
		return err
	}

	pipeSshKnownHost := expandPipelineSshKnownHost(d)
	log.Printf("[DEBUG] Pipeline Ssh Key Request: %#v", pipeSshKnownHost)
	_, _, err = pipeApi.UpdateRepositoryPipelineKnownHost(c.AuthContext, *pipeSshKnownHost, workspace, repo, uuid)

	if err != nil {
		return fmt.Errorf("error updating pipeline ssh known host: %w", err)
	}

	return resourcePipelineSshKnownHostsRead(d, m)
}

func resourcePipelineSshKnownHostsRead(d *schema.ResourceData, m interface{}) error {
	c := m.(Clients).genClient
	pipeApi := c.ApiClient.PipelinesApi

	workspace, repo, uuid, err := pipeSshKnownHostId(d.Id())
	if err != nil {
		return err
	}

	host, res, err := pipeApi.GetRepositoryPipelineKnownHost(c.AuthContext, workspace, repo, uuid)
	if err != nil {
		return fmt.Errorf("error reading Pipeline Ssh known host (%s): %w", d.Id(), err)
	}

	if res.StatusCode == http.StatusNotFound {
		log.Printf("[WARN] Pipeline Ssh known host (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if res.Body == nil {
		return fmt.Errorf("error getting Pipeline Ssh known host (%s): empty response", d.Id())
	}

	d.Set("repository", repo)
	d.Set("workspace", workspace)
	d.Set("hostname", host.Hostname)
	d.Set("uuid", host.Uuid)
	d.Set("public_key", flattenPipelineSshKnownHost(host.PublicKey))

	return nil
}

func resourcePipelineSshKnownHostsDelete(d *schema.ResourceData, m interface{}) error {
	c := m.(Clients).genClient
	pipeApi := c.ApiClient.PipelinesApi

	workspace, repo, uuid, err := pipeSshKnownHostId(d.Id())
	if err != nil {
		return err
	}
	_, err = pipeApi.DeleteRepositoryPipelineKnownHost(c.AuthContext, workspace, repo, uuid)

	if err != nil {
		return fmt.Errorf("error deleting Pipeline Ssh known host (%s): %w", d.Id(), err)
	}

	return err
}

func expandPipelineSshKnownHost(d *schema.ResourceData) *bitbucket.PipelineKnownHost {
	key := &bitbucket.PipelineKnownHost{
		Hostname:  d.Get("hostname").(string),
		PublicKey: expandPipelineSshKnownHostKey(d.Get("public_key").([]interface{})),
	}

	return key
}

func expandPipelineSshKnownHostKey(pubKey []interface{}) *bitbucket.PipelineSshPublicKey {
	tfMap, _ := pubKey[0].(map[string]interface{})

	key := &bitbucket.PipelineSshPublicKey{
		KeyType: tfMap["key_type"].(string),
		Key:     tfMap["key"].(string),
	}

	return key
}

func flattenPipelineSshKnownHost(rp *bitbucket.PipelineSshPublicKey) []interface{} {
	if rp == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"key_type":           rp.KeyType,
		"key":                rp.Key,
		"md5_fingerprint":    rp.Md5Fingerprint,
		"sha256_fingerprint": rp.Sha256Fingerprint,
	}

	return []interface{}{m}
}

func pipeSshKnownHostId(id string) (string, string, string, error) {
	parts := strings.Split(id, "/")

	if len(parts) != 3 {
		return "", "", "", fmt.Errorf("unexpected format of ID (%q), expected WORKSPACE-ID/REPO-ID/UUID", id)
	}

	return parts[0], parts[1], parts[2], nil
}
