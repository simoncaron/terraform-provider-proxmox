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

func TestAccResourceNotificationsEndpointSMTP(t *testing.T) {
	te := test.InitEnvironment(t)

	tests := []struct {
		name  string
		steps []resource.TestStep
	}{
		{
			"create and update SMTP endpoint",
			[]resource.TestStep{
				{
					Config: te.RenderConfig(`
					resource "proxmox_notifications_endpoint_smtp" "test" {
						name         = "acc-test-smtp"
						server       = "smtp.example.com"
						from_address = "pve@example.com"
						mailto       = ["admin@example.com"]
					}`),
					Check: resource.ComposeTestCheckFunc(
						test.ResourceAttributes("proxmox_notifications_endpoint_smtp.test", map[string]string{
							"id":           "acc-test-smtp",
							"name":         "acc-test-smtp",
							"server":       "smtp.example.com",
							"from_address": "pve@example.com",
							"mode":         "tls",
							"disable":      "false",
							"mailto.#":     "1",
						}),
					),
				},
				{
					Config: te.RenderConfig(`
					resource "proxmox_notifications_endpoint_smtp" "test" {
						name         = "acc-test-smtp"
						server       = "smtp.example.com"
						from_address = "pve@example.com"
						author       = "Proxmox Alerts"
						mode         = "starttls"
						port         = 587
						username     = "smtp-user"
						password     = "smtp-pass"
						mailto       = ["admin@example.com", "ops@example.com"]
					}`),
					Check: resource.ComposeTestCheckFunc(
						test.ResourceAttributes("proxmox_notifications_endpoint_smtp.test", map[string]string{
							"author":   "Proxmox Alerts",
							"mode":     "starttls",
							"port":     "587",
							"username": "smtp-user",
							"mailto.#": "2",
						}),
					),
				},
				{
					Config: te.RenderConfig(`
					resource "proxmox_notifications_endpoint_smtp" "test" {
						name         = "acc-test-smtp"
						server       = "smtp.example.com"
						from_address = "pve@example.com"
					}`),
					Check: resource.ComposeTestCheckFunc(
						test.NoResourceAttributesSet("proxmox_notifications_endpoint_smtp.test", []string{
							"author",
							"username",
							"port",
						}),
					),
				},
			},
		},
		{
			"import SMTP endpoint",
			[]resource.TestStep{
				{
					Config: te.RenderConfig(`
					resource "proxmox_notifications_endpoint_smtp" "import_test" {
						name         = "acc-test-smtp-import"
						server       = "smtp.example.com"
						from_address = "pve@example.com"
					}`),
				},
				{
					ResourceName:      "proxmox_notifications_endpoint_smtp.import_test",
					ImportState:       true,
					ImportStateVerify: false, // password cannot be imported
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
