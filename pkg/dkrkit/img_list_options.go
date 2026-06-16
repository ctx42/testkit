// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package dkrkit

// ImgListOption represents an option for [Docker.ImgLs] and [DockerT.ImgLs].
type ImgListOption func(*ImgListOptions)

// ImgListOptions represents Docker image listing options.
type ImgListOptions struct {
	filters []string // List of image filters.
}

// WithImgLsFilter is an option for defining image filters.
func WithImgLsFilter(filters ...string) ImgListOption {
	return func(opt *ImgListOptions) {
		opt.filters = append(opt.filters, filters...)
	}
}
