#!/usr/bin/env bash
#set -o errexit

PREFIX="" && [[ "$PWD" != *tests* ]] && PREFIX="tests/"


#These command will create a pod and run a shell command inside.
kubectl run demo-shell -n default --image=busybox --restart='Never' -- sh -c "sleep 600"
sleep 2
kubectl exec -i --tty demo-shell -n default -- sh -c "uptime" -n default

#This template will create a pod that will mount docker sock
kubectl create -f ${PREFIX}pod-demo-docker-sock.yaml -n default

#check3: This template will create a pod that will mount /etc as a hostPath Volume.
kubectl create -f ${PREFIX}pod-demo-hostpath.yaml -n default





