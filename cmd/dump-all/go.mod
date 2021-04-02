module kubesphere.io/openpitrix-jobs/dump-all

go 1.16

require (
	github.com/aws/aws-sdk-go v1.25.21
	openpitrix.io/openpitrix v0.4.8
	sigs.k8s.io/yaml v1.2.0 // indirect
)

replace github.com/gocraft/dbr => github.com/gocraft/dbr v0.0.0-20180507214907-a0fd650918f6

replace github.com/docker/docker => github.com/docker/engine v0.0.0-20190423201726-d2cfbce3f3b0

replace openpitrix.io/openpitrix => openpitrix.io/openpitrix v0.4.9-0.20200617102217-10d232395f06

replace k8s.io/client-go => k8s.io/client-go v0.17.3
