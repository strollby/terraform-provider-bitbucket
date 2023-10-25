package bitbucket

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccBitbucketWorkspaceHook_basic(t *testing.T) {
	var hook Hook
	resourceName := "bitbucket_workspace_hook.test"
	workspace := os.Getenv("BITBUCKET_TEAM")
	rName := acctest.RandomWithPrefix("tf-test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBitbucketWorkspaceHookDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBitbucketWorkspaceHookConfig(workspace, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBitbucketWorkspaceHookExists(resourceName, &hook),
					resource.TestCheckResourceAttr(resourceName, "description", "Test hook for terraform"),
					resource.TestCheckResourceAttr(resourceName, "workspace", workspace),
					resource.TestCheckResourceAttr(resourceName, "url", "https://httpbin.org"),
					resource.TestCheckResourceAttr(resourceName, "skip_cert_verification", "true"),
					resource.TestCheckResourceAttr(resourceName, "secret_set", "false"),
					resource.TestCheckResourceAttr(resourceName, "history_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "active", "true"),
					resource.TestCheckResourceAttr(resourceName, "events.#", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccBitbucketWorkspaceHookImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
			{
				Config: testAccBitbucketWorkspaceHookConfigUpdated(workspace, rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBitbucketWorkspaceHookExists(resourceName, &hook),
					resource.TestCheckResourceAttr(resourceName, "description", "Test hook for terraform Updated"),
					resource.TestCheckResourceAttr(resourceName, "workspace", workspace),
					resource.TestCheckResourceAttr(resourceName, "url", "https://httpbin.org"),
					resource.TestCheckResourceAttr(resourceName, "skip_cert_verification", "false"),
					resource.TestCheckResourceAttr(resourceName, "secret_set", "true"),
					resource.TestCheckResourceAttr(resourceName, "history_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "active", "false"),
					resource.TestCheckResourceAttr(resourceName, "events.#", "2"),
				),
			},
			{
				Config: testAccBitbucketWorkspaceHookConfig(workspace, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBitbucketWorkspaceHookExists(resourceName, &hook),
					resource.TestCheckResourceAttr(resourceName, "description", "Test hook for terraform"),
					resource.TestCheckResourceAttr(resourceName, "workspace", workspace),
					resource.TestCheckResourceAttr(resourceName, "url", "https://httpbin.org"),
					resource.TestCheckResourceAttr(resourceName, "skip_cert_verification", "true"),
					resource.TestCheckResourceAttr(resourceName, "secret_set", "false"),
					resource.TestCheckResourceAttr(resourceName, "history_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "active", "true"),
					resource.TestCheckResourceAttr(resourceName, "events.#", "1"),
				),
			},
		},
	})
}

func testAccCheckBitbucketWorkspaceHookDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(Clients).httpClient
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "bitbucket_workspace_hook" {
			continue
		}

		response, err := client.Get(fmt.Sprintf("2.0/workspaces/%s/hooks/%s", rs.Primary.Attributes["workspace"], url.PathEscape(rs.Primary.Attributes["uuid"])))

		if err == nil {
			return fmt.Errorf("The resource was found should have errored")
		}

		if response.StatusCode != http.StatusNotFound {
			return fmt.Errorf("Hook still exists")
		}

	}
	return nil
}

func testAccCheckBitbucketWorkspaceHookExists(n string, hook *Hook) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found %s", n)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No Hook ID is set")
		}
		return nil
	}
}

func testAccBitbucketWorkspaceHookConfig(workspace, rName string) string {
	return fmt.Sprintf(`
resource "bitbucket_workspace_hook" "test" {
  workspace              = %[1]q
  description            = "Test hook for terraform"
  url                    = "https://httpbin.org"
  skip_cert_verification = true

  events = [
  	"repo:push",
  ]
}
`, workspace, rName)
}

func testAccBitbucketWorkspaceHookConfigUpdated(workspace, rName string, enable bool) string {
	return fmt.Sprintf(`
resource "bitbucket_workspace_hook" "test" {
  workspace              = %[1]q
  description            = "Test hook for terraform Updated"
  url                    = "https://httpbin.org"
  skip_cert_verification = %[3]t
  active                 = %[3]t
  secret                 = %[2]q

  events = [
  	"repo:push",
    "repo:fork",
  ]
}
`, workspace, rName, enable)
}

func testAccBitbucketWorkspaceHookImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}
		return fmt.Sprintf("%s/%s", rs.Primary.Attributes["workspace"], rs.Primary.ID), nil
	}
}
