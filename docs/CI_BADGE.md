# CI Status Badge

Add the following badge to your README.md to display the CI build status:

## Badge Markdown

```markdown
[![CI](https://github.com/USERNAME/REPOSITORY/actions/workflows/ci.yml/badge.svg)](https://github.com/USERNAME/REPOSITORY/actions/workflows/ci.yml)
```

Replace `USERNAME` and `REPOSITORY` with your actual GitHub username and repository name.

## Badge with Branch

To show status for a specific branch:

```markdown
[![CI](https://github.com/USERNAME/REPOSITORY/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/USERNAME/REPOSITORY/actions/workflows/ci.yml)
```

## Multiple Badges

You can also add other useful badges:

```markdown
[![CI](https://github.com/USERNAME/REPOSITORY/actions/workflows/ci.yml/badge.svg)](https://github.com/USERNAME/REPOSITORY/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/USERNAME/REPOSITORY)](https://goreportcard.com/report/github.com/USERNAME/REPOSITORY)
[![Go Version](https://img.shields.io/github/go-mod/go-version/USERNAME/REPOSITORY)](https://github.com/USERNAME/REPOSITORY)
[![License](https://img.shields.io/github/license/USERNAME/REPOSITORY)](https://github.com/USERNAME/REPOSITORY/blob/main/LICENSE)
```

## Example README Section

```markdown
# Project Name

[![CI](https://github.com/USERNAME/REPOSITORY/actions/workflows/ci.yml/badge.svg)](https://github.com/USERNAME/REPOSITORY/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/USERNAME/REPOSITORY)](https://goreportcard.com/report/github.com/USERNAME/REPOSITORY)

Brief description of your project.

## Quality Gates

This project maintains the following quality standards:

- ✅ All tests pass with race detection enabled
- ✅ Code coverage ≥ 80%
- ✅ Passes golangci-lint with 30+ linters
- ✅ Docker image builds successfully

See [CI/CD documentation](docs/CI.md) for details.
```
