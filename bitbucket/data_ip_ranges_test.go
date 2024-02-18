package bitbucket

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceIPRanges_basic(t *testing.T) {
	dataSourceName := "data.bitbucket_ip_ranges.test"
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccBitbucketIPRangesConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckTypeSetElemNestedAttrs(dataSourceName, "ranges.*", map[string]string{
						"network":   "3.26.128.128",
						"mask_len":  "26",
						"cidr":      "3.26.128.128/26",
						"mask":      "255.255.255.192",
						"perimeter": "commercial",
					}),
				),
			},
		},
	})
}

func testAccBitbucketIPRangesConfig() string {
	return `
data "bitbucket_ip_ranges" "test" {}
`
}
