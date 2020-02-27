# cdsn

CDSN is a DNS Server for ESXi Resources

## Develop

Build local ssh server

```
docker build -t ssh .
docker run -it -p 2200:22 ssh
```

Check DNS server

```
dig @127.0.0.1 -p 5300 dev51.hoge
```

ref: https://gist.github.com/walm/0d67b4fb2d5daf3edd4fad3e13b162cb
