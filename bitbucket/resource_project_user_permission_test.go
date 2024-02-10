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

func TestAccBitbucketProjectUserPermission_basic(t *testing.T) {
	var projectUserPermission ProjectUserPermission
	resourceName := "bitbucket_project_user_permission.test"
	workspace := os.Getenv("BITBUCKET_TEAM")
	rName := acctest.RandomWithPrefix("tf-test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBitbucketProjectUserPermissionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBitbucketProjectUserPermissionConfig(workspace, rName, "read"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBitbucketProjectUserPermissionExists(resourceName, &projectUserPermission),
					resource.TestCheckResourceAttrPair(resourceName, "project_key", "bitbucket_project.test", "key"),
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
				Config: testAccBitbucketProjectUserPermissionConfig(workspace, rName, "write"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBitbucketProjectUserPermissionExists(resourceName, &projectUserPermission),
					resource.TestCheckResourceAttrPair(resourceName, "project_key", "bitbucket_project.test", "key"),
					resource.TestCheckResourceAttrPair(resourceName, "user_id", "data.bitbucket_current_user.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "workspace", workspace),
					resource.TestCheckResourceAttr(resourceName, "permission", "write"),
				),
			},
		},
	})
}

func testAccCheckBitbucketProjectUserPermissionDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(Clients).httpClient
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "bitbucket_project_user_permission" {
			continue
		}

		response, err := client.Get(fmt.Sprintf("2.0/repositories/%s/%s/permissions-config/users/%s", rs.Primary.Attributes["workspace"], rs.Primary.Attributes["project_key"], rs.Primary.Attributes["user_id"]))

		if err == nil {
			return fmt.Errorf("The resource was found should have errored")
		}

		if response.StatusCode != http.StatusNotFound {
			return fmt.Errorf("Project User Permission still exists")
		}

	}
	return nil
}

func testAccCheckBitbucketProjectUserPermissionExists(n string, projectUserPermission *ProjectUserPermission) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found %s", n)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No Project User Permission ID is set")
		}
		return nil
	}
}

func testAccBitbucketProjectUserPermissionConfig(workspace, rName, permission string) string {
	return fmt.Sprintf(`
resource "bitbucket_project" "test" {
  owner = %[1]q
  name  = %[2]q
  key   = "USERPERM"
}

data "bitbucket_current_user" "test" {}

resource "bitbucket_project_user_permission" "test" {
  workspace   = %[1]q
  project_key = bitbucket_project.test.key
  user_id     = data.bitbucket_current_user.test.id
  permission  = %[3]q
}
`, workspace, rName, permission)
}
