/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */

package nodes

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/bpg/terraform-provider-proxmox/proxmox/api"
)

// applySizeRequiresReplace centralises the logic that compares the stored original
// size with the current remote size and sets replacement/diagnostics accordingly.
func applySizeRequiresReplace(
	resp *planmodifier.Int64Response,
	originalStateSizeBytes []byte,
	stateSize int64,
	planOverwrite bool,
	resourceKind string,
) {
	if originalStateSizeBytes == nil {
		return
	}

	originalStateSize, err := strconv.ParseInt(string(originalStateSizeBytes), 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Unable to convert original state %s size to int64", resourceKind),
			"Unexpected error in parsing string to int64, key original_state_size. "+
				"Please retry the operation or report this issue to the provider developers.\n\n"+
				"Error: "+err.Error(),
		)

		return
	}

	if stateSize != originalStateSize && planOverwrite {
		resp.RequiresReplace = true
		resp.PlanValue = types.Int64Value(originalStateSize)

		resp.Diagnostics.AddWarning(
			fmt.Sprintf("The %s size in datastore has changed outside of terraform.", resourceKind),
			fmt.Sprintf(
				"Previous size: %d saved in state does not match current size from datastore: %d. "+
					"You can disable this behaviour by using overwrite=false",
				originalStateSize,
				stateSize,
			),
		)

		return
	}
}

// handleReadResult centralises the common read-time error handling used by
// resources after attempting to read the remote resource.
// Returns true if the error was handled and the caller should return.
func handleReadResult(ctx context.Context, resp *resource.ReadResponse, err error, notExistMessage string) bool {
	if err != nil {
		if strings.Contains(err.Error(), "failed to authenticate") {
			resp.Diagnostics.AddError("Failed to authenticate", err.Error())

			return true
		}

		resp.Diagnostics.AddWarning(notExistMessage, err.Error())
		resp.State.RemoveResource(ctx)

		return true
	}

	return false
}

// handleDatastoreDeleteError centralises the error handling for Delete operations
// on datastore files/resources.
func handleDatastoreDeleteError(resp *resource.DeleteResponse, err error, id string, itemKind string) {
	if err == nil || errors.Is(err, api.ErrResourceDoesNotExist) {
		return
	}

	if strings.Contains(err.Error(), "unable to parse") {
		resp.Diagnostics.AddWarning(
			"Datastore "+itemKind+" does not exist",
			fmt.Sprintf(
				"Could not delete datastore %s '%s', it does not exist or has been deleted outside of Terraform.",
				itemKind, id,
			),
		)
	} else {
		resp.Diagnostics.AddError(
			"Error deleting datastore "+itemKind,
			fmt.Sprintf("Could not delete datastore %s '%s', unexpected error: %s", itemKind, id, err.Error()),
		)
	}
}
