# Go language Web API Boilerplate
Fast, Small, Secure

## links

 * [httprouter](https://github.com/julienschmidt/httprouter)
 * [negroni](https://github.com/urfave/negroni)
 * [cors](https://github.com/rs/cors)
 * [secure](https://github.com/unrolled/secure)
 * [logrus](https://github.com/sirupsen/logrus)
 * [oauth2](https://golang.org/x/oauth2)
 * [jason](https://github.com/antonholmquist/jason)
 * [yaml](https://gopkg.in/yaml.v2)

## SetUp

Edit /etc/hosts or C:\Windows\system32\drivers\etc\hosts

```/etc/hosts
127.0.0.1 www.local.test
127.0.0.1 www2.local.test
127.0.0.1 api.local.test
```

## Run

Rename the directory to a simple name like "foobar".
For recommended naming conventions you should refer to ["The Go Blog"'s post](https://blog.golang.org/package-names).

Add your APIs to handler.go.

And run

```
go build && ./foobar
```


