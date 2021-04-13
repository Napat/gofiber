# Hello

``` bash
go run ./cmd/fiberserver/
```

Request GET plaintext

``` bash
$ curl -v http://127.0.0.1:3000/
...
< Content-Type: text/plain; charset=utf-8
...
Hello, World!

$ curl -v http://127.0.0.1:3000/q/xxx=1
```

Request GET JSON

``` bash
$ curl -sv http://127.0.0.1:3000/health | jq
...
< Content-Type: application/json
...
{
  "status": "ok"
}

$ curl -sv http://127.0.0.1:3000/greets | jq
...
< Content-Type: application/json
...
[
  {
    "uid": 1,
    "title": "Hello Foo",
    "message": "Hello To Mr.Foo"
  },
  {
    "uid": 2,
    "title": "Hello Bar",
    "message": "Hello To Ms.Bar"
  }
]
```

Request Static File/page

``` bash
$ curl http://127.0.0.1:3000/web
<html>
    <head>My Header</head>
    <body>
        Hello HTML
    </body>
</html>
```
