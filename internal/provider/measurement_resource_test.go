// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	//"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccMeasurementResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		//PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + testAccMeasurementResourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ripe-atlas_measurement.test", "description", "MyFirstTest"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "ripe-atlas_measurement.test",
				ImportState:       true,
				ImportStateVerify: true,
				// This is not normally necessary, but is here because this
				// example code does not have an actual upstream service.
				// Once the Read method is able to refresh information from
				// the upstream service, this can be removed.
				ImportStateVerifyIgnore: []string{"description", "MyFirstTest"},
			},
			// Update and Read testing
			{
				Config: providerConfig + testAccMeasurementResourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ripe-atlas_measurement.test", "description", "MyFirstTest2"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccMeasurementResourceConfig() string {
	return `
	resource "ripe-atlas_measurement" "test" {
		description = "MyFirstTest"
		type        = "ping"
	}
	`
}
