<!-- TOC -->
* [The `dkrkit` package](#the-dkrkit-package)
  * [Two APIs: Docker and DockerT](#two-apis-docker-and-dockert)
    * [Docker](#docker)
    * [DockerT](#dockert)
  * [Building images](#building-images)
    * [Build](#build)
    * [BuildTestImg](#buildtestimg)
    * [ImgPull](#imgpull)
    * [Build options](#build-options)
  * [Listing and removing images](#listing-and-removing-images)
    * [ImgLs](#imgls)
    * [ImgRm](#imgrm)
    * [ImgRm options](#imgrm-options)
  * [Inspecting labels and environment variables](#inspecting-labels-and-environment-variables)
    * [Labels / Label](#labels--label)
    * [Envs / Env](#envs--env)
  * [Running containers](#running-containers)
    * [CtrRun](#ctrrun)
    * [CtrRm / CtrKill / CtrExec](#ctrrm--ctrkill--ctrexec)
    * [CtrPs / CtrFile](#ctrps--ctrfile)
    * [CtrRun options](#ctrrun-options)
  * [Networks](#networks)
  * [Assertion helpers](#assertion-helpers)
    * [HasLabel / HasNoLabel / HasLabels](#haslabel--hasnolabel--haslabels)
    * [HasEnv / HasNoEnv / HasEnvs](#hasenv--hasnoenv--hasenvs)
    * [HasBuildArg](#hasbuildarg)
  * [Reference utilities](#reference-utilities)
    * [Ref](#ref)
    * [ShortID](#shortid)
    * [StripHashName](#striphashname)
    * [ToBuildArgs](#tobuildargs)
    * [RandName / RandTag / RandRef / RandNet](#randname--randtag--randref--randnet)
<!-- TOC -->

# The `dkrkit` package

`dkrkit` wraps the Docker CLI in Go helper functions designed for use
in integration tests. It covers the full lifecycle of test images and
containers: building, running, inspecting, and removing them.

The package exposes two layered APIs and a set of standalone assertion
helpers. The pure utility functions (`Ref`, `ShortID`, etc.) are fully
deterministic and carry no Docker dependency.

## Two APIs: Docker and DockerT

### Docker

`Docker` executes docker commands and returns errors to the caller.
Use it in shared test helpers, fixtures, or `TestMain` functions where
a `*testing.T` is not yet available:

```go
dkr := dkrkit.New()

ref, iid, err := dkr.Build(
    dkrkit.WithBuildPth("testdata/Dockerfile"),
    dkrkit.WithBuildArg("BASE", "busybox:1.38-uclibc"),
)
if err != nil {
    // handle …
}
defer dkr.ImgRm(iid)
```

To run docker commands with a custom environment instead of
`os.Environ`, pass `WithEnv`:

```go
dkr := dkrkit.New(dkrkit.WithEnv(
    append(os.Environ(), "DOCKER_HOST=tcp://remote:2376"),
))
```

### DockerT

`DockerT` wraps `Docker` and calls `t.Error` on failure, so methods
return results directly without an error value. It also registers
automatic cleanup for built images via `t.Cleanup`. Use it directly
inside test functions:

```go
dkr := dkrkit.NewT(t)

ref, iid := dkr.Build(
    dkrkit.WithBuildPth("testdata/Dockerfile"),
    dkrkit.WithBuildArg("BASE", "busybox:1.38-uclibc"),
)
// image is removed by t.Cleanup automatically

cid := dkr.CtrRun(ref)
defer dkr.CtrRm(cid)
```

`DockerT.Build` also injects the test name as the `authors` build
argument, so built images are traceable back to their test.

## Building images

### Build

`Docker.Build` builds an image and returns the image reference and
image ID. The reference is `name:tag`; both are randomly generated
unless overridden with `WithBuildName` / `WithBuildTag`. The image ID
has the `sha256:` prefix stripped.

```go
dkr := dkrkit.New()
ref, iid, err := dkr.Build(
    dkrkit.WithBuildPth("testdata/Dockerfile"),
    dkrkit.WithBuildArg("BASE", dkrkit.TestImageBaseRef),
    dkrkit.WithBuildLabel("com.example.suite", "integration"),
)
```

To supply a Dockerfile as a reader rather than a file path (useful for
embedded Dockerfiles):

```go
//go:embed testdata/Dockerfile
var myBld []byte

ref, iid, err := dkr.Build(
    dkrkit.WithBuildRdr(bytes.NewReader(myBld)),
)
```

### BuildTestImg

`BuildTestImg` builds the package's own embedded test image. It reads
`C42_IMG_CREATED` and `C42_IMG_REF_NAME` from the environment and
sets the remaining OCI labels to fixed test values:

```go
ref, iid, err := dkr.BuildTestImg()
```

### ImgPull

`ImgPull` pulls an image from a registry. It is a no-op when the
image is already present locally:

```go
err := dkr.ImgPull("busybox:1.38-uclibc")
```

### Build options

| Option             | Description                                             |
|--------------------|---------------------------------------------------------|
| `WithBuildName`    | Set the image name (default: random `ctx42-tst-img-*`). |
| `WithBuildTag`     | Set the image tag (default: random `ctx42-tst-tag-*`).  |
| `WithBuildLabel`   | Add a single label to the image.                        |
| `WithBuildLabels`  | Add multiple labels from a `map[string]string`.         |
| `WithBuildArg`     | Add a single `--build-arg`.                             |
| `WithBuildArgs`    | Add multiple build args from a `map[string]string`.     |
| `WithBuildPth`     | Path to the Dockerfile (mutually exclusive with Rdr).   |
| `WithBuildRdr`     | `io.Reader` supplying the Dockerfile contents.          |
| `WithBuildNoCache` | Pass `--no-cache` to docker build.                      |
| `WithBuildDryRun`  | Print the docker command instead of running it.         |

Setting `C42_BLD_NO_CACHE=1` in the environment has the same
effect as `WithBuildNoCache()` for all builds.

## Listing and removing images

### ImgLs

`ImgLs` returns all local images as an `Images` slice. Each `Image`
carries its ID, repository, and tag. Use `WithImgLsFilter` to narrow
results:

```go
ims, err := dkr.ImgLs(
    dkrkit.WithImgLsFilter("reference=ctx42-tst-img-*"),
)
img := ims.FindByRef("ctx42-tst-img-abc:ctx42-tst-tag-xyz")
```

`Images.FindByRef` and `Images.FindByID` return `nil` when nothing
matches; `FindByID` strips the `sha256:` prefix before comparing.

### ImgRm

`ImgRm` removes an image by reference or ID. It handles the common
case of a stopped container holding the image: it retries up to 20
times (configurable) with a 250 ms sleep between attempts, forcing
removal when docker allows it. Passing an empty string or a
non-existent reference is a no-op:

```go
err := dkr.ImgRm(iid)
```

### ImgRm options

| Option                  | Description                                            |
|-------------------------|--------------------------------------------------------|
| `WithImgRmSleep`        | Sleep between retries (default: 250 ms).               |
| `WithImgRmTries`        | Maximum retry attempts (default: 20).                  |
| `WithImgRmIgnoreErrors` | `DockerT` only: log errors via `t.Log`, not `t.Error`. |

## Inspecting labels and environment variables

### Labels / Label

`Labels` returns all labels for an image or container as a
`map[string]string`. `Label` returns a single label value and errors
if the label does not exist. Both accept an image ID, image reference,
or container ID:

```go
lbs, err := dkr.Labels(iid)
val, err := dkr.Label(ref, "com.example.version")
```

### Envs / Env

`Envs` returns all environment variables for an image or container as
a `map[string]string`. `Env` returns a single variable and errors if
the name is not set:

```go
envs, err := dkr.Envs(ref)
val, err := dkr.Env(ref, "APP_ENV")
```

## Running containers

### CtrRun

`CtrRun` runs a container and returns its stdout. By default the
container is removed when it exits (`--rm`). Pass the image reference
or ID as the first argument, followed by any extra docker-run
arguments via `WithCtrRunArgs`:

```go
out, err := dkr.CtrRun(ref, dkrkit.WithCtrRunArgs("echo", "hello"))
```

To retrieve the container ID while it is still running, use
`WithCtrRunCID`. It returns a channel that receives the ID once the
cidfile is created (within 1 s), then closes itself:

```go
cidCh, cidOpt := dkrkit.WithCtrRunCID()
out, err := dkr.CtrRun(ref,
    cidOpt,
    dkrkit.WithCtrRunNoRemove(),
    dkrkit.WithCtrRunDetach(),
)
cid, ok := <-cidCh  // ok is false on timeout
defer dkr.CtrRm(cid)
```

### CtrRm / CtrKill / CtrExec

`CtrRm` force-removes a container by ID. `CtrKill` sends SIGKILL to
a running container. `CtrExec` runs a command inside a running
container and returns its stdout:

```go
err := dkr.CtrRm(cid)
err  = dkr.CtrKill(cid)
out, err := dkr.CtrExec(cid, "cat", "/etc/hostname")
```

### CtrPs / CtrFile

`CtrPs` lists all containers (equivalent to `docker ps --all`). The
returned `Containers` value is lazy: full details (Env, Labels,
Networks, State) are populated only when `FindByImage` or `FindByID`
is called:

```go
ctrs, err := dkr.CtrPs()
ctr, err  := ctrs.FindByImage("myapp:latest")
ctr, err   = ctrs.FindByID(cid)
```

`CtrFile` extracts a single file from a running container and returns
its content as a string:

```go
content, err := dkr.CtrFile(cid, "/etc/hostname")
```

### CtrRun options

| Option               | Description                                        |
|----------------------|----------------------------------------------------|
| `WithCtrRunArgs`     | Append extra arguments after the image reference.  |
| `WithCtrRunLabel`    | Set a single label on the container.               |
| `WithCtrRunLabels`   | Set multiple labels from a `map[string]string`.    |
| `WithCtrRunCIDPth`   | Path for the cidfile (default: auto temp file).    |
| `WithCtrRunCID`      | Return a channel that receives the container ID.   |
| `WithCtrRunDetach`   | Run the container in the background (`--detach`).  |
| `WithCtrRunNoRemove` | Do not remove the container on exit (omit `--rm`). |

## Networks

`NetLs` returns all Docker networks as a `Networks` slice. Each
`Network` carries its ID, name, driver, attachable flag, and labels.
`NetRm` removes a network by ID or name and is a no-op when the
network does not exist:

```go
nets, err := dkr.NetLs()
net := nets.FindByName("my-test-net")
net  = nets.FindByID(netID)

err = dkr.NetRm("my-test-net")
```

## Assertion helpers

The package-level helpers call docker inspect internally, so they work
with any image ID, image reference, or container ID. All helpers
accept a `tester.T`, fail the test via `t.Error` on mismatch, and
return `false` so the calling test can continue:

### HasLabel / HasNoLabel / HasLabels

```go
dkrkit.HasLabel(t, iid, "com.example.version", "v1.2.3")
dkrkit.HasNoLabel(t, iid, "com.example.internal")
dkrkit.HasLabels(t, ref, map[string]string{
    "com.example.version": "v1.2.3",
    "com.example.env":     "test",
})
```

### HasEnv / HasNoEnv / HasEnvs

```go
dkrkit.HasEnv(t, ref, "APP_ENV", "test")
dkrkit.HasNoEnv(t, ref, "SECRET_KEY")
dkrkit.HasEnvs(t, ref, map[string]string{
    "APP_ENV":  "test",
    "APP_PORT": "8080",
})
```

### HasBuildArg

`HasBuildArg` checks that a `map[string]*string` (as returned by
`ToBuildArgs`) has a key set to the expected value. It is designed for
asserting `docker build --build-arg` arguments in tests that construct
them programmatically:

```go
args := dkrkit.ToBuildArgs(map[string]string{"VERSION": "v1.2.3"})
dkrkit.HasBuildArg(t, args, "v1.2.3", "VERSION")
```

## Reference utilities

### Ref

`Ref` assembles a Docker image reference from a registry/repo prefix,
an image name, and a tag. Any trailing slash on the repo is trimmed.
Omitting the repo produces `name:tag`; omitting the tag produces just
`name`:

<!-- gmdoceg:ExampleRef -->
```go
// Ref assembles an image reference from repo, name, and tag.
ref := dkrkit.Ref("example.com/repo", "myapp", "v1.2.3")
fmt.Println(ref)
// Output:
// example.com/repo/myapp:v1.2.3
```

<!-- gmdoceg:ExampleRef_noRepo -->
```go
// Without a repo, Ref returns name:tag.
ref := dkrkit.Ref("", "myapp", "v1.2.3")
fmt.Println(ref)
// Output:
// myapp:v1.2.3
```

### ShortID

`ShortID` truncates a hex ID to its first 12 characters, matching the
short-ID format shown by docker commands. Non-hex strings (image
references, names) are returned unchanged:

<!-- gmdoceg:ExampleShortID -->
```go
// ShortID truncates a long hex ID to its first 12 characters.
short := dkrkit.ShortID("785e9f61d4598b65c6c86c5f122830f7")
fmt.Println(short)
// Output:
// 785e9f61d459
```

<!-- gmdoceg:ExampleShortID_nonHex -->
```go
// Non-hex strings (e.g. image references) are returned unchanged.
ref := dkrkit.ShortID("myapp:v1.2.3")
fmt.Println(ref)
// Output:
// myapp:v1.2.3
```

### StripHashName

`StripHashName` removes the algorithm prefix from a content digest,
returning only the hex part. A bare hex string is returned unchanged:

<!-- gmdoceg:ExampleStripHashName -->
```go
// StripHashName removes the algorithm prefix from a digest.
id := dkrkit.StripHashName("sha256:b3aab1576e98b7f41f01fa")
fmt.Println(id)
// Output:
// b3aab1576e98b7f41f01fa
```

### ToBuildArgs

`ToBuildArgs` converts a `map[string]string` into the
`map[string]*string` form expected by Docker build-argument helpers.
Each value is copied so the original map is unaffected:

<!-- gmdoceg:ExampleToBuildArgs -->
```go
// ToBuildArgs converts a string map to the pointer map that Docker
// build arguments require.
args := dkrkit.ToBuildArgs(map[string]string{"VERSION": "v1.2.3"})
fmt.Println(*args["VERSION"])
// Output:
// v1.2.3
```

### RandName / RandTag / RandRef / RandNet

The `Rand*` helpers generate random identifiers for test images and
networks. Each appends a 12-character lowercase hex suffix to a
fixed prefix, making them recognisable and safe to clean up by prefix:

<!-- gmdoceg:ExampleRandName -->
```go
// RandName returns a unique name suitable for a test image.
name := dkrkit.RandName()
fmt.Println(strings.HasPrefix(name, "ctx42-tst-img-"))
// Output:
// true
```

| Helper     | Prefix                            | Suitable for   |
|------------|-----------------------------------|----------------|
| `RandName` | `ctx42-tst-img-`                  | Image name     |
| `RandTag`  | `ctx42-tst-tag-`                  | Image tag      |
| `RandRef`  | `ctx42-tst-img-*:ctx42-tst-tag-*` | Full reference |
| `RandNet`  | `ctx42-tst-net-`                  | Network name   |
