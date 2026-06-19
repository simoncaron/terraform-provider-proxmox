/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */

package notifications

import (
	"github.com/bpg/terraform-provider-proxmox/proxmox/types"
)

// SMTPData contains the data from an SMTP endpoint GET response.
// Note: password is never returned by the API.
type SMTPData struct {
	Name        string            `json:"name"                  url:"name"`
	Server      string            `json:"server"                url:"server"`
	FromAddress string            `json:"from-address"          url:"from-address"`
	Author      *string           `json:"author,omitempty"      url:"author,omitempty"`
	Comment     *string           `json:"comment,omitempty"     url:"comment,omitempty"`
	Disable     *types.CustomBool `json:"disable,omitempty"     url:"disable,omitempty,int"`
	Mailto      []string          `json:"mailto,omitempty"      url:"mailto,omitempty,comma"`
	MailtoUser  []string          `json:"mailto-user,omitempty" url:"mailto-user,omitempty,comma"`
	Mode        *string           `json:"mode,omitempty"        url:"mode,omitempty"`
	Port        *int64            `json:"port,omitempty"        url:"port,omitempty"`
	Username    *string           `json:"username,omitempty"    url:"username,omitempty"`
}

// SMTPResponseBody contains the body from an SMTP endpoint GET response.
type SMTPResponseBody struct {
	Data *SMTPData `json:"data,omitempty"`
}

// SMTPRequestData contains the data for an SMTP endpoint POST/PUT request.
type SMTPRequestData struct {
	SMTPData

	// Password is write-only and never returned by the API.
	Password *string  `url:"password,omitempty"`
	Delete   []string `url:"delete,omitempty,comma"`
}
