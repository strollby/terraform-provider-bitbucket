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

func TestAccBitbucketRepositoryGroupPermission_basic(t *testing.T) {
	var repositoryGroupPermission RepositoryGroupPermission
	resourceName := "bitbucket_repository_group_permission.test"
	workspace := os.Getenv("BITBUCKET_TEAM")
	rName := acctest.RandomWithPrefix("tf-test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBitbucketRepositoryGroupPermissionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBitbucketRepositoryGroupPermissionConfig(workspace, rName, "read"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBitbucketRepositoryGroupPermissionExists(resourceName, &repositoryGroupPermission),
					resource.TestCheckResourceAttrPair(resourceName, "repo_slug", "bitbucket_repository.test", "name"),
					resource.TestCheckResourceAttrPair(resourceName, "group_slug", "bitbucket_group.test", "slug"),
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
				Config: testAccBitbucketRepositoryGroupPermissionConfig(workspace, rName, "write"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBitbucketRepositoryGroupPermissionExists(resourceName, &repositoryGroupPermission),
					resource.TestCheckResourceAttrPair(resourceName, "repo_slug", "bitbucket_repository.test", "name"),
					resource.TestCheckResourceAttrPair(resourceName, "group_slug", "bitbucket_group.test", "slug"),
					resource.TestCheckResourceAttr(resourceName, "workspace", workspace),
					resource.TestCheckResourceAttr(resourceName, "permission", "write"),
				),
			},
		},
	})
}

func testAccCheckBitbucketRepositoryGroupPermissionDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(Clients).httpClient
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "bitbucket_repository_group_permission" {
			continue
		}

		response, err := client.Get(fmt.Sprintf("2.0/repositories/%s/%s/permissions-config/groups/%s", rs.Primary.Attributes["workspace"], rs.Primary.Attributes["repo_slug"], rs.Primary.Attributes["group_slug"]))

		if err == nil {
			return fmt.Errorf("The resource was found should have errored")
		}

		if response.StatusCode != http.StatusNotFound {
			return fmt.Errorf("Repository Group Permission still exists")
		}

	}
	return nil
}

func testAccCheckBitbucketRepositoryGroupPermissionExists(n string, repositoryGroupPermission *RepositoryGroupPermission) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found %s", n)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No Repository Group Permission ID is set")
		}
		return nil
	}
}

func testAccBitbucketRepositoryGroupPermissionConfig(workspace, rName, permission string) string {
	return fmt.Sprintf(`
resource "bitbucket_repository" "test" {
  owner = %[1]q
  name  = %[2]q
}

resource "bitbucket_group" "test" {
  workspace  = %[1]q
  name       = %[2]q
}

resource "bitbucket_repository_group_permission" "test" {
  workspace  = %[1]q
  repo_slug  = bitbucket_repository.test.name
  group_slug = bitbucket_group.test.slug
  permission = %[3]q
}
`, workspace, rName, permission)
}
