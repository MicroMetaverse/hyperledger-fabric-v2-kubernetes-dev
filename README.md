# Hyperledger Fabric v2 with Raft on Kubernetes

## Prerequisites

- Kubernetes cluster with at least 4GB memory and 2 vCPUs (tested on IBM Cloud free tier IKS)
- kubectl available on path and configured to use a cluster
- Fabric binaries available on path

## Architechture

- Three peer orgs and two orderer orgs. Peer orgs too run an orderer each.
- Each org components are deployed in org's own namespace
- crypto materials generated by cryptogen
- crypto materials and channel-artifacts are mounted as k8s Secret
- Fabric CA stores data in Postgres (in this demo in sqlite)
- Fabric peer uses couchdb as state db. CouchDB is deployed in a separate pod (in this demo same pod as peer itself)

![hyperledger-fabric-network](http://www.plantuml.com/plantuml/proxy?cache=no&src=https://raw.githubusercontent.com/blockchaind/hyperledger-fabric-v2-kubernetes-dev/master/network-diagram.puml)

## Network KV

Start

```bash
./hlf.sh up
```

Have peers joined to channel. Ensure all components are up and running.

```bash
./hlf.sh joinChannel
```

Chaincode lifecycle

```bash
./hlf.sh ccInstall
./hlf.sh ccApprove
./hlf.sh ccCommit
#
./hlf.sh explorerAndAPI

./hlf.sh ccInvoke         # Creates greeting="Hello, World!"
./hlf.sh ccQuery          # Reads greeting value
./hlf.sh ccInvokeUpdate   # Updates greeting="Hello, Blockchain!"
./hlf.sh ccQuery          # Reads greeting value to check update succeeded
```
## Network erc721
```shell
./hlf-erc721.sh up
./hlf-erc721.sh joinChannel

#
./hlf-erc721.sh ccInstall # 安装，通过
./hlf-erc721.sh ccApprove
./hlf-erc721.sh ccCommit

./hlf-erc721.sh explorerAndAPI
#
# 如'{"function":"MintWithTokenURI","Args":["101", "http://172.16.3.20:32000/test/000.jpg"]}'
./hlf-erc721.sh ccInvoke 'xxx' 
#
# 如 '{"function":"ClientAccountBalance","Args":[]}' 
# '{"function":"BalanceOf","Args":["xxx"]}' 
# '{"function":"ClientAccountID","Args":[]}' 
./hlf-erc721.sh ccQuery 'xxx' 

#
./hlf-erc721.sh down
```
## Explorer & Rest API

Start explorer db

explorer should now be available at <http://localhost:8080>

Access API Swagger UI at <http://localhost:3000/swagger>

## 加密问题
###1.通过加密机制实现自定义保护。
###2.TODO:链码自行完成加密解密。
