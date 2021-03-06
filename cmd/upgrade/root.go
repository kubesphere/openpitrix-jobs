package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	services3 "github.com/aws/aws-sdk-go/service/s3"
	"io"
	"kubesphere.io/openpitrix-jobs/pkg/s3"
	"kubesphere.io/openpitrix-jobs/pkg/types"
	"kubesphere.io/openpitrix-jobs/pkg/utils"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
)

const (
	openpitrixNamespace = "openpitrix-system"
	openpitrixDeploy    = "openpitrix-hyperpitrix-deployment"
	importAppPath       = "/usr/local/bin/import-app"
	dumpAllPath         = "/usr/local/bin/dump-all"
)

var kubeconfig string
var master string
var k8sClient *kubernetes.Clientset

func newRootCmd(out io.Writer, args []string) (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:   "upgrade",
		Short: "parse kubesphere-config then start dump-all and convert app",
		Run: func(cmd *cobra.Command, args []string) {
			utils.DumpConfig()
			config, err := types.TryLoadFromDisk()
			if err != nil {
				klog.Fatalf("load config failed, error: %s", err)
			}

			if config.OpenPitrixOptions == nil || config.OpenPitrixOptions.S3Options == nil || config.OpenPitrixOptions.S3Options.Endpoint == "" {
				klog.Infof("openpitrix s3 config is empty")
				return
			}

			err = createBucket(config.OpenPitrixOptions.S3Options)
			if err != nil {
				klog.Fatalf("create bucket error: %s", err)
			}

			if config.MySql == nil {
				klog.Warningf("mysql is empty, use the default config")
				config.MySql = &types.MySqlOptions{
					Host:     "mysql.kubesphere-system.svc:3306",
					Password: "password",
					Username: "root",
				}
			}

			// 1. dump legacy data
			hostAndPort := strings.Split(config.MySql.Host, ":")
			klog.Infof("start to dump legacy data")
			runCmd := exec.Cmd{
				Path:   dumpAllPath,
				Stdout: out,
				Stderr: out,
				Env: []string{
					"OPENPITRIX_GRPC_SHOW_ERROR_CAUSE=true",
					"OPENPITRIX_LOG_LEVEL=debug",
					"OPENPITRIX_ETCD_ENDPOINTS=etcd.kubesphere-system.svc:2379",
					fmt.Sprintf("OPENPITRIX_MYSQL_HOST=%s", hostAndPort[0]),
					fmt.Sprintf("OPENPITRIX_ATTACHMENT_ENDPOINT=%s", config.OpenPitrixOptions.S3Options.Endpoint),
					"OPENPITRIX_ATTACHMENT_BUCKET_NAME=openpitrix-attachment",
					fmt.Sprintf("OPENPITRIX_MYSQL_PASSWORD=%s", config.MySql.Password),
				},
			}
			runCmd.Args = make([]string, 0, 10)

			err = runCmd.Run()
			if err != nil {
				klog.Fatalf("run import app failed, error: %s", err)
			}
			klog.Infof("dump legacy data end")

			// 2. convert legacy data to k8s crd
			klog.Infof("start to convert legacy data to k8s crd")
			runCmd = exec.Cmd{
				Path:   importAppPath,
				Stdout: out,
				Stderr: out,
			}
			runCmd.Args = make([]string, 0, 10)
			runCmd.Args = append(runCmd.Args, importAppPath, "convert")

			if config.MultiClusterOptions != nil {
				runCmd.Args = append(runCmd.Args, fmt.Sprintf("--multi-cluster-enable=%t", config.MultiClusterOptions.Enable))
			}
			if master != "" {
				runCmd.Args = append(runCmd.Args, fmt.Sprintf("--master=%s", master))
			}

			klog.Infof("run cmd: %s", runCmd.String())
			err = runCmd.Run()

			if err != nil {
				klog.Fatalf("convert legacy data to k8s crd failed, error: %s", err)
			}
			klog.Infof("convert legacy data to k8s crd end")

		},
	}

	cobra.OnInitialize(func() {
		config, err := clientcmd.BuildConfigFromFlags(master, kubeconfig)
		if err != nil {
			klog.Fatalf("load kubeconfig failed, error: %s", err)
		}
		k8sClient = kubernetes.NewForConfigOrDie(config)
	})

	klog.SetOutput(out)
	flags := cmd.PersistentFlags()

	addKlogFlags(flags)
	flags.StringVar(&kubeconfig, "kubeconfig", "", "path to the kubeconfig file")
	flags.StringVar(&master, "master", "", "kubernetes master")

	flags.Parse(args)

	return cmd, nil
}

func createBucket(s3config *s3.Options) error {
	cred := credentials.NewStaticCredentials(s3config.AccessKeyID, s3config.SecretAccessKey, s3config.SessionToken)

	config := aws.Config{
		Region:           aws.String(s3config.Region),
		Endpoint:         aws.String(s3config.Endpoint),
		DisableSSL:       aws.Bool(s3config.DisableSSL),
		S3ForcePathStyle: aws.Bool(s3config.ForcePathStyle),
		Credentials:      cred,
	}

	s, err := session.NewSession(&config)
	if err != nil {
		klog.Error(err)
		return err
	}

	svc := services3.New(s, &config)
	_, err = svc.CreateBucket(&services3.CreateBucketInput{Bucket: aws.String(s3config.Bucket)})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok && aerr.Code() != services3.ErrCodeBucketAlreadyOwnedByYou {
			klog.Error(err)
			return err
		}
	}

	return nil
}
