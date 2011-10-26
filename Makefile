include $(GOROOT)/src/Make.inc

TARG=gough
GOFILES=\
	hough.go\
	geometry.go\
	edges.go\
	window.go\
	helper.go\
	common.go\
	storage.go\
	main.go\

GCIMPORTS=-I $(GOPATH)/pkg/linux_amd64
LDIMPORTS=-L $(GOPATH)/pkg/linux_amd64

include $(GOROOT)/src/Make.cmd
