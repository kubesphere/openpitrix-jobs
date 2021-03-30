package main

import (
	"github.com/spf13/cobra"
	"io"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
	"kubesphere.io/openpitrix-jobs/pkg/client/clientset/versioned"
	"kubesphere.io/openpitrix-jobs/pkg/s3"
)

var kubeconfig string
var versionedClient *versioned.Clientset
var k8sClient *kubernetes.Clientset
var s3Options *s3.Options

func newRootCmd(out io.Writer, args []string) (*cobra.Command, error) {
	s3Options = s3.NewS3Options()
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
		versionedClient, err = versioned.NewForConfig(config)
		if err != nil {
			klog.Fatalf("build config failed, error: %s", err)
		}
		k8sClient = kubernetes.NewForConfigOrDie(config)
	})

	flags := cmd.PersistentFlags()

	addKlogFlags(flags)
	flags.StringVar(&kubeconfig, "kubeconfig", "", "path to the kubeconfig file")

	s3Options.AddFlags(flags, s3Options)
	flags.Parse(args)
	// Add subcommands
	cmd.AddCommand(
		newImportCmd(),
		newConvertCmd(),
	)

	return cmd, nil
}
