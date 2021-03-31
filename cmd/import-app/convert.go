package main

import (
	"context"
	"encoding/json"
	"github.com/spf13/cobra"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/repo"
	"io/ioutil"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
	"kubesphere.io/openpitrix-jobs/pkg/apis/application/v1alpha1"
	clusterv1alpha1 "kubesphere.io/openpitrix-jobs/pkg/apis/cluster/v1alpha1"
	applicationv1alpha1 "kubesphere.io/openpitrix-jobs/pkg/client/clientset/versioned/typed/application/v1alpha1"
	versionedclusterv1alpha1 "kubesphere.io/openpitrix-jobs/pkg/client/clientset/versioned/typed/cluster/v1alpha1"
	"kubesphere.io/openpitrix-jobs/pkg/constants"
	"kubesphere.io/openpitrix-jobs/pkg/idutils"
	legacy_op "kubesphere.io/openpitrix-jobs/pkg/legacy-op"
	"kubesphere.io/openpitrix-jobs/pkg/s3"
	"kubesphere.io/openpitrix-jobs/pkg/types"
	"os"
	"path"
	"sigs.k8s.io/yaml"
	"strings"
	"time"
)

var (
	multiClusterEnabled   = false
	legacyDir             string
	nullChar              = "\u0000"
	legacyCreator         = "system"
	newCreator            = "admin"
	appFile               = "apps.json"
	appVersionFile        = "app_versions.json"
	repoFile              = "repos.json"
	clusterFile           = "clusters.json"
	categoriesFile        = "categories.json"
	categoryResourcesFile = "category_resources.json"
	repoLabelsFile        = "repo_labels.json"
)

func newConvertCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "convert",
		Short: "convert legacy OpenPitrix data to crd data",
		Run: func(cmd *cobra.Command, args []string) {
			var err error
			s3Client, err := s3.NewS3Client(s3Options)
			if err != nil {
				klog.Fatalf("connect s3 failed, error: %s", err)
			}

			cw := &ConvertWorkflow{
				s3Client:  s3Client,
				appClient: versionedClient.ApplicationV1alpha1(),
				//clusterClient: versionedClient.
				k8sClient: k8sClient,
				legacyDir: legacyDir,
			}
			cw.multiClusterEnabled = multiClusterEnabled

			// 1. load legacy data
			err = cw.LoadAllData()
			ctx := context.Background()

			// 2. create category crd
			err = cw.CreateCategories(ctx)
			if err != nil {
				klog.Fatalf("create categories failed, error: %s", err)
			}

			// 3. create app and appVersion crd
			err = cw.CreateAppsAndVersions(ctx)
			if err != nil {
				klog.Fatalf("create app and version failed, error: %s", err)
			}

			// 4. create repo crd
			err = cw.CreateRepos(ctx)
			if err != nil {
				klog.Fatalf("create repo failed, error: %s", err)
			}

			// 5. create release crd
			err = cw.CreateReleases(ctx)
			if err != nil {
				klog.Fatalf("create releases failed, error: %s", err)
			}

			if err != nil {
				klog.Fatalf("load data failed, error: %s", err)
			}
		},
	}

	f := cmd.Flags()

	f.StringVar(&legacyDir, "legacy-dir",
		"/tmp/op-dump",
		"dir of legacy openpitrix data")

	f.BoolVar(&multiClusterEnabled, "multi-cluster-enable", false, "multi cluster enable or not")

	return cmd
}

