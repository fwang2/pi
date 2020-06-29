# Parallel tree walks and tools 

This is a Go implementation of parallel tree walk and a suite of file system tools aiming for large-scale and performant profiling. Differing from Python-based [pcircle](http://github.com/fwang2/pcircle) and C++ based [fprof](http://github.com/fwang2/fprof), where both rely on MPI for inter-communication to implement cluster wide _work stealing_ and _distributed termination detection_, **pi** is meant to run interactively on a single machine with good scaling properties. 

On the OLCF's Summit production file system, a 250 PB, GPFS-based parallel file
system, we measured over 200,000 ops/seconds scanning rate, on a single IBM
POWER9 node running with 128 threads. It should be more than sufficient for
regular use. 

That said, HPC
file system is infamous for extreme cases, such a a single shared directory with
more 2 to 7 million files. It is difficult to handle this kind of shared
directory if PFS doesn't implement distributed directory striping such as
[Lustre's
DNE2](http://cdn.opensfs.org/wp-content/uploads/2015/04/Scalability-Testing-of-DNE2-in-Lustre-27_Simms_V2.pdf)
or GPFS's distributed meta node handling. It remains to be see if this is good
enough for a full system scan.

## Local install

Assuming you have golang installed and available on your `PATH`:

```
go get -u github.com/fwang2/pi
```

This will be the binary **pi** into your `GOPATH`, by default, it is your `$HOME/go/bin`.

## On Summit or Rhea

On Summit (POWER arch)
```
module use /sw/exp9/spack/modules/linux-rhel7-power91e
```

On Rhea (x86)
```
module use /sw/exp9/linux-rhel7-sandybridge
```

Assume above module use is okay, then:

```
module load pi
```

will make pi available to use.


## Examples

### List top n largest files and directories

```
▶ pi topn .
```

### Profiling and show file distributions

```
▶ pi profile --hist .
```

`--hist` is to build histogram of file distribution. It is turned off by default.

### Find all files of size greater than 100M, created within last 7 days

```
▶ pi find / --type f --size +10m --ctime 7d
```
### Create tar.gz 

```
▶ pi zip /path/to/project -o project.tar.gz
```

This can be helpful if you have large files and feels tar/zip is taking too long. The compression itself is parallelized, but subsequent tar still have room to improve though. 


### Check sparse file (Linux only)

```
▶ pi sparse-check /path/to/sparsefile
```


