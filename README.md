# blobd

> [AT Protocol Blob](https://atproto.com/specs/data-model#blob-type)-serving HTTP Server in Go

A web server that provides better access to binary data (called [blobs](https://atproto.com/specs/data-model#blob-type)) from the [AT Protocol](https://atproto.com/). It automatically locates the blob and the PDS that hosts it using the unique `did` identity and `cid` reference - then downloads it, performs the required transformations, and returns it to the user (and stores it in the cache for later).

### Features
- automatic blob discovery (using atscan api)
- file storage/cache

### TODO
- image transformations (resolutions, formats, incl. webp)
- transport compression
- storage/cache compression (zstd?)
- pre-fetching of blobs from PDS
- metrics (HTTP server, processing)
- custom PLCs or PDS

## Installation

```bash
go install github.com/atscan/blobd@latest
```

## Usage

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