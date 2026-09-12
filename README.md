# LSGo

[![build](https://github.com/mszalbach/lsgo/actions/workflows/ci.yaml/badge.svg)](https://github.com/mszalbach/lsgo/actions/workflows/ci.yaml)

## Description

LSGo lets you access files through a web browser. Like the Linux `ls` command, it is designed only to list files, with the added ability to preview and download them.
It is not intended for editing or creating files.

## Installation

Currently no packaging, will be delivered later in the Roadmap.


## Usage

:warning: LSGo makes your folder accesible via your Browser. Do not expose it directly to the internet.

The default settings try to aim to be as secure as possible. 
But keep in mind I am not perfect, I could produced bugs in the code or fail to document some special behaviour.

Run this unprivileged inside a container.

LSGo is configured via commandline flags. Run with `--help` to see the possible settings.

The most important are:


| flag     | description                                                                     |
| -------- | ------------------------------------------------------------------------------- |
| --addr   | defines where the server will listen too. In Default only listens on localhost. |
| --folder | denifes which folder is exposed via the web ui. Defaults to ./public            |


## Support

You can open a GitHub issue.

## Roadmap

* [ ] sort by name, size, or date
* [ ] 404 handling
* [ ] Web security improvements
* [ ] prevent overly large files from being rendered
* [ ] Packaging
  * Docker container
* [ ] Signed Docker image and SBOM attestations
* [ ] download folder/files as zip
* [ ] OIDC login
* [ ] dark mode support

## Contributing

This is a test project, and contributions are not currently planned. However, they are not prohibited—feel free to open an issue to discuss your ideas.

## Development

You need Go 1.27 installed.

For development, check the [Makefile](./Makefile) for instructions on running the formatter, linter, and tests.
The linter requires `golangci-lint`.

You can start the server via normal go command:

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
