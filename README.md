# gh-list-prs

A GitHub CLI Extension to list pull requests in specified repository.

## Installation

```bash
gh extension install suer/gh-list-prs
```

to upgrade

```bash
gh extension upgrade suer/gh-list-prs
```

## Usage

```bash
gh list-prs <org>
```

see `gh list-prs --help` for more information.

## For developers

to build and install

```bash
go build .
gh extension install .
```

