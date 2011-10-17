include $(GOROOT)/src/Make.inc

TARG=gough
GOFILES=\
	hough.go\
	geometry.go\
	edges.go\
	interface.go\
	helper.go\
	common.go\
	storage.go\
	main.go\

include $(GOROOT)/src/Make.cmd
