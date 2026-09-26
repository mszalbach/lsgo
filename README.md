# LSGo

[![build](https://github.com/mszalbach/lsgo/actions/workflows/ci.yaml/badge.svg)](https://github.com/mszalbach/lsgo/actions/workflows/ci.yaml)

## Description

LSGo lets you browse files in a web browser. Like the Linux `ls` command, it is designed to list files only, with the added ability to preview and download them.
It is not intended for editing or creating files.

## Installation

The easiest way to run LSGo is with the provided Docker image. Replace `VERSION` with the release you want to use:

```bash
docker pull ghcr.io/mszalbach/lsgo:VERSION
```
You can also download a binary from the [GitHub Releases](https://github.com/mszalbach/lsgo/releases) page and run it directly on your machine.

Before running LSGo, read the Usage section. Make sure the server is not reachable by untrusted users and that the folders you serve do not contain sensitive files.

### Verify the Docker image

Release images are signed with Cosign using GitHub Actions keyless signing:

```bash
cosign verify \
  --certificate-identity-regexp "https://github.com/mszalbach/lsgo/.*" \
  --certificate-oidc-issuer "https://token.actions.githubusercontent.com" \
  "ghcr.io/mszalbach/lsgo:VERSION"
```

Successful verification confirms the image signature was created by this repository's release workflow and has not been altered since signing.

### Verify the SBOM attestation

Release images also include a signed CycloneDX SBOM attestation. Verify it with:

```bash
cosign verify-attestation \
  --type cyclonedx \
  --certificate-identity-regexp "https://github.com/mszalbach/lsgo/.*" \
  --certificate-oidc-issuer "https://token.actions.githubusercontent.com" \
  "ghcr.io/mszalbach/lsgo:VERSION"
```

To inspect the verified SBOM, decode the attestation payload:

```bash
cosign verify-attestation \
  --type cyclonedx \
  --certificate-identity-regexp "https://github.com/mszalbach/lsgo/.*" \
  --certificate-oidc-issuer "https://token.actions.githubusercontent.com" \
  "ghcr.io/mszalbach/lsgo:VERSION" |
  jq -r '.payload' | base64 --decode | jq "." | less
```

Successful verification confirms the attestation was signed by this repository's release workflow and is associated with the image digest.

## Usage

> [!WARNING]
> LSGo makes the selected folder available in your web browser. Do not expose it directly to the internet.

The default settings listen on `localhost:8080`, so the server is only available from the local machine. If you need to make it available to other users, run it behind a reverse proxy with TLS and authentication, such as OIDC or mTLS.

Run LSGo as an unprivileged user, preferably inside a container. Only expose folders that are safe for the intended users to browse.

When running the Docker image, it defaults to the user/group `7777:0`. This means the mounted folder must be readable by that user, or you must override the runtime user with a different UID/GID. For example, if the host directory is owned by a different user, run the container with `--user` or adjust file permissions so the process can read the content before it starts serving files.

LSGo is configured with command-line flags. Run `lsgo --help` to see all available options.

The most important ones are:

| Flag                     | Description                                                                                                                               |
| ------------------------ | ----------------------------------------------------------------------------------------------------------------------------------------- |
| `--addr`                 | Address where the server listens. Defaults to `localhost:8080`.                                                                           |
| `--base-url`             | Base URL path for serving behind a reverse proxy. It is normalized to start and end with a slash, for example `/files/`. Defaults to `/`. |
| `--folder`               | Folder exposed through the web UI. Defaults to `./public`.                                                                                |
| `--max-inline-file-size` | Maximum file size to display inline in bytes. Larger files are available as downloads only. Defaults to `1_048_576`.                      |

### Run with Docker

The following command serves the host's `/tmp` directory at `http://localhost:8080`:

```bash
docker run --rm \
  --publish 8080:8080 \
  --volume /tmp:/app/public:ro \
  ghcr.io/mszalbach/lsgo:VERSION \
  --addr :8080
```

Open [http://localhost:8080](http://localhost:8080) in your browser.

### Run a downloaded binary

To serve a local directory with a downloaded binary, run:

```bash
./lsgo --folder /path/to/folder
```

Then open [http://localhost:8080](http://localhost:8080) in your browser.

## Support

You can open a GitHub issue.

## Roadmap

* [x] listing folders and files on the web
* [x] sort by name, size, or date
* [x] 404 handling
* [x] forcing insecure media types to always be downloaded
* [x] prevent overly large files from being rendered
* [x] Packaging
  * Docker container
* [x] configurable base path for proxy usage
* [x] dark mode support
* [x] download folder/files as zip
* [x] Signed Docker image + SBOM attestation
* [ ] OIDC login

## Contributing

Contributions are welcome. Please open an issue to discuss a larger change before submitting a pull request.

See [CONTRIBUTING.md](./CONTRIBUTING.md) for commit and pull request conventions, as well as the release process.

## Documentation

See the [architecture documentation](./docs/README.md), including the project's architecture decisions and ADRs.

## Development

You need Go 1.27 installed.

For development, check the [Makefile](./Makefile) for instructions on running the formatter, linter, and tests.
The linter requires `golangci-lint`.

You can start the server using the standard Go command:

```bash
go run ./...
```

or you can use [air](https://github.com/air-verse/air) to automatically reload on changes:

```bash
air
```

TL;DR:

```bash
make check # fmt, lint, test
air
```

## License

This project is licensed under the Mozilla Public License 2.0.

See [LICENSE](./LICENSE).
