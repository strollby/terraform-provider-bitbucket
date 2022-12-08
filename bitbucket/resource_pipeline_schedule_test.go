package bitbucket

import (
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccBitbucketPipelineSchedule_basic(t *testing.T) {
	resourceName := "bitbucket_pipeline_schedule.test"

	workspace := os.Getenv("BITBUCKET_TEAM")
	//because the schedule resource requires a pipe already defined we are passing here a bootstrapped repo
	repo := os.Getenv("BITBUCKET_PIPELINED_REPO")

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckPipeSchedule(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBitbucketPipelineScheduleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBitbucketPipelineScheduleConfig(workspace, repo, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBitbucketPipelineScheduleExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "workspace", workspace),
					resource.TestCheckResourceAttr(resourceName, "repository", repo),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "cron_pattern", "0 30 * * * ? *"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccBitbucketPipelineScheduleConfig(workspace, repo, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBitbucketPipelineScheduleExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "workspace", workspace),
					resource.TestCheckResourceAttr(resourceName, "repository", repo),
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "cron_pattern", "0 30 * * * ? *"),
				),
			},
		},
	})
}

func TestAccBitbucketPipelineSchedule_disabled(t *testing.T) {
	resourceName := "bitbucket_pipeline_schedule.test"

	workspace := os.Getenv("BITBUCKET_TEAM")
	//because the schedule resource requires a pipe already defined we are passing here a bootstrapped repo
	repo := os.Getenv("BITBUCKET_PIPELINED_REPO")

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckPipeSchedule(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBitbucketPipelineScheduleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBitbucketPipelineScheduleConfig(workspace, repo, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBitbucketPipelineScheduleExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "workspace", workspace),
					resource.TestCheckResourceAttr(resourceName, "repository", repo),
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "cron_pattern", "0 30 * * * ? *"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccBitbucketPipelineScheduleConfig(workspace, repo, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBitbucketPipelineScheduleExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "workspace", workspace),
					resource.TestCheckResourceAttr(resourceName, "repository", repo),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "cron_pattern", "0 30 * * * ? *"),
				),
			},
		},
	})
}

func testAccCheckBitbucketPipelineScheduleDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(Clients).genClient
	pipeApi := client.ApiClient.PipelinesApi

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "bitbucket_pipeline_schedule" {
			continue
		}

		workspace, repo, uuid, err := pipeScheduleId(rs.Primary.ID)
		if err != nil {
			return err
		}

		_, res, err := pipeApi.GetRepositoryPipelineSchedule(client.AuthContext, workspace, repo, uuid)

		if err == nil {
			return fmt.Errorf("The resource was found should have errored")
		}

		if res.StatusCode != http.StatusNotFound {
			return fmt.Errorf("Pipeline Schedule still exists")
		}

	}
	return nil
}

func testAccCheckBitbucketPipelineScheduleExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found %s", n)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No Pipeline Schedule ID is set")
		}
		return nil
	}
}

func testAccBitbucketPipelineScheduleConfig(workspace, repo string, enabled bool) string {
	return fmt.Sprintf(`
resource "bitbucket_pipeline_schedule" "test" {
  workspace    = %[1]q
  repository   = %[2]q
  enabled      = %[3]t
  cron_pattern = "0 30 * * * ? *"

  target {
    ref_name = "master"
	ref_type = "branch"
	selector {
      pattern = "staging"
	}
  }
}
`, workspace, repo, enabled)
}
