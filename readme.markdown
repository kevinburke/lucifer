# Lucifer

This is a tool that will help you reload your Javascript based test suite
really fast.

### Server 

#### Usage

Lucifer exposes one function with the following signature:

```javascript
var lucifer = function(initTestEnvironment, isTestFile, opts)
```

**initTestEnvironment**: function(cb) A function which sets up your test
environment. For us this is `sails.lift` and a few other things. Should take a
`callback` as an argument, and callback with a single error argument if there
was one.

**isTestFile**: function(file) -> bool. A function which takes a filename as
input and returns true if the file is a test file. Necessary so we know which
files to run, and because you don't want to re-require test files until you are
ready to run them.

**opts**: dictionary. Currently the dictionary supports two keys:

  - **directory**: When you make requests to the Lucifer server, assume they
    are relative to this directory.
  - **slow**: A slow test threshold, in milliseconds

So a simple example would be:

```javascript
var lucifer = require('./lucifer');
var s = require('sails');

var startTestEnvironment = function(cb) {
  var start = Date.now();
  s.lift({}, function(err) {
    if (err) {
      return cb(err);
    }
    console.log('Sails ready.. booted in ' + (Date.now() - start) + 'ms');
    return cb();
  });
};

var isTestFile = function(file) {
  return file.indexOf('.test.') >= 0;
};

var opts = { directory: __dirname, slow: 2 };
lucifer(startTestEnvironment, isTestFile, opts);
```

#### Installation

Copy the `lucifer-server.js` file to a place you can import it from.

#### Server Dependencies

The server is an instance of [the Express web server][express] and runs tests
with [Mocha][mocha].

[express]: http://expressjs.com/
[mocha]: http://mochajs.org/

### Client 

#### Installation

```bash
go get github.com/kevinburke/lucifer    # Or 'make install'
```

I should really package these up / distribute them as binaries...

#### Usage

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

[slow]: https://kev.inburke.com/kevin/node-require-is-dog-slow/
