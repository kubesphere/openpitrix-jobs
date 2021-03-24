#!/bin/bash

set -x
workdir=$1
chart_dir=${workdir}/charts
chartsUrl='https://raw.githubusercontent.com/openpitrix/helm-package-repository/master/package/urls.txt'

crd_resource=helmapplications.application.kubesphere.io
kubectl get crd ${crd_resource} -oname
if [[ $? -eq "1" ]]; then
  echo ${crd_resource} not found
  exit 1
fi

crd_resource=helmapplicationversions.application.kubesphere.io
kubectl get crd ${crd_resource} -oname

if [[ $? -eq "1" ]]; then
  echo ${crd_resource} not found
  exit 1
fi

function parse_chart_name_version() {
  line=${1}
  name_ver=${line##*/}
  name=${name_ver%-*}
  ver=${name_ver##*-}
  ver=${ver%.tgz}
  echo ${name}/${ver}
}
strSrc="0123456789abcdefghijklmnopqrstuvwxyz"

function random_id() {
  r=""
  for i in `seq 14`; do
        n=`echo "$RANDOM%36" | bc`
        r=${r}"${strSrc:$n:1}"
  done
  echo -n $r
}


function upload_chart() {
  chart_name=${1}
  attachmet_id="appv-$(random_id)"
  echo ${attachment_id}

}

upload_chart


exit

echo "download chart"
data=$(curl ${chartsUrl})
for line in ${data}; do
  chart_ver=$(parse_chart_name_version $line)
  echo "download chart data, ${chart_ver}"
  curl -LO ${line}
  appResource="
apiVersion: application.kubesphere.io/v1alpha1
kind: HelmApplication
metadata:
  annotations:
    kubesphere.io/creator: admin
  name: app-j2yvz76591m43q
spec:
  description: DEPRECATED - Ubiquiti Network's Unifi Controller
  name: unifi
"
  appVerResource="
apiVersion: application.kubesphere.io/v1alpha1
kind: HelmApplicationVersion
metadata:
  annotations:
    kubesphere.io/creator: admin
  labels:
    application.kubesphere.io/app-id: app-j2yvz76591m43q
    kubesphere.io/workspace: test-repo
  name: appv-5pmnlrolwwn15k
  ownerReferences:
  - apiVersion: application.kubesphere.io/v1alpha1
    kind: HelmApplication
    name: app-j2yvz76591m43q
    uid: bc6b536a-9a4b-4bfb-befc-89ef9b5cc834
spec:
  appVersion: 5.12.32
  created: "2021-03-19T07:45:24Z"
  dataKey: appv-5pmnlrolwwn15k
  description: DEPRECATED - Ubiquiti Network's Unifi Controller
  home: https://github.com/jacobalberty/unifi-docker
  icon: https://blog.ubnt.com/wp-content/uploads/2016/10/unifi-app-logo.png
  name: unifi
  version: 1.30.2
  "


done