type ConvertWorkflow struct {
	multiClusterEnabled bool
	k8sClient           kubernetes.Interface
	clusterClient       versionedclusterv1alpha1.ClusterInterface
	appClient           applicationv1alpha1.ApplicationV1alpha1Interface
	s3Client            s3.Interface

	legacyDir string

	apps             []legacy_op.OpenpitrixApp
	appVersions      []legacy_op.OpenpitrixAppVersion
	categories       []legacy_op.OpenpitrixCategory
	categoryResource []legacy_op.CategoryResource
	clusters         []legacy_op.OpenpitrixCluster
	repos            []legacy_op.OpenpitrixRepo
	repoLabels       []legacy_op.RepoLabel

	versionIdToVersion map[string]*legacy_op.OpenpitrixAppVersion
	appIdToVersions    map[string][]*legacy_op.OpenpitrixAppVersion
	// appIdToApp map app id to the instance of app
	appIdToApp map[string]*legacy_op.OpenpitrixApp
	// repoIdToAppIds map repo id of the repo to the app id list
	repoIdToAppIds map[string][]string
	appIdToCtgId   map[string]string
	// repoWorkspace map repo id to workspace
	repoWorkspace map[string]string
}

func (cw *ConvertWorkflow) LoadAllData() error {
	err := cw.loadApps()
	if err != nil {
		klog.Errorf("load apps failed, error: %s", err)
		return err
	} else {
		klog.Infof("load apps success")
	}

	err = cw.loadAppVersion()
	if err != nil {
		klog.Errorf("load app version failed, error: %s", err)
		return err
	} else {
		klog.Infof("load app version success")
	}

	err = cw.loadCategories()
	if err != nil {
		klog.Errorf("load categories failed, error: %s", err)
		return err
	} else {
		klog.Infof("load categories success")
	}

	err = cw.loadCategoryResource()
	if err != nil {
		klog.Errorf("load resource categories failed, error: %s", err)
		return err
	} else {
		klog.Infof("load resource categories success")
	}

	err = cw.loadRepos()
	if err != nil {
		klog.Errorf("load repos failed, error: %s", err)
		return err
	} else {
		klog.Infof("load repos success")
	}

	err = cw.loadRepoLabels()
	if err != nil {
		klog.Errorf("load repo labels failed, error: %s", err)
		return err
	} else {
		klog.Infof("load repo labels success")
	}

	err = cw.loadClusters()
	if err != nil {
		klog.Errorf("load clusters failed, error: %s", err)
		return err
	} else {
		klog.Infof("load clusters success")
	}

	cw.buildRelation()

	return nil
}

type legacyCred struct {
	AccessKeyId string `json:"access_key_id"`
	SecretKeyId string `json:"secret_key_id"`
}

func buildCredential(cred string) (*v1alpha1.HelmRepoCredential, error) {
	oldCred := &legacyCred{}
	err := json.Unmarshal([]byte(cred), oldCred)
	if err != nil {
		return nil, err
	}

	newCred := &v1alpha1.HelmRepoCredential{
		S3Config: v1alpha1.S3Config{
			AccessKeyID:     oldCred.AccessKeyId,
			SecretAccessKey: oldCred.SecretKeyId,
		},
	}
	return newCred, nil

}

// create apps and appVersions
func (cw *ConvertWorkflow) CreateRepos(ctx context.Context) error {
	klog.Infof("start create repos")
	var err error

	for _, oldRepo := range cw.repos {
		if oldRepo.RepoID == "repo-vmbased" || oldRepo.RepoID == "repo-helm" {
			klog.Infof("skip repo %s", oldRepo.RepoID)
			continue
		}

		cred := &v1alpha1.HelmRepoCredential{}

		// "{}" is an empty credential
		if len(oldRepo.Credential) > 2 {
			cred, err = buildCredential(oldRepo.Credential)
			if err != nil {
				return err
			}
		}
		lowerRepoId := strings.ToLower(oldRepo.RepoID)

		newRepo := &v1alpha1.HelmRepo{
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{
					constants.CreatorAnnotationKey: convertCreator(oldRepo.Owner),
				},
				Labels: map[string]string{
					constants.WorkspaceLabelKey: cw.repoWorkspace[oldRepo.RepoID],
				},
				Name: lowerRepoId,
			},
			Spec: v1alpha1.HelmRepoSpec{
				Name:        oldRepo.Name,
				Url:         oldRepo.URL,
				Description: oldRepo.Description,
				Credential:  *cred,
				Version:     1,
			},
		}

		newRepo, err = cw.appClient.HelmRepos().Create(ctx, newRepo, metav1.CreateOptions{})
		if err != nil {
			if apierrors.IsAlreadyExists(err) {
				klog.Infof("repo %s exists, id: %s", oldRepo.Name, oldRepo.RepoID)
			} else {
				klog.Errorf("create repo %s failed, id: %s error: %s", oldRepo.Name, oldRepo.RepoID, err)
				return err
			}
		}

		err = cw.updateRepoStatus(ctx, &oldRepo)
		if err != nil {
			return err
		}
	}

	klog.Infof("create repos end")
	return nil
}

