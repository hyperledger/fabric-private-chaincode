package handlers

import (
	"encoding/hex"
	"fmt"
	"net/http"
	
	"github.com/gin-gonic/gin"
	"github.com/hyperledger-labs/ccapi/chaincode"
	"github.com/hyperledger-labs/ccapi/common"
	protos "github.com/hyperledger/fabric-protos-go-apiv2/common"
	queryresultprotos "github.com/hyperledger/fabric-protos-go-apiv2/ledger/queryresult"
	rwsetprotos "github.com/hyperledger/fabric-protos-go-apiv2/ledger/rwset"
	mspprotos "github.com/hyperledger/fabric-protos-go-apiv2/msp"
	peerprotos "github.com/hyperledger/fabric-protos-go-apiv2/peer"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
)

func QueryQSCC(c *gin.Context) {
	channelName := c.Param("channelName")
	txname := c.Param("txname")

	switch txname {
	case "getBlockByNumber":
		getBlockByNumber(c, channelName)
	case "getBlockByHash":
	getBlockByHash(c, channelName)
	case "getTransactionByID":
		getTransactionByID(c, channelName)
	case "getChainInfo":
		getChainInfo(c, channelName)
	case "getBlockByTxID":
		getBlockByTxID(c, channelName)
	default:
		common.Abort(c, http.StatusNotFound, fmt.Errorf("unknown endpoint call"))
	}
}

func getChainInfo(c *gin.Context, channelName string) {
	// Query
	user := c.GetHeader("User")
	if user == "" {
		user = "Admin"
	}

	result, err := chaincode.QueryGateway(channelName, "qscc", "GetChainInfo", user, []string{channelName})
	if err != nil {
		err, status := common.ParseError(err)
		common.Abort(c, status, err)
		return
	}

	var chainInfo protos.BlockchainInfo

	err = proto.Unmarshal(result, &chainInfo)
	if err != nil {
		common.Abort(c, http.StatusInternalServerError, err)
		return
	}

	chainInfoRes := map[string]interface{}{
		"height":              chainInfo.Height,
		"current_block_hash":  hex.EncodeToString(chainInfo.CurrentBlockHash),
		"previous_block_hash": hex.EncodeToString(chainInfo.PreviousBlockHash),
	}

	common.Respond(c, chainInfoRes, http.StatusOK, nil)
}

func getBlockByNumber(c *gin.Context, channelName string) {
	// Query
	user := c.GetHeader("User")
	if user == "" {
		user = "Admin"
	}

	number, ok := c.GetQuery("number")
	if !ok {
		common.Abort(c, http.StatusBadRequest, fmt.Errorf("missing number"))
		return
	}

	result, err := chaincode.QueryGateway(channelName, "qscc", "GetBlockByNumber", user, []string{channelName, number})
	if err != nil {
		err, status := common.ParseError(err)
		common.Abort(c, status, err)
		return
	}

	blockMap, err := decodeBlock(result)
	if err != nil {
		common.Abort(c, http.StatusInternalServerError, err)
		return
	}

	common.Respond(c, blockMap, http.StatusOK, nil)
}

func getBlockByTxID(c *gin.Context, channelName string) {
	// Query
	user := c.GetHeader("User")
	if user == "" {
		user = "Admin"
	}

	txid, ok := c.GetQuery("txid")
	if !ok {
		common.Abort(c, http.StatusBadRequest, fmt.Errorf("missing number"))
		return
	}

	result, err := chaincode.QueryGateway(channelName, "qscc", "GetBlockByTxID", user, []string{channelName, txid})
	if err != nil {
		err, status := common.ParseError(err)
		common.Abort(c, status, err)
		return
	}

	blockMap, err := decodeBlock(result)
	if err != nil {
		common.Abort(c, http.StatusInternalServerError, err)
		return
	}

	common.Respond(c, blockMap, http.StatusOK, nil)
}

func getBlockByHash(c *gin.Context, channelName string) {
	// Query
	user := c.GetHeader("User")
	if user == "" {
		user = "Admin"
	}

	hash, ok := c.GetQuery("hash")
	if !ok {
		common.Abort(c, http.StatusBadRequest, fmt.Errorf("missing hash"))
		return
	}

	hashBytes, err := hex.DecodeString(hash)

	if err != nil {
		common.Abort(c, http.StatusBadRequest, fmt.Errorf("invalid hash format: %s", hash))
		return
	}

	result, err := chaincode.QueryGateway(channelName, "qscc", "GetBlockByHash", user, []string{channelName, string(hashBytes)})
	
	if err != nil {
		err, status := common.ParseError(err)
		common.Abort(c, status, err)
		return
	}

	blockMap, err := decodeBlock(result)
	if err != nil {
		common.Abort(c, http.StatusInternalServerError, err)
		return
	}

	common.Respond(c, blockMap, http.StatusOK, nil)
}

