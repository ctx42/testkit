// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package dkrkit

import (
	"context"
	"encoding/json"
	"time"

	"github.com/ctx42/testing/pkg/notice"

	"github.com/ctx42/testkit/pkg/oskit"
)

// State holds the runtime state of a Docker container.
type State struct {
	Status     string `json:"Status"`
	Running    bool   `json:"Running"`
	Paused     bool   `json:"Paused"`
	Restarting bool   `json:"Restarting"`
	Dead       bool   `json:"Dead"`
	ExitCode   int    `json:"ExitCode"`
	Error      string `json:"Error"`
}

// Container holds the metadata and runtime state of a Docker container.
type Container struct {
	ID       string            `json:"ID"`
	Image    string            `json:"Image"`
	Created  time.Time         `json:"Created"`
	State    State             `json:"State"`
	Env      map[string]string `json:"-"`
	Labels   map[string]string `json:"-"`
	Networks map[string]string `json:"-"`
}

func (ctr *Container) UnmarshalJSON(data []byte) error {
	type T1 Container
	t1 := struct {
		*T1
		Config struct {
			Image  string            `json:"Image"`
			Env    []string          `json:"Env"`
			Labels map[string]string `json:"Labels"`
		} `json:"Config"`
		NetworkSettings struct {
			Networks map[string]struct {
				NetworkID string `json:"NetworkID"`
			} `json:"Networks"`
		} `json:"NetworkSettings"`
	}{
		T1: (*T1)(ctr),
	}

	if err := json.Unmarshal(data, &t1); err != nil {
		return err
	}
	t1.Image = t1.Config.Image
	t1.Env = oskit.EnvSplit(t1.Config.Env)
	t1.Labels = t1.Config.Labels

	nets := make(map[string]string)
	for name, network := range t1.NetworkSettings.Networks {
		nets[name] = network.NetworkID
	}
	t1.Networks = nets
	return nil
}

// Containers is a collection of [Container] values. It carries the Docker
// environment used to query container details on demand.
type Containers struct {
	env  []string
	ctrs []*Container
}

// FindByImage returns the first container whose image matches ref
// (image-name:image-tag), inspecting it to populate full details and verify
// it still exists. Returns nil, nil if no container in the collection matches.
// Returns an error if a match is found, but the container no longer exists.
func (cts Containers) FindByImage(ref string) (*Container, error) {
	for _, ctr := range cts.ctrs {
		if ctr.Image == ref {
			if err := inspect(cts.env, ctr); err != nil {
				return nil, err
			}
			return ctr, nil
		}
	}
	return nil, nil
}

// FindByID returns the container with the given ID, inspecting it to populate
// full details and verify it still exists. Returns nil, nil if no container
// in the collection has that ID. Returns an error if a match is found but
// the container no longer exists.
func (cts Containers) FindByID(cid string) (*Container, error) {
	for _, ctr := range cts.ctrs {
		if ctr.ID == cid {
			if err := inspect(cts.env, ctr); err != nil {
				return nil, err
			}
			return ctr, nil
		}
	}
	return nil, nil
}

// inspect fetches full container details from the Docker daemon and writes
// them into ctr. Returns an error if the container has been removed from the
// daemon since it appeared in [Docker.CtrPs] output, or if parsing fails.
func inspect(env []string, ctr *Container) error {
	ctx := context.Background()
	args := []string{"inspect", "--format", "json", ctr.ID}
	sout, _, err := dockerCmd(ctx, env, args)
	if err != nil {
		return notice.From(err, "inspecting container")
	}
	wrap := []*Container{ctr}
	if err = json.Unmarshal([]byte(sout), &wrap); err != nil {
		return notice.From(err, "inspecting container")
	}
	return nil
}
