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

type endpointSendmailModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Author      types.String `tfsdk:"author"`
	Comment     types.String `tfsdk:"comment"`
	Disable     types.Bool   `tfsdk:"disable"`
	FromAddress types.String `tfsdk:"from_address"`
	Mailto      types.List   `tfsdk:"mailto"`
	MailtoUser  types.List   `tfsdk:"mailto_user"`
}

// fromAPI populates the model from a SendmailData API response.
func (m *endpointSendmailModel) fromAPI(name string, data *notifications.SendmailData) {
	m.ID = types.StringValue(name)
	m.Name = types.StringValue(name)
	m.Author = types.StringPointerValue(data.Author)
	m.Comment = types.StringPointerValue(data.Comment)
	m.Disable = boolOrDefault(data.Disable)
	m.FromAddress = types.StringPointerValue(data.FromAddress)
	m.Mailto = stringSliceToList(data.Mailto)
	m.MailtoUser = stringSliceToList(data.MailtoUser)
}

// toAPI converts the model to a SendmailRequestData for POST/PUT requests.
func (m *endpointSendmailModel) toAPI(ctx context.Context, diags *diag.Diagnostics) *notifications.SendmailRequestData {
	return &notifications.SendmailRequestData{
		SendmailData: notifications.SendmailData{
			Name:        m.Name.ValueString(),
			Author:      attribute.StringPtrFromValue(m.Author),
			Comment:     attribute.StringPtrFromValue(m.Comment),
			Disable:     attribute.CustomBoolPtrFromValue(m.Disable),
			FromAddress: attribute.StringPtrFromValue(m.FromAddress),
			Mailto:      listToStringSlice(ctx, m.Mailto, diags),
			MailtoUser:  listToStringSlice(ctx, m.MailtoUser, diags),
		},
	}
}
