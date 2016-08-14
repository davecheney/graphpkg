# Installation

First install [graphviz](http://graphviz.org/Download.php) for your OS, then

	go get github.com/davecheney/graphpkg

# Usage

To graph the dependencies of the net package:

	graphpkg net

# Filtering

graphpkg can also filter out packages that do not match the supplied regex, this may improve the readability of some graphs by excluding the std library:

	graphpkg -match 'launchpad.net' launchpad.net/goamz/s3

# Output

By default graphpkg shows the graph in your browser, you can choose to print the resulting svg to standard output:

	graphpkg -stdout -match 'github.com' github.com/davecheney/graphpkg
