package bitbucket

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/DrFaust92/bitbucket-go-client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceProjectDefaultReviewers() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceProjectDefaultReviewersCreate,
		ReadWithoutTimeout:   resourceProjectDefaultReviewersRead,
		UpdateWithoutTimeout: resourceProjectDefaultReviewersUpdate,
		DeleteWithoutTimeout: resourceProjectDefaultReviewersDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"workspace": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"project": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"reviewers": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Required: true,
			},
		},
	}
}

func resourceProjectDefaultReviewersCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(Clients).genClient
	projectsApi := c.ApiClient.ProjectsApi

	workspace := d.Get("workspace").(string)
	project := d.Get("project").(string)

	for _, user := range d.Get("reviewers").(*schema.Set).List() {
		userName := user.(string)
		_, _, err := projectsApi.WorkspacesWorkspaceProjectsProjectKeyDefaultReviewersSelectedUserPut(c.AuthContext, project, userName, workspace)
		if err := handleClientError(err); err != nil {
			return diag.FromErr(err)
		}
	}

	d.SetId(fmt.Sprintf("%s/%s", workspace, project))
	return resourceProjectDefaultReviewersRead(ctx, d, m)
}

func resourceProjectDefaultReviewersRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(Clients).httpClient

	workspace, project, err := defaultProjectReviewersId(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	resourceURL := fmt.Sprintf("2.0/workspaces/%s/projects/%s/default-reviewers", workspace, project)

	res, err := client.Get(resourceURL)
	if err != nil {
		return diag.FromErr(err)
	}

	if res.StatusCode == http.StatusNotFound {
		log.Printf("[WARN] Project Default Reviewers (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	var reviewers bitbucket.PaginatedDefaultReviewerAndType
	var terraformReviewers []string

	for {
		reviewersResponse, err := client.Get(resourceURL)
		if err != nil {
			return diag.FromErr(err)
		}

		decoder := json.NewDecoder(reviewersResponse.Body)
		err = decoder.Decode(&reviewers)
		if err != nil {
			return diag.FromErr(err)
		}

		for _, reviewer := range reviewers.Values {
			terraformReviewers = append(terraformReviewers, reviewer.User.Uuid)
		}

		if reviewers.Next != "" {
			nextPage := reviewers.Page + 1
			resourceURL = fmt.Sprintf("2.0/workspaces/%s/projects/%s/default-reviewers?page=%d", workspace, project, nextPage)
			reviewers = bitbucket.PaginatedDefaultReviewerAndType{}
		} else {
			break
		}
	}

	d.Set("workspace", workspace)
	d.Set("project", project)
	d.Set("reviewers", terraformReviewers)

	return nil
}

func resourceProjectDefaultReviewersUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(Clients).genClient

	projectsApi := c.ApiClient.ProjectsApi
	oraw, nraw := d.GetChange("reviewers")
	o := oraw.(*schema.Set)
	n := nraw.(*schema.Set)

	add := n.Difference(o)
	remove := o.Difference(n)
	project := d.Get("project").(string)
	workspace := d.Get("workspace").(string)

	for _, user := range add.List() {
		userName := user.(string)
		_, _, err := projectsApi.WorkspacesWorkspaceProjectsProjectKeyDefaultReviewersSelectedUserPut(c.AuthContext, project, userName, workspace)
		if err := handleClientError(err); err != nil {
			return diag.FromErr(err)
		}
	}

	for _, user := range remove.List() {
		userName := user.(string)
		_, err := projectsApi.WorkspacesWorkspaceProjectsProjectKeyDefaultReviewersSelectedUserDelete(c.AuthContext, project, userName, workspace)
		if err := handleClientError(err); err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceProjectDefaultReviewersRead(ctx, d, m)
}

func resourceProjectDefaultReviewersDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(Clients).genClient
	projectsApi := c.ApiClient.ProjectsApi

	project := d.Get("project").(string)
	workspace := d.Get("workspace").(string)
	for _, user := range d.Get("reviewers").(*schema.Set).List() {
		userName := user.(string)
		_, err := projectsApi.WorkspacesWorkspaceProjectsProjectKeyDefaultReviewersSelectedUserDelete(c.AuthContext, project, userName, workspace)
		if err := handleClientError(err); err != nil {
			return diag.FromErr(err)
		}
	}
	return nil
}

func defaultProjectReviewersId(id string) (string, string, error) {
	parts := strings.Split(id, "/")

	if len(parts) != 2 {
		return "", "", fmt.Errorf("unexpected format of ID (%q), expected workspace/project", id)
	}

	return parts[0], parts[1], nil
}
