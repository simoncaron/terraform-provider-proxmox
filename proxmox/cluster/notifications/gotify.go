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

// GetGotify retrieves a gotify endpoint by name.
func (c *Client) GetGotify(ctx context.Context, name string) (*GotifyData, error) {
	resBody := &GotifyResponseBody{}

	err := c.DoRequest(ctx, http.MethodGet, c.ExpandPath("endpoints/gotify/"+name), nil, resBody)
	if err != nil {
		return nil, fmt.Errorf("error reading gotify endpoint %q: %w", name, err)
	}

	if resBody.Data == nil {
		return nil, api.ErrNoDataObjectInResponse
	}

	return resBody.Data, nil
}

// CreateGotify creates a new gotify endpoint.
func (c *Client) CreateGotify(ctx context.Context, data *GotifyRequestData) error {
	err := c.DoRequest(ctx, http.MethodPost, c.ExpandPath("endpoints/gotify/"+data.Name), data, nil)
	if err != nil {
		return fmt.Errorf("error creating gotify endpoint %q: %w", data.Name, err)
	}

	return nil
}

// UpdateGotify updates an existing gotify endpoint.
func (c *Client) UpdateGotify(ctx context.Context, data *GotifyRequestData) error {
	err := c.DoRequest(ctx, http.MethodPut, c.ExpandPath("endpoints/gotify/"+data.Name), data, nil)
	if err != nil {
		return fmt.Errorf("error updating gotify endpoint %q: %w", data.Name, err)
	}

	return nil
}

// DeleteGotify deletes a gotify endpoint.
func (c *Client) DeleteGotify(ctx context.Context, name string) error {
	err := c.DoRequest(ctx, http.MethodDelete, c.ExpandPath("endpoints/gotify/"+name), nil, nil)
	if err != nil {
		return fmt.Errorf("error deleting gotify endpoint %q: %w", name, err)
	}

	return nil
}
