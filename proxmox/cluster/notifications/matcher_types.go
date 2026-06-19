/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */

package notifications

import (
	"github.com/bpg/terraform-provider-proxmox/proxmox/types"
)

// MatcherData contains the data from a matcher GET response.
// MatchField elements use format "(regex|exact):<field>=<value>".
type MatcherData struct {
	Name          string            `json:"name"                     url:"name"`
	Comment       *string           `json:"comment,omitempty"        url:"comment,omitempty"`
	Disable       *types.CustomBool `json:"disable,omitempty"        url:"disable,omitempty,int"`
	InvertMatch   *types.CustomBool `json:"invert-match,omitempty"   url:"invert-match,omitempty,int"`
	MatchCalendar []string          `json:"match-calendar,omitempty" url:"match-calendar,omitempty,comma"`
	MatchField    []string          `json:"match-field,omitempty"    url:"match-field,omitempty,comma"`
	MatchSeverity []string          `json:"match-severity,omitempty" url:"match-severity,omitempty,comma"`
	Mode          *string           `json:"mode,omitempty"           url:"mode,omitempty"`
	Target        []string          `json:"target,omitempty"         url:"target,omitempty,comma"`
}

// MatcherResponseBody contains the body from a matcher GET response.
type MatcherResponseBody struct {
	Data *MatcherData `json:"data,omitempty"`
}

// MatcherRequestData contains the data for a matcher POST/PUT request.
type MatcherRequestData struct {
	MatcherData

	Delete []string `url:"delete,omitempty,comma"`
}
