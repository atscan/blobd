# blobd

> AT Protocol Blob-serving HTTP Server in Go

A high-performance web server that provides better access to binary data (called
[blobs](https://atproto.com/specs/data-model#blob-type)) from the
[AT Protocol](https://atproto.com/). It automatically locates the blob and the
PDS that hosts it using the unique `did` identity and `cid` reference - then
downloads it, verify it, performs the required transformations, and returns it
to the user (and stores it in the cache for later).

It is recommended to have blobd running behind a reverse proxy such as
[Caddy](https://caddyserver.com/) or [Nginx](https://www.nginx.com/) if you want
features such as SSL or load balancing.

You can try the application at [blob.atscan.net](https://blob.atscan.net/did:plc:z72i7hdynmk6r22z27h6tvur/bafkreic5kmqlhrhbfnh2bx6fsetvkra4noqja5ngsnnadrvubd6jcoc3ae), which is a publicly accessible instance.

### Features

- automatic blob discovery (using atscan api)
- MIME type detection
- on-fly image transfomations (webp, different resolutions)
- file storage/cache
- blob inspection endpoint

### TODO

- transport compression
- storage/cache compression (zstd?)
- pre-fetching of blobs from PDS
- metrics (HTTP server, processing)
- custom PLCs or PDS

## Endpoints

| Method | Path         | Name         | Examples |
| ------ | ------------ | ------------ | -------- |
| GET    | `/<did>/<cid>` | Get the blob (original) | [(1)](https://blob.atscan.net/did:plc:z72i7hdynmk6r22z27h6tvur/bafkreic5kmqlhrhbfnh2bx6fsetvkra4noqja5ngsnnadrvubd6jcoc3ae), [(2)](https://blob.atscan.net/did:plc:ewvi7nxzyoun6zhxrhs64oiz/bafkreibjfgx2gprinfvicegelk5kosd6y2frmqpqzwqkg7usac74l3t2v4) |
| GET    | `/<did>/<cid>?format=webp&width=200` | Get transcoded image | [(1)](https://blob.atscan.net/did:plc:z72i7hdynmk6r22z27h6tvur/bafkreic5kmqlhrhbfnh2bx6fsetvkra4noqja5ngsnnadrvubd6jcoc3ae?format=webp&width=200), [(2)](https://blob.atscan.net/did:plc:ewvi7nxzyoun6zhxrhs64oiz/bafkreibjfgx2gprinfvicegelk5kosd6y2frmqpqzwqkg7usac74l3t2v4?format=webp&width=200) |
| GET    | `/<did>/<cid>/inspect` | Inspect blob | [(1)](https://blob.atscan.net/did:plc:z72i7hdynmk6r22z27h6tvur/bafkreic5kmqlhrhbfnh2bx6fsetvkra4noqja5ngsnnadrvubd6jcoc3ae/inspect), [(2)](https://blob.atscan.net/did:plc:ewvi7nxzyoun6zhxrhs64oiz/bafkreibjfgx2gprinfvicegelk5kosd6y2frmqpqzwqkg7usac74l3t2v4/inspect) |

## Usage

Requires:

- go 1.20+
- libwebp

### Install Dependencies

#### Linux
```bash
sudo apt-get update
sudo apt-get install libwebp-dev
```
#### Mac OS
```bash
brew install webp
```

### Install

You can install the application using this command:

```bash
go install github.com/atscan/blobd@latest
```

### How to start

Starting the server on port `3000` and caching the blobs in `/path/to/data`:

```bash
blobd -d /path/to/data -p 3000
```

Try it out to see if it works:

```bash
http localhost:3000/did:plc:ewvi7nxzyoun6zhxrhs64oiz/bafkreibjfgx2gprinfvicegelk5kosd6y2frmqpqzwqkg7usac74l3t2v4
```

Result:

```httpie
HTTP/1.1 200 OK
Content-Length: 86984
Content-Type: image/jpeg
Date: Fri, 14 Jul 2023 09:38:55 GMT



+-----------------------------------------+
| NOTE: binary data not shown in terminal |
+-----------------------------------------+
```

## Authors

- [tree 🌴](https://bsky.app/profile/did:plc:524tuhdhh3m7li5gycdn6boe)

## License

MIT
