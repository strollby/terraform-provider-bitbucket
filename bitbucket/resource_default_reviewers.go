package bitbucket

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Reviewer is teh default reviewer you want
type Reviewer struct {
	DisplayName string `json:"display_name,omitempty"`
	UUID        string `json:"uuid,omitempty"`
	Type        string `json:"type,omitempty"`
}

// PaginatedReviewers is a paginated list that the bitbucket api returns
type PaginatedReviewers struct {
	Values []Reviewer `json:"values,omitempty"`
	Page   int        `json:"page,omitempty"`
	Size   int        `json:"size,omitempty"`
	Next   string     `json:"next,omitempty"`
}

func resourceDefaultReviewers() *schema.Resource {
	return &schema.Resource{
		Create: resourceDefaultReviewersCreate,
		Read:   resourceDefaultReviewersRead,
		Update: resourceDefaultReviewersUpdate,
		Delete: resourceDefaultReviewersDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"owner": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"repository": {
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

func resourceDefaultReviewersCreate(d *schema.ResourceData, m interface{}) error {
	c := m.(Clients).genClient
	prApi := c.ApiClient.PullrequestsApi

	repo := d.Get("repository").(string)
	workspace := d.Get("owner").(string)
	for _, user := range d.Get("reviewers").(*schema.Set).List() {
		userName := user.(string)
		_, reviewerResp, err := prApi.RepositoriesWorkspaceRepoSlugDefaultReviewersTargetUsernamePut(c.AuthContext, repo, userName, workspace)

		if err != nil {
			return err
		}

		if reviewerResp.StatusCode != 200 {
			return fmt.Errorf("failed to create reviewer %s got code %d", userName, reviewerResp.StatusCode)
		}
	}

	d.SetId(fmt.Sprintf("%s/%s/reviewers", workspace, repo))
	return resourceDefaultReviewersRead(d, m)
}

func resourceDefaultReviewersRead(d *schema.ResourceData, m interface{}) error {
	client := m.(Clients).httpClient

	owner, repo, err := defaultReviewersId(d.Id())
	if err != nil {
		return err
	}
	resourceURL := fmt.Sprintf("2.0/repositories/%s/%s/default-reviewers", owner, repo)

	res, err := client.Get(resourceURL)
	if err != nil {
		return err
	}

	if res.StatusCode == http.StatusNotFound {
		log.Printf("[WARN] Default Reviewers (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	var reviewers PaginatedReviewers
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
			terraformReviewers = append(terraformReviewers, reviewer.UUID)
		}

		if reviewers.Next != "" {
			nextPage := reviewers.Page + 1
			resourceURL = fmt.Sprintf("2.0/repositories/%s/%s/default-reviewers?page=%d", owner, repo, nextPage)
			reviewers = PaginatedReviewers{}
		} else {
			break
		}
	}

	d.Set("owner", owner)
	d.Set("repository", repo)
	d.Set("reviewers", terraformReviewers)

	return nil
}

func resourceDefaultReviewersUpdate(d *schema.ResourceData, m interface{}) error {
	c := m.(Clients).genClient

	prApi := c.ApiClient.PullrequestsApi
	oraw, nraw := d.GetChange("reviewers")
	o := oraw.(*schema.Set)
	n := nraw.(*schema.Set)

	add := n.Difference(o)
	remove := o.Difference(n)
	repo := d.Get("repository").(string)
	workspace := d.Get("owner").(string)

	for _, user := range add.List() {
		userName := user.(string)
		_, reviewerResp, err := prApi.RepositoriesWorkspaceRepoSlugDefaultReviewersTargetUsernamePut(c.AuthContext, repo, userName, workspace)

		if err != nil {
			return err
		}

		if reviewerResp.StatusCode != 200 {
			return fmt.Errorf("failed to create reviewer %s got code %d", userName, reviewerResp.StatusCode)
		}
	}

	for _, user := range remove.List() {
		userName := user.(string)
		reviewerResp, err := prApi.RepositoriesWorkspaceRepoSlugDefaultReviewersTargetUsernameDelete(c.AuthContext, repo, userName, workspace)

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

	return resourceDefaultReviewersRead(d, m)
}

func resourceDefaultReviewersDelete(d *schema.ResourceData, m interface{}) error {
	c := m.(Clients).genClient
	prApi := c.ApiClient.PullrequestsApi

	repo := d.Get("repository").(string)
	workspace := d.Get("owner").(string)
	for _, user := range d.Get("reviewers").(*schema.Set).List() {
		userName := user.(string)
		reviewerResp, err := prApi.RepositoriesWorkspaceRepoSlugDefaultReviewersTargetUsernameDelete(c.AuthContext, repo, userName, workspace)

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

func defaultReviewersId(id string) (string, string, error) {
	parts := strings.Split(id, "/")

	if len(parts) != 3 {
		return "", "", fmt.Errorf("unexpected format of ID (%q), expected OWNER/REPOSITORY/reviewers", id)
	}

	return parts[0], parts[1], nil
}
