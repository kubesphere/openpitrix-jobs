package legacy_op

import (
	"github.com/go-openapi/strfmt"
	"time"
)

type RepoLabel struct {
	RepoLabelId string
	RepoId      string
	LabelKey    string
	LabelValue  string

	CreateTime time.Time
}

type CategoryResource struct {
	CategoryId string `json:"CategoryId"`
	ResourceId string `json:"ResourceId"`
	// enabled or disabled
	Status     string    `json:"status"`
	CreateTime time.Time `json:"CreateTime"`
	StatusTime time.Time `json:"StatusTime"`
}

type OpenpitrixApp struct {

	// abstraction of app
	Abstraction string `json:"Abstraction,omitempty"`

	// whether there is a released version in the app
	Active bool `json:"Active,omitempty"`

	// app id
	AppID string `json:"AppId,omitempty"`

	// app version types eg.[vmbased|helm]
	AppVersionTypes string `json:"app_version_types,omitempty"`

	// chart name of app
	ChartName string `json:"ChartName,omitempty"`

	// company website
	CompanyWebsite string `json:"company_website,omitempty"`

	// the time when app create
	CreateTime strfmt.DateTime `json:"CreateTime,omitempty"`

	// app description
	Description string `json:"Description,omitempty"`

	// app home page
	Home string `json:"Home,omitempty"`

	// app icon
	Icon string `json:"Icon,omitempty"`

	// the isv user who create the app
	Isv string `json:"Isv,omitempty"`

	// app key words
	Keywords string `json:"keywords,omitempty"`

	// latest version of app
	LatestAppVersion *OpenpitrixAppVersion `json:"latest_app_version,omitempty"`

	// app maintainers
	Maintainers string `json:"maintainers,omitempty"`

	// app name
	Name string `json:"name,omitempty"`

	// owner of app
	Owner string `json:"Owner,omitempty"`

	// owner path of the app, concat string group_path:user_id
	OwnerPath string `json:"owner_path,omitempty"`

	// app instructions
	Readme string `json:"readme,omitempty"`

	// repository(store app package) id
	RepoID string `json:"RepoId,omitempty"`

	// app screenshots
	Screenshots string `json:"Screenshots,omitempty"`

	// sources of app
	Sources string `json:"Sources,omitempty"`

	// status eg.[modify|submit|review|cancel|release|delete|pass|reject|suspend|recover]
	Status string `json:"Status,omitempty"`

	// record status changed time
	StatusTime strfmt.DateTime `json:"status_time,omitempty"`

	// tos of app
	Tos string `json:"tos,omitempty"`

	// the time when app update
	UpdateTime strfmt.DateTime `json:"update_time,omitempty"`
}

type OpenpitrixAppVersion struct {

	// active or not
	Active bool `json:"Active,omitempty"`

	// app id
	AppID string `json:"AppId,omitempty"`

	// the time when app version create
	CreateTime strfmt.DateTime `json:"create_time,omitempty"`

	// description of app of specific version
	Description string `json:"Description,omitempty"`

	// home of app of specific version
	Home string `json:"Home,omitempty"`

	// icon of app of specific version
	Icon string `json:"Icon,omitempty"`

	// keywords of app of specific version
	Keywords string `json:"keywords,omitempty"`

	// maintainers of app of specific version
	Maintainers string `json:"maintainers,omitempty"`

	// message path of app of specific version
	Message string `json:"Message,omitempty"`

	// version name
	Name string `json:"Name,omitempty"`

	// owner
	Owner string `json:"Owner,omitempty"`

	// owner path of app of specific version, concat string group_path:user_id
	OwnerPath string `json:"owner_path,omitempty"`

	// package name of app of specific version
	PackageName string `json:"PackageName,omitempty"`

	// readme of app of specific version
	Readme string `json:"readme,omitempty"`

	// review id of app of specific version
	ReviewID string `json:"ReviewId,omitempty"`

	// screenshots of app of specific version
	Screenshots string `json:"Screenshots,omitempty"`

	// sequence of app of specific version
	Sequence int64 `json:"sequence,omitempty"`

	// sources of app of specific version
	Sources string `json:"Sources,omitempty"`

	// status of app of specific version eg.[draft|submitted|passed|rejected|active|in-review|deleted|suspended]
	Status string `json:"Status,omitempty"`

	// record status changed time
	StatusTime strfmt.DateTime `json:"StatusTime,omitempty"`

	// type of app of specific version
	Type string `json:"Type,omitempty"`

	// the time when app version update
	UpdateTime strfmt.DateTime `json:"UpdateTime,omitempty"`

	// version id of app
	VersionID string `json:"VersionId,omitempty"`
}

type OpenpitrixAttachment struct {

	// filename map to content
	AttachmentContent map[string]strfmt.Base64 `json:"attachment_content,omitempty"`

	// attachment id
	AttachmentID string `json:"attachment_id,omitempty"`

	// the time attachment create
	CreateTime strfmt.DateTime `json:"create_time,omitempty"`
}

type OpenpitrixCategory struct {

	// category id
	CategoryID string `json:"CategoryID,omitempty"`

	// the time when category create
	CreateTime strfmt.DateTime `json:"CreateTime,omitempty"`

	// category description
	Description string `json:"Description,omitempty"`

	// category icon
	Icon string `json:"Icon,omitempty"`

	// the i18n of this category, json format, sample: {"zh_cn": "数据库", "en": "database"}
	Locale string `json:"Locale,omitempty"`

	// category name,app belong to a category,eg.[AI|Firewall|cache|...]
	Name string `json:"Name,omitempty"`

	// owner
	Owner string `json:"Owner,omitempty"`

	// owner path, concat string group_path:user_id
	OwnerPath string `json:"OwnerPath,omitempty"`

	// the time when category update
	UpdateTime strfmt.DateTime `json:"UpdateTime,omitempty"`
}

