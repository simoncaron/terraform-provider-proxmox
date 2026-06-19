/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */

package notifications

import (
	"github.com/bpg/terraform-provider-proxmox/proxmox/types"
)

// WebhookData contains the data from a webhook endpoint GET response.
// Header elements use format "name=<name>,value=<base64 of value>".
// Secret values are not returned by the API.
// Body is base64-encoded.
type WebhookData struct {
	Name    string            `json:"name"              url:"name"`
	URL     string            `json:"url"               url:"url"`
	Method  string            `json:"method"            url:"method"`
	Body    *string           `json:"body,omitempty"    url:"body,omitempty"`
	Comment *string           `json:"comment,omitempty" url:"comment,omitempty"`
	Disable *types.CustomBool `json:"disable,omitempty" url:"disable,omitempty,int"`
	// Header elements each have format "name=<name>,value=<base64 of value>".
	// Sent as repeated parameters (no comma tag) because values contain commas.
	Header []string `json:"header,omitempty" url:"header,omitempty"`
}

// WebhookResponseBody contains the body from a webhook endpoint GET response.
type WebhookResponseBody struct {
	Data *WebhookData `json:"data,omitempty"`
}

// WebhookRequestData contains the data for a webhook endpoint POST/PUT request.
type WebhookRequestData struct {
	WebhookData

	// Secret elements each have format "name=<name>,value=<base64 of value>".
	// Write-only: never returned by the API.
	Secret []string `url:"secret,omitempty"`
	Delete []string `url:"delete,omitempty,comma"`
}