func (cw *ConvertWorkflow) updateRepoStatus(ctx context.Context, oldRepo *legacy_op.OpenpitrixRepo) error {

	newRepoId := strings.ToLower(oldRepo.RepoID)
	retry := 5
	for i := 0; i < retry; i++ {
		newRepo, err := cw.appClient.HelmRepos().Get(ctx, newRepoId, metav1.GetOptions{})
		if err != nil {
			return err
		}

		savedIndex := &types.SavedIndex{
			Applications: map[string]*types.Application{},
		}
		appIds := cw.repoIdToAppIds[oldRepo.RepoID]
		for _, appId := range appIds {
			oldApp := cw.appIdToApp[appId]
			app := types.Application{
				Name:          oldApp.Name,
				ApplicationId: strings.ToLower(appId),
				Description:   oldApp.Description,
				Icon:          oldApp.Icon,
			}
			chartVersions := make([]*types.ChartVersion, 0, 10)
			for _, oldAppVer := range cw.appIdToVersions[appId] {
				chartName, chatVersion, chatAppVersion := cw.getChartInfo(oldAppVer)
				ver := types.ChartVersion{
					ApplicationVersionId: strings.ToLower(oldAppVer.VersionID),
					ChartVersion: repo.ChartVersion{
						Metadata: &chart.Metadata{
							Name:        chartName,
							Icon:        oldAppVer.Icon,
							Description: oldAppVer.Description,
							Version:     chatVersion,
							APIVersion:  chatAppVersion,
						},
						// todo parse url
						URLs: []string{oldAppVer.PackageName},
					},
				}
				chartVersions = append(chartVersions, &ver)
			}
			app.Charts = chartVersions

			savedIndex.Applications[oldApp.Name] = &app
		}

		data, err := savedIndex.Bytes()
		if err != nil {
			return err
		}

		newRepo.Status.Data = string(data)
		newRepo.Status.Version = newRepo.Spec.Version
		newRepo.Status.State = v1alpha1.RepoStateSuccessful
		now := metav1.NewTime(time.Now())
		state := append([]v1alpha1.HelmRepoSyncState{{
			State:    v1alpha1.RepoStateSuccessful,
			SyncTime: &now,
		}}, newRepo.Status.SyncState...)

		newRepo.Status.LastUpdateTime = &now
		if len(state) > v1alpha1.HelmRepoSyncStateLen {
			state = state[0:v1alpha1.HelmRepoSyncStateLen]
		}
		newRepo.Status.SyncState = state

		newRepo, err = cw.appClient.HelmRepos().UpdateStatus(ctx, newRepo, metav1.UpdateOptions{})
		if err != nil {
			if apierrors.IsConflict(err) {
				continue
			}
			return err
		}
	}

	return nil
}
func (cw *ConvertWorkflow) uploadAttachments(oldApp *legacy_op.OpenpitrixApp) (icon string, attachments []string) {
	if oldApp.Icon != "" && strings.HasPrefix(oldApp.Icon, v1alpha1.HelmAttachmentPrefix) {
		icon = ""
		f, err := os.Open(path.Join(cw.legacyDir, oldApp.Icon, "raw"))
		if err != nil {
			klog.Warningf("open icon %s file failed, error: %s", oldApp.AppID, err)
		} else {
			icon = idutils.GetUuid36(v1alpha1.HelmAttachmentPrefix)
			err = cw.s3Client.Upload(icon, icon, f)
		}
	}

	attachments = make([]string, 0, 4)
	if len(oldApp.Screenshots) != 0 {
		parts := strings.Split(oldApp.Screenshots, ",")
		for _, oldId := range parts {
			f, err := os.Open(path.Join(cw.legacyDir, oldId, "raw"))
			if err != nil {
				klog.Warningf("open screenshot %s file failed, error: %s", oldApp.AppID, err)
			} else {
				newId := idutils.GetUuid36(v1alpha1.HelmAttachmentPrefix)
				err = cw.s3Client.Upload(newId, newId, f)
				if err != nil {
					klog.Errorf("upload screenshot %s failed, error: %s", oldApp.AppID, err)
				} else {
					attachments = append(attachments, newId)
				}
			}
		}
	}

	return
}

