/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */

package notifications

import (
	"github.com/bpg/terraform-provider-proxmox/proxmox/types"
)

// GotifyData contains the data from a gotify endpoint GET response.
// Note: token is never returned by the API.
type GotifyData struct {
	Name    string            `json:"name"              url:"name"`
	Server  string            `json:"server"            url:"server"`
	Comment *string           `json:"comment,omitempty" url:"comment,omitempty"`
	Disable *types.CustomBool `json:"disable,omitempty" url:"disable,omitempty,int"`
}

// GotifyResponseBody contains the body from a gotify endpoint GET response.
type GotifyResponseBody struct {
	Data *GotifyData `json:"data,omitempty"`
}

// GotifyRequestData contains the data for a gotify endpoint POST/PUT request.
type GotifyRequestData struct {
	GotifyData

	// Token is write-only and never returned by the API.
	Token  *string  `url:"token,omitempty"`
	Delete []string `url:"delete,omitempty,comma"`
}
