A sub-command which displays to the console the query which is going to be made by `etrds`.

The principal action is to print as many `queryERTDS` as given in the arguments. Since this is a wrapper, we are going to use another useful command which is `batcat`.

>[!tip] Building Recommendations
> Try to replicate the manual page of commands you use. It is a way to get in touch and keep in mind the standards.

# Brief `batcat` Description
```less
BATCAT(1)				           General commands Manual  				         BATCAT(1)

NAME
	batcat - a cat(1)  clone with syntax highlightling and Git integration.

USAGE
	batcat [OPTIONS] [FILE]...
	
	batcat cache [CACHE-OPTIONS] [--build|--clear]

DESCRIPTION
       batcat prints the syntax-highlighted content of a collection of FILEs to the terminal.
       If no FILE is specified, or when FILE is '-', it reads from standard input.

       batcat  supports  a  large number of programming and markup languages.  It also
       communicates with git(1) to show modifications with respect to the git index.  batcat
       automatically pipes its  out‐put through a pager (by default: less).

       Whenever  the  output of batcat goes to a non-interactive terminal, i.e. when the 
       output is piped into another process or into a file, batcat will act as a drop-in
       replacement for cat(1) and fall back to printing the plain file contents.
```

# Goal Overall
+ [x] Use `batcat` to print with syntax-highlighted content of a collection of FILEs to the terminal. Including the implementations of standard input. [[catFunctionality|The implementation details are here!!]]
+ [ ] Implement `batcat` feature of displaying differences using `git`. Extend the capabilities of the OPTION **-d, --diff** from `batcat`. [[gitDiffFunctionality|The implementation details are here!!]]