imports
=======

print the imports from go source


Install
=======

	go get github.com/vbatts/imports


Usage
=====

See the imports recursively of the current directory:

	import -r .


See the imports of the source in the current directory:

	imports


See the imports of a specific directory:

	imports ~/go/src/pkg/net


See the imports of library in GOROOT or GOPATH:

	import net/http


