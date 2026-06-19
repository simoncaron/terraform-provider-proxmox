/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */

package notifications

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/bpg/terraform-provider-proxmox/fwprovider/attribute"
	"github.com/bpg/terraform-provider-proxmox/proxmox/cluster/notifications"
)

type matcherMatchFieldModel struct {
	Type  types.String `tfsdk:"type"`
	Field types.String `tfsdk:"field"`
	Value types.String `tfsdk:"value"`
}

type matcherModel struct {
	ID            types.String             `tfsdk:"id"`
	Name          types.String             `tfsdk:"name"`
	Comment       types.String             `tfsdk:"comment"`
	Disable       types.Bool               `tfsdk:"disable"`
	InvertMatch   types.Bool               `tfsdk:"invert_match"`
	MatchCalendar types.List               `tfsdk:"match_calendar"`
	MatchField    []matcherMatchFieldModel `tfsdk:"match_field"`
	MatchSeverity types.List               `tfsdk:"match_severity"`
	Mode          types.String             `tfsdk:"mode"`
	Target        types.List               `tfsdk:"target"`
}

// fromAPI populates the model from a MatcherData API response.
func (m *matcherModel) fromAPI(name string, data *notifications.MatcherData) {
	m.ID = types.StringValue(name)
	m.Name = types.StringValue(name)
	m.Comment = types.StringPointerValue(data.Comment)
	m.Disable = boolOrDefault(data.Disable)
	m.InvertMatch = types.BoolPointerValue(data.InvertMatch.PointerBool())
	m.MatchCalendar = stringSliceToList(data.MatchCalendar)
	m.MatchField = decodeMatchFields(data.MatchField)
	m.MatchSeverity = stringSliceToList(data.MatchSeverity)
	m.Mode = types.StringPointerValue(data.Mode)
	m.Target = stringSliceToList(data.Target)
}

// toAPI converts the model to a MatcherRequestData for POST/PUT requests.
func (m *matcherModel) toAPI(ctx context.Context, diags *diag.Diagnostics) *notifications.MatcherRequestData {
	return &notifications.MatcherRequestData{
		MatcherData: notifications.MatcherData{
			Name:          m.Name.ValueString(),
			Comment:       attribute.StringPtrFromValue(m.Comment),
			Disable:       attribute.CustomBoolPtrFromValue(m.Disable),
			InvertMatch:   attribute.CustomBoolPtrFromValue(m.InvertMatch),
			MatchCalendar: listToStringSlice(ctx, m.MatchCalendar, diags),
			MatchField:    encodeMatchFields(m.MatchField),
			MatchSeverity: listToStringSlice(ctx, m.MatchSeverity, diags),
			Mode:          attribute.StringPtrFromValue(m.Mode),
			Target:        listToStringSlice(ctx, m.Target, diags),
		},
	}
}

// decodeMatchFields parses "(regex|exact):<field>=<value>" strings into model structs.
func decodeMatchFields(raw []string) []matcherMatchFieldModel {
	result := make([]matcherMatchFieldModel, 0, len(raw))

	for _, s := range raw {
		matchType, field, value, ok := parseMatchField(s)
		if !ok {
			continue
		}

		result = append(result, matcherMatchFieldModel{
			Type:  types.StringValue(matchType),
			Field: types.StringValue(field),
			Value: types.StringValue(value),
		})
	}

	return result
}

// encodeMatchFields encodes model structs into "(regex|exact):<field>=<value>" strings.
func encodeMatchFields(fields []matcherMatchFieldModel) []string {
	if len(fields) == 0 {
		return nil
	}

	out := make([]string, len(fields))
	for i, f := range fields {
		out[i] = fmt.Sprintf("%s:%s=%s", f.Type.ValueString(), f.Field.ValueString(), f.Value.ValueString())
	}

	return out
}
