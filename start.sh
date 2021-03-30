#!/bin/sh

op_deploy=$(kubectl get deploy -n openpitrix-system -lapp=openpitrix,component=openpitrix-hyperpitrix -oNAME)

if [[ "Xop_deploy" -eq "X" ]]; then
  echo "import app"
  import-app import $@
else
  echo "convert app"
  import-app convert $@
fi