// createApp create a helm application crd
// if the crd exists, get and return it
func (cw *ConvertWorkflow) createApp(ctx context.Context, oldApp *legacy_op.OpenpitrixApp) (*v1alpha1.HelmApplication, error) {
	labels := make(map[string]string, 1)
	if oldApp.Isv == "" || oldApp.Isv == nullChar {
		// builtin app
		labels[builtinKey] = "true"
	} else {
		labels[constants.WorkspaceLabelKey] = oldApp.Isv
	}

	// add ctg to app
	if ctgId, exists := cw.appIdToCtgId[oldApp.AppID]; exists {
		labels[constants.CategoryIdLabelKey] = strings.ToLower(ctgId)
	}

	// upload icon and attachments if exists
	icon, attachments := cw.uploadAttachments(oldApp)

	appId := strings.ToLower(oldApp.AppID)
	newApp := &v1alpha1.HelmApplication{
		ObjectMeta: metav1.ObjectMeta{
			Name:   appId,
			Labels: labels,
			Annotations: map[string]string{
				constants.CreatorAnnotationKey: convertCreator(oldApp.Owner),
			},
		},
		Spec: v1alpha1.HelmApplicationSpec{
			Abstraction: oldApp.Abstraction,
			Name:        oldApp.Name,
			Description: oldApp.Description,
			Icon:        icon,
			AppHome:     oldApp.Home,
			Attachments: attachments,
		},
	}

	newApp, err := cw.appClient.HelmApplications().Create(ctx, newApp, metav1.CreateOptions{})

	if err != nil {
		if !apierrors.IsAlreadyExists(err) {
			klog.Errorf("create app %s failed, error: %s", oldApp.AppID, err)
			return nil, err
		} else {
			// if app exists, get the latest helm application
			newApp, err = cw.appClient.HelmApplications().Get(ctx, appId, metav1.GetOptions{})
		}
	}
	return newApp, err
}

func (cw *ConvertWorkflow) updateAppVerStatus(ctx context.Context, newVerName string, oldAppVer *legacy_op.OpenpitrixAppVersion) error {

	retry := 5
	for i := 0; i < retry; i++ {
		newAppVer, err := cw.appClient.HelmApplicationVersions().Get(ctx, newVerName, metav1.GetOptions{})
		if err != nil {
			return err
		}

		if newAppVer.Status.State != oldAppVer.Status {
			newAppVer.Status.State = oldAppVer.Status
			newAppVer.Status.Audit = append([]v1alpha1.Audit{{
				State:    oldAppVer.Status,
				Time:     metav1.NewTime(time.Now()),
				Operator: convertCreator(oldAppVer.Owner),
			}}, newAppVer.Status.Audit...)
		}

		_, err = cw.appClient.HelmApplicationVersions().UpdateStatus(ctx, newAppVer, metav1.UpdateOptions{})
		if err != nil {
			if !apierrors.IsConflict(err) {
				return err
			} else {
				klog.Infof("update app version %s conflict, retry: %d", newVerName, i)
				time.Sleep(1 * time.Second)
			}
		} else {
			break
		}
	}

	return nil
}

