This is a library to implement a reverse ssh tunnel proxy 

*The Reuqest Flow*
```
           
                                 External IP Addr

+------------+      +-------------+       +--------------+
|            |   4  |             |  3    |              |
|  Local     <------+    Remote   <-------+    External  |
|            |      |             |       |              |
+------------+      +----------+--+       +--------------+
                1                 |  2
          +------------------->+>Listener
           SSH To Remote


```

Usage:

```bash
go mod vendor
go build reverse_tunnel.go
chmod u+x reverse_tunnel
./reverse_tunnel -h
```


Equal to use ssh command on linux or Mac
```bash
ssh -R 3000:127.0.0.1:3000 -N 10.10.10.10
```