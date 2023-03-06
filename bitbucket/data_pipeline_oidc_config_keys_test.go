package bitbucket

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourcePipelineOidcConfigKeys_basic(t *testing.T) {
	dataSourceName := "data.bitbucket_pipeline_oidc_config_keys.test"
	workspace := os.Getenv("BITBUCKET_TEAM")
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccBitbucketPipelineOidcConfigKeysConfig(workspace),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "workspace", workspace),
					resource.TestCheckResourceAttrSet(dataSourceName, "keys"),
				),
			},
		},
	})
}

func testAccBitbucketPipelineOidcConfigKeysConfig(workspace string) string {
	return fmt.Sprintf(`
data "bitbucket_pipeline_oidc_config_keys" "test" {
  workspace = %[1]q
}
`, workspace)
}
