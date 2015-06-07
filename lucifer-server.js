// Lucifer Server
//
// Copyright 2015 Kevin Burke
//
// MIT Licensed

var express = require('express');
var Mocha = require('mocha');
var path = require('path');

var version = '0.1';

module.exports = function(initTestEnvironment, isTestFile, opts) {
  var app = express();
  app.use(express.logger('dev'));
  app.disable('x-powered-by');

  var mocha = new Mocha({
    ui: 'bdd',
    reporter: 'spec',
    timeout: 15000,
  });

  // Middleware for reading the request body - since we want to parse as JSON,
  // we need the whole thing, streaming doesn't work.
  app.use(function(req, res, next) {
    req.rawBody = '';
    req.setEncoding('utf8');

    req.on('data', function(chunk) {
      req.rawBody += chunk;
    });

    req.on('end', function() {
      next();
    });
  });

  app.post('/v1/cache/invalidate', function(req, res) {
    res.set('Server', 'lucifer/' + version);
      try {
        req.body = JSON.parse(req.rawBody);
      } catch (err) {
        console.log("lucifer: ", err);
        res.set('Content-Type', 'application/problem+json');
        var er = {
          title: 'Invalid post body (should be a JSON object)',
          type:  'https://github.com/kevinburke/lucifer',
        };
        res.status(400).json(er);
        return;
      }
    var files = req.body.files;
    if (files === null || files === undefined) {
      res.set('Content-Type', 'application/problem+json');
      var err = {
        title: 'Please include files to invalidate',
        type:  'https://github.com/kevinburke/lucifer',
      };
      res.status(400).json(err);
      return;
    }
    res.set('Content-Type', 'application/json');
    for (var i = 0; i < files.length; i++) {
      var file = files[i];
      // XXX check absolute before joining
      var absFilepath = path.join(opts.directory, file);
      if (absFilepath.indexOf(__filename) >= 0) {
        console.log("lucifer: not reloading Lucifer server from disk!");
        continue;
      }
      if (isTestFile(absFilepath)) {
        console.log("lucifer: not invalidating " + absFilepath + "yet, because it's a test file");
        continue;
      }
      try {
        delete require.cache[require.resolve(absFilepath)];
        require(absFilepath);
        console.log("lucifer: reloaded " + file + " from disk");
      } catch (e) { }
    }
    res.status(200).json({message: 'OK'}).end();
  });

  app.post('/v1/test_runs', function(req, res) {
    try {
      try {
        req.body = JSON.parse(req.rawBody);
      } catch (err) {
        console.log("lucifer: ", err);
        res.set('Content-Type', 'application/problem+json');
        var er = {
          title: 'Invalid post body (should be a JSON object)',
          type:  'https://github.com/kevinburke/lucifer',
        };
        res.status(400).json(er);
        return;
      }
      res.set('Content-Type', 'application/json');
      var files = req.body.files;
      if (files === null || files === undefined) {
        res.set('Content-Type', 'application/problem+json');
        var err = {
          title: 'Please include files to invalidate',
          type:  'https://github.com/kevinburke/lucifer',
        };
        res.status(400).json(err);
        return;
      }
      // Reset mocha to its initial state
      mocha.files = [];
      mocha.suite.suites = [];

      // These arguments taken from the _mocha runner
      if (req.body.bail === true) {
        mocha.suite.bail(true);
      } else {
        mocha.suite.bail(false);
      }
      if (req.body.grep) {
        mocha.grep(new RegExp(req.body.grep));
      } else {
        // Allow anything, taken from the Runner constructor
        mocha.grep(/.*/);
      }

      for (var i = 0; i < files.length; i++) {
        var file = files[i];
        // XXX check absolute before joining
        var absFilepath = path.join(opts.directory, file);
        if (isTestFile(absFilepath)) {
          try {
            delete require.cache[require.resolve(absFilepath)];
            require(absFilepath);
          } catch (e) { }
          mocha.addFile(absFilepath);
        } else {
          console.log("lucifer: not running file " + file + " because it's not a test file");
        }
      }
      mocha.suite.slow(opts.slow);
      mocha.run();
      res.status(201).send({status: "queued"});
    } catch (e) {
      console.log(e);
      res.status(500).send('server error').end();
    }
  });
  console.log("lucifer: loading test environment...");
  initTestEnvironment(function(err) {
    if (err) {
      console.log(err);
      process.exit(1);
    }
    var server = app.listen(11666, function() {
      var host = server.address().address;
      var port = server.address().port;
      console.log('lucifer: listening on http://%s:%s', host, port);
    });
  });
};
