package main

import (
	"k8s.io/klog/v2"
	"log"
	"os"
)

func init() {
	log.SetFlags(log.Lshortfile)
}

func main() {
	cmd, err := newRootCmd(os.Stdout, os.Args[1:])
	if err != nil {
		os.Exit(1)
	}

	if err := cmd.Execute(); err != nil {
		klog.Errorf("run command failed, error: %s", err)
	}
}
