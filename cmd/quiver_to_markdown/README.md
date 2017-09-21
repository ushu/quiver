# quiver_to_markdown

Converts a [Quiver] library to a set of markdown files.

#### Rationale

While I enjoy using [Quiver] a lot to take all sort of programming notes, I faced two issues with the original tool:

- Dropbox/iCloud save don't offer the fine-grained versioning git provides
- it is not so easy to "share" notes with collaborators

This tool allows me to convert my notes to a set of Markdown files and save them on Github.
This way I can have a fine-grained history of my changes, and a format that easy to share
and is visible on the web.

### Installing

```sh
$ go get -u github.com/ushu/quiver/cmd/quiver_to_markdown
```

#### `go` dependency

You need a working install of `go` to use this package.
Provided you are on a Mac (since [Quiver] is a Mac app 😁), the simplest way to install `go`
is probably to use [Homebrew]:

```sh
$ brew install go
```

**Note**: you might want to add the `$GOPATH/bin` to your shell `PATH`,
(in `.profile` or equivalent) to have it available as a shell command, something like

```sh
# set GOPATH to whatever you feel like
export GOPATH="$HOME"
# go binalries will end up in $GOPATH/bin, so we update the path
export PATH="$PATH:$GOPATH/bin"
```

(see the [go docs](https://golang.org/doc/code.html#GOPATH) for more details)


### Usage

```sh
# Convert an existing Quiver library to Markdown
$ quiver_to_markdown /path/to/Quiver.qvlibrary /output/path

# Print version
$ quiver_to_markdown -v
```

Then all the notes are available in `/output/path` as Markdown files.

## License

This project is licensed under the MIT License - see the [LICENSE](../../LICENSE) file for details

## TODO

* [x] Add support for note links
* [ ] Allow to convert back saved notes to [Quiver]

[Quiver]: https://itunes.apple.com/app/id866773894
[Homebrew]: https://brew.io

