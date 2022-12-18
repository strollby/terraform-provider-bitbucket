package bitbucket

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceCurrentUser_basic(t *testing.T) {
	dataSourceName := "data.bitbucket_current_user.test"
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccBitbucketCurrentUserConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "uuid"),
					resource.TestCheckResourceAttrSet(dataSourceName, "username"),
					resource.TestCheckResourceAttrSet(dataSourceName, "display_name"),
					resource.TestCheckTypeSetElemNestedAttrs(dataSourceName, "email.*", map[string]string{
						"is_primary": "true",
					}),
				),
			},
		},
	})
}

func testAccBitbucketCurrentUserConfig() string {
	return `
data "bitbucket_current_user" "test" {}
`
}
