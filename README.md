# LSGo

[![build](https://github.com/mszalbach/lsgo/actions/workflows/ci.yaml/badge.svg)](https://github.com/mszalbach/lsgo/actions/workflows/ci.yaml)

## Description

LSGo lets you browse files through a web browser. Like the Linux `ls` command, it is designed only to list files, with the added ability to preview and download them.
It is not intended for editing or creating files.

## Installation

Packaging is not available yet; it will be added later in the roadmap.

## Usage

:warning: LSGo makes your folder accessible via your browser. Do not expose it directly to the internet.

The default settings aim to be as secure as possible. However, keep in mind that I am not perfect; I may introduce bugs into the code or fail to document some edge cases.

Run this as an unprivileged user inside a container. If you want to expose it outside your PC, run it behind a proxy with at least TLS, preferably with authentication such as OIDC or mTLS.

LSGo is configured via command-line flags. Run with `--help` to see the available settings.

The most important ones are:

| flag     | description                                                                      |
| -------- | -------------------------------------------------------------------------------- |
| --addr   | Defines where the server will listen. By default, it only listens on localhost.   |
| --folder | Defines which folder is exposed via the web UI. Defaults to `./public`.          |


## Support

You can open a GitHub issue.

## Roadmap

* [ ] sort by name, size, or date
* [x] 404 handling
* [ ] Web security improvements
* [ ] prevent overly large files from being rendered
* [ ] Packaging
  * Docker container
* [ ] Signed Docker image and SBOM attestations
* [ ] download folder/files as zip
* [ ] OIDC login
* [ ] dark mode support

## Contributing

This is a small test project, and contributions are not currently planned. However, they are not prohibited—feel free to open an issue to discuss your ideas.

## Development

You need Go 1.27 installed.

For development, check the [Makefile](./Makefile) for instructions on running the formatter, linter, and tests.
The linter requires `golangci-lint`.

You can start the server via the normal Go command:

```bash
go run ./...
```

or you can use [air](https://github.com/air-verse/air) to automatically reload on changes:

```bash
air
```

TL;DR

```bash
make check # fmt, lint, test
air
```

## License

This project is licensed under Mozilla Public License 2.0.

See [LICENSE](./LICENSE).
