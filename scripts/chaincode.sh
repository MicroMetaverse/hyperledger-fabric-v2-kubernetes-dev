packageAndInstall() {
  CCURL=$1
  CCNAME=$2
  # CCURL="github.com/blockchaind/hyperledger-fabric-v2-kubernetes-dev/key-value-chaincode"
  # CCNAME="keyval"
  LANG=golang
  LABEL=${CCNAME}v1
  #>=go 1.16
  # LOCAL_CHAINCODE_PATH=/go/src/${CCURL}
  LOCAL_CHAINCODE_PATH=/go/hyperledger-fabric-v2-kubernetes-dev/key-value-chaincode

  cat <<EOF
echo "---getting chaincode for ${ORG} ${PEER}"
go version
echo "set go goproxy"
go env -w GO111MODULE=on

#go env -w GOPROXY=https://goproxy.io,direct
go env -w GOPROXY=https://goproxy.cn,direct


go env | grep GOPROXY
echo "set go goproxy,end"

echo "---go get -d ${CCURL}"
go get -d ${CCURL}
echo "---go get -d ${CCURL}"

pwd
# go get or 源码编译二选一
# no code in src,  git clone xxx
rm -rf hyperledger-fabric-v2-kubernetes-dev
git clone  https://gitclone.com/github.com/smallverse/hyperledger-fabric-v2-kubernetes-dev.git

echo "---ls---"
ls
echo "---ls---"

# https://blog.csdn.net/bean_business/article/details/110008244
cd ${LOCAL_CHAINCODE_PATH}
pwd
go env -w GOPROXY=https://goproxy.io,direct
go env -w GO111MODULE=on
go mod vendor
cd -
pwd

peer lifecycle chaincode package ${CCNAME}.tar.gz --path ${LOCAL_CHAINCODE_PATH} --lang ${LANG} --label ${LABEL}
peer lifecycle chaincode install ${CCNAME}.tar.gz

EOF
}

approve() {
  CCNAME=$1
  CHANNEL_ID=$2
  LABEL=${CCNAME}v1

  cat <<EOF
PACKAGE_ID=\$(peer lifecycle chaincode queryinstalled | awk '/${LABEL}/ {print substr(\$3, 1, length(\$3)-1)}')
echo "Package ID: \${PACKAGE_ID}"
peer lifecycle chaincode approveformyorg --package-id \${PACKAGE_ID} \
  --signature-policy "AND('Org1MSP.member','Org2MSP.member','Org3MSP.member')" \
  -C ${CHANNEL_ID} -n ${CCNAME} -v 1.0  --sequence 1 \
  --tls true --cafile \$ORDERER_TLS_ROOTCERT_FILE --waitForEvent
EOF
}

checkCommitReadiness() {
  CCNAME=$1
  CHANNEL_ID=$2
  cat <<EOF
peer lifecycle chaincode checkcommitreadiness \
--name ${CCNAME} --channelID ${CHANNEL_ID} \
--signature-policy "AND('Org1MSP.member', 'Org2MSP.member', 'Org3MSP.member')" \
--version 1.0 --sequence 1
EOF
}

commit() {
  CCNAME=$1
  CHANNEL_ID=$2
  cat <<EOF
echo "Committing smart contract"
peer lifecycle chaincode commit \
  --channelID ${CHANNEL_ID} \
  --name ${CCNAME} \
  --version 1.0 \
  --signature-policy "AND('Org1MSP.member','Org2MSP.member', 'Org3MSP.member')" \
  --sequence 1 --waitForEvent \
  --peerAddresses peer0.org1:7051 \
  --peerAddresses peer0.org2:7051 \
  --peerAddresses peer0.org3:7051  \
  --tlsRootCertFiles /etc/hyperledger/fabric-peer/client-root-tlscas/tlsca.org1-cert.pem \
  --tlsRootCertFiles /etc/hyperledger/fabric-peer/client-root-tlscas/tlsca.org2-cert.pem \
  --tlsRootCertFiles /etc/hyperledger/fabric-peer/client-root-tlscas/tlsca.org3-cert.pem  \
  --tls true --cafile \$ORDERER_TLS_ROOTCERT_FILE \
  -o orderer.org1:7050
EOF
}

invoke() {
  CCNAME=$1
  CHANNEL_ID=$2
  CTOR=$3

  cat <<EOF
echo "Submitting invoketransaction to smart contract on ${CHANNEL_ID}"
peer chaincode invoke \
  --channelID ${CHANNEL_ID} \
  --name ${CCNAME} \
  --ctor '${CTOR}' \
  --waitForEvent \
  --waitForEventTimeout 300s \
  --cafile \$ORDERER_TLS_ROOTCERT_FILE \
  --tls true -o orderer.org1:7050 \
  --peerAddresses peer0.org1:7051 \
  --peerAddresses peer0.org2:7051 \
  --peerAddresses peer0.org3:7051  \
  --tlsRootCertFiles /etc/hyperledger/fabric-peer/client-root-tlscas/tlsca.org1-cert.pem \
  --tlsRootCertFiles /etc/hyperledger/fabric-peer/client-root-tlscas/tlsca.org2-cert.pem \
  --tlsRootCertFiles /etc/hyperledger/fabric-peer/client-root-tlscas/tlsca.org3-cert.pem
EOF
}

query() {
  CCNAME=$1
  CHANNEL_ID=$2
  CTOR=$3

  cat <<EOF
peer chaincode query --name ${CCNAME} \
--channelID ${CHANNEL_ID} \
--ctor '${CTOR}' \
--tls --cafile \$ORDERER_TLS_ROOTCERT_FILE
EOF

}
