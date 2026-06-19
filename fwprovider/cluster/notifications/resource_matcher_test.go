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

func TestAccResourceNotificationsMatcher(t *testing.T) {
	te := test.InitEnvironment(t)

	tests := []struct {
		name  string
		steps []resource.TestStep
	}{
		{
			"create and update notification matcher",
			[]resource.TestStep{
				{
					Config: te.RenderConfig(`
					resource "proxmox_notifications_matcher" "test" {
						name = "acc-test-matcher"
					}`),
					Check: resource.ComposeTestCheckFunc(
						test.ResourceAttributes("proxmox_notifications_matcher.test", map[string]string{
							"id":      "acc-test-matcher",
							"name":    "acc-test-matcher",
							"disable": "false",
							"mode":    "all",
						}),
					),
				},
				{
					Config: te.RenderConfig(`
					resource "proxmox_notifications_matcher" "test" {
						name          = "acc-test-matcher"
						comment       = "test matcher"
						mode          = "any"
						match_severity = ["warning", "error"]
						match_field = [
							{
								type  = "exact"
								field = "type"
								value = "vzdump"
							}
						]
					}`),
					Check: resource.ComposeTestCheckFunc(
						test.ResourceAttributes("proxmox_notifications_matcher.test", map[string]string{
							"comment":             "test matcher",
							"mode":                "any",
							"match_severity.#":    "2",
							"match_field.#":       "1",
							"match_field.0.type":  "exact",
							"match_field.0.field": "type",
							"match_field.0.value": "vzdump",
						}),
					),
				},
				{
					Config: te.RenderConfig(`
					resource "proxmox_notifications_matcher" "test" {
						name = "acc-test-matcher"
					}`),
					Check: resource.ComposeTestCheckFunc(
						test.NoResourceAttributesSet("proxmox_notifications_matcher.test", []string{
							"comment",
						}),
					),
				},
			},
		},
		{
			"import notification matcher",
			[]resource.TestStep{
				{
					Config: te.RenderConfig(`
					resource "proxmox_notifications_matcher" "import_test" {
						name    = "acc-test-matcher-import"
						comment = "import test"
					}`),
				},
				{
					ResourceName:      "proxmox_notifications_matcher.import_test",
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
