package bitbucket

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/DrFaust92/bitbucket-go-client"
	"github.com/antihax/optional"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceRepository() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceRepositoryCreate,
		UpdateWithoutTimeout: resourceRepositoryUpdate,
		ReadWithoutTimeout:   resourceRepositoryRead,
		DeleteWithoutTimeout: resourceRepositoryDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"scm": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "git",
				ValidateFunc: validation.StringInSlice([]string{"hg", "git"}, false),
			},
			"has_wiki": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"has_issues": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"website": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"clone_ssh": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"clone_https": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"project_key": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"is_private": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"pipelines_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"fork_policy": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "allow_forks",
				ValidateFunc: validation.StringInSlice([]string{"allow_forks", "no_public_forks", "no_forks"}, false),
			},
			"language": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"owner": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"slug": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return computeSlug(old) == computeSlug(new)
				},
			},
			"uuid": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"link": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"avatar": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"href": {
										Type:     schema.TypeString,
										Optional: true,
										DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
											return strings.HasPrefix(old, "https://bytebucket.org/ravatar/")
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func newRepositoryFromResource(d *schema.ResourceData) *bitbucket.Repository {
	repo := &bitbucket.Repository{
		Name:        d.Get("name").(string),
		Language:    d.Get("language").(string),
		IsPrivate:   d.Get("is_private").(bool),
		Description: d.Get("description").(string),
		ForkPolicy:  d.Get("fork_policy").(string),
		HasWiki:     d.Get("has_wiki").(bool),
		HasIssues:   d.Get("has_issues").(bool),
		Scm:         d.Get("scm").(string),
	}

	if v, ok := d.GetOk("link"); ok && len(v.([]interface{})) > 0 && v.([]interface{}) != nil {
		repo.Links = expandLinks(v.([]interface{}))
	}

	if v, ok := d.GetOk("project_key"); ok && v.(string) != "" {
		project := &bitbucket.Project{
			Key: v.(string),
		}
		repo.Project = project
	}

	return repo
}

func resourceRepositoryUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(Clients).genClient
	repoApi := c.ApiClient.RepositoriesApi
	pipeApi := c.ApiClient.PipelinesApi

	repository := newRepositoryFromResource(d)

	var repoSlug string
	repoSlug = d.Get("slug").(string)
	if repoSlug == "" {
		repoSlug = d.Get("name").(string)
	}
	repoSlug = computeSlug(repoSlug)
	repoBody := &bitbucket.RepositoriesApiRepositoriesWorkspaceRepoSlugPutOpts{
		Body: optional.NewInterface(repository),
	}

	workspace := d.Get("owner").(string)
	_, _, err := repoApi.RepositoriesWorkspaceRepoSlugPut(c.AuthContext, repoSlug, workspace, repoBody)
	if err := handleClientError(err); err != nil {
		return diag.FromErr(err)
	}

	pipelinesEnabled := d.Get("pipelines_enabled").(bool)
	pipelinesConfig := &bitbucket.PipelinesConfig{Enabled: pipelinesEnabled}

	_, _, err = pipeApi.UpdateRepositoryPipelineConfig(c.AuthContext, *pipelinesConfig, workspace, repoSlug)
	if err := handleClientError(err); err != nil {
		return diag.FromErr(err)
	}

	return resourceRepositoryRead(ctx, d, m)
}

func resourceRepositoryCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(Clients).genClient
	repoApi := c.ApiClient.RepositoriesApi
	pipeApi := c.ApiClient.PipelinesApi
	repo := newRepositoryFromResource(d)

	var repoSlug string
	repoSlug = d.Get("slug").(string)
	if repoSlug == "" {
		repoSlug = d.Get("name").(string)
	}
	repoSlug = computeSlug(repoSlug)

	workspace := d.Get("owner").(string)

	repoBody := &bitbucket.RepositoriesApiRepositoriesWorkspaceRepoSlugPostOpts{
		Body: optional.NewInterface(repo),
	}

	_, _, err := repoApi.RepositoriesWorkspaceRepoSlugPost(c.AuthContext, repoSlug, workspace, repoBody)
	if err := handleClientError(err); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(string(fmt.Sprintf("%s/%s", d.Get("owner").(string), repoSlug)))

	pipelinesEnabled := d.Get("pipelines_enabled").(bool)
	pipelinesConfig := &bitbucket.PipelinesConfig{Enabled: pipelinesEnabled}

	_, _, err = pipeApi.UpdateRepositoryPipelineConfig(c.AuthContext, *pipelinesConfig, workspace, repoSlug)
	if err := handleClientError(err); err != nil {
		return diag.FromErr(err)
	}

	return resourceRepositoryRead(ctx, d, m)
}

func resourceRepositoryRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	if id != "" {
		idparts := strings.Split(id, "/")
		if len(idparts) == 2 {
			d.Set("owner", idparts[0])
			d.Set("slug", idparts[1])
		} else {
			return diag.Errorf("incorrect ID format, should match `owner/slug`")
		}
	}

	var repoSlug string
	repoSlug = d.Get("slug").(string)
	if repoSlug == "" {
		repoSlug = d.Get("name").(string)
	}
	repoSlug = computeSlug(repoSlug)

	workspace := d.Get("owner").(string)
	c := m.(Clients).genClient
	repoApi := c.ApiClient.RepositoriesApi
	pipeApi := c.ApiClient.PipelinesApi

	repoRes, res, err := repoApi.RepositoriesWorkspaceRepoSlugGet(c.AuthContext, repoSlug, workspace)

	if res.StatusCode == http.StatusNotFound {
		log.Printf("[WARN] Repository (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err := handleClientError(err); err != nil {
		return diag.FromErr(err)
	}

	d.Set("scm", repoRes.Scm)
	d.Set("is_private", repoRes.IsPrivate)
	d.Set("has_wiki", repoRes.HasWiki)
	d.Set("has_issues", repoRes.HasIssues)
	d.Set("name", repoRes.Name)
	d.Set("slug", repoRes.Slug)
	d.Set("language", repoRes.Language)
	d.Set("fork_policy", repoRes.ForkPolicy)
	// d.Set("website", repoRes.Website)
	d.Set("description", repoRes.Description)
	d.Set("project_key", repoRes.Project.Key)
	d.Set("uuid", repoRes.Uuid)

	for _, cloneURL := range repoRes.Links.Clone {
		if cloneURL.Name == "https" {
			d.Set("clone_https", cloneURL.Href)
		} else {
			d.Set("clone_ssh", cloneURL.Href)
		}
	}

	d.Set("link", flattenLinks(repoRes.Links))

	pipelinesConfigReq, res, err := pipeApi.GetRepositoryPipelineConfig(c.AuthContext, workspace, repoSlug)
	if err := handleClientError(err); err != nil {
		return diag.FromErr(err)
	}

	if res.StatusCode == 200 {
		d.Set("pipelines_enabled", pipelinesConfigReq.Enabled)
	} else if res.StatusCode == http.StatusNotFound {
		d.Set("pipelines_enabled", false)
	}

	return nil
}

func resourceRepositoryDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	var repoSlug string
	repoSlug = d.Get("slug").(string)
	if repoSlug == "" {
		repoSlug = d.Get("name").(string)
	}

	c := m.(Clients).genClient
	repoApi := c.ApiClient.RepositoriesApi

	_, err := repoApi.RepositoriesWorkspaceRepoSlugDelete(c.AuthContext, repoSlug, d.Get("owner").(string), nil)
	if err := handleClientError(err); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

var slugForbiddenCharacters *regexp.Regexp = regexp.MustCompile(`[\W-]`)

func computeSlug(repoName string) string {
	slug := slugForbiddenCharacters.ReplaceAllString(repoName, "-")
	return strings.ToLower(slug)
}

func splitFullName(repoFullName string) (string, string, error) {
	fullNameParts := strings.Split(repoFullName, "/")
	if len(fullNameParts) < 2 {
		return "", "", fmt.Errorf("Error parsing repo name (%s)", repoFullName)
	}
	owner := fullNameParts[0]
	repoSlug := strings.Join(fullNameParts[1:], "/")
	return owner, repoSlug, nil
}

func expandLinks(l []interface{}) *bitbucket.RepositoryLinks {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})

	if !ok {
		return nil
	}

	rp := &bitbucket.RepositoryLinks{}

	if v, ok := tfMap["avatar"].([]interface{}); ok && len(v) > 0 {
		rp.Avatar = expandLink(v)
	}

	return rp
}

func flattenLinks(rp *bitbucket.RepositoryLinks) []interface{} {
	if rp == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"avatar": flattenLink(rp.Avatar),
	}

	return []interface{}{m}
}

func expandLink(l []interface{}) *bitbucket.Link {

	tfMap, _ := l[0].(map[string]interface{})

	rp := &bitbucket.Link{}

	if v, ok := tfMap["href"].(string); ok {
		rp.Href = v
	}

	return rp
}

func flattenLink(rp *bitbucket.Link) []interface{} {
	m := map[string]interface{}{
		"href": rp.Href,
	}

	return []interface{}{m}
}
