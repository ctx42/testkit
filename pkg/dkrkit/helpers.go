// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package dkrkit

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/ctx42/testing/pkg/notice"

	"github.com/ctx42/testkit/pkg/oskit"
)

// imgUsedRX matches the docker error message emitted when a container
// references an image being removed (running or stopped).
var imgUsedRX = regexp.MustCompile("container ([a-fA-F0-9]+) is using its " +
	"referenced image ([a-fA-F0-9]+)")

// imgUsedRunningRx matches the docker error message emitted when a running
// container is using the image being removed.
var imgUsedRunningRx = regexp.MustCompile("image is being used by running " +
	"container ([a-fA-F0-9]+)")

// imgUsedStoppedRx matches the docker error message emitted when a stopped
// container is using the image being removed.
var imgUsedStoppedRx = regexp.MustCompile("image is being used by stopped " +
	"container ([a-fA-F0-9]+)")

// imgRmCase classifies why a docker image rm command failed.
type imgRmCase int

const (
	imgRmWait    imgRmCase = iota // running container — sleep and retry
	imgRmForce                    // stopped/referenced container — force retry
	imgRmUnknown                  // unrecognised error
)

// imgRmState holds the result of classifying a failed image removal.
type imgRmState struct {
	action imgRmCase
	cid    string
	iid    string
	force  bool
}

// classifyImgRmErr classifies the stderr of a failed docker image rm command
// and returns the next action [Docker.ImgRm] should take.
func classifyImgRmErr(eout string) imgRmState {
	mustForce := strings.Contains(eout, "(must force)") ||
		strings.Contains(eout, "(must be forced)")

	if m := imgUsedRunningRx.FindStringSubmatch(eout); len(m) == 2 {
		return imgRmState{action: imgRmWait, cid: ShortID(m[1])}
	}
	if m := imgUsedStoppedRx.FindStringSubmatch(eout); len(m) == 2 {
		return imgRmState{
			action: imgRmForce,
			cid:    ShortID(m[1]),
			force:  mustForce,
		}
	}
	if m := imgUsedRX.FindStringSubmatch(eout); len(m) == 3 {
		return imgRmState{
			action: imgRmForce,
			cid:    ShortID(m[1]),
			iid:    ShortID(m[2]),
			force:  mustForce,
		}
	}
	return imgRmState{action: imgRmUnknown}
}

// Ref returns a Docker image reference composed of repo, name, and tag.
func Ref(repo, name, tag string) string {
	ref := strings.TrimRight(repo, "/")
	if name == "" {
		return ref
	}
	if ref != "" {
		ref += "/"

	}
	ref += name
	if tag == "" {
		return ref
	}
	return ref + ":" + tag
}

// StripHashName strips the algorithm prefix from a digest, returning only the
// hex part.
// Example:
//
//	sha256:b3aab1576e98b7f41f01fa -> b3aab1576e98b7f41f01fa
func StripHashName(id string) string {
	if idx := strings.Index(id, ":"); idx != -1 {
		if idx == len(id)-1 {
			return ""
		}
		return id[idx+1:]
	}
	return id
}

// ShortID returns the first 12 characters of a hexadecimal ID. Non-hex
// strings are returned unchanged.
func ShortID(id string) string {
	if !isHex(id) {
		return id
	}
	if len(id) > 12 {
		return id[:12]
	}
	return id
}

// isHex reports whether s is a non-empty hexadecimal string.
func isHex(s string) bool {
	if s == "" {
		return false
	}
	dst := make([]byte, hex.DecodedLen(len(s)))
	if _, err := hex.Decode(dst, []byte(s)); err != nil {
		return false
	}
	return true
}

// randString returns a random 12-character lowercase hex string.
func randString() string {
	buf := make([]byte, 6)
	if _, err := rand.Read(buf); err != nil {
		panic(err)
	}
	src := hex.EncodeToString(buf)
	return src[:12]
}

// RandName returns a random Docker image name prefixed with "ctx42-tst-img-"
// and a 12-character hex suffix.
func RandName() string { return "ctx42-tst-img-" + randString() }

// RandTag returns a random Docker image tag prefixed with "ctx42-tst-tag-"
// and a 12-character hex suffix.
func RandTag() string { return "ctx42-tst-tag-" + randString() }

// RandRef returns a random Docker image reference (name:tag) built from
// [RandName] and [RandTag].
func RandRef() string { return RandName() + ":" + RandTag() }

// RandNet returns a random Docker network name prefixed with "ctx42-tst-net-"
// and a 12-character hex suffix.
func RandNet() string { return "ctx42-tst-net-" + randString() }

// ToBuildArgs converts a string-to-string map to a string-to-pointer map
// suitable for use as Docker build arguments.
func ToBuildArgs(ba map[string]string) map[string]*string {
	args := make(map[string]*string, len(ba))
	for k, v := range ba {
		args[k] = &v
	}
	return args
}

// getLabels returns the labels of the image or container identified by ref.
func getLabels(
	ctx context.Context,
	env []string,
	ref string,
) (map[string]string, error) {

	args := []string{"inspect", "--format", "{{ json .Config.Labels }}", ref}
	sout, _, err := dockerCmd(ctx, env, args)
	if err != nil {
		return nil, notice.From(err, "getting labels")
	}
	var dst map[string]string
	if err = json.Unmarshal([]byte(sout), &dst); err != nil {
		err = notice.New("[getting labels] error unmarshaling").
			Append("ref", "%q", ref).
			Append("err", "%s", err).
			Wrap(err)
		return nil, err
	}
	return dst, nil
}

// getEnvs returns the environment variables of the image or container
// identified by ref.
func getEnvs(
	ctx context.Context,
	env []string,
	ref string,
) (map[string]string, error) {

	args := []string{"inspect", "--format", "{{ json .Config.Env }}", ref}
	sout, _, err := dockerCmd(ctx, env, args)
	if err != nil {
		return nil, notice.From(err, "getting environment variables")
	}
	var dst []string
	if err = json.Unmarshal([]byte(sout), &dst); err != nil {
		mHeader := "[getting environment variables] error unmarshaling"
		err = notice.New(mHeader).
			Append("ref", "%q", ref).
			Append("err", "%s", err).
			Wrap(err)
		return nil, err
	}
	return oskit.EnvSplit(dst), nil
}

// formatMapKeys returns sorted, quoted map keys joined by newlines, for use
// in error messages.
func formatMapKeys(m map[string]string) string {
	names := make([]string, 0, len(m))
	for name := range m {
		names = append(names, strconv.Quote(name))
	}
	sort.Strings(names)
	return strings.Join(names, "\n")
}

// envGet returns the value of env variable key, or an empty string if not set.
func envGet(env []string, key string) string {
	return oskit.EnvSplit(env)[key]
}