func parseChartVersionName(name string) (version, appVersion string) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", ""
	}

	parts := strings.Split(name, "[")
	if len(parts) == 1 {
		return parts[0], ""
	}

	version = strings.TrimSpace(parts[0])

	appVersion = strings.Trim(parts[1], "]")
	appVersion = strings.TrimSpace(appVersion)
	return
}

func (cw *ConvertWorkflow) getChartInfo(oldAppVer *legacy_op.OpenpitrixAppVersion) (name, version, appVersion string) {
	if strings.HasPrefix(oldAppVer.PackageName, "att-") {
		chrt, err := cw.loadVersionAttachment(oldAppVer.PackageName)
		if err == nil {
			return chrt.Name(), chrt.Metadata.Version, chrt.AppVersion()
		}
	}

	parts := strings.Split(oldAppVer.PackageName, "/")
	if len(parts) > 0 {
		ver, appVer := parseChartVersionName(oldAppVer.Name)
		namePart := parts[len(parts)-1]
		nameInd := strings.Index(namePart, ver)
		if nameInd > 0 {
			return namePart[0:nameInd], ver, appVer
		}
	}

	return "", "", ""
}

func (cw *ConvertWorkflow) createAppVer(ctx context.Context, app *v1alpha1.HelmApplication, oldAppVer *legacy_op.OpenpitrixAppVersion) (*v1alpha1.HelmApplicationVersion, error) {
	appVerId := strings.ToLower(oldAppVer.VersionID)

	chrt, err := cw.loadVersionAttachment(oldAppVer.PackageName)
	if err != nil {
		return nil, err
	}

	name, err := chartutil.Save(chrt, "/tmp")
	if err != nil {
		klog.Errorf("save chart %s failed, error: %s", chrt.Name(), err)
		return nil, err
	}

	file, err := os.Open(name)
	if err != nil {
		return nil, err
	}

	err = cw.s3Client.Upload(path.Join(app.GetWorkspace(), appVerId), appVerId, file)
	if err != nil {
		klog.Errorf("upload chart to s3 failed, error: %s", err)
		return nil, err
	}
	file.Close()

	newAppVer := &v1alpha1.HelmApplicationVersion{
		ObjectMeta: metav1.ObjectMeta{
			Name: appVerId,
			Labels: map[string]string{
				constants.WorkspaceLabelKey:          app.GetWorkspace(),
				constants.ChartApplicationIdLabelKey: app.Name,
			},
			Annotations: map[string]string{
				constants.CreatorAnnotationKey: convertCreator(oldAppVer.Owner),
			},
			OwnerReferences: []metav1.OwnerReference{
				{
					UID:        app.UID,
					Name:       app.Name,
					APIVersion: v1alpha1.SchemeGroupVersion.String(),
					Kind:       v1alpha1.ResourceKindHelmApplication,
				},
			},
		},
		Spec: v1alpha1.HelmApplicationVersionSpec{
			Metadata: &v1alpha1.Metadata{
				Name:       chrt.Name(),
				Version:    chrt.Metadata.Version,
				AppVersion: chrt.Metadata.AppVersion,
			},
			DataKey: appVerId,
		},
	}

	newAppVer, err = cw.appClient.HelmApplicationVersions().Create(ctx, newAppVer, metav1.CreateOptions{})
	if err != nil {
		if !apierrors.IsAlreadyExists(err) {
			klog.Errorf("create app version %s/%s failed, error: %s", oldAppVer.AppID, oldAppVer.VersionID, err)
			return nil, err
		} else {
			newAppVer, err = cw.appClient.HelmApplicationVersions().Get(ctx, appVerId, metav1.GetOptions{})
		}
	}

	// update status
	return newAppVer, err
}

