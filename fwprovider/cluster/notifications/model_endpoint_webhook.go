/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */

package notifications

import (
	"encoding/base64"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/bpg/terraform-provider-proxmox/fwprovider/attribute"
	"github.com/bpg/terraform-provider-proxmox/proxmox/cluster/notifications"
)

type webhookHeaderModel struct {
	Name  types.String `tfsdk:"name"`
	Value types.String `tfsdk:"value"`
}

type webhookSecretModel struct {
	Name  types.String `tfsdk:"name"`
	Value types.String `tfsdk:"value"`
}

type endpointWebhookModel struct {
	ID      types.String         `tfsdk:"id"`
	Name    types.String         `tfsdk:"name"`
	URL     types.String         `tfsdk:"url"`
	Method  types.String         `tfsdk:"method"`
	Body    types.String         `tfsdk:"body"`
	Comment types.String         `tfsdk:"comment"`
	Disable types.Bool           `tfsdk:"disable"`
	Header  []webhookHeaderModel `tfsdk:"header"`
	Secret  []webhookSecretModel `tfsdk:"secret"`
}

// fromAPI populates the model from a WebhookData API response.
// Secret values are intentionally not set — they are write-only and must be preserved from plan/state.
// Header values are decoded from base64 format returned by the API.
// Body is decoded from base64.
func (m *endpointWebhookModel) fromAPI(name string, data *notifications.WebhookData) {
	m.ID = types.StringValue(name)
	m.Name = types.StringValue(name)
	m.URL = types.StringValue(data.URL)
	m.Method = types.StringValue(data.Method)
	m.Comment = types.StringPointerValue(data.Comment)
	m.Disable = boolOrDefault(data.Disable)

	if data.Body != nil {
		if decoded, err := base64.StdEncoding.DecodeString(*data.Body); err == nil {
			m.Body = types.StringValue(string(decoded))
		} else {
			m.Body = types.StringPointerValue(data.Body)
		}
	} else {
		m.Body = types.StringNull()
	}

	m.Header = decodeWebhookHeaders(data.Header)
	// Secret is intentionally not set from the API response.
}

// toAPI converts the model to a WebhookRequestData for POST/PUT requests.
func (m *endpointWebhookModel) toAPI() *notifications.WebhookRequestData {
	req := &notifications.WebhookRequestData{
		WebhookData: notifications.WebhookData{
			Name:    m.Name.ValueString(),
			URL:     m.URL.ValueString(),
			Method:  m.Method.ValueString(),
			Comment: attribute.StringPtrFromValue(m.Comment),
			Disable: attribute.CustomBoolPtrFromValue(m.Disable),
			Header:  encodeWebhookHeaders(m.Header),
		},
		Secret: encodeWebhookSecrets(m.Secret),
	}

	if !m.Body.IsNull() && !m.Body.IsUnknown() {
		encoded := base64.StdEncoding.EncodeToString([]byte(m.Body.ValueString()))
		req.Body = &encoded
	}

	return req
}

// decodeWebhookHeaders decodes API-format headers ("name=<n>,value=<b64>") into model structs.
func decodeWebhookHeaders(raw []string) []webhookHeaderModel {
	result := make([]webhookHeaderModel, 0, len(raw))

	for _, s := range raw {
		name, b64val, ok := parseWebhookNameValue(s)
		if !ok {
			continue
		}

		decoded, err := base64.StdEncoding.DecodeString(b64val)
		if err != nil {
			continue
		}

		result = append(result, webhookHeaderModel{
			Name:  types.StringValue(name),
			Value: types.StringValue(string(decoded)),
		})
	}

	return result
}

// encodeWebhookHeaders encodes model structs into API-format headers.
func encodeWebhookHeaders(headers []webhookHeaderModel) []string {
	if len(headers) == 0 {
		return nil
	}

	out := make([]string, len(headers))
	for i, h := range headers {
		out[i] = encodeWebhookNameValue(h.Name.ValueString(), h.Value.ValueString())
	}

	return out
}

// encodeWebhookSecrets encodes model structs into API-format secrets.
func encodeWebhookSecrets(secrets []webhookSecretModel) []string {
	if len(secrets) == 0 {
		return nil
	}

	out := make([]string, len(secrets))
	for i, s := range secrets {
		out[i] = encodeWebhookNameValue(s.Name.ValueString(), s.Value.ValueString())
	}

	return out
}
