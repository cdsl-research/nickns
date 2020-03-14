# NickNS

NickNS is DNS Server. This server resolve VM Name into IP Address on VMware ESXi.

<img src="https://raw.githubusercontent.com/cdsl-research/nickns/master/overview.png" width=600>

## Installation

### go get

```
$ go get github.com/x-motemen/ghq
```

## Usage

### [0] Enable SSH on ESXi

Generate SSH Key on laptop.

```
$ ssh-keygen -t rsa -b 4096
```

Copy SSH key from from laptop to ESXi.

```
$ sftp root@esxi.example.com
Connected to root@esxi.example.com.
sftp> puts /path/to/id_rsa.pub /etc/ssh/keys-root/authorized_keys
```

Test Connecting to SSH on Terminal.

```
$ ssh -i /path/to/id_rsa root@esxi.example.com
```

See also: https://kb.vmware.com/s/article/1002866

### [1] Edit Config files

Set ESXi Host on **hosts.toml**.

```
[host A]
address = "esxi.example.com"
port = "22"
user = "root"
identity_file = "/path/to/id_rsa"

[host B]
address = "esxi.example.com"
port = "22"
user = "root"
identity_file = "~/path/to/id_rsa"
```

Set NickNS running options on **config.toml**.

```
port = 5310
domains = ["local.", "example.com."]
ttl = 3600
```

### [2] start server

```
$ nickns
2020/03/14 21:40:25 NickNS Starting at 5310/udp
2020/03/14 21:40:27 [QueryHit] elastic5.local. => 192.168.0.36
2020/03/14 21:40:30 [QueryHit] elastic5.example.com. => 192.168.0.36
```

## For Develop

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

