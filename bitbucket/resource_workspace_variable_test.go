package bitbucket

import (
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccBitbucketWorkspaceVariable_basic(t *testing.T) {
	workspace := os.Getenv("BITBUCKET_TEAM")
	resourceName := "bitbucket_workspace_variable.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBitbucketWorkspaceVariableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBitbucketWorkspaceVariableConfig(workspace, "test", false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBitbucketWorkspaceVariableExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "workspace", workspace),
					resource.TestCheckResourceAttr(resourceName, "key", "test"),
					resource.TestCheckResourceAttr(resourceName, "value", "test"),
					resource.TestCheckResourceAttr(resourceName, "secured", "false"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccBitbucketWorkspaceVariableConfig(workspace, "test-2", false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBitbucketWorkspaceVariableExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "workspace", workspace),
					resource.TestCheckResourceAttr(resourceName, "key", "test"),
					resource.TestCheckResourceAttr(resourceName, "value", "test-2"),
					resource.TestCheckResourceAttr(resourceName, "secured", "false"),
				),
			},
		},
	})
}

func TestAccBitbucketWorkspaceVariable_manyVars(t *testing.T) {
	workspace := os.Getenv("BITBUCKET_TEAM")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBitbucketWorkspaceVariableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBitbucketWorkspaceVariableManyConfig(workspace, "test", false),
			},
		},
	})
}

func TestAccBitbucketWorkspaceVariable_secure(t *testing.T) {
	workspace := os.Getenv("BITBUCKET_TEAM")
	resourceName := "bitbucket_workspace_variable.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBitbucketWorkspaceVariableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBitbucketWorkspaceVariableConfig(workspace, "test", true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBitbucketWorkspaceVariableExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "workspace", workspace),
					resource.TestCheckResourceAttr(resourceName, "key", "test"),
					resource.TestCheckResourceAttr(resourceName, "value", "test"),
					resource.TestCheckResourceAttr(resourceName, "secured", "true"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"value"},
			},
			{
				Config: testAccBitbucketWorkspaceVariableConfig(workspace, "test", false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBitbucketWorkspaceVariableExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "workspace", workspace),
					resource.TestCheckResourceAttr(resourceName, "key", "test"),
					resource.TestCheckResourceAttr(resourceName, "value", "test"),
					resource.TestCheckResourceAttr(resourceName, "secured", "false"),
				),
			},
			{
				Config: testAccBitbucketWorkspaceVariableConfig(workspace, "test", true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBitbucketWorkspaceVariableExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "workspace", workspace),
					resource.TestCheckResourceAttr(resourceName, "key", "test"),
					resource.TestCheckResourceAttr(resourceName, "value", "test"),
					resource.TestCheckResourceAttr(resourceName, "secured", "true"),
				),
			},
		},
	})
}

func testAccCheckBitbucketWorkspaceVariableDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(Clients).genClient
	pipeApi := client.ApiClient.PipelinesApi
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "bitbucket_workspace_variable" {
			continue
		}

		workspace, uuid, err := workspaceVarId(rs.Primary.ID)
		if err != nil {
			return err
		}

		_, res, err := pipeApi.GetPipelineVariableForWorkspace(client.AuthContext, workspace, uuid)

		if err == nil {
			return fmt.Errorf("The resource was found should have errored")
		}

		if res.StatusCode != http.StatusNotFound {
			return fmt.Errorf("Workspace Variable still exists")
		}
	}
	return nil
}

func testAccCheckBitbucketWorkspaceVariableExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("Not found %s", n)
		}

		return nil
	}
}

func testAccBitbucketWorkspaceVariableConfig(workspace, val string, secure bool) string {
	return fmt.Sprintf(`
resource "bitbucket_workspace_variable" "test" {
  key       = "test"
  value     = %[2]q
  workspace = %[1]q
  secured   = %[3]t
}
`, workspace, val, secure)
}

func testAccBitbucketWorkspaceVariableManyConfig(workspace, val string, secure bool) string {
	return fmt.Sprintf(`
resource "bitbucket_workspace_variable" "test" {
  count = 50

  key       = "test${count.index}"
  value     = %[2]q
  workspace = %[1]q
  secured   = %[3]t
}
`, workspace, val, secure)
}
