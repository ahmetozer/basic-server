# Basic Server

A basic file server automatically generates self certificates and serves the given folder.

## Options

```bash
Usage of /basic-server:
  -client-cert string
        Allow only given client certificate
  -listen string
        Listen addr (default ":443")
  -path string
        Share Path (default "/web/") # Note if you not run in container, default path will be current path
  -server-cert string
        Server cert
  -server-key string
        Server key
  -server-name string
        Server name check from TLS
```

## Examples

### Run on local

```cmd
C:\Users\Ahmet\Desktop\basic-server.exe -h
2021/05/04 02:46:55 github.com/ahmetozer/basic-server
2021/05/04 02:46:55 Current name WORKSTATION\Ahmet
2021/05/04 02:46:55 ./key.pem exist
2021/05/04 02:46:55 ./cert.pem exist
2021/05/04 02:29:51 Starting HTTPS server on :443 at C:\Users\Ahmet\Desktop\
```

### Run in container

This configuration accepts all https clients with all domains.

```bash
docker run -it --rm -p 443:443 --mount type=bind,source="/my/path/",target=/web/,readonly  ghcr.io/ahmetozer/basic-server
```

### Run in container with client cert control

To allow incoming request from only cloudflare enable 'Authenticated Origin Pulls' in cloudflare.

```bash
docker run -it --rm -p 443:443 --mount type=bind,source="/my/path/",target=/web/,readonly ghcr.io/ahmetozer/basic-server \
--server-name mydomain.test  --client-cert /config/client-cloudflare.pem
```

### Run in different user

You can also run in different user id for access shared path.

```bash
docker run -it --rm -p 443:443 -u 1249:1249 --mount type=bind,source="/secret/my/path/",target=/web/,readonly ghcr.io/ahmetozer/basic-server \
--server-name mydomain.test  --client-cert config/client-cloudflare.pem
```

## Dump

You can also inspect your HTTP request with visiting /dump page.
