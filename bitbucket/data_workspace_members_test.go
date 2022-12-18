package bitbucket

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceWorkspaceMembers_basic(t *testing.T) {
	dataSourceName := "data.bitbucket_workspace_members.test"
	workspace := os.Getenv("BITBUCKET_TEAM")
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccBitbucketWorkspaceMembersConfig(workspace),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "members.#"),
				),
			},
		},
	})
}

func testAccBitbucketWorkspaceMembersConfig(workspace string) string {
	return fmt.Sprintf(`
data "bitbucket_workspace_members" "test" {
  workspace = %[1]q
}
`, workspace)
}
