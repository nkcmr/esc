# esc
a utility to automatically escape shell characters

## install
```
go install -v code.nkcmr.net/esc@latest
```

## usage
this (for me) really reduces the mental overhead of trying to escape things with `xargs` since it is really picky:
```
# list files with names with spaces/single-quotes/double/quotes
find . -type f \
  | xargs -I% echo '%'

# xargs will break and complain about unterminated quotes n such ☝️

esc --help
Usage:
  esc [flags]

Flags:
  -d, --double-quoted   signifies that the strings should be escaped for an double quoted evaluation
  -h, --help            help for esc
  -l, --per-line        will escape each input line individually (default true)
  -s, --single-quoted   signifies that the strings should be escaped for an single quoted evaluation
  -u, --unquoted        signifies that the strings should be escaped for an unquoted evaluation

# piping through esc and setting -u (unquoted) will properly escape by line
find . type f \
  | esc -u \
  | xargs -I% echo '%'
# note the quotes   ☝️ here are interpreted first so to xargs, it is unquoted
# (see! confusing!)
```
