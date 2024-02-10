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

func TestAccBitbucketProjectGroupPermission_basic(t *testing.T) {
	var projectGroupPermission ProjectGroupPermission
	resourceName := "bitbucket_project_group_permission.test"
	workspace := os.Getenv("BITBUCKET_TEAM")
	rName := acctest.RandomWithPrefix("tf-test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBitbucketProjectGroupPermissionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBitbucketProjectGroupPermissionConfig(workspace, rName, "read"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBitbucketProjectGroupPermissionExists(resourceName, &projectGroupPermission),
					resource.TestCheckResourceAttrPair(resourceName, "project_key", "bitbucket_project.test", "key"),
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
				Config: testAccBitbucketProjectGroupPermissionConfig(workspace, rName, "write"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBitbucketProjectGroupPermissionExists(resourceName, &projectGroupPermission),
					resource.TestCheckResourceAttrPair(resourceName, "project_key", "bitbucket_project.test", "key"),
					resource.TestCheckResourceAttrPair(resourceName, "group_slug", "bitbucket_group.test", "slug"),
					resource.TestCheckResourceAttr(resourceName, "workspace", workspace),
					resource.TestCheckResourceAttr(resourceName, "permission", "write"),
				),
			},
		},
	})
}

func testAccCheckBitbucketProjectGroupPermissionDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(Clients).httpClient
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "bitbucket_project_group_permission" {
			continue
		}

		response, err := client.Get(fmt.Sprintf("2.0/repositories/%s/%s/permissions-config/groups/%s", rs.Primary.Attributes["workspace"], rs.Primary.Attributes["project_key"], rs.Primary.Attributes["group_slug"]))

		if err == nil {
			return fmt.Errorf("The resource was found should have errored")
		}

		if response.StatusCode != http.StatusNotFound {
			return fmt.Errorf("Project Group Permission still exists")
		}

	}
	return nil
}

func testAccCheckBitbucketProjectGroupPermissionExists(n string, projectGroupPermission *ProjectGroupPermission) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found %s", n)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No Project Group Permission ID is set")
		}
		return nil
	}
}

func testAccBitbucketProjectGroupPermissionConfig(workspace, rName, permission string) string {
	return fmt.Sprintf(`
resource "bitbucket_project" "test" {
  owner = %[1]q
  name  = %[2]q
  key   = "GRPPERM"
}

resource "bitbucket_group" "test" {
  workspace  = %[1]q
  name       = %[2]q
}

resource "bitbucket_project_group_permission" "test" {
  workspace   = %[1]q
  project_key = bitbucket_project.test.key
  group_slug  = bitbucket_group.test.slug
  permission  = %[3]q
}
`, workspace, rName, permission)
}
