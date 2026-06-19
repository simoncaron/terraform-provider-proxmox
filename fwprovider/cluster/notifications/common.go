/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */

package notifications

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	proxmoxtypes "github.com/bpg/terraform-provider-proxmox/proxmox/types"
)

// boolOrDefault converts a *CustomBool to types.Bool, returning false when nil.
// PVE omits boolean fields from GET responses when they equal the server default.
func boolOrDefault(b *proxmoxtypes.CustomBool) types.Bool {
	if v := b.PointerBool(); v != nil {
		return types.BoolValue(*v)
	}

	return types.BoolValue(false)
}

// stringSliceToList converts a []string to a types.List of strings.
func stringSliceToList(s []string) types.List {
	if len(s) == 0 {
		return types.ListValueMust(types.StringType, []attr.Value{})
	}

	vals := make([]attr.Value, len(s))
	for i, v := range s {
		vals[i] = types.StringValue(v)
	}

	return types.ListValueMust(types.StringType, vals)
}

// listToStringSlice converts a types.List of strings to a []string.
func listToStringSlice(ctx context.Context, l types.List, diags *diag.Diagnostics) []string {
	if l.IsNull() || l.IsUnknown() {
		return nil
	}

	var out []string

	diags.Append(l.ElementsAs(ctx, &out, false)...)

	return out
}

// checkListDelete appends apiKey to toDelete when the plan list is empty/null
// and the state list was non-empty.
func checkListDelete(plan, state types.List, toDelete *[]string, apiKey string) {
	planEmpty := plan.IsNull() || len(plan.Elements()) == 0
	stateHadValues := !state.IsNull() && len(state.Elements()) > 0

	if planEmpty && stateHadValues {
		*toDelete = append(*toDelete, apiKey)
	}
}

// parseWebhookNameValue parses a "name=<name>,value=<base64>" string.
func parseWebhookNameValue(s string) (string, string, bool) {
	// Split into exactly two parts: "name=<n>" and "value=<b64>"
	parts := strings.SplitN(s, ",", 2)
	if len(parts) != 2 {
		return "", "", false
	}

	nameParts := strings.SplitN(parts[0], "=", 2)
	if len(nameParts) != 2 || nameParts[0] != "name" {
		return "", "", false
	}

	valueParts := strings.SplitN(parts[1], "=", 2)
	if len(valueParts) != 2 || valueParts[0] != "value" {
		return "", "", false
	}

	return nameParts[1], valueParts[1], true
}

// encodeWebhookNameValue encodes a name and raw value into "name=<name>,value=<base64>" format.
func encodeWebhookNameValue(name, rawValue string) string {
	return fmt.Sprintf("name=%s,value=%s", name, base64.StdEncoding.EncodeToString([]byte(rawValue)))
}

// parseMatchField parses a "(regex|exact):<field>=<value>" string.
func parseMatchField(s string) (string, string, string, bool) {
	matchType, rest, ok := strings.Cut(s, ":")
	if !ok {
		return "", "", "", false
	}

	field, value, ok := strings.Cut(rest, "=")
	if !ok {
		return "", "", "", false
	}

	return matchType, field, value, true
}
