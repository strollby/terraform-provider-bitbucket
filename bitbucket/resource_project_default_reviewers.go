package bitbucket

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/DrFaust92/bitbucket-go-client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// // Reviewer is teh default reviewer you want
// type Reviewer struct {
// 	DisplayName string `json:"display_name,omitempty"`
// 	UUID        string `json:"uuid,omitempty"`
// 	Type        string `json:"type,omitempty"`
// }

// // PaginatedReviewers is a paginated list that the bitbucket api returns
// type PaginatedReviewers struct {
// 	Values []Reviewer `json:"values,omitempty"`
// 	Page   int        `json:"page,omitempty"`
// 	Size   int        `json:"size,omitempty"`
// 	Next   string     `json:"next,omitempty"`
// }

func resourceProjectDefaultReviewers() *schema.Resource {
	return &schema.Resource{
		Create: resourceProjectDefaultReviewersCreate,
		Read:   resourceProjectDefaultReviewersRead,
		Update: resourceProjectDefaultReviewersUpdate,
		Delete: resourceProjectDefaultReviewersDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
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

func resourceProjectDefaultReviewersCreate(d *schema.ResourceData, m interface{}) error {
	c := m.(Clients).genClient
	projectsApi := c.ApiClient.ProjectsApi

	workspace := d.Get("workspace").(string)
	project := d.Get("project").(string)

	for _, user := range d.Get("reviewers").(*schema.Set).List() {
		userName := user.(string)
		_, reviewerResp, err := projectsApi.WorkspacesWorkspaceProjectsProjectKeyDefaultReviewersSelectedUserPut(c.AuthContext, project, userName, workspace)

		if err != nil {
			return err
		}

		if reviewerResp.StatusCode != 200 {
			return fmt.Errorf("failed to create reviewer %s got code %d", userName, reviewerResp.StatusCode)
		}
	}

	d.SetId(fmt.Sprintf("%s/%s", workspace, project))
	return resourceProjectDefaultReviewersRead(d, m)
}

func resourceProjectDefaultReviewersRead(d *schema.ResourceData, m interface{}) error {
	client := m.(Clients).httpClient

	workspace, project, err := defaultProjectReviewersId(d.Id())
	if err != nil {
		return err
	}

	resourceURL := fmt.Sprintf("2.0/workspaces/%s/projects/%s/default-reviewers", workspace, project)

	res, err := client.Get(resourceURL)
	if err != nil {
		return err
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
			return err
		}

		decoder := json.NewDecoder(reviewersResponse.Body)
		err = decoder.Decode(&reviewers)
		if err != nil {
			return err
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

func resourceProjectDefaultReviewersUpdate(d *schema.ResourceData, m interface{}) error {
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
		_, reviewerResp, err := projectsApi.WorkspacesWorkspaceProjectsProjectKeyDefaultReviewersSelectedUserPut(c.AuthContext, project, userName, workspace)

		if err != nil {
			return err
		}

		if reviewerResp.StatusCode != 200 {
			return fmt.Errorf("failed to create reviewer %s got code %d", userName, reviewerResp.StatusCode)
		}
	}

	for _, user := range remove.List() {
		userName := user.(string)
		reviewerResp, err := projectsApi.WorkspacesWorkspaceProjectsProjectKeyDefaultReviewersSelectedUserDelete(c.AuthContext, project, userName, workspace)

		if err != nil {
			return err
		}

		if reviewerResp.StatusCode != 204 {
			return fmt.Errorf("[%d] Could not delete %s from default reviewers",
				reviewerResp.StatusCode,
				userName,
			)
		}
	}

	return resourceProjectDefaultReviewersRead(d, m)
}

func resourceProjectDefaultReviewersDelete(d *schema.ResourceData, m interface{}) error {
	c := m.(Clients).genClient
	projectsApi := c.ApiClient.ProjectsApi

	project := d.Get("project").(string)
	workspace := d.Get("workspace").(string)
	for _, user := range d.Get("reviewers").(*schema.Set).List() {
		userName := user.(string)
		reviewerResp, err := projectsApi.WorkspacesWorkspaceProjectsProjectKeyDefaultReviewersSelectedUserDelete(c.AuthContext, project, userName, workspace)

		if err != nil {
			return err
		}

		if reviewerResp.StatusCode != 204 {
			return fmt.Errorf("[%d] Could not delete %s from default reviewer",
				reviewerResp.StatusCode,
				userName,
			)
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
