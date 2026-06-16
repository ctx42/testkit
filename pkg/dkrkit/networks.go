// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package dkrkit

// Network represents a Docker network.
type Network struct {
	ID         string            `json:"ID"`
	Driver     string            `json:"Driver"`
	Name       string            `json:"Name"`
	Attachable bool              `json:"Attachable"`
	Labels     map[string]string `json:"Labels"`
}

// Networks is a collection of [Network] values.
type Networks []*Network

// FindByName finds the network with a given name in the collection. Returns
// nil when the network has not been found.
func (nts Networks) FindByName(name string) *Network {
	for _, net := range nts {
		if name == net.Name {
			return net
		}
	}
	return nil
}

// FindByID finds the network with the given ID in the collection. Returns nil
// when the network has not been found.
func (nts Networks) FindByID(id string) *Network {
	for _, net := range nts {
		if id == net.ID {
			return net
		}
	}
	return nil
}
