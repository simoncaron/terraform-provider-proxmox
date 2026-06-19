/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */

package notifications

import (
	"github.com/bpg/terraform-provider-proxmox/proxmox/types"
)

// SendmailData contains the data from a sendmail endpoint GET response.
type SendmailData struct {
	Name        string            `json:"name"                   url:"name"`
	Author      *string           `json:"author,omitempty"       url:"author,omitempty"`
	Comment     *string           `json:"comment,omitempty"      url:"comment,omitempty"`
	Disable     *types.CustomBool `json:"disable,omitempty"      url:"disable,omitempty,int"`
	FromAddress *string           `json:"from-address,omitempty" url:"from-address,omitempty"`
	Mailto      []string          `json:"mailto,omitempty"       url:"mailto,omitempty,comma"`
	MailtoUser  []string          `json:"mailto-user,omitempty"  url:"mailto-user,omitempty,comma"`
}

// SendmailResponseBody contains the body from a sendmail endpoint GET response.
type SendmailResponseBody struct {
	Data *SendmailData `json:"data,omitempty"`
}

// SendmailRequestData contains the data for a sendmail endpoint POST/PUT request.
type SendmailRequestData struct {
	SendmailData

	Delete []string `url:"delete,omitempty,comma"`
}