func getTransactionByID(c *gin.Context, channelName string) {
	// Query
	user := c.GetHeader("User")
	if user == "" {
		user = "Admin"
	}

	fmt.Println("getting txid")
	txid, ok := c.GetQuery("txid")
	if !ok {
		common.Abort(c, http.StatusBadRequest, fmt.Errorf("missing txid"))
		return
	}

	fmt.Println("calling GetTransactionByID")
	result, err := chaincode.QueryGateway(channelName, "qscc", "GetTransactionByID", user, []string{channelName, txid})
	if err != nil {
		fmt.Println("error calling GetTransactionByID: ", err)
		err, status := common.ParseError(err)
		common.Abort(c, status, err)
		return
	}

	fmt.Println("decoding transaction")
	m, err := decodeProcessedTransaction(result)
	if err != nil {
		fmt.Println("error decoding transaction: ", err)
		common.Abort(c, http.StatusInternalServerError, err)
		return
	}

	fmt.Println("responding")
	common.Respond(c, m, http.StatusOK, nil)
}

func decodeBlock(b []byte) (map[string]interface{}, error) {
	var block protos.Block

	err := proto.Unmarshal(b, &block)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal block")
	}

	blockDataProto := block.GetData()
	dataProto := blockDataProto.GetData()

	dataList := make([]interface{}, 0)

	for _, dataP := range dataProto {
		var envelope protos.Envelope

		err := proto.Unmarshal(dataP, &envelope)
		if err != nil {
			return nil, errors.Wrap(err, "failed to unmarshal envelope")
		}

		var payload protos.Payload

		err = proto.Unmarshal(envelope.Payload, &payload)
		if err != nil {
			return nil, errors.Wrap(err, "failed to unmarshal payload")
		}

		var channelHeader protos.ChannelHeader

		err = proto.Unmarshal(payload.Header.ChannelHeader, &channelHeader)
		if err != nil {
			return nil, errors.Wrap(err, "failed to unmarshal channel header")
		}

		var tx interface{}
		if channelHeader.Type == int32(protos.HeaderType_ENDORSER_TRANSACTION) {
			tx, err = decodeTransaction(payload.Data)
			if err != nil {
				return nil, errors.Wrap(err, "failed to decode transaction")
			}
		} else {
			tx = payload.Data
		}

		dataList = append(dataList, map[string]interface{}{
			"payload": map[string]interface{}{
				"header": map[string]interface{}{
					"channel_header":   &channelHeader,
					"signature_header": payload.Header.SignatureHeader,
				},
				"data": tx,
			},
			"signature": envelope.Signature,
		})

	}

	blockMap := map[string]interface{}{
		"header": map[string]interface{}{
			"number":        block.Header.Number,
			"previous_hash": hex.EncodeToString(block.Header.PreviousHash),
			"data_hash":     hex.EncodeToString(block.Header.DataHash),
		},
		"metadata": block.Metadata,
		"data":     dataList,
	}
	return blockMap, nil
}