func (cw *ConvertWorkflow) loadVersionAttachment(attId string) (*chart.Chart, error) {
	chrt, err := loader.LoadDir(path.Join(cw.legacyDir, attId))
	return chrt, err
}

// create apps and appVersions
func (cw *ConvertWorkflow) CreateAppsAndVersions(ctx context.Context) error {
	klog.Infof("start create apps")

	for _, oldApp := range cw.apps {
		if oldApp.RepoID == "" {
			klog.Infof("start to create app %s", oldApp.AppID)
			newApp, err := cw.createApp(ctx, &oldApp)
			if err != nil {
				klog.Errorf("create app %s failed, error: %s", oldApp.AppID, err)
				return err
			} else {
				klog.Infof("create app version %s success", oldApp.AppID)
			}

			for _, oldVer := range cw.appIdToVersions[oldApp.AppID] {
				klog.Infof("start to create app version %s/%s", oldApp.RepoID, oldVer.VersionID)
				newAppVer, err := cw.createAppVer(ctx, newApp, oldVer)
				if err != nil {
					klog.Errorf("create app version %s/%s failed, error: %s", oldApp.AppID, oldVer.VersionID, err)
					//return err
					continue
				} else {
					klog.Infof("create app version %s/%s success", oldApp.AppID, oldVer.VersionID)
				}

				err = cw.updateAppVerStatus(ctx, newAppVer.Name, oldVer)
				if err != nil {
					klog.Errorf("update app version %s/%s failed, error: %s", oldApp.AppID, oldVer.VersionID, err)
					continue
				}

				klog.Infof("create app version %s/%s success", oldApp.RepoID, oldVer.VersionID)
			}
		}
	}

	klog.Infof("create apps end")
	return nil
}

func convertCreator(oldCreator string) string {
	var creator string
	if oldCreator == legacyCreator {
		creator = newCreator
	} else {
		creator = oldCreator
	}
	return creator
}

func isHostCluster(cluster *clusterv1alpha1.Cluster) bool {
	if _, ok := cluster.Labels[clusterv1alpha1.HostCluster]; ok {
		return true
	}
	return false
}

// getWorkspace get the workspace to which the ns belong
func (cw *ConvertWorkflow) getWorkspace(ctx context.Context, cluster, ns string) (workspace string, err error) {
	if cw.multiClusterEnabled {
		clusterInstance, err := cw.clusterClient.Get(ctx, cluster, metav1.GetOptions{})
		if err != nil {
			return "", nil
		}

		var client kubernetes.Interface
		if !isHostCluster(clusterInstance) {
			kubeConfig := clusterInstance.Spec.Connection.KubeConfig
			config, err := clientcmd.RESTConfigFromKubeConfig(kubeConfig)
			if err != nil {
				return "", err
			}

			client, err = kubernetes.NewForConfig(config)
			if err != nil {
				return "", nil
			}
		} else {
			// get ns from host cluster
			client = cw.k8sClient
		}

		namespace, err := client.CoreV1().Namespaces().Get(ctx, ns, metav1.GetOptions{})
		if err != nil {
			return "", err
		}
		return namespace.Labels[constants.WorkspaceLabelKey], nil
	} else {
		// just one cluster, get info from current k8s cluster
		namespace, err := cw.k8sClient.CoreV1().Namespaces().Get(ctx, ns, metav1.GetOptions{})
		if err != nil {
			return "", err
		}
		return namespace.Labels[constants.WorkspaceLabelKey], nil
	}
}

