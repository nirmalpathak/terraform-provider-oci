// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

const (
	PrivateIpRequiredOnlyResource = PrivateIpResourceDependencies + `
resource "oci_core_private_ip" "test_private_ip" {
	#Required
	vnic_id = "${lookup(data.oci_core_vnic_attachments.t.vnic_attachments[0], "vnic_id")}"
}
`

	PrivateIpResourceConfig = PrivateIpResourceDependencies + `
resource "oci_core_private_ip" "test_private_ip" {
	#Required
	vnic_id = "${lookup(data.oci_core_vnic_attachments.t.vnic_attachments[0], "vnic_id")}"

	#Optional
	defined_tags = "${map("${oci_identity_tag_namespace.tag-namespace1.name}.${oci_identity_tag.tag1.name}", "${var.private_ip_defined_tags_value}")}"
	display_name = "${var.private_ip_display_name}"
	freeform_tags = "${var.private_ip_freeform_tags}"
	hostname_label = "${var.private_ip_hostname_label}"
	ip_address = "${var.private_ip_ip_address}"
}
`
	PrivateIpPropertyVariables = `
variable "private_ip_defined_tags_value" { default = "value" }
variable "private_ip_display_name" { default = "displayName" }
variable "private_ip_freeform_tags" { default = {"Department"= "Finance"} }
variable "private_ip_hostname_label" { default = "privateiptestinstance" }
variable "private_ip_ip_address" { default = "10.0.1.5" }

`
	PrivateIpResourceDependencies = instanceDnsConfig + `
	data "oci_core_vnic_attachments" "t" {
		availability_domain = "${data.oci_identity_availability_domains.test_availability_domains.availability_domains.0.name}"
		compartment_id = "${var.compartment_id}"
		instance_id = "${oci_core_instance.t.id}"
	}

` + AvailabilityDomainConfig
)

func TestCorePrivateIpResource_basic(t *testing.T) {
	provider := testAccProvider
	config := testProviderConfig()

	compartmentId := getRequiredEnvSetting("compartment_ocid")
	compartmentIdVariableStr := fmt.Sprintf("variable \"compartment_id\" { default = \"%s\" }\n", compartmentId)

	resourceName := "oci_core_private_ip.test_private_ip"
	datasourceName := "data.oci_core_private_ips.test_private_ips"

	var resId, resId2 string

	resource.Test(t, resource.TestCase{
		Providers: map[string]terraform.ResourceProvider{
			"oci": provider,
		},
		Steps: []resource.TestStep{
			// verify create
			{
				ImportState:       true,
				ImportStateVerify: true,
				Config:            config + PrivateIpPropertyVariables + compartmentIdVariableStr + PrivateIpRequiredOnlyResource,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "subnet_id"),
					resource.TestCheckResourceAttrSet(resourceName, "vnic_id"),

					func(s *terraform.State) (err error) {
						resId, err = fromInstanceState(s, resourceName, "id")
						return err
					},
				),
			},

			// delete before next create
			{
				Config: config + compartmentIdVariableStr + PrivateIpResourceDependencies,
			},
			// verify create with optionals
			{
				Config: config + PrivateIpPropertyVariables + compartmentIdVariableStr + PrivateIpResourceConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "defined_tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "display_name", "displayName"),
					resource.TestCheckResourceAttr(resourceName, "freeform_tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "hostname_label", "privateiptestinstance"),
					resource.TestCheckResourceAttr(resourceName, "ip_address", "10.0.1.5"),
					resource.TestCheckResourceAttrSet(resourceName, "subnet_id"),
					resource.TestCheckResourceAttrSet(resourceName, "vnic_id"),

					func(s *terraform.State) (err error) {
						resId, err = fromInstanceState(s, resourceName, "id")
						return err
					},
				),
			},

			// verify updates to updatable parameters
			{
				Config: config + `
variable "private_ip_defined_tags_value" { default = "updatedValue" }
variable "private_ip_display_name" { default = "displayName2" }
variable "private_ip_freeform_tags" { default = {"Department"= "Accounting"} }
variable "private_ip_hostname_label" { default = "privateiptestinstance2" }
variable "private_ip_ip_address" { default = "10.0.1.5" }

                ` + compartmentIdVariableStr + PrivateIpResourceConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "defined_tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "display_name", "displayName2"),
					resource.TestCheckResourceAttr(resourceName, "freeform_tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "hostname_label", "privateiptestinstance2"),
					resource.TestCheckResourceAttr(resourceName, "ip_address", "10.0.1.5"),
					resource.TestCheckResourceAttrSet(resourceName, "subnet_id"),
					resource.TestCheckResourceAttrSet(resourceName, "vnic_id"),

					func(s *terraform.State) (err error) {
						resId2, err = fromInstanceState(s, resourceName, "id")
						if resId != resId2 {
							return fmt.Errorf("Resource recreated when it was supposed to be updated.")
						}
						return err
					},
				),
			},
			// verify datasource
			{
				Config: config + `
variable "private_ip_defined_tags_value" { default = "updatedValue" }
variable "private_ip_display_name" { default = "displayName2" }
variable "private_ip_freeform_tags" { default = {"Department"= "Accounting"} }
variable "private_ip_hostname_label" { default = "privateiptestinstance2" }
variable "private_ip_ip_address" { default = "10.0.1.5" }

data "oci_core_private_ips" "test_private_ips" {

	#Optional
	vnic_id = "${lookup(data.oci_core_vnic_attachments.t.vnic_attachments[0], "vnic_id")}"

    filter {
    	name = "id"
    	values = ["${oci_core_private_ip.test_private_ip.id}"]
    }
}
                ` + compartmentIdVariableStr + PrivateIpResourceConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(datasourceName, "vnic_id"),

					resource.TestCheckResourceAttr(datasourceName, "private_ips.#", "1"),
					resource.TestCheckResourceAttr(datasourceName, "private_ips.0.defined_tags.%", "1"),
					resource.TestCheckResourceAttr(datasourceName, "private_ips.0.display_name", "displayName2"),
					resource.TestCheckResourceAttr(datasourceName, "private_ips.0.freeform_tags.%", "1"),
					resource.TestCheckResourceAttr(datasourceName, "private_ips.0.hostname_label", "privateiptestinstance2"),
					resource.TestCheckResourceAttr(datasourceName, "private_ips.0.ip_address", "10.0.1.5"),
					resource.TestCheckResourceAttrSet(datasourceName, "private_ips.0.subnet_id"),
					resource.TestCheckResourceAttrSet(datasourceName, "private_ips.0.vnic_id"),
				),
			},
		},
	})
}
