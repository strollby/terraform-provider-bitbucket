package bitbucket

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type PaginatedUserEmails struct {
	Values []UserEmail `json:"values,omitempty"`
	Page   int         `json:"page,omitempty"`
	Size   int         `json:"size,omitempty"`
	Next   string      `json:"next,omitempty"`
}

type UserEmail struct {
	Email       string `json:"email"`
	IsPrimary   bool   `json:"is_primary"`
	IsConfirmed bool   `json:"is_confirmed"`
}

func dataCurrentUser() *schema.Resource {
	return &schema.Resource{
		Read: dataReadCurrentUser,

		Schema: map[string]*schema.Schema{
			"username": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"display_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"uuid": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"email": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"is_confirmed": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"is_primary": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"email": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataReadCurrentUser(d *schema.ResourceData, m interface{}) error {
	c := m.(Clients).genClient
	httpClient := m.(Clients).httpClient
	usersApi := c.ApiClient.UsersApi

	curUser, curUserRes, err := usersApi.UserGet(c.AuthContext)
	if err != nil {
		return fmt.Errorf("error reading current user: %w", err)
	}

	if curUserRes.StatusCode == http.StatusNotFound {
		return fmt.Errorf("user not found")
	}

	if curUserRes.StatusCode >= http.StatusInternalServerError {
		return fmt.Errorf("internal server error fetching user")
	}

	log.Printf("[DEBUG] Current User: %#v", curUser)

	curUserEmails, err := httpClient.Get("2.0/user/emails")
	if err != nil {
		return err
	}

	emailBody, readerr := io.ReadAll(curUserEmails.Body)
	if readerr != nil {
		return readerr
	}

	log.Printf("[DEBUG] Current User Emails Response JSON: %v", string(emailBody))

	var emails PaginatedUserEmails

	decodeerr := json.Unmarshal(emailBody, &emails)
	if decodeerr != nil {
		return decodeerr
	}

	log.Printf("[DEBUG] Current User Emails Response Decoded: %#v", emails)

	d.SetId(curUser.Uuid)
	d.Set("uuid", curUser.Uuid)
	d.Set("username", curUser.Username)
	d.Set("display_name", curUser.DisplayName)
	d.Set("email", flattenUserEmails(emails.Values))

	return nil
}

func flattenUserEmails(userEmails []UserEmail) []interface{} {
	if len(userEmails) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, btRaw := range userEmails {
		log.Printf("[DEBUG] User Email Response Decoded: %#v", btRaw)

		branchType := map[string]interface{}{
			"email":        btRaw.Email,
			"is_confirmed": btRaw.IsConfirmed,
			"is_primary":   btRaw.IsPrimary,
		}

		tfList = append(tfList, branchType)
	}

	return tfList
}
