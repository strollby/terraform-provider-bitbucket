package bitbucket

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceDataGroups_basic(t *testing.T) {
	dataSourceName := "data.bitbucket_groups.test"

	workspace := os.Getenv("BITBUCKET_TEAM")

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccBitbucketGroupsConfig(workspace),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "groups.#"),
				),
			},
		},
	})
}

func testAccBitbucketGroupsConfig(workspace string) string {
	return fmt.Sprintf(`
data "bitbucket_workspace" "test" {
  workspace = %[1]q
}

data "bitbucket_groups" "test" {
  workspace = data.bitbucket_workspace.test.id
}
`, workspace)
}
