# LSGo

## Description

LSGo lets you access files through your web browser. Like the Linux `ls` command, it is designed only to list files, with the additional ability to preview and download them.
It is not intended for editing or creating files.

## Installation


## Usage


## Support

You can open a GitHub issue.

## Roadmap

* [ ] UI
  * human-readable sizes
  * sort by name, size, or date
  * dark mode support
* [ ] Packaging
  * docker container
* [ ] download folder as zip
* [ ] prevent overly large files from being rendered
* [ ] signed Docker + SBOM attestations
* [ ] OIDC login
* [ ] Web security stuff

## Contributing

This is a test project and contributions are not currently planned. However, contributions are not forbidden — feel free to open an issue to discuss your ideas.

## Development

You need Go 1.27 installed.

For development, check the [Makefile](./Makefile) for instructions on running the formatter, linter, and tests.
The linter requires `golangci-lint`.

TL;DR

```bash
make check # fmt, lint, test
go run ./...
```

## License

This project is licensed under Mozilla Public License 2.0.

See [LICENSE](./LICENSE).
