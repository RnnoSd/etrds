Printing with highlighting for queries is the main functionality of `etrds show` command which work with any arguments, none or at the end -(hyphen) as an argument.

Also, it must implement the redirection in a pipe like `cat` does.

It is important for our implementation to manage the way our command will display `stdin` as a stream like in the use of `-` and as an stream from a pipe.
# Tests
Indeed, the test are the most essential part of this work. Given this is a fully *cli implementation* the best to test is by using `BASH`. This is done through an embedding into go test architecture.
## Main Bash test
+ [x] Display files
```include bash
@/internal/show/test/testDisplay.sh
```
+ [x] Display `stdin` from arguments
```include bash
@/internal/show/test/testStdinArguments.sh
```
+ [x] Display `stdin` from a pipe
```include bash
@/internal/show/test/testStdinPipe.sh
```