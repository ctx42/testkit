// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package dkrkit

// Image holds the metadata of a Docker image.
type Image struct {
	ID         string `json:"ID"`
	Repository string `json:"Repository"`
	Tag        string `json:"Tag"`
}

// Images is a collection of [Image] values.
type Images []*Image

// FindByRef returns the image in the collection whose ref
// (image-name:image-tag) matches. Returns nil if not found.
func (ims Images) FindByRef(ref string) *Image {
	for _, img := range ims {
		if ref == img.Repository+":"+img.Tag {
			return img
		}
	}
	return nil
}

// FindByID returns the image in the collection with the given ID.
// Returns nil if not found.
func (ims Images) FindByID(iid string) *Image {
	for _, img := range ims {
		if StripHashName(iid) == StripHashName(img.ID) {
			return img
		}
	}
	return nil
}
