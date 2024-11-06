terraform {
  required_providers {
    ripe-atlas = {
      source = "registry.terraform.io/lavaburn/ripe-atlas"
    }
  }
}

provider "ripe-atlas" {
  api_key = "a404ba38-497a-47b0-b8f9-5cb6000dc33f"
}

resource "ripe-atlas_measurement" "test" {
  count       = 0
  description = "MyFirstTest1"
  type        = "ping"
  target      = "105.235.209.25"

  probe_set = [
    { 
        type   = "country"
        value  = "BE"
        number = 1
    },
    { 
        type   = "country"
        value  = "NL"
        number = 1
    },
    { 
        type   = "country"
        value  = "SS"
        number = 1
    },
  ]
}

#data "ripe-atlas_measurement" "mine" {
#    hidden = false
#}

#output "my_measurements" {
#  value = data.ripe-atlas_measurement.mine
#}

data "ripe-atlas_credits" "mine" {
    
}

output "my_credits" {
  value = data.ripe-atlas_credits.mine
}
