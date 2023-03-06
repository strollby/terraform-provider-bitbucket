package bitbucket

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func dataUser() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataReadUser,

		Schema: map[string]*schema.Schema{
			"username": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"display_name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"uuid": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
		},
	}
}

func dataReadUser(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(Clients).genClient
	usersApi := c.ApiClient.UsersApi
	var selectedUser string

	if v, ok := d.GetOk("uuid"); ok && v.(string) != "" {
		selectedUser = v.(string)
	}

	user, _, err := usersApi.UsersSelectedUserGet(c.AuthContext, selectedUser)
	if err := handleClientError(err); err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] User: %#v", user)

	d.SetId(user.Uuid)
	d.Set("uuid", user.Uuid)
	d.Set("username", user.Username)
	d.Set("display_name", user.DisplayName)

	return nil
}
