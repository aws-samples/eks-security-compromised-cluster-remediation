#!/usr/bin/env bash
#set -o errexit

NEWLINE=$'\n'
CMD_KUBECTL="kubectl"
PREFIX="" && [[ "$PWD" != *tests* ]] && PREFIX="tests/"
clear

${CMD_KUBECTL} apply -f ${PREFIX}0-ns.yaml

echo "${NEWLINE}"

echo ">>> 1. Good config..."
${CMD_KUBECTL} apply -f ${PREFIX}1-ok.yaml
sleep 2
${CMD_KUBECTL} delete -f ${PREFIX}1-ok.yaml
sleep 2

echo "${NEWLINE}"

echo ">>> 2. Missing deployment spec template container security context..."
${CMD_KUBECTL} apply -f ${PREFIX}2-dep-sec-cont.yaml
sleep 2

echo "${NEWLINE}"

echo ">>> 3. Missing deployment spec template container security context elements..."
${CMD_KUBECTL} apply -f ${PREFIX}3-dep-sec-cont.yaml
sleep 2

echo "${NEWLINE}"

echo ">>> 4. Deployment spec template container security context allowPrivilegeEscalation ..."
${CMD_KUBECTL} apply -f ${PREFIX}4-dep-sec-cont.yaml
sleep 2

echo "${NEWLINE}"

echo ">>> 5. Missing deployment spec template container security context runAsUser & readOnlyRootFilesystem..."
${CMD_KUBECTL} apply -f ${PREFIX}5-dep-sec-cont.yaml
sleep 2

echo "${NEWLINE}"

echo ">>> 6. Deployment spec template container security context runAsUser = 0"
${CMD_KUBECTL} apply -f ${PREFIX}6-dep-sec-cont.yaml
sleep 2

echo "${NEWLINE}"

echo ">>> 7. Missing deployment spec template container security context readOnlyRootFilesystem"
${CMD_KUBECTL} apply -f ${PREFIX}7-dep-sec-cont.yaml
sleep 2

echo "${NEWLINE}"

echo ">>> 8. Deployment spec template container security context readOnlyRootFilesystem = false"
${CMD_KUBECTL} apply -f ${PREFIX}8-dep-sec-cont.yaml
sleep 2

echo "${NEWLINE}"

echo ">>> 9. Deployment host network"
${CMD_KUBECTL} apply -f ${PREFIX}9-dep-host-network.yaml
sleep 2

echo "${NEWLINE}"

echo ">>> 10. Pod host network"
${CMD_KUBECTL} apply -f ${PREFIX}10-pod-host-network.yaml
sleep 2

echo "${NEWLINE}"

echo ">>> 11. Pod security context"
${CMD_KUBECTL} apply -f ${PREFIX}11-pod-sec-cont.yaml
sleep 2

echo "${NEWLINE}"

echo ">>> 12. Pod host filesystem"
${CMD_KUBECTL} apply -f ${PREFIX}12-pod-host-filesystem.yaml
sleep 2

echo "${NEWLINE}"

echo ">>> 13. Deployment host filesystem"
${CMD_KUBECTL} apply -f ${PREFIX}13-dep-host-filesystem.yaml
sleep 2

echo "${NEWLINE}"

${CMD_KUBECTL} delete -f ${PREFIX}0-ns.yaml
















