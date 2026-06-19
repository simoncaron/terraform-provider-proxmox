//go:build acceptance || all

/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */

//testacc:tier=light
//testacc:resource=misc

package notifications_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/bpg/terraform-provider-proxmox/fwprovider/test"
)

func TestAccResourceNotificationsEndpointSendmail(t *testing.T) {
	te := test.InitEnvironment(t)

	tests := []struct {
		name  string
		steps []resource.TestStep
	}{
		{
			"create and update sendmail endpoint",
			[]resource.TestStep{
				{
					Config: te.RenderConfig(`
					resource "proxmox_notifications_endpoint_sendmail" "test" {
						name         = "acc-test-sendmail"
						from_address = "pve@example.com"
						mailto       = ["admin@example.com"]
					}`),
					Check: resource.ComposeTestCheckFunc(
						test.ResourceAttributes("proxmox_notifications_endpoint_sendmail.test", map[string]string{
							"id":           "acc-test-sendmail",
							"name":         "acc-test-sendmail",
							"from_address": "pve@example.com",
							"disable":      "false",
							"mailto.#":     "1",
							"mailto.0":     "admin@example.com",
						}),
					),
				},
				{
					Config: te.RenderConfig(`
					resource "proxmox_notifications_endpoint_sendmail" "test" {
						name         = "acc-test-sendmail"
						from_address = "pve@example.com"
						author       = "Proxmox Alerts"
						comment      = "test comment"
						mailto       = ["admin@example.com", "ops@example.com"]
					}`),
					Check: resource.ComposeTestCheckFunc(
						test.ResourceAttributes("proxmox_notifications_endpoint_sendmail.test", map[string]string{
							"author":   "Proxmox Alerts",
							"comment":  "test comment",
							"mailto.#": "2",
						}),
					),
				},
				{
					Config: te.RenderConfig(`
					resource "proxmox_notifications_endpoint_sendmail" "test" {
						name = "acc-test-sendmail"
					}`),
					Check: resource.ComposeTestCheckFunc(
						test.NoResourceAttributesSet("proxmox_notifications_endpoint_sendmail.test", []string{
							"author",
							"comment",
							"from_address",
						}),
					),
				},
			},
		},
		{
			"import sendmail endpoint",
			[]resource.TestStep{
				{
					Config: te.RenderConfig(`
					resource "proxmox_notifications_endpoint_sendmail" "import_test" {
						name   = "acc-test-sendmail-import"
						mailto = ["admin@example.com"]
					}`),
				},
				{
					ResourceName:      "proxmox_notifications_endpoint_sendmail.import_test",
					ImportState:       true,
					ImportStateVerify: true,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resource.ParallelTest(t, resource.TestCase{
				ProtoV6ProviderFactories: te.AccProviders,
				Steps:                    tt.steps,
			})
		})
	}
}
