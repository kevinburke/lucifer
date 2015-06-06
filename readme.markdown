# Lucifer

This is a tool that will help you reload your Javascript based test suite
really fast.

### Usage

```
The Lucifer binary makes requests to the Lucifer server.

Usage:

    lucifer command [arguments]

The commands are:

    invalidate      Invalidate the cache for a given file
    run             Run tests for a given file

Use "lucifer help [command]" for more information about a command.
```

### Why?

Node's `require` is [really, really slow][slow]. If you need to load a large
app with an unfortunately large number of dependencies to run your test suite,
you're looking at a 5-10 second penalty to run a single test. 

Instead of requiring every file every time you want to run your test suite,
load all of them once and listen on a socket for incoming test run requests.

##### But a long running server won't take into account the changes I make

Yes! Luckily it's not too difficult to invalidate the Node cache for a module.
If you call `lucifer invalidate [file]` the binary will make a request to the
server to invalidate the cache for that module; this way you can ensure the
server is running tests against the version of the module on your file system.

### Install

Run `make install`. You'll want to have Go installed to install the command
line client.

[slow]: https://kev.inburke.com/kevin/node-require-is-dog-slow/
