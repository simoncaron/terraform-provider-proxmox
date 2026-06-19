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

func TestAccResourceNotificationsEndpointWebhook(t *testing.T) {
	te := test.InitEnvironment(t)

	tests := []struct {
		name  string
		steps []resource.TestStep
	}{
		{
			"create and update webhook endpoint",
			[]resource.TestStep{
				{
					Config: te.RenderConfig(`
					resource "proxmox_notifications_endpoint_webhook" "test" {
						name   = "acc-test-webhook"
						url    = "https://webhook.example.com/notify"
						method = "post"
					}`),
					Check: resource.ComposeTestCheckFunc(
						test.ResourceAttributes("proxmox_notifications_endpoint_webhook.test", map[string]string{
							"id":      "acc-test-webhook",
							"name":    "acc-test-webhook",
							"url":     "https://webhook.example.com/notify",
							"method":  "post",
							"disable": "false",
						}),
					),
				},
				{
					Config: te.RenderConfig(`
					resource "proxmox_notifications_endpoint_webhook" "test" {
						name    = "acc-test-webhook"
						url     = "https://webhook.example.com/notify"
						method  = "post"
						comment = "test webhook"
						body    = "{\"text\": \"{{message}}\"}"
						header = [
							{
								name  = "Content-Type"
								value = "application/json"
							}
						]
						secret = [
							{
								name  = "api_key"
								value = "super-secret"
							}
						]
					}`),
					Check: resource.ComposeTestCheckFunc(
						test.ResourceAttributes("proxmox_notifications_endpoint_webhook.test", map[string]string{
							"comment":       "test webhook",
							"header.#":      "1",
							"header.0.name": "Content-Type",
						}),
					),
				},
				{
					Config: te.RenderConfig(`
					resource "proxmox_notifications_endpoint_webhook" "test" {
						name   = "acc-test-webhook"
						url    = "https://webhook.example.com/notify"
						method = "post"
					}`),
					Check: resource.ComposeTestCheckFunc(
						test.NoResourceAttributesSet("proxmox_notifications_endpoint_webhook.test", []string{
							"comment",
							"body",
						}),
					),
				},
			},
		},
		{
			"import webhook endpoint",
			[]resource.TestStep{
				{
					Config: te.RenderConfig(`
					resource "proxmox_notifications_endpoint_webhook" "import_test" {
						name   = "acc-test-webhook-import"
						url    = "https://webhook.example.com/notify"
						method = "post"
					}`),
				},
				{
					ResourceName:      "proxmox_notifications_endpoint_webhook.import_test",
					ImportState:       true,
					ImportStateVerify: false, // secret values cannot be imported
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
