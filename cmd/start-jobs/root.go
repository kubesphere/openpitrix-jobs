package main

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"io"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
	"kubesphere.io/openpitrix-jobs/cmd/start-jobs/types"
	"os/exec"
	"strings"
)

const (
	openpitrixNamespace = "openpitrix-system"
	openpitrixDeploy    = "openpitrix-hyperpitrix-deployment"
	importAppPath       = "/usr/local/bin/import-app"
	dumpAllPath         = "/usr/local/bin/dump-all"
)

var kubeconfig string
var k8sClient *kubernetes.Clientset

func newRootCmd(out io.Writer, args []string) (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:   "start-app",
		Short: "parse kubesphere-config then start dump-all and import-app",
		Run: func(cmd *cobra.Command, args []string) {
			config, err := types.TryLoadFromDisk()
			if err != nil {
				klog.Fatalf("load config failed, error: %s", err)
			}

			_, err = k8sClient.AppsV1().Deployments(openpitrixNamespace).Get(context.TODO(), openpitrixDeploy, metav1.GetOptions{})
			if err != nil {
				if !apierrors.IsNotFound(err) {
					klog.Fatalf("get openpitrix deploy failed, error: %s", err)
				} else {
					// import app
					klog.Infof("start to import app")

					cmd := exec.Cmd{
						Path:   importAppPath,
						Stdout: out,
						Stderr: out,
					}

					cmd.Args = make([]string, 0, 10)
					cmd.Args = append(cmd.Args, importAppPath, "import")
					cmd.Args = appendS3Param(cmd.Args, config)

					err = cmd.Run()

					if err != nil {
						klog.Fatalf("run import app failed, error: %s", err)
					}
					klog.Infof("import app ends")
				}
			} else {
				// openpitrix-hyperpitrix-deployment deploy exists, convert legacy app to k8s crd
				// 1. dump legacy data
				hostAndPort := strings.Split(config.MySql.Host, ":")
				klog.Infof("start to dump legacy data")
				cmd := exec.Cmd{
					Path:   dumpAllPath,
					Stdout: out,
					Stderr: out,
					Env: []string{
						"OPENPITRIX_GRPC_SHOW_ERROR_CAUSE=true",
						"OPENPITRIX_LOG_LEVEL=debug",
						"OPENPITRIX_ETCD_ENDPOINTS=etcd.kubesphere-system.svc:2379",
						fmt.Sprintf("OPENPITRIX_MYSQL_HOST=%s", hostAndPort[0]),
						fmt.Sprintf("OPENPITRIX_ATTACHMENT_ENDPOINT=%s", config.S3Options.Endpoint),
						"OPENPITRIX_ATTACHMENT_BUCKET_NAME=openpitrix-attachment",
						fmt.Sprintf("OPENPITRIX_MYSQL_PASSWORD=%s", config.MySql.Password),
					},
				}
				cmd.Args = make([]string, 0, 10)

				err = cmd.Run()
				if err != nil {
					klog.Fatalf("run import app failed, error: %s", err)
				}
				klog.Infof("dump legacy data end")

				// 2. convert legacy data to k8s crd
				klog.Infof("start to convert legacy data to k8s crd")
				cmd = exec.Cmd{
					Path:   importAppPath,
					Stdout: out,
					Stderr: out,
				}
				cmd.Args = make([]string, 0, 10)
				cmd.Args = append(cmd.Args, importAppPath, "convert")
				cmd.Args = appendS3Param(cmd.Args, config)
				if config.MultiClusterOptions != nil {
					cmd.Args = append(cmd.Args, fmt.Sprintf("--multi-cluster-enable=%t", config.MultiClusterOptions.Enable))
				}

				klog.Infof("run cmd: %s", cmd.String())
				err = cmd.Run()

				if err != nil {
					klog.Fatalf("convert legacy data to k8s crd failed, error: %s", err)
				}
				klog.Infof("convert legacy data to k8s crd end")
			}

		},
	}

	cobra.OnInitialize(func() {
		config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			klog.Fatalf("load kubeconfig failed, error: %s", err)
		}
		k8sClient = kubernetes.NewForConfigOrDie(config)
	})

	klog.SetOutput(out)
	flags := cmd.PersistentFlags()

	addKlogFlags(flags)
	flags.StringVar(&kubeconfig, "kubeconfig", "", "path to the kubeconfig file")

	flags.Parse(args)

	return cmd, nil
}

func appendS3Param(args []string, config *types.Config) []string {
	args = append(args,
		fmt.Sprintf("--s3-endpoint=%s", config.OpenPitrixOptions.S3Options.Endpoint),
		fmt.Sprintf("--s3-access-key-id=%s", config.OpenPitrixOptions.S3Options.AccessKeyID),
		fmt.Sprintf("--s3-bucket=%s", config.OpenPitrixOptions.S3Options.Bucket),
		fmt.Sprintf("--s3-disable-SSL=%t", config.OpenPitrixOptions.S3Options.DisableSSL),
		fmt.Sprintf("--s3-force-path-style=%t", config.OpenPitrixOptions.S3Options.ForcePathStyle),
		fmt.Sprintf("--s3-secret-access-key=%s", config.OpenPitrixOptions.S3Options.SecretAccessKey))

	if config.OpenPitrixOptions.S3Options.Region != "" {
		args = append(args, fmt.Sprintf("--s3-region=%s", config.OpenPitrixOptions.S3Options.Region))
	}
	if config.OpenPitrixOptions.S3Options.SessionToken != "" {
		args = append(args, fmt.Sprintf("--s3-session-token=%s", config.OpenPitrixOptions.S3Options.SessionToken))
	}

	return args
}
