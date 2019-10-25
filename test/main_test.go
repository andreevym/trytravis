package main

import (
	"testing"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric-sdk-go/certstore"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/ledger"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/stretchr/testify/require"
)

var invoke = channel.Request{
	ChaincodeID: "mycc",
	Fcn:         "invoke",
	Args:        [][]byte{[]byte("a"), []byte("b"), []byte("10")},
}

var query = channel.Request{
	ChaincodeID: "mycc",
	Fcn:         "query",
	Args:        [][]byte{[]byte("a")},
}

var channelPeers = channel.WithTargetEndpoints("peer0.org1.example.com", "peer0.org2.example.com")
var ledgerPeers = ledger.WithTargetEndpoints("peer0.org1.example.com", "peer0.org2.example.com")

func TestMain(t *testing.T) {
	sdk, err := fabsdk.New(config.FromFile("./config.yml"))
	require.NoError(t, err)
	defer sdk.Close()
	cCtx := sdk.ChannelContext("mychannel", fabsdk.WithUser("Admin"), fabsdk.WithOrg("Org1"))
	client, err := channel.New(cCtx)
	require.NoError(t, err)
	ledger, err := ledger.New(cCtx)
	require.NoError(t, err)

	t.Run("Check value before tests", queryTest(client, "90"))
	t.Run("Invoke with cert", invokeTest(client, ledger, true))
	t.Run("Check value between tests", queryTest(client, "80"))
	t.Run("Invoke without cert", invokeTest(client, ledger, false))
	t.Run("Check value after tests", queryTest(client, "70"))
}

func queryTest(client *channel.Client, expected string) func(*testing.T) {
	return func(t *testing.T) {
		resp, err := client.Query(query, channelPeers)
		require.NoError(t, err)
		require.Equal(t, expected, string(resp.Payload))
		checkEndorsersResponse(t, resp.Responses, false)
	}
}

func invokeTest(client *channel.Client, ledger *ledger.Client, withCert bool) func(*testing.T) {
	return func(t *testing.T) {
		var resp channel.Response
		var err error
		if withCert {
			resp, err = client.Execute(invoke, channelPeers)
		} else {
			resp, err = client.Execute(invoke, channel.WithoutCert(), channelPeers)
		}
		require.NoError(t, err)
		checkEndorsersResponse(t, resp.Responses, false)
		time.Sleep(time.Second * 3)
		t.Run("Check certificates in transaction", checkTX(ledger, resp.TransactionID, withCert))
	}
}

func checkTX(ledger *ledger.Client, txID fab.TransactionID, creatorCert bool) func(*testing.T) {
	return func(t *testing.T) {
		tx, err := ledger.QueryTransaction(txID, ledgerPeers)
		require.NoError(t, err)
		pld := &common.Payload{}
		err = proto.Unmarshal(tx.TransactionEnvelope.Payload, pld)
		require.NoError(t, err)
		checkSignatureHeader(t, pld.Header.SignatureHeader, creatorCert)
		txd := &peer.Transaction{}
		err = proto.Unmarshal(pld.Data, txd)
		require.NoError(t, err)
		ta := txd.Actions[0]
		checkSignatureHeader(t, ta.Header, creatorCert)
		ca := &peer.ChaincodeActionPayload{}
		err = proto.Unmarshal(ta.Payload, ca)
		require.NoError(t, err)

		for _, e := range ca.Action.Endorsements {
			checkSerializedIdentity(t, e.Endorser, false)
		}
	}
}

func checkEndorsersResponse(t *testing.T, resp []*fab.TransactionProposalResponse, withCert bool) {
	for _, r := range resp {
		require.Equal(t, r.Status, int32(200))
		require.Equal(t, r.ChaincodeStatus, int32(200))
		checkSerializedIdentity(t, r.Endorsement.Endorser, withCert)
	}
}

func checkSignatureHeader(t *testing.T, data []byte, withCert bool) {
	sh := &common.SignatureHeader{}
	err := proto.Unmarshal(data, sh)
	require.NoError(t, err)
	checkSerializedIdentity(t, sh.Creator, withCert)
}

func checkSerializedIdentity(t *testing.T, data []byte, withCert bool) {
	si := &certstore.SerializedIdentityWithRef{}
	err := proto.Unmarshal(data, si)
	require.NoError(t, err)
	if withCert {
		require.NotEmpty(t, si.IdBytes)
		return
	}
	require.Empty(t, si.IdBytes)
	require.NotEmpty(t, si.IdRef)
}
