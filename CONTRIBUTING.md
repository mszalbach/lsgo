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

1. **Check main branch status**
   Ensure all tests are passing on `main`.

2. **Determine the next version**
   Follow [Semantic Versioning](https://semver.org/) (`vMAJOR.MINOR.PATCH`):
   - `PATCH` for backwards-compatible bug fixes (`fix:`)
   - `MINOR` for backwards-compatible features (`feat:`)
   - `MAJOR` for breaking changes (`feat!:` or `BREAKING CHANGE:`)

3. **Create and push a Git tag**
   ```bash
   # Create tag locally
   git tag -a v1.2.0 -m "Release v1.2.0"

   # Push tag to GitHub
   git push origin v1.2.0
    ```