package main

import (
	"github.com/spf13/cobra"
	"io"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
	"kubesphere.io/openpitrix-jobs/pkg/client/clientset/versioned"
	"kubesphere.io/openpitrix-jobs/pkg/s3"
	"kubesphere.io/openpitrix-jobs/pkg/types"
	"kubesphere.io/openpitrix-jobs/pkg/utils"
	"os"
	"time"
)

var kubeconfig string
var master string
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
		var ksConfig *types.Config
		var err error

		// https://kubernetes.io/docs/tasks/configure-pod-container/configure-pod-configmap/#mounted-configmaps-are-updated-automatically
		// Mounted ConfigMaps are updated automatically
		retry := 100
		for i := 0; ; i++ {
			if i == retry {
				klog.Errorf("load openpitrix config failed")
				os.Exit(1)
			}
			utils.DumpConfig()
			ksConfig, err = types.TryLoadFromDisk()
			if err != nil {
				klog.Errorf("load config failed, error: %s", err)
			} else {
				if ksConfig.OpenPitrixOptions == nil {
					klog.Errorf("openpitrix config is empty, please wait a minute")
				} else if ksConfig.OpenPitrixOptions.S3Options == nil {
					klog.Errorf("openpitrix s3 config is empty, please wait a minute")
				} else {
					break
				}
			}

			time.Sleep(5 * time.Second)
		}

		s3Options = ksConfig.OpenPitrixOptions.S3Options

		config, err := clientcmd.BuildConfigFromFlags(master, kubeconfig)
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
	flags.StringVar(&master, "master", "", "kubernetes master")

	flags.Parse(args)
	cmd.AddCommand(
		newImportCmd(),
		newConvertCmd(),
	)

	return cmd, nil
}
