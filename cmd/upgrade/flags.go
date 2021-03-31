package main

import (
	"flag"
	"strings"

	"github.com/spf13/pflag"
	"k8s.io/klog/v2"
)

// addKlogFlags adds flags from k8s.io/klog
// marks the flags as hidden to avoid polluting the help text
func addKlogFlags(fs *pflag.FlagSet) {
	local := flag.NewFlagSet("klog", flag.ExitOnError)
	klog.InitFlags(local)
	local.VisitAll(func(fl *flag.Flag) {
		fl.Name = normalize(fl.Name)
		if fs.Lookup(fl.Name) != nil {
			return
		}
		newflag := pflag.PFlagFromGoFlag(fl)
		newflag.Hidden = true
		fs.AddFlag(newflag)
	})
}

// normalize replaces underscores with hyphens
func normalize(s string) string {
	return strings.ReplaceAll(s, "_", "-")
}
