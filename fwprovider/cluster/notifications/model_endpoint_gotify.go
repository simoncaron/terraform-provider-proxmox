/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */

package notifications

import (
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/bpg/terraform-provider-proxmox/fwprovider/attribute"
	"github.com/bpg/terraform-provider-proxmox/proxmox/cluster/notifications"
)

type endpointGotifyModel struct {
	ID      types.String `tfsdk:"id"`
	Name    types.String `tfsdk:"name"`
	Server  types.String `tfsdk:"server"`
	Token   types.String `tfsdk:"token"`
	Comment types.String `tfsdk:"comment"`
	Disable types.Bool   `tfsdk:"disable"`
}

// fromAPI populates the model from a GotifyData API response.
// Token is intentionally not set here — it is write-only and must be preserved from plan/state.
func (m *endpointGotifyModel) fromAPI(name string, data *notifications.GotifyData) {
	m.ID = types.StringValue(name)
	m.Name = types.StringValue(name)
	m.Server = types.StringValue(data.Server)
	m.Comment = types.StringPointerValue(data.Comment)
	m.Disable = boolOrDefault(data.Disable)
}

// toAPI converts the model to a GotifyRequestData for POST/PUT requests.
func (m *endpointGotifyModel) toAPI() *notifications.GotifyRequestData {
	return &notifications.GotifyRequestData{
		GotifyData: notifications.GotifyData{
			Name:    m.Name.ValueString(),
			Server:  m.Server.ValueString(),
			Comment: attribute.StringPtrFromValue(m.Comment),
			Disable: attribute.CustomBoolPtrFromValue(m.Disable),
		},
		Token: attribute.StringPtrFromValue(m.Token),
	}
}
