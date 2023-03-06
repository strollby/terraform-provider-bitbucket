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

func TestAccBitbucketRepositoryUserPermission_basic(t *testing.T) {
	var repositoryUserPermission RepositoryUserPermission
	resourceName := "bitbucket_repository_user_permission.test"
	workspace := os.Getenv("BITBUCKET_TEAM")
	rName := acctest.RandomWithPrefix("tf-test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBitbucketRepositoryUserPermissionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBitbucketRepositoryUserPermissionConfig(workspace, rName, "read"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBitbucketRepositoryUserPermissionExists(resourceName, &repositoryUserPermission),
					resource.TestCheckResourceAttrPair(resourceName, "repo_slug", "bitbucket_repository.test", "name"),
					resource.TestCheckResourceAttrPair(resourceName, "user_id", "data.bitbucket_current_user.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "workspace", workspace),
					resource.TestCheckResourceAttr(resourceName, "permission", "read"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccBitbucketRepositoryUserPermissionConfig(workspace, rName, "write"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBitbucketRepositoryUserPermissionExists(resourceName, &repositoryUserPermission),
					resource.TestCheckResourceAttrPair(resourceName, "repo_slug", "bitbucket_repository.test", "name"),
					resource.TestCheckResourceAttrPair(resourceName, "user_id", "data.bitbucket_current_user.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "workspace", workspace),
					resource.TestCheckResourceAttr(resourceName, "permission", "write"),
				),
			},
		},
	})
}

func testAccCheckBitbucketRepositoryUserPermissionDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(Clients).httpClient
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "bitbucket_repository_user_permission" {
			continue
		}

		response, err := client.Get(fmt.Sprintf("2.0/repositories/%s/%s/permissions-config/users/%s", rs.Primary.Attributes["workspace"], rs.Primary.Attributes["repo_slug"], rs.Primary.Attributes["user_id"]))

		if err == nil {
			return fmt.Errorf("The resource was found should have errored")
		}

		if response.StatusCode != http.StatusNotFound {
			return fmt.Errorf("Repository User Permission still exists")
		}

	}
	return nil
}

func testAccCheckBitbucketRepositoryUserPermissionExists(n string, repositoryUserPermission *RepositoryUserPermission) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found %s", n)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No Repository User Permission ID is set")
		}
		return nil
	}
}

func testAccBitbucketRepositoryUserPermissionConfig(workspace, rName, permission string) string {
	return fmt.Sprintf(`
resource "bitbucket_repository" "test" {
  owner = %[1]q
  name  = %[2]q
}

data "bitbucket_current_user" "test" {}

resource "bitbucket_repository_user_permission" "test" {
  workspace  = %[1]q
  repo_slug  = bitbucket_repository.test.name
  user_id    = data.bitbucket_current_user.test.id
  permission = %[3]q
}
`, workspace, rName, permission)
}
