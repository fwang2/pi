package fs

/*

This note is to implement copy more efficiently.

source, err := os.Open("source.txt")
destination, err := os.OpenFile("dest.txt", os.O_RDWR|os.O_CREATE, 0666)
_, err = io.Copy(destination, source)  // Copy (dst Writer, src Reader)

Above code is fine EXCEPT:

* the single file is extremely large
* the single file is sparse

The first case can be addressed by:

* breaking the file into chunks
* create a thread pool, and workers copy chunks in parallel
* master will close the file when all is done.


*/
