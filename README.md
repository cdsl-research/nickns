# NickNS

![](https://github.com/cdsl-research/nickns/workflows/build/badge.svg)

NickNS is a DNS Server for ESXi Resources

<img src="https://raw.githubusercontent.com/cdsl-research/nickns/master/overview.png" width=600>

## Develop

Build local ssh server

```
docker build -t ssh .
docker run -it -p 2200:22 ssh
```

Run DNS Server

```
go run main.go
```

Check DNS server

```
# Type A
$ dig A +short @127.0.0.1 -p 5300 unbound.local
192.168.0.35
$ dig A +short @127.0.0.1 -p 5300 unbound.local
192.168.0.35

# Type PTR
$ dig +short @127.0.0.1 -p 5300 -x 192.168.0.35
unbound.local.
```

Server Log

```
2020/02/29 18:47:25 Starting at 5300
2020/02/29 18:47:38 [QueryHit] unbound.local. => 192.168.0.35
2020/02/29 18:47:41 [CacheHit] unbound.local.
2020/02/29 18:47:42 [QueryHit] unbound.local. => 192.168.0.35
2020/02/29 18:47:52 [QueryHit] 35.0.168.192.in-addr.arpa. => unbound.local.
```

ref: https://gist.github.com/walm/0d67b4fb2d5daf3edd4fad3e13b162cb
