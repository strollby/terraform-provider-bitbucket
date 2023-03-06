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

func TestAccBitbucketProjectBranchingModel_basic(t *testing.T) {
	var branchRestriction BranchingModel
	rName := acctest.RandomWithPrefix("tf-test")
	workspace := os.Getenv("BITBUCKET_TEAM")
	resourceName := "bitbucket_project_branching_model.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBitbucketProjectBranchingModelDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBitbucketProjectBranchingModelConfig(workspace, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBitbucketProjectBranchingModelExists(resourceName, &branchRestriction),
					resource.TestCheckResourceAttrPair(resourceName, "project", "bitbucket_project.test", "key"),
					resource.TestCheckResourceAttr(resourceName, "development.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "development.0.use_mainbranch", "true"),
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

func TestAccBitbucketProjectBranchingModel_production(t *testing.T) {
	var branchRestriction BranchingModel
	rName := acctest.RandomWithPrefix("tf-test")
	workspace := os.Getenv("BITBUCKET_TEAM")
	resourceName := "bitbucket_project_branching_model.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBitbucketProjectBranchingModelDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBitbucketProjectBranchingModelProdConfig(workspace, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBitbucketProjectBranchingModelExists(resourceName, &branchRestriction),
					resource.TestCheckResourceAttrPair(resourceName, "project", "bitbucket_project.test", "key"),
					resource.TestCheckResourceAttr(resourceName, "development.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "development.0.use_mainbranch", "true"),
					resource.TestCheckResourceAttr(resourceName, "production.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "production.0.use_mainbranch", "true"),
					resource.TestCheckResourceAttr(resourceName, "production.0.enabled", "true"),
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

func TestAccBitbucketProjectBranchingModel_branchTypes(t *testing.T) {
	var branchRestriction BranchingModel
	rName := acctest.RandomWithPrefix("tf-test")
	workspace := os.Getenv("BITBUCKET_TEAM")
	resourceName := "bitbucket_project_branching_model.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBitbucketProjectBranchingModelDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBitbucketProjectBranchingModelBranchTypesConfig1(workspace, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBitbucketProjectBranchingModelExists(resourceName, &branchRestriction),
					resource.TestCheckResourceAttrPair(resourceName, "project", "bitbucket_project.test", "key"),
					resource.TestCheckResourceAttr(resourceName, "development.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "development.0.use_mainbranch", "true"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "branch_type.*", map[string]string{
						"kind":   "feature",
						"prefix": "test/",
					}),
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

func testAccBitbucketProjectBranchingModelConfig(workspace, rName string) string {
	return fmt.Sprintf(`
resource "bitbucket_project" "test" {
  owner = %[1]q
  name  = %[2]q
  key   = "DDDDDD"
}

resource "bitbucket_project_branching_model" "test" {
  workspace      = %[1]q
  project = bitbucket_project.test.key

  development {
    use_mainbranch = true
  }
}
`, workspace, rName)
}

func testAccBitbucketProjectBranchingModelProdConfig(workspace, rName string) string {
	return fmt.Sprintf(`
resource "bitbucket_project" "test" {
  owner = %[1]q
  name  = %[2]q
  key   = "EEEEE"
}

resource "bitbucket_project_branching_model" "test" {
  workspace      = %[1]q
  project = bitbucket_project.test.key

  development {
    use_mainbranch = true
  }

  production {
    use_mainbranch = true
	enabled        = true
  }
}
`, workspace, rName)
}

func testAccBitbucketProjectBranchingModelBranchTypesConfig1(workspace, rName string) string {
	return fmt.Sprintf(`
resource "bitbucket_project" "test" {
  owner = %[1]q
  name  = %[2]q
  key   = "FFFFF"
}

resource "bitbucket_project_branching_model" "test" {
  workspace      = %[1]q
  project = bitbucket_project.test.key

  development {
    use_mainbranch = true
  }

  branch_type {
    enabled = true
	kind    = "feature"
	prefix  = "test/"
  }

  branch_type {
    enabled = true
	kind    = "hotfix"
	prefix  = "hotfix/"
  }
 
  branch_type {
    enabled = true
	kind    = "release"
	prefix  = "release/"
  }
 
  branch_type {
    enabled = true
	kind    = "bugfix"
	prefix  = "bugfix/"
  }   
}
`, workspace, rName)
}

func testAccCheckBitbucketProjectBranchingModelDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(Clients).httpClient
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "bitbucket_project_branching_model" {
			continue
		}
		response, err := client.Get(fmt.Sprintf("2.0/workspaces/%s/projects/%s/branching-model", rs.Primary.Attributes["workspace"], rs.Primary.Attributes["project"]))

		if err == nil {
			return fmt.Errorf("The resource was found should have errored")
		}

		if response.StatusCode != http.StatusNotFound {
			return fmt.Errorf("Project Branching Model still exists")
		}
	}

	return nil
}

func testAccCheckBitbucketProjectBranchingModelExists(n string, branchRestriction *BranchingModel) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found %s", n)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No BranchingModel ID is set")
		}
		return nil
	}
}
