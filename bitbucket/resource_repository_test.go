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

func TestAccBitbucketRepository_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-test")
	testUser := os.Getenv("BITBUCKET_TEAM")
	resourceName := "bitbucket_repository.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBitbucketRepositoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBitbucketRepoConfig(testUser, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBitbucketRepositoryExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "owner", testUser),
					resource.TestCheckResourceAttr(resourceName, "scm", "git"),
					resource.TestCheckResourceAttr(resourceName, "has_wiki", "false"),
					resource.TestCheckResourceAttrSet(resourceName, "uuid"),
					resource.TestCheckResourceAttr(resourceName, "fork_policy", "allow_forks"),
					resource.TestCheckResourceAttr(resourceName, "language", ""),
					resource.TestCheckResourceAttr(resourceName, "has_issues", "false"),
					resource.TestCheckResourceAttr(resourceName, "slug", rName),
					resource.TestCheckResourceAttr(resourceName, "is_private", "true"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "link.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "link.0.avatar.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "link.0.avatar.0.href"),
					resource.TestCheckResourceAttrSet(resourceName, "project_key"),
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

func TestAccBitbucketRepository_project(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-test")
	testUser := os.Getenv("BITBUCKET_TEAM")
	resourceName := "bitbucket_repository.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBitbucketRepositoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBitbucketRepoProjectConfig(testUser, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBitbucketRepositoryExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "owner", testUser),
					resource.TestCheckResourceAttrPair(resourceName, "project_key", "bitbucket_project.test", "key"),
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

func TestAccBitbucketRepository_avatar(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-test")
	testUser := os.Getenv("BITBUCKET_TEAM")
	resourceName := "bitbucket_repository.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBitbucketRepositoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBitbucketRepoAvatarConfig(testUser, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBitbucketRepositoryExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "link.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "link.0.avatar.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "link.0.avatar.0.href"),
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

func TestAccBitbucketRepository_slug(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-test")
	rSlug := acctest.RandomWithPrefix("tf-test")
	testUser := os.Getenv("BITBUCKET_TEAM")
	resourceName := "bitbucket_repository.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBitbucketRepositoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBitbucketRepoSlugConfig(testUser, rName, rSlug),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBitbucketRepositoryExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "slug", rSlug),
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

func testAccBitbucketRepoConfig(testUser, rName string) string {
	return fmt.Sprintf(`
resource "bitbucket_repository" "test" {
  owner = %[1]q
  name  = %[2]q
}
`, testUser, rName)
}

func testAccBitbucketRepoProjectConfig(testUser, rName string) string {
	return fmt.Sprintf(`
resource "bitbucket_project" "test" {
  owner = %[1]q
  name  = %[2]q
  key   = "AAAAAAA"
}
	
resource "bitbucket_repository" "test" {
  owner       = %[1]q
  name        = %[2]q
  project_key = bitbucket_project.test.key
}
`, testUser, rName)
}

func testAccBitbucketRepoAvatarConfig(testUser, rName string) string {
	return fmt.Sprintf(`
resource "bitbucket_repository" "test" {
  owner = %[1]q
  name  = %[2]q

  link {
    avatar {
      href = "https://d301sr5gafysq2.cloudfront.net/dfb18959be9c/img/repo-avatars/python.png"
	}
  }  
}
`, testUser, rName)
}

func testAccBitbucketRepoSlugConfig(testUser, rName, rSlug string) string {
	return fmt.Sprintf(`
resource "bitbucket_repository" "test" {
  owner = %[1]q
  name  = %[2]q
  slug  = %[3]q
}
`, testUser, rName, rSlug)
}

func testAccCheckBitbucketRepositoryDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(Clients).genClient
	repoApi := client.ApiClient.RepositoriesApi

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "bitbucket_repository" {
			continue
		}
		_, res, err := repoApi.RepositoriesWorkspaceRepoSlugGet(client.AuthContext,
			rs.Primary.Attributes["name"], rs.Primary.Attributes["owner"])

		if err == nil {
			return fmt.Errorf("The repository was found should have errored")
		}

		if res.StatusCode != http.StatusNotFound {
			return fmt.Errorf("Repository still exists")
		}
	}
	return nil
}

func testAccCheckBitbucketRepositoryExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found %s", n)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No repository ID is set")
		}
		return nil
	}
}