func (cw *ConvertWorkflow) CreateReleases(ctx context.Context) error {
	klog.Infof("start create releases")
	for ind := range cw.clusters {
		oldCluster := &cw.clusters[ind]
		klog.Infof("start create release %s, id: %s", oldCluster.Name, oldCluster.ClusterID)

		creator := convertCreator(oldCluster.Owner)
		labels := map[string]string{
			constants.ChartApplicationIdLabelKey:        strings.ToLower(oldCluster.AppID),
			constants.ChartApplicationVersionIdLabelKey: strings.ToLower(oldCluster.VersionID),
			constants.NamespaceLabelKey:                 oldCluster.Zone,
		}

		app := cw.appIdToApp[oldCluster.AppID]
		if app.RepoID != "" {
			labels[constants.ChartRepoIdLabelKey] = strings.ToLower(app.RepoID)
		}

		if cw.multiClusterEnabled {
			labels[constants.ClusterNameLabelKey] = oldCluster.RuntimeID
		}

		ws, err := cw.getWorkspace(ctx, oldCluster.RuntimeID, oldCluster.Zone)
		if err != nil {
			continue
		}
		labels[constants.WorkspaceLabelKey] = ws

		version := cw.versionIdToVersion[oldCluster.VersionID]
		if version == nil {
			klog.Warningf("app version %s not exists, cluster id: %s", oldCluster.VersionID, oldCluster.ClusterID)
			continue
		}
		chartName, chartVer, chartAppVer := cw.getChartInfo(version)

		values, err := yaml.JSONToYAML([]byte(oldCluster.Env))
		if err != nil {
			klog.Warningf("convert env format failed, id: %s, error: %s", oldCluster.ClusterID, err)
			continue
		}

		release := &v1alpha1.HelmRelease{
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{
					constants.CreatorAnnotationKey: creator,
				},
				Labels: labels,
				Name:   strings.ToLower(oldCluster.ClusterID),
			},
			Spec: v1alpha1.HelmReleaseSpec{
				Name:                 oldCluster.Name,
				Description:          oldCluster.Description,
				ChartName:            chartName,
				ChartVersion:         chartVer,
				ChartAppVersion:      chartAppVer,
				ApplicationId:        strings.ToLower(oldCluster.AppID),
				ApplicationVersionId: strings.ToLower(oldCluster.VersionID),
				RepoId:               strings.ToLower(app.RepoID),
				Version:              1,
				Values:               values,
			},
		}

		release, err = cw.appClient.HelmReleases().Create(ctx, release, metav1.CreateOptions{})
		if err != nil {
			if apierrors.IsAlreadyExists(err) {
				klog.Infof("release %s exists, id: %s", oldCluster.Name, oldCluster.ClusterID)
			} else {
				klog.Errorf("create release %s failed, id: %s error: %s", oldCluster.Name, oldCluster.ClusterID, err)
				return err
			}
		}
	}

	klog.Infof("create releases end")
	return nil
}

func (cw *ConvertWorkflow) CreateCategories(ctx context.Context) error {
	klog.Infof("start create categories")
	for ind := range cw.categories {
		oldCtg := &cw.categories[ind]
		klog.Infof("start create categories %s, id: %s", oldCtg.Name, oldCtg.CategoryID)

		creator := convertCreator(oldCtg.Owner)

		ctg := &v1alpha1.HelmCategory{
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{
					constants.CreatorAnnotationKey: creator,
				},
				Name: strings.ToLower(oldCtg.CategoryID),
			},
			Spec: v1alpha1.HelmCategorySpec{
				Name:        oldCtg.Name,
				Description: oldCtg.Description,
				Locale:      oldCtg.Locale,
			},
		}

		ctg, err := cw.appClient.HelmCategories().Create(ctx, ctg, metav1.CreateOptions{})
		if err != nil {
			if apierrors.IsAlreadyExists(err) {
				klog.Infof("category %s exists, id: %s", oldCtg.Name, oldCtg.CategoryID)
			} else {
				klog.Errorf("create category %s failed, id: %s error: %s", oldCtg.Name, oldCtg.CategoryID, err)
				return err
			}
		}
	}

	klog.Infof("create categories end")
	return nil
}

func (cw *ConvertWorkflow) loadApps() error {
	filePath := path.Join(cw.legacyDir, appFile)
	fileData, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}

	err = json.Unmarshal(fileData, &cw.apps)
	if err != nil {
		return err
	}

	return nil
}

