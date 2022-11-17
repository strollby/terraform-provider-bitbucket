package bitbucket

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type PaginatedIPRanges struct {
	Items     []IPRange `json:"items,omitempty"`
	SyncToken int       `json:"syncToken,omitempty"`
}

type IPRange struct {
	Network    string   `json:"network"`
	MaskLen    int      `json:"mask_len"`
	CIDR       string   `json:"cidr"`
	Mask       string   `json:"mask"`
	Regions    []string `json:"region"`
	Products   []string `json:"product"`
	Directions []string `json:"direction"`
}

func dataIPRanges() *schema.Resource {
	return &schema.Resource{
		Read: dataReadIPRanges,

		Schema: map[string]*schema.Schema{
			"ranges": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"network": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"mask_len": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"cidr": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"mask": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"regions": {
							Type:     schema.TypeSet,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Computed: true,
						},
						"products": {
							Type:     schema.TypeSet,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Computed: true,
						},
						"directions": {
							Type:     schema.TypeSet,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataReadIPRanges(d *schema.ResourceData, m interface{}) error {

	req, err := http.Get("https://ip-ranges.atlassian.com/")
	if err != nil {
		return err
	}

	if req.StatusCode == http.StatusNotFound {
		return fmt.Errorf("IP whitelist not found")
	}

	body, readerr := io.ReadAll(req.Body)
	if readerr != nil {
		return readerr
	}

	log.Printf("[DEBUG] IP Ranges Response JSON: %v", string(body))

	var pageIpRanges PaginatedIPRanges

	decodeerr := json.Unmarshal(body, &pageIpRanges)
	if decodeerr != nil {
		return decodeerr
	}

	log.Printf("[DEBUG] IP Ranges Decoded: %#v", pageIpRanges)

	d.SetId(fmt.Sprintf("%d", pageIpRanges.SyncToken))
	d.Set("ranges", flattenIPRanges(pageIpRanges.Items))

	return nil
}

func flattenIPRanges(ranges []IPRange) []interface{} {
	if len(ranges) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, btRaw := range ranges {
		log.Printf("[DEBUG] IP Range Response Decoded: %#v", btRaw)

		ipRange := map[string]interface{}{
			"cidr":       btRaw.CIDR,
			"mask":       btRaw.Mask,
			"mask_len":   btRaw.MaskLen,
			"network":    btRaw.Network,
			"directions": btRaw.Directions,
			"products":   btRaw.Products,
			"regions":    btRaw.Regions,
		}

		tfList = append(tfList, ipRange)
	}

	return tfList
}
