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

func TestAccBitbucketDefaultReviewers_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-test")
	owner := os.Getenv("BITBUCKET_TEAM")
	resourceName := "bitbucket_default_reviewers.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBitbucketDefaultReviewersDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBitbucketDefaultReviewersConfig(owner, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBitbucketDefaultReviewersExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "repository", "bitbucket_repository.test", "name"),
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

func testAccBitbucketDefaultReviewersConfig(owner, rName string) string {
	return fmt.Sprintf(`
data "bitbucket_current_user" "test" {}

resource "bitbucket_repository" "test" {
  owner = %[1]q
  name  = %[2]q
}

resource "bitbucket_default_reviewers" "test" {
  owner      = %[1]q
  repository = bitbucket_repository.test.name
  reviewers  = [data.bitbucket_current_user.test.uuid]
}
`, owner, rName)
}

func testAccCheckBitbucketDefaultReviewersDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(Clients).genClient
	prApi := client.ApiClient.PullrequestsApi
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "bitbucket_default_reviewers" {
			continue
		}
		_, response, _ := prApi.RepositoriesWorkspaceRepoSlugDefaultReviewersGet(client.AuthContext, rs.Primary.Attributes["repository"], rs.Primary.Attributes["owner"], nil)

		if response.StatusCode != http.StatusNotFound {
			return fmt.Errorf("Defaults Reviewer still exists")
		}
	}
	return nil
}

func testAccCheckBitbucketDefaultReviewersExists(n string) resource.TestCheckFunc {
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
