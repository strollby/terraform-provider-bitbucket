package bitbucket

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceDataGroup_basic(t *testing.T) {
	dataSourceName := "data.bitbucket_group.test"
	groupResourceName := "bitbucket_group.test"

	workspace := os.Getenv("BITBUCKET_TEAM")
	rName := acctest.RandomWithPrefix("tf-test")

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccBitbucketGroupDataConfig(workspace, rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "workspace", groupResourceName, "workspace"),
					resource.TestCheckResourceAttrPair(dataSourceName, "slug", groupResourceName, "slug"),
					resource.TestCheckResourceAttrPair(dataSourceName, "name", groupResourceName, "name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "auto_add", groupResourceName, "auto_add"),
					resource.TestCheckResourceAttrPair(dataSourceName, "permission", groupResourceName, "permission"),
				),
			},
		},
	})
}

func testAccBitbucketGroupDataConfig(workspace, rName string) string {
	return fmt.Sprintf(`
data "bitbucket_workspace" "test" {
  workspace = %[1]q
}

resource "bitbucket_group" "test" {
  workspace  = data.bitbucket_workspace.test.id
  name       = %[2]q
}

data "bitbucket_group" "test" {
  workspace = data.bitbucket_workspace.test.id
  slug      = bitbucket_group.test.slug
}
`, workspace, rName)
}
