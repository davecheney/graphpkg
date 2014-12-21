# Installation

```sh
# davecheney's original
go get github.com/davecheney/graphpkg

# or jbenet's fork
go get github.com/davecheney/graphpkg
```

# Usage

To graph the dependencies of the net package:

```
x-www-browser $(graphpkg net)
```


```
> graphpkg
usage: graphpkg [flags] <package name>

  -browser=false: open a browser with the output
  -format="dot-svg": format: {dot, dot-*, d3json}
  -match=".*": filter packages
```

# Filtering

graphpkg can also filter out packages that do not match the supplied regex, this may improve the readability of some graphs by excluding the std library:

	graphpkg -match 'launchpad.net' launchpad.net/goamz/s3


# Examples

```
> graphpkg -format=dot-svg -browset
opening in your browser...

> graphpkg -format=dot runtime
digraph {
  N0 [label="unsafe",shape=box];
  N1 [label="runtime",shape=box];
  N1 -> N0 [weight=1];
}

> graphpkg -format=d3json runtime
[{"name":"runtime","size":1000,"imports":["unsafe"]}
,{"name":"unsafe","size":1000}
]
```
