package bitbucket

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceDataGroupMembers_basic(t *testing.T) {
	dataSourceName := "data.bitbucket_group_members.test"
	groupResourceName := "bitbucket_group.test"

	workspace := os.Getenv("BITBUCKET_TEAM")
	rName := acctest.RandomWithPrefix("tf-test")

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccBitbucketGroupMembersDataConfig(workspace, rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "workspace", groupResourceName, "workspace"),
					resource.TestCheckResourceAttrPair(dataSourceName, "slug", groupResourceName, "slug"),
					resource.TestCheckResourceAttr(dataSourceName, "members.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "members.*", "data.bitbucket_current_user.test", "uuid"),
				),
			},
		},
	})
}

func testAccBitbucketGroupMembersDataConfig(workspace, rName string) string {
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

data "bitbucket_group_members" "test" {
  workspace = data.bitbucket_workspace.test.id
  slug      = bitbucket_group_membership.test.group_slug
}
`, workspace, rName)
}