func decodeTransaction(b []byte) (map[string]interface{}, error) {
	var transaction peerprotos.Transaction

	err := proto.Unmarshal(b, &transaction)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal transaction")
	}

	actions := transaction.GetActions()

	actionList := make([]interface{}, 0)
	for _, action := range actions {
		headerB := action.GetHeader()
		payloadB := action.GetPayload()

		var sigHeader protos.SignatureHeader
		err = proto.Unmarshal(headerB, &sigHeader)
		if err != nil {
			return nil, errors.Wrap(err, "failed to unmarshal signature header")
		}

		var creator mspprotos.SerializedIdentity
		err = proto.Unmarshal(sigHeader.Creator, &creator)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("failed to unmarshal creator, %s", string(sigHeader.Creator)))
		}

		var ccActionPayload peerprotos.ChaincodeActionPayload
		err = proto.Unmarshal(payloadB, &ccActionPayload)
		if err != nil {
			return nil, errors.Wrap(err, "failed to unmarshal chaincode action payload")
		}

		var ccProposalPayload peerprotos.ChaincodeProposalPayload
		err = proto.Unmarshal(ccActionPayload.ChaincodeProposalPayload, &ccProposalPayload)
		if err != nil {
			return nil, errors.Wrap(err, "failed to unmarshal chaincode proposal payload")
		}

		var input peerprotos.ChaincodeInvocationSpec
		err = proto.Unmarshal(ccProposalPayload.Input, &input)
		if err != nil {
			return nil, errors.Wrap(err, "failed to unmarshal chaincode invocation spec")
		}

		args := input.ChaincodeSpec.Input.Args

		inputList := make([]string, 0)
		for _, arg := range args {
			inputList = append(inputList, string(arg))
		}

		ccEndorsedAction := ccActionPayload.Action

		var proposalResponsePayload peerprotos.ProposalResponsePayload
		err = proto.Unmarshal(ccEndorsedAction.ProposalResponsePayload, &proposalResponsePayload)
		if err != nil {
			return nil, errors.Wrap(err, "failed to unmarshal proposal response payload")
		}

		extension := proposalResponsePayload.Extension
		proposalHash := proposalResponsePayload.ProposalHash

		var chaincodeAction peerprotos.ChaincodeAction
		err = proto.Unmarshal(extension, &chaincodeAction)
		if err != nil {
			return nil, errors.Wrap(err, "failed to unmarshal chaincode action")
		}

		var txRWSet rwsetprotos.TxReadWriteSet
		err = proto.Unmarshal(chaincodeAction.Results, &txRWSet)
		if err != nil {
			return nil, errors.Wrap(err, "failed to unmarshal tx read write set")
		}

		nsRWList := make([]interface{}, 0)
		for _, nsRWSet := range txRWSet.NsRwset {
			var kvSet queryresultprotos.KV
			err = proto.Unmarshal(nsRWSet.Rwset, &kvSet)
			if err != nil {
				return nil, errors.Wrap(err, "failed to unmarshal kv read write set")
			}

			nsRWList = append(nsRWList, map[string]interface{}{
				"namespace": nsRWSet.Namespace,
				"rwset": map[string]interface{}{
					"key":       kvSet.Key,
					"value":     string(kvSet.Value),
					"namespace": kvSet.Namespace,
				},
				"collections": nsRWSet.CollectionHashedRwset,
			})
		}

		endorsements := ccEndorsedAction.Endorsements

		endorsementList := make([]interface{}, 0)
		for _, endorsement := range endorsements {
			var endorser mspprotos.SerializedIdentity
			err = proto.Unmarshal(endorsement.Endorser, &endorser)
			if err != nil {
				return nil, errors.Wrap(err, "failed to unmarshal endorser")
			}

			endorsementList = append(endorsementList, map[string]interface{}{
				"endorser":  &endorser,
				"signature": endorsement.Signature,
			})
		}

		actionList = append(actionList, map[string]interface{}{
			"header": map[string]interface{}{
				"creator": &creator,
				"nonce":   sigHeader.Nonce,
			},
			"payload": map[string]interface{}{
				"chaincode_proposal_payload": map[string]interface{}{
					"chaincode_id": input.ChaincodeSpec.ChaincodeId,
					// "type":         input.ChaincodeSpec.Type,
					// "timeout":      input.ChaincodeSpec.Timeout,
					"input": inputList,
				},
				"action": map[string]interface{}{
					"proposal_response_payload": map[string]interface{}{
						"proposal_hash": proposalHash,
						"extension": map[string]interface{}{
							"results": map[string]interface{}{
								"ns_rwset":   nsRWList,
								"data_model": txRWSet.DataModel,
							},
							"response": map[string]interface{}{
								"status":  chaincodeAction.Response.Status,
								"message": chaincodeAction.Response.Message,
								"payload": string(chaincodeAction.Response.Payload),
							},
							"chaincode_id": chaincodeAction.ChaincodeId,
							"events":       chaincodeAction.Events,
						},
					},
					"endorsements": endorsementList,
				},
			},
		})
	}

	transactionMap := map[string]interface{}{
		"actions": actionList,
	}

	return transactionMap, nil
}

func decodeProcessedTransaction(t []byte) (map[string]interface{}, error) {
	var processedTransaction peerprotos.ProcessedTransaction
	err := proto.Unmarshal(t, &processedTransaction)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal transaction")
	}

	transactionEnv := processedTransaction.TransactionEnvelope
	transactionPayload := transactionEnv.GetPayload()

	var payload protos.Payload

	err = proto.Unmarshal(transactionPayload, &payload)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal payload")
	}

	transaction, err := decodeTransaction(payload.Data)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode transaction")
	}

	processedTransactionMap := map[string]interface{}{
		"payload":   transaction,
		"signature": transactionEnv.Signature,
	}

	return processedTransactionMap, nil
}