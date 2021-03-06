#!/bin/bash
echo "Helm Version"
helm version

echo "printenv"
printenv

set -x
if [[ $HELM_REPO_URL == "" ]]; then
  echo "HELM_REPO_URL is not set. Please set HELM_REPO_URL and deploy again";
  exit 1;
fi

if [[ $HELM_REPO_NAME == "" ]]; then
  HELM_REPO_NAME=$(basename $HELM_REPO_URL)
  echo $HELM_REPO_NAME
fi

if [[ $SAT_CHART_NAME == "" ]]; then
  SAT_CHART_NAME="helm-chart"
  echo $SAT_CHART_NAME
fi

echo "install the plugin"
helm repo add $HELM_REPO_NAME $HELM_REPO_URL
helm repo update
helm repo list

echo "HELM_OPTIONS"
echo $HELM_OPTIONS
helm_options=${HELM_OPTIONS// /}  #trim whitespaces in the HELM_OPTIONS string
helm_options=$(echo $helm_options | tr "," "\n")  #split HELM_OPTIONS by comma

if [[ $SAT_CHART_NAME == "ibm-object-storage-plugin" ]]; then
  if [[ $SAT_CHART_VERSION == "" ]];then
    helm fetch --untar $HELM_REPO_NAME/$SAT_CHART_NAME
  else
    helm fetch --untar $HELM_REPO_NAME/$SAT_CHART_NAME --version $SAT_CHART_VERSION
  fi
  cd $SAT_CHART_NAME/
  chmod 755 helm-ibmc/ibmc.sh
  helm plugin install helm-ibmc
  helm ibmc --help
  set +x
  if [[ $SAT_CHART_VERSION == "" ]];then
    helm_install_command="helm ibmc install $SAT_CHART_NAME ./ --namespace $SAT_NAMESPACE"
    helm_update_command="helm ibmc upgrade $SAT_CHART_NAME  $HELM_REPO_NAME/$SAT_CHART_NAME --namespace $SAT_NAMESPACE" 
  else 
    helm_install_command="helm ibmc install $SAT_CHART_NAME ./ --namespace $SAT_NAMESPACE --version $SAT_CHART_VERSION"
    helm_update_command="helm ibmc upgrade $SAT_CHART_NAME  $HELM_REPO_NAME/$SAT_CHART_NAME --namespace $SAT_NAMESPACE --version $SAT_CHART_VERSION" 
  fi
# elif [[ $SAT_CHART_NAME == "aws-ebs-csi-driver" ]]; then
#     if [[ $SAT_CHART_VERSION == "" ]];then
#     helm_install_command="helm upgrade --install $SAT_CHART_NAME $HELM_REPO_NAME/$SAT_CHART_NAME --namespace $SAT_NAMESPACE"
#   else
#     helm_install_command="helm upgrade --install $SAT_CHART_NAME $HELM_REPO_NAME/$SAT_CHART_NAME --namespace $SAT_NAMESPACE --version $SAT_CHART_VERSION"
#   fi
else
  if [[ $SAT_CHART_VERSION == "" ]];then
    helm_install_command="helm upgrade --install $SAT_CHART_NAME $HELM_REPO_NAME/$SAT_CHART_NAME --namespace $SAT_NAMESPACE"
  else
    helm_install_command="helm upgrade --install $SAT_CHART_NAME $HELM_REPO_NAME/$SAT_CHART_NAME --namespace $SAT_NAMESPACE --version $SAT_CHART_VERSION"
  fi
fi

for option in $helm_options
do
helm_install_command="${helm_install_command} --set $option"
done

echo "final helm install cmd"
echo ${helm_install_command}

#check if the plugin is already installed
set -x
helm ls -A --all | awk '{print $1}' | grep $SAT_CHART_NAME
if [ $? == 0 ]; then
  # if $UPDATE_CHART == "true"; then
  #   echo "update the pre-installed chart"
  #   ${helm_install_command}
  #   tail -f /dev/null
  if $UPDATE_CHART == "true"; then 
    if [[ $SAT_CHART_NAME == "ibm-object-storage-plugin" ]]; then 
      for option in $helm_options
      do
      helm_update_command="${helm_update_command} --set $option"
      done

       echo "update the pre-installed chart"
       ${helm_update_command}
       tail -f /dev/null
    else 
      echo "update the pre-installed chart"
      ${helm_install_command}
      tail -f /dev/null
    fi
  else 
    echo "chart already installed. please delete it and deploy the chart again";
    exit 1;
    set +x
  fi
else
  #helm_install_command="${helm_install_command} --debug"
  ${helm_install_command}
  tail -f /dev/null
fi