type OpenpitrixClusterCommon struct {

	// action of cluster support.eg.[change_vxnet|scale_horizontal]
	AdvancedActions string `json:"advanced_actions,omitempty"`

	// agent install or not
	AgentInstalled bool `json:"agent_installed,omitempty"`

	// policy of backup
	BackupPolicy string `json:"backup_policy,omitempty"`

	// backup service config, a json string
	BackupService string `json:"backup_service,omitempty"`

	// cluster id
	ClusterID string `json:"cluster_id,omitempty"`

	// custom metadata script, a json string
	CustomMetadataScript string `json:"custom_metadata_script,omitempty"`

	// custom service config, a json string
	CustomService string `json:"custom_service,omitempty"`

	// delete snapshot service config, a json string
	DeleteSnapshotService string `json:"delete_snapshot_service,omitempty"`

	// destroy service config, a json string
	DestroyService string `json:"destroy_service,omitempty"`

	// health check config,a json string
	HealthCheck string `json:"health_check,omitempty"`

	// hypervisor.eg.[docker|kvm|...]
	Hypervisor string `json:"hypervisor,omitempty"`

	// image id
	ImageID string `json:"image_id,omitempty"`

	// support incremental backup or not
	IncrementalBackupSupported bool `json:"incremental_backup_supported,omitempty"`

	// init service config, a json string
	InitService string `json:"init_service,omitempty"`

	// monitor config,a json string
	Monitor string `json:"monitor,omitempty"`

	// passphraseless
	Passphraseless string `json:"passphraseless,omitempty"`

	// restart service config, a json string
	RestartService string `json:"restart_service,omitempty"`

	// restore service config, a json string
	RestoreService string `json:"restore_service,omitempty"`

	// cluster role
	Role string `json:"role,omitempty"`

	// scale in service config, a json string
	ScaleInService string `json:"scale_in_service,omitempty"`

	// scale out service config, a json string
	ScaleOutService string `json:"scale_out_service,omitempty"`

	// bound of server id(index number), some service(zookeeper) need the index to be bounded
	ServerIDUpperBound int64 `json:"server_id_upper_bound,omitempty"`

	// start service config, a json string
	StartService string `json:"start_service,omitempty"`

	// stop service config, a json string
	StopService string `json:"stop_service,omitempty"`

	// upgrade service config, a json string
	UpgradeService string `json:"upgrade_service,omitempty"`

	// vertical scaling policy.eg.[parallel|sequential]
	VerticalScalingPolicy string `json:"vertical_scaling_policy,omitempty"`
}
type OpenpitrixClusterClusterCommonSet []*OpenpitrixClusterCommon

type OpenpitrixCluster struct {

	// additional info
	AdditionalInfo string `json:"additional_info,omitempty"`

	// id of app run in cluster
	AppID string `json:"AppId,omitempty"`

	// cluster id
	ClusterID string `json:"ClusterId,omitempty"`

	// cluster type, frontgate or normal cluster
	ClusterType int64 `json:"cluster_type,omitempty"`

	// the time when cluster create
	CreateTime strfmt.DateTime `json:"CreateTime,omitempty"`

	// cluster used to debug or not
	Debug bool `json:"Debug,omitempty"`

	// cluster description
	Description string `json:"Description,omitempty"`

	// endpoint of cluster
	Endpoints string `json:"endpoints,omitempty"`

	// cluster env
	Env string `json:"Env,omitempty"`

	// cluster name
	Name string `json:"Name,omitempty"`

	// owner
	Owner string `json:"Owner,omitempty"`

	// owner path, concat string group_path:user_id
	OwnerPath string `json:"owner_path,omitempty"`

	// cluster runtime id
	RuntimeID string `json:"RuntimeId,omitempty"`

	// cluster status eg.[active|used|enabled|disabled|deleted|stopped|ceased]
	Status string `json:"Status,omitempty"`

	// record status changed time
	StatusTime strfmt.DateTime `json:"status_time,omitempty"`

	// cluster upgraded time
	UpgradeTime strfmt.DateTime `json:"upgrade_time,omitempty"`

	// id of version of app run in cluster
	VersionID string `json:"VersionId,omitempty"`

	// zone of cluster eg.[pek3a|pek3b]
	Zone string `json:"Zone,omitempty"`
}

type OpenpitrixRepo struct {

	// app default status eg[active|draft]
	AppDefaultStatus string `json:"app_default_status,omitempty"`

	// controller, value 0 for self resource, value 1 for openpitrix resource
	Controller int32 `json:"controller,omitempty"`

	// the time when repository create
	CreateTime strfmt.DateTime `json:"create_time,omitempty"`

	// credential of visiting the repository
	Credential string `json:"Credential,omitempty"`

	// repository description
	Description string `json:"Description,omitempty"`

	// repository name
	Name string `json:"Name,omitempty"`

	// owner
	Owner string `json:"Owner,omitempty"`

	// owner path, concat string group_path:user_id
	OwnerPath string `json:"owner_path,omitempty"`

	// runtime provider eg.[qingcloud|aliyun|aws|kubernetes]
	Providers []string `json:"providers"`

	// repository id
	RepoID string `json:"RepoID,omitempty"`

	// status eg.[active|deleted]
	Status string `json:"status,omitempty"`

	// record status changed time
	StatusTime strfmt.DateTime `json:"status_time,omitempty"`

	// type of repository eg.[http|https|s3]
	Type string `json:"Type,omitempty"`

	// url of visiting the repository
	URL string `json:"URL,omitempty"`

	// visibility.eg:[public|private]
	Visibility string `json:"visibility,omitempty"`
}
