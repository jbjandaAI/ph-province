# ph-province

`ph-province` draws any of the 82 Philippine provinces in a terminal using
Unicode half-block characters. It is network-free at runtime and supports
direct and interactive use.

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

Multiword province names do not need quotes:

```console
$ ph-province agusan del norte
```

Start the prompt when you want to make repeated queries:

```console
$ ph-province
province> cebu
```

Province names are case-insensitive, and punctuation, hyphens, and repeated
spaces are normalized. Common former names such as `Compostela Valley`,
`North Cotabato`, and `Western Samar` are also recognized. Mistyped names
receive suggestions instead of silently selecting a province.

List all supported names with:

```console
$ ph-province --list
```

Within interactive mode, enter `list` to show the same catalog. Enter `quit`,
`exit`, or Ctrl-D to leave.

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

The 82 embedded province geometries were extracted from a pinned, simplified
PSA/NAMRIA-derived dataset. See [THIRD_PARTY_NOTICES.md](THIRD_PARTY_NOTICES.md)
for the exact release, checksums, provenance, and attribution.

The file `cebu_shape.png` is a user-provided, watermarked visual reference. It
is intentionally ignored by Git and is not read, embedded, or redistributed by
this application.
