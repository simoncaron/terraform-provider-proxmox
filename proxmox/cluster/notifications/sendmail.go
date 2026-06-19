/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */

package notifications //nolint:dupl

import (
	"context"
	"fmt"
	"net/http"

	"github.com/bpg/terraform-provider-proxmox/proxmox/api"
)

// GetSendmail retrieves a sendmail endpoint by name.
func (c *Client) GetSendmail(ctx context.Context, name string) (*SendmailData, error) {
	resBody := &SendmailResponseBody{}

	err := c.DoRequest(ctx, http.MethodGet, c.ExpandPath("endpoints/sendmail/"+name), nil, resBody)
	if err != nil {
		return nil, fmt.Errorf("error reading sendmail endpoint %q: %w", name, err)
	}

	if resBody.Data == nil {
		return nil, api.ErrNoDataObjectInResponse
	}

	return resBody.Data, nil
}

// CreateSendmail creates a new sendmail endpoint.
func (c *Client) CreateSendmail(ctx context.Context, data *SendmailRequestData) error {
	err := c.DoRequest(ctx, http.MethodPost, c.ExpandPath("endpoints/sendmail/"+data.Name), data, nil)
	if err != nil {
		return fmt.Errorf("error creating sendmail endpoint %q: %w", data.Name, err)
	}

	return nil
}

// UpdateSendmail updates an existing sendmail endpoint.
func (c *Client) UpdateSendmail(ctx context.Context, data *SendmailRequestData) error {
	err := c.DoRequest(ctx, http.MethodPut, c.ExpandPath("endpoints/sendmail/"+data.Name), data, nil)
	if err != nil {
		return fmt.Errorf("error updating sendmail endpoint %q: %w", data.Name, err)
	}

	return nil
}

// DeleteSendmail deletes a sendmail endpoint.
func (c *Client) DeleteSendmail(ctx context.Context, name string) error {
	err := c.DoRequest(ctx, http.MethodDelete, c.ExpandPath("endpoints/sendmail/"+name), nil, nil)
	if err != nil {
		return fmt.Errorf("error deleting sendmail endpoint %q: %w", name, err)
	}

	return nil
}
