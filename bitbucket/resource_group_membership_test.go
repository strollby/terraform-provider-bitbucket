package bitbucket

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccBitbucketGroupMembership_basic(t *testing.T) {
	var group UserGroup
	resourceName := "bitbucket_group_membership.test"
	grpResourceName := "bitbucket_group.test"

	workspace := os.Getenv("BITBUCKET_TEAM")
	rName := acctest.RandomWithPrefix("tf-test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBitbucketGroupMembershipDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBitbucketGroupMembershipConfig(workspace, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBitbucketGroupMembershipExists(resourceName, &group),
					resource.TestCheckResourceAttrPair(resourceName, "workspace", grpResourceName, "workspace"),
					resource.TestCheckResourceAttrPair(resourceName, "group_slug", grpResourceName, "slug"),
					resource.TestCheckResourceAttrPair(resourceName, "uuid", "data.bitbucket_current_user.test", "id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckBitbucketGroupMembershipDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(Clients).httpClient
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "bitbucket_group_membership" {
			continue
		}

		workspace, slug, uuid, err := groupMemberId(rs.Primary.ID)
		if err != nil {
			return err
		}

		response, _ := client.Get(fmt.Sprintf("1.0/groups/%s/%s/members",
			workspace, slug))

		if response.StatusCode == http.StatusNotFound {
			continue
		}

		if err != nil {
			return err
		}

		var members []*UserGroupMembership
		body, readerr := io.ReadAll(response.Body)
		if readerr != nil {
			return readerr
		}

		log.Printf("[DEBUG] Group Membership Response Test JSON: %v", string(body))

		decodeerr := json.Unmarshal(body, &members)
		if decodeerr != nil {
			return decodeerr
		}

		log.Printf("[DEBUG] Group Membership Response Test Decoded: %#v", members)

		var member *UserGroupMembership
		for _, mbr := range members {
			if mbr.UUID == uuid {
				member = mbr
				continue
			}
		}

		if member != nil {
			return fmt.Errorf("Group Member still exists")
		}
	}
	return nil
}

func testAccCheckBitbucketGroupMembershipExists(n string, group *UserGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found %s", n)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No Group Membership ID is set")
		}
		return nil
	}
}

func testAccBitbucketGroupMembershipConfig(workspace, rName string) string {
	return fmt.Sprintf(`
data "bitbucket_workspace" "test" {
  workspace = %[1]q
}

resource "bitbucket_group" "test" {
  workspace  = data.bitbucket_workspace.test.id
  name       = %[2]q
}

data "bitbucket_current_user" "test" {}

resource "bitbucket_group_membership" "test" {
  workspace  = bitbucket_group.test.workspace
  group_slug = bitbucket_group.test.slug
  uuid       = data.bitbucket_current_user.test.id
}
`, workspace, rName)
}
