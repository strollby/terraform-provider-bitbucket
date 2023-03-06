package bitbucket

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceDeployment_basic(t *testing.T) {
	dataSourceName := "data.bitbucket_deployment.test"
	resourceName := "bitbucket_deployment.test"

	rName := acctest.RandomWithPrefix("tf-test")

	workspace := os.Getenv("BITBUCKET_TEAM")
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccBitbucketDeploymentConfig(workspace, rName, rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "repository", "bitbucket_repository.test", "name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "uuid", resourceName, "uuid"),
					resource.TestCheckResourceAttr(dataSourceName, "workspace", workspace),
				),
			},
		},
	})
}

func testAccBitbucketDeploymentConfig(workspace, repoName, deployName string) string {
	return testAccBitbucketDeployment(workspace, repoName, deployName) + fmt.Sprintf(`
data "bitbucket_deployment" "test" {
  workspace  = %[1]q
  uuid       = bitbucket_deployment.test.uuid
  repository = bitbucket_repository.test.name
}
`, workspace)
}
