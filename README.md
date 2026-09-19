# ph-province

`ph-province` draws the shape of Cebu in a terminal using Unicode half-block
characters. It is network-free at runtime and supports direct and interactive
use.

## Usage

```console
$ ph-province cebu
             ▄
             █
            ▄█
            ██
           ▄██
           ███
             ⋮
```

Start the prompt when you want to make repeated queries:

```console
$ ph-province
province> cebu
```

Province names are case-insensitive. Version 1 supports Cebu. Enter `quit`,
`exit`, or Ctrl-D to leave interactive mode.

## Install

Download the archive for your Linux or macOS architecture from the GitHub
release, extract it, and place the executable somewhere on `PATH`:

```bash
mkdir -p "$HOME/.local/bin"
tar -xzf ph-province_linux_amd64.tar.gz
mv ph-province "$HOME/.local/bin/"
ph-province cebu
```

If `$HOME/.local/bin` is not already on `PATH`, add it to your shell profile.
No Go installation or data files are needed to run a release binary.

## Develop

Go 1.22 or newer is required.

```bash
go test ./...
go build -o ph-province ./cmd/ph-province
./ph-province cebu
```

Release binaries are built with `CGO_ENABLED=0` for Linux and macOS on amd64
and arm64.

## Boundary data

The embedded Cebu geometry was extracted from the simplified geoBoundaries
Philippines ADM2 dataset pinned at commit `41af8f1`. See
[THIRD_PARTY_NOTICES.md](THIRD_PARTY_NOTICES.md) for source and license details.

The file `cebu_shape.png` is a user-provided, watermarked visual reference. It
is intentionally ignored by Git and is not read, embedded, or redistributed by
this application.
