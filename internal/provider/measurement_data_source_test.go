// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccMeasurementDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		//PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: providerConfig + testAccMeasurementDataSourceConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.ripe-atlas_measurement.mine", "measurements.0.id", "TEST"),
				),
			},
		},
	})
}

const testAccMeasurementDataSourceConfig = `
data "ripe-atlas_measurement" "mine" {
	#mine = true
}
`
