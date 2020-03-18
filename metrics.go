package main

type _metrics struct {
	requested int
	handled   int
	rejected  int
	failed    int
}

var metrics = _metrics{}
