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

// GetWebhook retrieves a webhook endpoint by name.
func (c *Client) GetWebhook(ctx context.Context, name string) (*WebhookData, error) {
	resBody := &WebhookResponseBody{}

	err := c.DoRequest(ctx, http.MethodGet, c.ExpandPath("endpoints/webhook/"+name), nil, resBody)
	if err != nil {
		return nil, fmt.Errorf("error reading webhook endpoint %q: %w", name, err)
	}

	if resBody.Data == nil {
		return nil, api.ErrNoDataObjectInResponse
	}

	return resBody.Data, nil
}

// CreateWebhook creates a new webhook endpoint.
func (c *Client) CreateWebhook(ctx context.Context, data *WebhookRequestData) error {
	err := c.DoRequest(ctx, http.MethodPost, c.ExpandPath("endpoints/webhook/"+data.Name), data, nil)
	if err != nil {
		return fmt.Errorf("error creating webhook endpoint %q: %w", data.Name, err)
	}

	return nil
}

// UpdateWebhook updates an existing webhook endpoint.
func (c *Client) UpdateWebhook(ctx context.Context, data *WebhookRequestData) error {
	err := c.DoRequest(ctx, http.MethodPut, c.ExpandPath("endpoints/webhook/"+data.Name), data, nil)
	if err != nil {
		return fmt.Errorf("error updating webhook endpoint %q: %w", data.Name, err)
	}

	return nil
}

// DeleteWebhook deletes a webhook endpoint.
func (c *Client) DeleteWebhook(ctx context.Context, name string) error {
	err := c.DoRequest(ctx, http.MethodDelete, c.ExpandPath("endpoints/webhook/"+name), nil, nil)
	if err != nil {
		return fmt.Errorf("error deleting webhook endpoint %q: %w", name, err)
	}

	return nil
}
