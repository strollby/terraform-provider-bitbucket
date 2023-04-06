package bitbucket

import (
	"bytes"
	"context"
	"fmt"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/DrFaust92/bitbucket-go-client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceCommitFile() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceCommitFilePut,
		ReadWithoutTimeout:   resourceCommitFileRead,
		DeleteWithoutTimeout: resourceCommitFileDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"workspace": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"repo_slug": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"content": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"filename": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"branch": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"commit_message": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The SHA of the commit that modified the file",
				ForceNew:    true,
			},
			"commit_author": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The SHA of the commit that modified the file",
				ForceNew:    true,
			},
			"commit_sha": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The SHA of the commit that modified the file",
			},
		},
	}
}

func resourceCommitFilePut(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(Clients).httpClient

	repoSlug := d.Get("repo_slug").(string)
	workspace := d.Get("workspace").(string)
	content := d.Get("content").(string)
	filename := d.Get("filename").(string)
	branch := d.Get("branch").(string)
	commitMessage := d.Get("commit_message").(string)
	commitAuthor := d.Get("commit_author").(string)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile(filename, filename)
	part.Write([]byte(content))
	writer.Close()

	messageFormField, _ := writer.CreateFormField("message")
	messageFormField.Write([]byte(commitMessage))

	authorFormField, _ := writer.CreateFormField("author")
	authorFormField.Write([]byte(commitAuthor))

	branchFormField, _ := writer.CreateFormField("branch")
	branchFormField.Write([]byte(branch))

	response, err := client.PostWithContentType(fmt.Sprintf("2.0/repositories/%s/%s/src",
		workspace,
		repoSlug,
	), writer.FormDataContentType(), body)

	if err := handleClientError(err); err != nil {
		return diag.FromErr(err)
	}

	if response.StatusCode != http.StatusCreated {
		return diag.FromErr(fmt.Errorf(""))
	}

	d.SetId(string(fmt.Sprintf("%s/%s/%s", workspace, repoSlug, filename)))

	location, _ := response.Location()
	splitPath := strings.Split(location.Path, "/")
	d.Set("commit_sha", splitPath[len(splitPath)-1])

	return resourceCommitFileRead(ctx, d, m)
}

func resourceCommitFileRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(Clients).genClient
	sourceApi := c.ApiClient.SourceApi

	repoSlug := d.Get("repo_slug").(string)
	workspace := d.Get("workspace").(string)
	filename := d.Get("filename").(string)
	commit := d.Get("commit_sha").(string)

	_, _, err := sourceApi.RepositoriesWorkspaceRepoSlugSrcCommitPathGet(c.AuthContext, commit, filename, repoSlug, workspace, &bitbucket.SourceApiRepositoriesWorkspaceRepoSlugSrcCommitPathGetOpts{})

	if err := handleClientError(err); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceCommitFileDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return nil
}
