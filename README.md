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

For example, you can get all the pull requests listed in the suer orgazation that have the author suer with the following command:

```bash
gh list-prs suer -a suer
# gh-list-prs
#19 suer 2024-12-29 example usage
```

see `gh list-prs --help` for more information.

## For developers

to build and install

```bash
go build .
gh extension install .
```

