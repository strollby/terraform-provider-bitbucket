package bitbucket

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceUser_uuid(t *testing.T) {
	dataSourceName := "data.bitbucket_user.test"
	currUserDataSource := "data.bitbucket_current_user.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccBitbucketUserUUIDConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "uuid", currUserDataSource, "uuid"),
					resource.TestCheckResourceAttrPair(dataSourceName, "display_name", currUserDataSource, "display_name"),
				),
			},
		},
	})
}

func testAccBitbucketUserUUIDConfig() string {
	return `
data "bitbucket_current_user" "test" {}

data "bitbucket_user" "test" {
  uuid = data.bitbucket_current_user.test.uuid
}
`
}
