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

func TestAccResourceNotificationsEndpointGotify(t *testing.T) {
	te := test.InitEnvironment(t)

	tests := []struct {
		name  string
		steps []resource.TestStep
	}{
		{
			"create and update gotify endpoint",
			[]resource.TestStep{
				{
					Config: te.RenderConfig(`
					resource "proxmox_notifications_endpoint_gotify" "test" {
						name   = "acc-test-gotify"
						server = "https://gotify.example.com"
						token  = "test-token"
					}`),
					Check: resource.ComposeTestCheckFunc(
						test.ResourceAttributes("proxmox_notifications_endpoint_gotify.test", map[string]string{
							"id":      "acc-test-gotify",
							"name":    "acc-test-gotify",
							"server":  "https://gotify.example.com",
							"disable": "false",
						}),
					),
				},
				{
					Config: te.RenderConfig(`
					resource "proxmox_notifications_endpoint_gotify" "test" {
						name    = "acc-test-gotify"
						server  = "https://gotify.example.com"
						token   = "test-token"
						comment = "updated comment"
						disable = true
					}`),
					Check: resource.ComposeTestCheckFunc(
						test.ResourceAttributes("proxmox_notifications_endpoint_gotify.test", map[string]string{
							"comment": "updated comment",
							"disable": "true",
						}),
					),
				},
				{
					Config: te.RenderConfig(`
					resource "proxmox_notifications_endpoint_gotify" "test" {
						name   = "acc-test-gotify"
						server = "https://gotify.example.com"
						token  = "test-token"
					}`),
					Check: resource.ComposeTestCheckFunc(
						test.NoResourceAttributesSet("proxmox_notifications_endpoint_gotify.test", []string{
							"comment",
						}),
					),
				},
			},
		},
		{
			"import gotify endpoint",
			[]resource.TestStep{
				{
					Config: te.RenderConfig(`
					resource "proxmox_notifications_endpoint_gotify" "import_test" {
						name   = "acc-test-gotify-import"
						server = "https://gotify.example.com"
						token  = "test-token"
					}`),
				},
				{
					ResourceName:      "proxmox_notifications_endpoint_gotify.import_test",
					ImportState:       true,
					ImportStateVerify: false, // token cannot be imported
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
