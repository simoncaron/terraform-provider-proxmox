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

// GetSMTP retrieves an SMTP endpoint by name.
func (c *Client) GetSMTP(ctx context.Context, name string) (*SMTPData, error) {
	resBody := &SMTPResponseBody{}

	err := c.DoRequest(ctx, http.MethodGet, c.ExpandPath("endpoints/smtp/"+name), nil, resBody)
	if err != nil {
		return nil, fmt.Errorf("error reading SMTP endpoint %q: %w", name, err)
	}

	if resBody.Data == nil {
		return nil, api.ErrNoDataObjectInResponse
	}

	return resBody.Data, nil
}

// CreateSMTP creates a new SMTP endpoint.
func (c *Client) CreateSMTP(ctx context.Context, data *SMTPRequestData) error {
	err := c.DoRequest(ctx, http.MethodPost, c.ExpandPath("endpoints/smtp/"+data.Name), data, nil)
	if err != nil {
		return fmt.Errorf("error creating SMTP endpoint %q: %w", data.Name, err)
	}

	return nil
}

// UpdateSMTP updates an existing SMTP endpoint.
func (c *Client) UpdateSMTP(ctx context.Context, data *SMTPRequestData) error {
	err := c.DoRequest(ctx, http.MethodPut, c.ExpandPath("endpoints/smtp/"+data.Name), data, nil)
	if err != nil {
		return fmt.Errorf("error updating SMTP endpoint %q: %w", data.Name, err)
	}

	return nil
}

// DeleteSMTP deletes an SMTP endpoint.
func (c *Client) DeleteSMTP(ctx context.Context, name string) error {
	err := c.DoRequest(ctx, http.MethodDelete, c.ExpandPath("endpoints/smtp/"+name), nil, nil)
	if err != nil {
		return fmt.Errorf("error deleting SMTP endpoint %q: %w", name, err)
	}

	return nil
}
