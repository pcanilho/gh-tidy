# GitHub `tidy` 🧹extension

[![Actions Status: Release](https://github.com/pcanilho/gh-tidy/workflows/release/badge.svg)](https://github.com/pcanilho/gh-tidy/actions?query=release)

The `gh-tidy` project is an extension for the standard `gh` cli that aims at offering tidy/cleanup operations on existing `refs`
(in either `branch` or `PR` format) by providing rules, such as `stale` status based on last commit date for a given branch, PR activity and others.

🚀 This project is entirely built upon GitHub's `graphql` API offered via the https://github.com/shurcooL/githubv4 project.

---

## Using `gh-tidy`
0. <ins>Expose</ins> a `GITHUB_TOKEN` environment variable with `repo:read` privileges or `repo:write` if you wish to use the `Delete` features. (*)
1. <ins>Install</ins> the `gh` cli available [here](https://github.com/cli/cli#installation).
2. <ins>Install</ins> the extension:
    ```shell
    $ gh extension install pcanilho/gh-tidy
    ```
   or <ins>upgrade</ins> to `latest` version:
    ```shell
    $ gh extension upgrade pcanilho/gh-tidy
    ```
3. 🚀 <ins>Use</ins> the extension:
   ```shell
   $ gh tidy --help
   ```

\* This can be a `PAT`, a GitHub App installation `access_token` or any other format compatible with the `oauth2.StaticTokenSource` OAuth2 client.

**Note**: Authentication through direct GitHub App PEM is not (yet) supported.
### Usage
```shell
Examples:
$ gh tidy stale branches <owner/repo> -t 72h
$ gh tidy stale prs      <owner/repo> -t 72h -s OPEN -s MERGED

Flags:
  -f, --force               If specified, all interactive operations will be disabled
  -h, --help                help for stale
  -t, --threshold duration  The stale threshold value. [1 month] (default 672h0m0s)
  -s, --state stringArray   The PR state. Supported values are: OPEN, MERGED or CLOSED (default [OPEN])
      --rm                  If specified, this flag enable the removal mode of the correlated sub-command

Global Flags:
      --format string       The desired output format. Supported values are: yaml, json (default "yaml")
  -o, --owner string        The GitHub owner value. (Automatically set if the repository is given in the 'owner/repository' format
```

### Examples

#### `List`

* <ins>List</ins> all branches with `stale` commits for the last `128 hours` in `yaml` format:
   ```shell
   $ gh tidy stale branches <owner/repository> -t 128h -f yaml
   ```

* <ins>Filter</ins> results using `jq`:
   ```shell
   $ gh tidy <command> -f json | jq <query>
   ```

* <ins>List</ins> all PRs with `stale` commits for the last `128 hours`, that are in `OPEN` state, in `yaml` format:
   ```shell
   $ gh tidy stale prs <owner/repository> -t 128h -f yaml -s OPEN
   ```

#### `Delete`

* <ins>Delete</ins> all branches with `stale` commits for the last `128 hours`:
   ```shell
   $ gh tidy stale branches <owner/repository> -t 128h --rm
   ```

* <ins>Delete</ins> all branches with `stale` commits for the last `128 hours` excluding branch names with a pattern (regex):
   ```shell
   $ gh tidy stale branches <owner/repository> -t 128h --exclude '<regex>' --rm
   ```

#### `Close`

* <ins>Close</ins> all PRs with `stale` commits for the last `128 hours`:
   ```shell
   $ gh tidy stale prs <owner/repository> -t 128h --rm
   ```