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

func TestBitbucketRepository_ComputeSlug(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		Name           string
		Input          string
		ExpectedOutput string
	}{
		{
			Name:           "Left hand en-dash normalization",
			Input:          "------a_",
			ExpectedOutput: "a_",
		},
		{
			Name:           "Right hand en-dash normalization",
			Input:          "a_------",
			ExpectedOutput: "a_",
		},
		{
			Name:           "En-dash consecutive normalization",
			Input:          "a---b_---a",
			ExpectedOutput: "a-b_-a",
		},
		{
			Name:           "En-dash consecutive and begin/end normalization",
			Input:          "--a---b_---a----",
			ExpectedOutput: "a-b_-a",
		},
		{
			Name:           "Allow dot character",
			Input:          "test.repository",
			ExpectedOutput: "test.repository",
		},
		{
			Name:           `Replace & with en-dash`,
			Input:          "test&repository",
			ExpectedOutput: "test-repository",
		},
		{
			Name:           `Replace multiple ; with en-dash`,
			Input:          "test;repository;;",
			ExpectedOutput: "test-repository",
		},
		{
			Name:           `Truncate long slug with en-dash`,
			Input:          "my-very-long-repository-name-that-is-over-the-max-allow-characters",
			ExpectedOutput: "my-very-long-repository-name-that-is-over-the-max-allow",
		},
		{
			Name:           `Do not truncate long slug without en-dash`,
			Input:          "myverylongrepositorynamethatisoverthemaxallowcharactersmyverylongrepositorynamethatisoverthemaxallowcharacters",
			ExpectedOutput: "myverylongrepositorynamethatisoverthemaxallowcharactersmyverylongrepositorynamethatisoverthemaxallowcharacters",
		},
	}

	for i, testCase := range testCases {
		i, testCase := i, testCase

		t.Run(testCase.Name, func(t *testing.T) {
			t.Parallel()

			result := computeSlug(testCase.Input)
			if result != testCase.ExpectedOutput {
				t.Fatalf("%d: expected result (%s), received: %s", i, testCase.ExpectedOutput, result)
			}
		})
	}
}

func TestAccBitbucketRepository_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-test")
	workspace := os.Getenv("BITBUCKET_TEAM")
	resourceName := "bitbucket_repository.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBitbucketRepositoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBitbucketRepoConfig(workspace, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBitbucketRepositoryExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "owner", workspace),
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
					resource.TestCheckResourceAttr(resourceName, "inherit_default_merge_strategy", "true"),
					resource.TestCheckResourceAttr(resourceName, "inherit_branching_model", "true"),
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
	workspace := os.Getenv("BITBUCKET_TEAM")
	resourceName := "bitbucket_repository.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBitbucketRepositoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBitbucketRepoProjectConfig(workspace, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBitbucketRepositoryExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "owner", workspace),
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
	workspace := os.Getenv("BITBUCKET_TEAM")
	resourceName := "bitbucket_repository.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBitbucketRepositoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBitbucketRepoAvatarConfig(workspace, rName),
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
	workspace := os.Getenv("BITBUCKET_TEAM")
	resourceName := "bitbucket_repository.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBitbucketRepositoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBitbucketRepoSlugConfig(workspace, rName, rSlug),
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

func TestAccBitbucketRepository_inherit(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-test")
	workspace := os.Getenv("BITBUCKET_TEAM")
	resourceName := "bitbucket_repository.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBitbucketRepositoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBitbucketRepoInheritConfig(workspace, rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBitbucketRepositoryExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "inherit_default_merge_strategy", "true"),
					resource.TestCheckResourceAttr(resourceName, "inherit_branching_model", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccBitbucketRepoInheritConfig(workspace, rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBitbucketRepositoryExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "inherit_default_merge_strategy", "false"),
					resource.TestCheckResourceAttr(resourceName, "inherit_branching_model", "true"),
				),
			},
			{
				Config: testAccBitbucketRepoInheritConfig(workspace, rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBitbucketRepositoryExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "inherit_default_merge_strategy", "true"),
					resource.TestCheckResourceAttr(resourceName, "inherit_branching_model", "true"),
				),
			},
		},
	})
}

func testAccBitbucketRepoInheritConfig(workspace, rName string, enable bool) string {
	return fmt.Sprintf(`
resource "bitbucket_repository" "test" {
  owner                          = %[1]q
  name                           = %[2]q
  inherit_default_merge_strategy = %[3]t
}
`, workspace, rName, enable)
}

func testAccBitbucketRepoConfig(workspace, rName string) string {
	return fmt.Sprintf(`
resource "bitbucket_repository" "test" {
  owner = %[1]q
  name  = %[2]q
}
`, workspace, rName)
}

func testAccBitbucketRepoProjectConfig(workspace, rName string) string {
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
`, workspace, rName)
}

func testAccBitbucketRepoAvatarConfig(workspace, rName string) string {
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
`, workspace, rName)
}

func testAccBitbucketRepoSlugConfig(workspace, rName, rSlug string) string {
	return fmt.Sprintf(`
resource "bitbucket_repository" "test" {
  owner = %[1]q
  name  = %[2]q
  slug  = %[3]q
}
`, workspace, rName, rSlug)
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
