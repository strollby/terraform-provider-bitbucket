package bitbucket

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourcePipelineOidcConfig_basic(t *testing.T) {
	dataSourceName := "data.bitbucket_pipeline_oidc_config.test"
	workspace := os.Getenv("BITBUCKET_TEAM")
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccBitbucketPipelineOidcConfigConfig(workspace),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "workspace", workspace),
					resource.TestCheckResourceAttrSet(dataSourceName, "oidc_config"),
				),
			},
		},
	})
}

func testAccBitbucketPipelineOidcConfigConfig(workspace string) string {
	return fmt.Sprintf(`
data "bitbucket_pipeline_oidc_config" "test" {
  workspace = %[1]q
}
`, workspace)
}
