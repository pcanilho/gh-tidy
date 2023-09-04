# GitHub `tidy` 🧹extension

[![CI: Tests](https://github.com/pcanilho/gh-tidy/workflows/ci/badge.svg)](https://github.com/pcanilho/gh-tidy/actions?query=ci)
[![CD: Release](https://github.com/pcanilho/gh-tidy/workflows/release/badge.svg)](https://github.com/pcanilho/gh-tidy/actions?query=release)

<img src="https://github.com/pcanilho/gh-tidy/blob/main/docs/logo.png?raw=true" width="92">

The `gh-tidy` project is a tiny & simple extension for the standard `gh` cli that aims at offering tidy/cleanup operations on existing `refs`
(in either `branch`, `tag` or `PR` format) by providing rules, such as `stale` status based on HEAD commit date for a given branch, tag, PR activity and others.

🚀 Supports:
* **Enterprise** and **Public** GitHub API endpoints are supported.
* **Automatic authentication** using the environment variable `GITHUB_TOKEN`.
* **Automatic** GitHub API **limit handling** where requests are restarted after the `X-RateLimit-Reset` timer expires.
* **Automatic** API **batching** to avoid unnecessary collisions with the internal API (_defaults to `20`_).
* **Listing** & **Deletion** of branches with a stale HEAD commit based on time duration.
* **Listing** & **Deletion** of tags with a stale commit based on time duration.
* **Closing** of PRs with a stale branch HEAD commit based on time duration & PR state.

ℹ️ This is a utility project that I have been extending when needed on a best-effort basis. Feel free to contribute with a PR
or open an Issue on GitHub!

📝 TODOs (for lack of time...):
* API:
  * Support GitHub APP `pem` direct authentication.
* [stale] Branches:
  * Support optional detected if the provided branch has already been merged to the repository default branch.

---

## Using `gh-tidy` 
_...**locally** or through a CI system like **Jenkins**, **GitHub actions** or any other..._

0. <ins>Expose</ins> a `GITHUB_TOKEN` environment variable with `repo:read` privileges or `repo:admin` if you wish to use the `delete` features. (*)
1. <ins>Install</ins> the `gh` cli available [here](https://github.com/cli/cli#installation).
2. <ins>Install</ins> the extension:
    ```shell
    $ gh extension install pcanilho/gh-tidy
    ```
   or <ins>upgrade</ins> to the `latest` version:
    ```shell
    $ gh extension upgrade pcanilho/gh-tidy
    ```
3. 🚀 <ins>Use</ins> the extension:
   ```shell
   $ gh tidy --help
   ```

\* This can be a `PAT`, a GitHub App installation `access_token` or any other format that allows API access via `Bearer` token.

**Note**: Authentication through direct GitHub App PEM is not (yet) supported.
### Usage
```shell
Examples:
$ gh tidy stale branches <owner/repo> -t 72h
$ gh tidy stale prs      <owner/repo> -t 72h -s OPEN -s MERGED
$ gh tidy stale tags     <owner/repo> -t 72h
$ gh tidy delete         <owner/repo> -t 72h --ref <branch_name> --ref <tag_name>

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

* <ins>Delete</ins> all tags with a `stale` ref for the last `128 hours`:
   ```shell
   $ gh tidy stale tags <owner/repository> -t 128h --rm
   ```

#### `Close`

* <ins>Close</ins> all PRs with `stale` commits for the last `128 hours`:
   ```shell
   $ gh tidy stale prs <owner/repository> -t 128h --rm
   ```