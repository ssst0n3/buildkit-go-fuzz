# fuzz

## FuzzParseAttributes

```shell
root@ubuntu:~/fuzz_project/buildkit-go-fuzz# ./fuzz.sh ./util/compression/ -fuzz=FuzzParseAttributes
+++ pwd
++ docker run -tid -v /root/fuzz_project/buildkit-go-fuzz:/go/src/github.com/moby/buildkit/ -w /go/src/github.com/moby/buildkit/ golang:1.25 go test '-run=^$' ./util/compression/ -fuzz=FuzzParseAttributes
+ CID=b31d665e35a95ca97313ae58d3252182c0a1c1160ae1d735f18dcb09408467dc
+ docker logs -f b31d665e35a95ca97313ae58d3252182c0a1c1160ae1d735f18dcb09408467dc
fuzz: elapsed: 0s, gathering baseline coverage: 0/5 completed
fuzz: elapsed: 0s, gathering baseline coverage: 5/5 completed, now fuzzing with 24 workers
fuzz: elapsed: 3s, execs: 482866 (160944/sec), new interesting: 15 (total: 20)
...
fuzz: elapsed: 15s, execs: 3068468 (209289/sec), new interesting: 15 (total: 20)
fuzz: elapsed: 18s, execs: 3719502 (217004/sec), new interesting: 17 (total: 22)
...
fuzz: elapsed: 1m21s, execs: 17964689 (226980/sec), new interesting: 17 (total: 22)
fuzz: elapsed: 1m24s, execs: 18632379 (222429/sec), new interesting: 18 (total: 23)
...
fuzz: elapsed: 5m57s, execs: 38029699 (108650/sec), new interesting: 19 (total: 24)
...
fuzz: elapsed: 19h39m0s, execs: 5590951140 (225669/sec), new interesting: 19 (total: 24)
```

## exporter/containerimage
### FuzzResolve

```shell
root@ubuntu:~/fuzz_project/buildkit-go-fuzz# ./fuzz.sh ./exporter/containerimage/ -fuzz=FuzzResolve
+++ pwd
++ docker run -tid -v /root/fuzz_project/buildkit-go-fuzz:/go/src/github.com/moby/buildkit/ -w /go/src/github.com/moby/buildkit/ golang:1.25 go test '-run=^$' ./exporter/containerimage/ -fuzz=FuzzResolve
+ CID=fc777850804ca0f4f083a9d76fe4cdd0def9505f85530d61f99e6ccbb8fb9cdc
+ docker logs -f fc777850804ca0f4f083a9d76fe4cdd0def9505f85530d61f99e6ccbb8fb9cdc
fuzz: elapsed: 0s, gathering baseline coverage: 0/5 completed
fuzz: elapsed: 0s, gathering baseline coverage: 5/5 completed, now fuzzing with 24 workers
fuzz: elapsed: 3s, execs: 6498 (2162/sec), new interesting: 6 (total: 11)
fuzz: elapsed: 6s, execs: 24645 (6055/sec), new interesting: 16 (total: 21)
...
fuzz: elapsed: 16h59m54s, execs: 708223601 (10546/sec), new interesting: 66 (total: 71)
...
fuzz: elapsed: 19h29m12s, execs: 853080947 (23581/sec), new interesting: 66 (total: 71)
```

### FuzzLoad

```shell
root@ubuntu:~/fuzz_project/buildkit-go-fuzz# ./fuzz.sh ./exporter/containerimage/ -fuzz=FuzzLoad$
+++ pwd
++ docker run -tid -v /root/fuzz_project/buildkit-go-fuzz:/go/src/github.com/moby/buildkit/ -w /go/src/github.com/moby/buildkit/ golang:1.25 go test '-run=^$' ./exporter/containerimage/ '-fuzz=FuzzLoad$'
+ CID=33a0b796dea922406c6036f54d46fcb1b555a0a59ad966b7375a22afa8557b01
+ docker logs -f 33a0b796dea922406c6036f54d46fcb1b555a0a59ad966b7375a22afa8557b01
fuzz: elapsed: 0s, gathering baseline coverage: 0/8 completed
fuzz: elapsed: 1s, gathering baseline coverage: 8/8 completed, now fuzzing with 24 workers
fuzz: elapsed: 3s, execs: 6366 (2120/sec), new interesting: 3 (total: 11)
...
fuzz: elapsed: 13h52m21s, execs: 668550624 (7553/sec), new interesting: 245 (total: 253)
...
fuzz: elapsed: 18h56m12s, execs: 894706349 (26947/sec), new interesting: 245 (total: 253)
```

### FuzzCommit

```shell
root@ubuntu:~/fuzz_project/buildkit-go-fuzz# ./fuzz.sh ./exporter/containerimage/ -fuzz=FuzzCommit$
+++ pwd
++ docker run -tid -v /root/fuzz_project/buildkit-go-fuzz:/go/src/github.com/moby/buildkit/ -w /go/src/github.com/moby/buildkit/ golang:1.25 go test '-run=^$' ./exporter/containerimage/ '-fuzz=FuzzCommit$'
+ CID=15a02e712fabe046481637aa749f8e15e9290102e8deed22f177dcef8586b511
+ docker logs -f 15a02e712fabe046481637aa749f8e15e9290102e8deed22f177dcef8586b511
fuzz: elapsed: 0s, gathering baseline coverage: 0/4 completed
fuzz: elapsed: 0s, gathering baseline coverage: 4/4 completed, now fuzzing with 24 workers
fuzz: elapsed: 3s, execs: 443 (148/sec), new interesting: 3 (total: 7)
fuzz: elapsed: 6s, execs: 1649 (402/sec), new interesting: 8 (total: 12)
fuzz: elapsed: 9s, execs: 3047 (466/sec), new interesting: 10 (total: 14)
fuzz: elapsed: 12s, execs: 4797 (583/sec), new interesting: 10 (total: 14)
...
fuzz: elapsed: 22h30m0s, execs: 36227932 (5389/sec), new interesting: 25 (total: 29)
...
fuzz: elapsed: 23h15m48s, execs: 50895939 (4921/sec), new interesting: 25 (total: 29)
```

