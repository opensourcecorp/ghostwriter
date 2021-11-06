# ghostwriter

<!-- badges: start -->
![Github Actions](https://github.com/opensourcecorp/ghostwriter/actions/workflows/main.yaml/badge.svg)

[![Support OpenSourceCorp on Ko-Fi!](https://img.shields.io/badge/Ko--fi-F16061?style=for-the-badge&logo=ko-fi&logoColor=white)](https://ko-fi.com/ryapric)
<!-- badges: end -->

Generate code, config, IaC, dependency files and more from template files -- all
using a single master config file & [Go text
templates](https://pkg.go.dev/text/template).

Similar in spirit to [HasiCorp Consul
Template](https://github.com/hashicorp/consul-template), but does not rely on
another service to manage the config values for you.

## Installation

`ghostwriter` is distributed as a single Go binary named `ghostwrite` (a verb,
without the trailing "r"), and can be easily installed a number of ways.

### Download Releases

The easiest way to get `ghostwriter` is to download the latest release from the
[Releases page](https://github.com/opensourcecorp/ghostwriter/releases), and
extract the `ghostwrite` binary to somewhere on your `PATH`, e.g.
`/usr/local/bin`.

### Install from Source

If you have a working Go installation, you can install with `go` itself:

    go install github.com/opensourcecorp/ghostwriter@<latest|vX>

## Usage

Assuming you're ok with the default behavior (documented below), and your config
file has data in it, you just need to run:

    ghostwrite

from the directory that you want to render templates from. That's it!

Files are written out with the same name the same as their input file, just in
your output directory instead. So for example, assuming the default behavior and
with a directory tree that looks like this:

    .
    |-- ghostwriter.yaml
    |-- some
    |-- files
    |-- to
    |-- render

then calling `ghostwrite` from the root will result in the following tree,
having rendered your templates in-flight:

    .
    |-- ghostwriter.yaml
    |-- some
    |-- files
    |-- to
    |-- render
    |-- gw-rendered/
        |-- some
        |-- files
        |-- to
        |-- render

### CLI Options

The `ghostwrite` CLI supports the following flags:

- `-config-file`: The config file containing your render data. Default:
  `ghostwriter.yaml`

- `-input-dir`: The input directory to render templates from. Default: `.`

- `-output-dir`: The output directory where the rendered tree will be
  reconstructed. Default: `gw-rendered`

### The .gwignore file

Using entries in a file at your directory root called `.gwignore`, you can use
regular expressions to exclude certain path patterns from `ghostwrite`'s
rendering. For example, to ignore all paths containing `secret`, you can put a
line containing `secret` into the `.gwignore`. If you want to be sure to ignore
***any*** path containing "secret" at all, you can instead put `.*secret.*`.
Note that these are ***regular expressions, not shell globs*** such as you might
find in a `.gitignore` (i.e. they're processed by [Go's regexp
package](https://pkg.go.dev/regexp)).

## Roadmap Notes

- Explore if only touching `*.gw` files as templates makes sense (that's how it
  was in the previous Python version)

- Allow outputs to be added to the repo's `.gitignore`.

- Allow users to pass single files as render sources/targets, vs. requiring
  directories.
