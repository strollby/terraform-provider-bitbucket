package bitbucket

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccBitbucketGroup_basic(t *testing.T) {
	var group UserGroup
	resourceName := "bitbucket_group.test"

	workspace := os.Getenv("BITBUCKET_TEAM")
	rName := acctest.RandomWithPrefix("tf-test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBitbucketGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBitbucketGroupConfig(workspace, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBitbucketGroupExists(resourceName, &group),
					resource.TestCheckResourceAttrPair(resourceName, "workspace", "data.bitbucket_workspace.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "slug"),
					resource.TestCheckResourceAttr(resourceName, "auto_add", "false"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccBitbucketGroupAutoAddConfig(workspace, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBitbucketGroupExists(resourceName, &group),
					resource.TestCheckResourceAttrPair(resourceName, "workspace", "data.bitbucket_workspace.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "slug"),
					resource.TestCheckResourceAttr(resourceName, "auto_add", "true"),
					resource.TestCheckResourceAttr(resourceName, "permission", "read"),
					resource.TestCheckResourceAttr(resourceName, "email_forwarding_disabled", "true"),
				),
			},
		},
	})
}

func testAccCheckBitbucketGroupDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(Clients).httpClient
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "bitbucket_group" {
			continue
		}

		response, err := client.Get(fmt.Sprintf("1.0/groups/%s/%s",
			rs.Primary.Attributes["workspace"], rs.Primary.Attributes["slug"]))

		if response.StatusCode == http.StatusNotFound {
			continue
		}

		if err != nil {
			return err
		}

		var group *UserGroup
		body, readerr := io.ReadAll(response.Body)
		if readerr != nil {
			return readerr
		}

		log.Printf("[DEBUG] Group Response Test JSON: %v", string(body))

		decodeerr := json.Unmarshal(body, &group)
		if decodeerr != nil {
			return decodeerr
		}

		log.Printf("[DEBUG] Group Response Test Decoded: %#v", group)

		if group != nil {
			return fmt.Errorf("Group still exists")
		}

	}
	return nil
}

func testAccCheckBitbucketGroupExists(n string, group *UserGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found %s", n)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No Group ID is set")
		}
		return nil
	}
}

func testAccBitbucketGroupConfig(workspace, rName string) string {
	return fmt.Sprintf(`
data "bitbucket_workspace" "test" {
  workspace = %[1]q
}

resource "bitbucket_group" "test" {
  workspace  = data.bitbucket_workspace.test.id
  name       = %[2]q
}
`, workspace, rName)
}

func testAccBitbucketGroupAutoAddConfig(workspace, rName string) string {
	return fmt.Sprintf(`
data "bitbucket_workspace" "test" {
  workspace = %[1]q
}

resource "bitbucket_group" "test" {
  workspace                 = data.bitbucket_workspace.test.id
  name                      = %[2]q
  auto_add                  = true
  permission                = "read"
  email_forwarding_disabled = true
}
`, workspace, rName)
}
