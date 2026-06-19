/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */

package notifications

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/bpg/terraform-provider-proxmox/fwprovider/attribute"
	"github.com/bpg/terraform-provider-proxmox/proxmox/cluster/notifications"
)

type endpointSMTPModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Server      types.String `tfsdk:"server"`
	FromAddress types.String `tfsdk:"from_address"`
	Author      types.String `tfsdk:"author"`
	Comment     types.String `tfsdk:"comment"`
	Disable     types.Bool   `tfsdk:"disable"`
	Mailto      types.List   `tfsdk:"mailto"`
	MailtoUser  types.List   `tfsdk:"mailto_user"`
	Mode        types.String `tfsdk:"mode"`
	Port        types.Int64  `tfsdk:"port"`
	Username    types.String `tfsdk:"username"`
	Password    types.String `tfsdk:"password"`
}

// fromAPI populates the model from an SMTPData API response.
// Password is intentionally not set here — it is write-only and must be preserved from plan/state.
func (m *endpointSMTPModel) fromAPI(name string, data *notifications.SMTPData) {
	m.ID = types.StringValue(name)
	m.Name = types.StringValue(name)
	m.Server = types.StringValue(data.Server)
	m.FromAddress = types.StringValue(data.FromAddress)
	m.Author = types.StringPointerValue(data.Author)
	m.Comment = types.StringPointerValue(data.Comment)
	m.Disable = boolOrDefault(data.Disable)
	m.Mailto = stringSliceToList(data.Mailto)
	m.MailtoUser = stringSliceToList(data.MailtoUser)
	m.Mode = types.StringPointerValue(data.Mode)
	m.Port = types.Int64PointerValue(data.Port)
	m.Username = types.StringPointerValue(data.Username)
}

// toAPI converts the model to an SMTPRequestData for POST/PUT requests.
func (m *endpointSMTPModel) toAPI(ctx context.Context, diags *diag.Diagnostics) *notifications.SMTPRequestData {
	return &notifications.SMTPRequestData{
		SMTPData: notifications.SMTPData{
			Name:        m.Name.ValueString(),
			Server:      m.Server.ValueString(),
			FromAddress: m.FromAddress.ValueString(),
			Author:      attribute.StringPtrFromValue(m.Author),
			Comment:     attribute.StringPtrFromValue(m.Comment),
			Disable:     attribute.CustomBoolPtrFromValue(m.Disable),
			Mailto:      listToStringSlice(ctx, m.Mailto, diags),
			MailtoUser:  listToStringSlice(ctx, m.MailtoUser, diags),
			Mode:        attribute.StringPtrFromValue(m.Mode),
			Port:        attribute.Int64PtrFromValue(m.Port),
			Username:    attribute.StringPtrFromValue(m.Username),
		},
		Password: attribute.StringPtrFromValue(m.Password),
	}
}
