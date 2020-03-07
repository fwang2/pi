# Parallel tree walks and tools 

This is a Go implementation of parallel tree walk and a suite of file system tools aiming for large-scale and performant profiling and tooling. Differ from Python-based [fprof](http://github.com/fwang2/pcircle) and C++ based [fprof2](http://github.com/fwang2/fprof), which both rely on MPI for inter-communication to implement cluster wide _work stealing_ and _distributed termination detection_, the **pi** is meant to run interactively on a single machine. 

On OLCF's Summit production file system, a 250PB, GPFS-based parallel file system, we observed over 200,000 ops/seconds scanning rate, with 128 threads on a Power9 node. It should be more than sufficient for regular use. That said, HPC file system is infamous for extreme cases, such a a single shared directory with more 2 to 7 million files. It is difficult to handle this kind of shared directory if PFS doesn't implement distributed directory striping such as Lustre's DNE2 or GPFS's distributed meta node handling. It remains to be see if this is good enough for a full system scan.

## Install

```
go get -u github.com/fwang2/pi

```

This will be the binary **pi** into your `GOPATH`, by default, it is your `$HOME/go/bin`.

## Demo

![](doc/tty.gif)

## On naming

I thought about using "pie" as the name, who doesn't like a piece of pie? But, "pi" is one letter short, and that is more important :grin:



