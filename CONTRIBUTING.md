## Commit Messages & Pull Requests

This repository follows the [Conventional Commits](https://www.conventionalcommits.org/) specification for commit messages and PR titles.

### PR Title Format
Every Pull Request title must follow this format: `<type>(<scope>): <short summary>`

Common types:
- `feat`: A new feature
- `fix`: A bug fix
- `docs`: Documentation changes
- `refactor`: Code changes that neither fix a bug nor add a feature
- `chore`: Maintenance tasks, dependency updates, CI changes
- `test`: Fixing or improving the tests
- `ci`: Changes to the GitHub Actions or release automation

### Examples
- `feat(cli): add --output-json flag`
- `fix(parser): resolve panic on empty input`
- `docs: update installation instructions in README`

> **Note:** PR titles are automatically linted by CI. If you use `Squash and Merge`, your PR title will become the commit message on `main` and will appear in the release notes.


# Release Process

Releases are using [GoReleaser](https://goreleaser.com/) and GitHub Actions.

## Prerequisites
- Write access to the repository.
- Local `main` branch synced with upstream.

## Cutting a New Release

1. Manually run the Semantic Release workflow
   In GitHub, open **Actions**, select the **Release** workflow, and click **Run workflow** against `main`. If you are unsure about the proposed release, leave **preview without publishing** enabled (the default) and inspect the run output first.
2. Publish the release
   Once you have reviewed the preview, run the workflow again against `main` with **preview without publishing** disabled. The workflow is manually triggered; do not create or push a release tag yourself.
