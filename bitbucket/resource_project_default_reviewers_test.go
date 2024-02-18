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

func TestAccBitbucketProjectDefaultReviewers_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-test")
	workspace := os.Getenv("BITBUCKET_TEAM")
	resourceName := "bitbucket_project_default_reviewers.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBitbucketProjectDefaultReviewersDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBitbucketProjectDefaultReviewersConfig(workspace, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBitbucketProjectDefaultReviewersExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "project", "bitbucket_project.test", "key"),
					resource.TestCheckResourceAttr(resourceName, "reviewers.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "reviewers.*", "data.bitbucket_current_user.test", "uuid"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccBitbucketProjectDefaultReviewersConfig(workspace, rName string) string {
	return fmt.Sprintf(`
data "bitbucket_current_user" "test" {}

resource "bitbucket_project" "test" {
  owner = %[1]q
  name  = %[2]q
  key   = "CCCCCCCC"
}

resource "bitbucket_project_default_reviewers" "test" {
  workspace = %[1]q
  project   = bitbucket_project.test.key
  reviewers = [data.bitbucket_current_user.test.uuid]
}
`, workspace, rName)
}

func testAccCheckBitbucketProjectDefaultReviewersDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(Clients).genClient
	projectsApi := client.ApiClient.ProjectsApi

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "bitbucket_project_default_reviewers" {
			continue
		}

		_, response, _ := projectsApi.WorkspacesWorkspaceProjectsProjectKeyDefaultReviewersGet(client.AuthContext, rs.Primary.Attributes["project"], rs.Primary.Attributes["workspace"], nil)

		if response.StatusCode != http.StatusNotFound {
			return fmt.Errorf("Project Defaults Reviewer still exists")
		}
	}
	return nil
}

func testAccCheckBitbucketProjectDefaultReviewersExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("Not found %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No default reviewers ID is set")
		}

		return nil
	}
}
