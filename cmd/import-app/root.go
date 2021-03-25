package main

import (
	"github.com/spf13/cobra"
	"io"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
	"kubesphere.io/openpitrix-jobs/pkg/client/clientset/versioned"
)

var kubeconfig string
var appClient *versioned.Clientset

func newRootCmd(out io.Writer, args []string) (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:          "import-app",
		Short:        "import builtin app into kubesphere",
		SilenceUsage: true,
	}

	cobra.OnInitialize(func() {
		config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			klog.Fatalf("load kubeconfig failed, error: %s", err)
		}
		appClient, err = versioned.NewForConfig(config)
		if err != nil {
			klog.Fatalf("build config failed, error: %s", err)
		}
	})

	flags := cmd.PersistentFlags()

	addKlogFlags(flags)
	flags.StringVar(&kubeconfig, "kubeconfig", "", "path to the kubeconfig file")

	flags.Parse(args)

	// Add subcommands
	cmd.AddCommand(
		newImportCmd(),
	)

	return cmd, nil
}
