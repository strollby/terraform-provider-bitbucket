package bitbucket

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceHookTypes_basic(t *testing.T) {
	dataSourceName := "data.bitbucket_hook_types.test"
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccBitbucketHookTypesConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "subject_type", "workspace"),
					resource.TestCheckTypeSetElemNestedAttrs(dataSourceName, "hook_types.*", map[string]string{
						"event":       "repo:transfer",
						"category":    "Repository",
						"label":       "Transfer accepted",
						"description": "Whenever a repository transfer is accepted",
					}),
				),
			},
		},
	})
}

func testAccBitbucketHookTypesConfig() string {
	return `
data "bitbucket_hook_types" "test" {
  subject_type = "workspace"
}
`
}