### FuzzCommitAttestationsManifest

```shell
root@ubuntu:~/fuzz_project/buildkit-go-fuzz# ./fuzz.sh ./exporter/containerimage/ -fuzz=FuzzCommitAttestationsManifest
+++ pwd
++ docker run -tid -v /root/fuzz_project/buildkit-go-fuzz:/go/src/github.com/moby/buildkit/ -w /go/src/github.com/moby/buildkit/ golang:1.25 go test '-run=^$' ./exporter/containerimage/ -fuzz=FuzzCommitAttestationsManifest
+ CID=7ada8e0523a6d337f865e5675cef0390cfea65a09d046c3fc68ef24cfbe94cee
+ docker logs -f 7ada8e0523a6d337f865e5675cef0390cfea65a09d046c3fc68ef24cfbe94cee
fuzz: elapsed: 0s, gathering baseline coverage: 0/4 completed
fuzz: elapsed: 0s, gathering baseline coverage: 4/4 completed, now fuzzing with 24 workers
fuzz: elapsed: 3s, execs: 60 (20/sec), new interesting: 0 (total: 4)
fuzz: elapsed: 6s, execs: 60 (0/sec), new interesting: 0 (total: 4)
fuzz: elapsed: 9s, execs: 60 (0/sec), new interesting: 0 (total: 4)
fuzz: elapsed: 12s, execs: 759 (233/sec), new interesting: 1 (total: 5)
...
fuzz: elapsed: 23h52m42s, execs: 16224040 (337/sec), new interesting: 497 (total: 501)
...
fuzz: elapsed: 23h56m6s, execs: 16307588 (1/sec), new interesting: 497 (total: 501)
```

### FuzzCommitDistributionManifest

```shell
root@ubuntu:~/fuzz_project/buildkit-go-fuzz# ./fuzz.sh ./exporter/containerimage/ -fuzz=FuzzCommitDistributionManifest
+++ pwd
++ docker run -tid -v /root/fuzz_project/buildkit-go-fuzz:/go/src/github.com/moby/buildkit/ -w /go/src/github.com/moby/buildkit/ golang:1.25 go test '-run=^$' ./exporter/containerimage/ -fuzz=FuzzCommitDistributionManifest
+ CID=4ec3406c52cea36c9f540c685342e6b1487e3a2f9c7bc0a0fa57608931042ad7
+ docker logs -f 4ec3406c52cea36c9f540c685342e6b1487e3a2f9c7bc0a0fa57608931042ad7
fuzz: elapsed: 0s, gathering baseline coverage: 0/3 completed
fuzz: elapsed: 0s, gathering baseline coverage: 3/3 completed, now fuzzing with 24 workers
fuzz: elapsed: 3s, execs: 3500 (1167/sec), new interesting: 7 (total: 10)
fuzz: elapsed: 6s, execs: 29645 (8713/sec), new interesting: 23 (total: 26)
...
fuzz: elapsed: 24h7m6s, execs: 13254432 (701/sec), new interesting: 579 (total: 583)
...
fuzz: elapsed: 24h20m24s, execs: 13536844 (0/sec), new interesting: 579 (total: 583)
```

### FuzzParseAnnotations

```shell
root@ubuntu:~/fuzz_project/buildkit-go-fuzz# ./fuzz.sh ./exporter/containerimage/ -fuzz=FuzzParseAnnotations
+++ pwd
++ docker run -tid -v /root/fuzz_project/buildkit-go-fuzz:/go/src/github.com/moby/buildkit/ -w /go/src/github.com/moby/buildkit/ golang:1.25 go test '-run=^$' ./exporter/containerimage/ -fuzz=FuzzParseAnnotations
+ CID=5905df11a3e220165d5b0cd486f49ad8cd4f0bcbed9c89d2bb56aa9bf7ab8f11
+ docker logs -f 5905df11a3e220165d5b0cd486f49ad8cd4f0bcbed9c89d2bb56aa9bf7ab8f11
fuzz: elapsed: 0s, gathering baseline coverage: 0/9 completed
fuzz: elapsed: 0s, gathering baseline coverage: 9/9 completed, now fuzzing with 24 workers
fuzz: elapsed: 3s, execs: 7905 (2635/sec), new interesting: 8 (total: 17)
fuzz: elapsed: 6s, execs: 69002 (20364/sec), new interesting: 18 (total: 27)
...
fuzz: elapsed: 3h16m48s, execs: 525917938 (45755/sec), new interesting: 135 (total: 144)
...
fuzz: elapsed: 3h53m3s, execs: 625696567 (26741/sec), new interesting: 135 (total: 144)
```