# fuzz

## FuzzPatchImageConfig

```shell
root@ubuntu:~/research_project/buildkit-go-fuzz# ./fuzz.sh ./exporter/containerimage -fuzz=FuzzPatchImageConfig
++ pwd
+ docker run -tid -v /home/st0n3/fuzz_project/buildkit-go-fuzz:/go/src/github.com/moby/buildkit/ -w /go/src/github.com/moby/buildkit/ golang:1.25 go test '-run=^$' ./exporter/containerimage -fuzz=FuzzPatchImageConfig
d3c5f110d8651f3930b8741a06e7e59c580ff2d9b896fb119eec114e58289a62 
root@ubuntu:~/research_project/buildkit-go-fuzz# docker logs -f d3c
fuzz: elapsed: 13m0s, execs: 10640681 (13070/sec), new interesting: 572 (total: 574)
fuzz: minimizing 97-byte failing input file
fuzz: elapsed: 13m2s, minimizing
--- FAIL: FuzzPatchImageConfig (782.20s)
    --- FAIL: FuzzPatchImageConfig (0.00s)
        testing.go:1825: panic: assignment to entry in nil map
            goroutine 1358284 [running]:
            runtime/debug.Stack()
            	/usr/local/go/src/runtime/debug/stack.go:26 +0x9b
            testing.tRunner.func1()
            	/usr/local/go/src/testing/testing.go:1825 +0x1d0
            panic({0x12bd8c0?, 0x1b86120?})
            	/usr/local/go/src/runtime/panic.go:783 +0x132
            github.com/moby/buildkit/exporter/containerimage.patchImageConfig({0xc0026c2d80, 0x4, 0x8}, {0x0, 0x0, 0xc001e1e7c0?}, {0x0, 0x0, 0x0}, {0xc0026c2dc8, ...}, ...)
            	/go/src/github.com/moby/buildkit/exporter/containerimage/writer.go:586 +0x3e5
            github.com/moby/buildkit/exporter/containerimage.FuzzPatchImageConfig.func1(0xc000694cc0?, {0xc0026c2d80, 0x4, 0x8}, {0xc0026c2d98, 0x1, 0x8}, {0xc0026c2db0, 0x1, 0x8}, ...)
            	/go/src/github.com/moby/buildkit/exporter/containerimage/writer_test.go:74 +0x25d
            reflect.Value.call({0x12d9300?, 0x142f158?, 0x13?}, {0x13e690a, 0x4}, {0xc0022fdb00, 0x7, 0x8?})
            	/usr/local/go/src/reflect/value.go:581 +0xcc6
            reflect.Value.Call({0x12d9300?, 0x142f158?, 0x1b31d60?}, {0xc0022fdb00?, 0x13e5b20?, 0x199edf0?})
            	/usr/local/go/src/reflect/value.go:365 +0xb9
            testing.(*F).Fuzz.func1.1(0xc002842e00?)
            	/usr/local/go/src/testing/fuzz.go:341 +0x32a
            testing.tRunner(0xc002842e00, 0xc0025d85a0)
            	/usr/local/go/src/testing/testing.go:1934 +0xea
            created by testing.(*F).Fuzz.func1 in goroutine 19
            	/usr/local/go/src/testing/fuzz.go:328 +0x637
            
    
    Failing input written to testdata/fuzz/FuzzPatchImageConfig/b302afdfe365c152
    To re-run:
    go test -run=FuzzPatchImageConfig/b302afdfe365c152
FAIL
exit status 1
FAIL	github.com/moby/buildkit/exporter/containerimage	782.228s
```

## FuzzCommitDistributionManifest

```shell
fuzz: elapsed: 26m0s, execs: 43451423 (30567/sec), new interesting: 699 (total: 702)
fuzz: minimizing 120-byte failing input file
fuzz: elapsed: 26m1s, minimizing
--- FAIL: FuzzCommitDistributionManifest (1561.22s)
    --- FAIL: FuzzCommitDistributionManifest (0.00s)
        testing.go:1825: panic: assignment to entry in nil map
            goroutine 1341918 [running]:
            runtime/debug.Stack()
            	/usr/local/go/src/runtime/debug/stack.go:26 +0x9b
            testing.tRunner.func1()
            	/usr/local/go/src/testing/testing.go:1825 +0x1d0
            panic({0x12e2e80?, 0x1bc1140?})
            	/usr/local/go/src/runtime/panic.go:783 +0x132
            github.com/moby/buildkit/exporter/containerimage.patchImageConfig({0xc001b72fc8, 0x4, 0x8}, {0x0, 0x0, 0x4a600000000?}, {0x0, 0x0, 0x0}, {0x1c0c7e0, ...}, ...)
            	/go/src/github.com/moby/buildkit/exporter/containerimage/writer.go:586 +0x3e5
            github.com/moby/buildkit/exporter/containerimage.(*ImageWriter).commitDistributionManifest(0xc0013f76b8, {0x1550a30, 0x1c0c7e0}, 0xc0013f7660, {0x0, 0x0}, {0xc001b72fc8, 0x4, 0x8}, 0xc0013f75c8, ...)
            	/go/src/github.com/moby/buildkit/exporter/containerimage/writer.go:345 +0x28b
            github.com/moby/buildkit/exporter/containerimage.FuzzCommitDistributionManifest.func1(0xc001b8f6c0, 0x0, {0xc001b72fc8, 0x4, 0x8}, {0xc001b72fe0, 0x1, 0x8}, {0xc001b72ff8, 0x1, ...}, ...)
            	/go/src/github.com/moby/buildkit/exporter/containerimage/writer_fuzz_test.go:107 +0x6ba
            reflect.Value.call({0x1321140?, 0x1457b18?, 0x13?}, {0x140e4d3, 0x4}, {0xc001bba780, 0x9, 0x10?})
            	/usr/local/go/src/reflect/value.go:581 +0xcc6
            reflect.Value.Call({0x1321140?, 0x1457b18?, 0x1b6cd60?}, {0xc001bba780?, 0x140d6e0?, 0x19d5a38?})
            	/usr/local/go/src/reflect/value.go:365 +0xb9
            testing.(*F).Fuzz.func1.1(0xc001b8f6c0?)
            	/usr/local/go/src/testing/fuzz.go:341 +0x32a
            testing.tRunner(0xc001b8f6c0, 0xc001ba1d40)
            	/usr/local/go/src/testing/testing.go:1934 +0xea
            created by testing.(*F).Fuzz.func1 in goroutine 53
            	/usr/local/go/src/testing/fuzz.go:328 +0x637
            
    
    Failing input written to testdata/fuzz/FuzzCommitDistributionManifest/59a19893b9ae43e6
    To re-run:
    go test -run=FuzzCommitDistributionManifest/59a19893b9ae43e6
FAIL
exit status 1
FAIL	github.com/moby/buildkit/exporter/containerimage	1561.272s
```