func (cw *ConvertWorkflow) loadAppVersion() error {
	filePath := path.Join(cw.legacyDir, appVersionFile)
	fileData, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}

	err = json.Unmarshal(fileData, &cw.appVersions)
	if err != nil {
		return err
	}

	return nil
}

func (cw *ConvertWorkflow) loadCategoryResource() error {
	filePath := path.Join(cw.legacyDir, categoryResourcesFile)
	fileData, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}

	err = json.Unmarshal(fileData, &cw.categoryResource)
	if err != nil {
		return err
	}

	return nil

}

func (cw *ConvertWorkflow) loadCategories() error {
	filePath := path.Join(cw.legacyDir, categoriesFile)
	fileData, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}

	err = json.Unmarshal(fileData, &cw.categories)
	if err != nil {
		return err
	}

	return nil
}

func (cw *ConvertWorkflow) loadRepos() error {
	filePath := path.Join(cw.legacyDir, repoFile)
	fileData, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}

	err = json.Unmarshal(fileData, &cw.repos)
	if err != nil {
		return err
	}

	return nil
}

func (cw *ConvertWorkflow) loadRepoLabels() error {
	filePath := path.Join(cw.legacyDir, repoLabelsFile)
	fileData, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}

	err = json.Unmarshal(fileData, &cw.repoLabels)
	if err != nil {
		return err
	}

	return nil
}

func (cw *ConvertWorkflow) buildRelation() {
	appIdToVersions := make(map[string][]*legacy_op.OpenpitrixAppVersion, len(cw.apps))
	versionIdToVersion := make(map[string]*legacy_op.OpenpitrixAppVersion)

	for ind := range cw.appVersions {
		ver := &cw.appVersions[ind]
		if ver.Active == false && (ver.Status == v1alpha1.StateActive || ver.Status == v1alpha1.StateSuspended) {
			continue
		}

		versionIdToVersion[ver.VersionID] = ver
		if items, exists := appIdToVersions[ver.AppID]; !exists {
			appIdToVersions[ver.AppID] = []*legacy_op.OpenpitrixAppVersion{ver}
		} else {
			appIdToVersions[ver.AppID] = append(items, ver)
		}
	}
	cw.appIdToVersions = appIdToVersions
	cw.versionIdToVersion = versionIdToVersion

	appIdToApp := make(map[string]*legacy_op.OpenpitrixApp)
	repoIdToAppIds := make(map[string][]string, len(cw.repos))
	for ind, app := range cw.apps {
		if app.Active == false && (app.Status == v1alpha1.StateActive || app.Status == v1alpha1.StateSuspended) {
			continue
		}

		appIdToApp[app.AppID] = &cw.apps[ind]
		if app.RepoID != "" {
			if items, exists := repoIdToAppIds[app.RepoID]; !exists {
				repoIdToAppIds[app.RepoID] = []string{app.AppID}
			} else {
				repoIdToAppIds[app.RepoID] = append(items, app.AppID)
			}
		}
	}
	cw.appIdToApp = appIdToApp
	cw.repoIdToAppIds = repoIdToAppIds

	cw.appIdToCtgId = make(map[string]string)
	for _, ctg := range cw.categoryResource {
		if ctg.Status == "enabled" {
			cw.appIdToCtgId[ctg.ResourceId] = ctg.CategoryId
		}
	}

	repoWs := make(map[string]string)
	for _, attr := range cw.repoLabels {
		if attr.LabelKey == "workspace" {
			repoWs[attr.RepoId] = attr.LabelValue
		}
	}

	cw.repoWorkspace = repoWs
}

func (cw *ConvertWorkflow) loadClusters() error {
	filePath := path.Join(cw.legacyDir, clusterFile)
	fileData, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}

	err = json.Unmarshal(fileData, &cw.clusters)
	if err != nil {
		return err
	}

	return nil
}
