#!/usr/bin/env bash
export CHANNEL_PROFILE=AllOrgsChannel # defined in configtx.yaml
export CHANNEL_ID=erc721             # anything 暂时不可行Error: Invalid channel create transaction : mismatched channel ID allorgs != xxx

export CCURL=github.com/smallverse/hyperledger-fabric-v2-kubernetes-dev/erc721-chaincode
export CCNAME=erc721-chaincode
#
source ./scripts/generate-crypto-artifacts.sh
source ./scripts/create-secrets.sh
source ./scripts/chaincode-erc721.sh
#


createNamespaces() {
	for NS in org1 org2 org3 org4 org5; do
		kubectl create ns $NS
	done
}

deleteNamespaces() {
	kubectl delete ns org1 org2 org3 org4 org5
	exit $?
}

ccpgen() {
	sh scripts/ccp-generate.sh
}

###
case $1 in
#
up)
	genCrypto
	genChannelArtifacts ${CHANNEL_PROFILE} ${CHANNEL_ID}
	createNamespaces
	for ORG in org1 org2 org3 org4 org5; do
		creareOrdererSecret ${ORG}
		sh templates/orderer.sh ${ORG} | kubectl apply -f -
	done

	for ORG in org1 org2 org3; do
		createCASecrets ${ORG}
		createAdminSecret ${ORG} ${PEER}
		createChannelArtifactsSecrets ${ORG} ${CHANNEL_ID}
		createOrgRootTLSCAsSecret ${ORG}
		sh templates/ca.sh ${ORG} | kubectl apply -f -

		for PEER in peer0; do
			createPeerSecret ${ORG} ${PEER}
			sh templates/peer.sh ${ORG} ${PEER} | kubectl apply -f -
			sh templates/admin.sh ${ORG} ${PEER} ${CHANNEL_ID} | kubectl apply -f -
		done
	done
	;;

#
down)
	deleteNamespaces
	;;
#
explorerAndAPI)
	sh templates/explorer.sh ${CHANNEL_ID} | kubectl -n org1 apply -f -
	kubectl -n org1 apply -f explorer/explorerdb.yaml
  kubectl -n org1 apply -f api/api-k8s.yaml
	;;
#
joinChannel)
	for ORG in org1 org2 org3; do
		for PEER in peer0; do
			./scripts/join-channel.sh ${ORG} ${PEER} ${CHANNEL_ID} | sh -c "kubectl --namespace ${ORG} \
        exec -i $(kubectl -n ${ORG} get pod -l app=admin -o name) -- sh -"
		done
	done
	;;

#
ccInstall)
	for ORG in org1 org2 org3; do
		echo "Installing on ${ORG} peers"
		for PEER in peer0; do
			echo "Installing ${CCNAME} on ${PEER}"
			packageAndInstall ${CCURL} ${CCNAME} | sh -c "kubectl --namespace ${ORG} \
        exec -i $(kubectl -n ${ORG} get pod -l app=admin -o name) -- sh -"
		done
	done
	;;

#
ccApprove)
	for ORG in org1 org2 org3; do
		echo "Installing on ${ORG} peers"
		for PEER in peer0; do
			echo "Approving ${CCNAME} for ${ORG}"
			approve ${CCNAME} ${CHANNEL_ID} | sh -c "kubectl --namespace ${ORG} \
        exec -i $(kubectl -n ${ORG} get pod -l app=admin -o name) -- sh -"
		done
	done
	;;

#
ccCommit)
	ORG=org1
	PEER=peer0
	commit ${CCNAME} ${CHANNEL_ID} | sh -c "kubectl --namespace ${ORG} \
    exec -i $(kubectl -n ${ORG} get pod -l app=admin -o name) -- sh -"
	;;

#
ccInvoke)
  CTOR=$2
  echo "---ccInvoke:${CTOR}"
	invoke ${CCNAME} ${CHANNEL_ID} "${CTOR}" | sh -c "kubectl --namespace org1 exec -i $(kubectl -n org1 get pod -l app=admin -o name) -- sh -"
	;;

#
ccQuery)
  CTOR=$2
  echo "---ccQuery:${CTOR}"
	for ORG in org1 org2 org3; do
		for PEER in peer0; do
			echo "Quering on ${PEER}.${ORG}"
			query ${CCNAME} ${CHANNEL_ID} "${CTOR}" | sh -c "kubectl --namespace ${ORG} exec -i $(kubectl -n ${ORG} get pod -l app=admin -o name) -- sh -"
		done
	done
	;;

#
esac
