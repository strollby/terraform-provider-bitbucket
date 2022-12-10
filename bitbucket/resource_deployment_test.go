package bitbucket

import (
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccBitbucketDeployment_basic(t *testing.T) {
	var deploy Deployment

	resourceName := "bitbucket_deployment.test"
	rName := acctest.RandomWithPrefix("tf-test")
	owner := os.Getenv("BITBUCKET_TEAM")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBitbucketDeploymentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBitbucketDeployment(owner, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBitbucketDeploymentExists(resourceName, &deploy),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "stage", "Staging"),
					resource.TestCheckResourceAttrPair(resourceName, "repository", "bitbucket_repository.test", "id"),
				),
			},
		},
	})
}

func testAccCheckBitbucketDeploymentDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(Clients).httpClient
	rs, ok := s.RootModule().Resources["bitbucket_deployment.test"]
	if !ok {
		return fmt.Errorf("Not found %s", "bitbucket_deployment.test")
	}

	response, _ := client.Get(fmt.Sprintf("2.0/repositories/%s/%s", rs.Primary.Attributes["owner"], rs.Primary.Attributes["name"]))

	if response.StatusCode != http.StatusNotFound {
		return fmt.Errorf("Deployment still exists")
	}

	return nil
}

func testAccCheckBitbucketDeploymentExists(n string, deployment *Deployment) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found %s", n)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No deployment ID is set")
		}
		return nil
	}
}

func testAccBitbucketDeployment(workspace, rName string) string {
	return fmt.Sprintf(`
resource "bitbucket_repository" "test" {
  owner = %[1]q
  name  = %[2]q
}

resource "bitbucket_deployment" "test" {
  name       = %[2]q
  stage      = "Staging"
  repository = bitbucket_repository.test.id
}
`, workspace, rName)
}
