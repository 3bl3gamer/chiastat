package network

import (
	"chiastat/chia/types"
	"chiastat/chia/utils"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"log"
	"os"
	"sync"
	"time"

	"github.com/ansel1/merry"
	"github.com/gorilla/websocket"
)

const NETWORK_ID = "mainnet"
const PROTOCOL_VERSION = "0.0.32"
const SOFTWARE_VERSION = "1.1.6"

func MakeTSLConfigFromFiles(caCertPath, nodeCertPath, nodeKeyPath string) (*tls.Config, error) {
	caCertBuf, err := os.ReadFile(caCertPath)
	if err != nil {
		return nil, merry.Wrap(err)
	}
	nodeCertBuf, err := os.ReadFile(nodeCertPath)
	if err != nil {
		return nil, merry.Wrap(err)
	}
	nodeKeyBuf, err := os.ReadFile(nodeKeyPath)
	if err != nil {
		return nil, merry.Wrap(err)
	}
	return MakeTSLConfigFromBytes(caCertBuf, nodeCertBuf, nodeKeyBuf)
}

func MakeTSLConfigFromBytes(caCertBuf, nodeCertBuf, nodeKeyBuf []byte) (*tls.Config, error) {
	data, _ := pem.Decode(caCertBuf)
	caCert, err := x509.ParseCertificate(data.Bytes)
	if err != nil {
		return nil, merry.Wrap(err)
	}
	rootCAs := x509.NewCertPool()
	rootCAs.AddCert(caCert)

	nodeCert, err := tls.X509KeyPair(nodeCertBuf, nodeKeyBuf)
	if err != nil {
		return nil, merry.Wrap(err)
	}

	return MakeTSLConfig(rootCAs, nodeCert), nil
}

func MakeTSLConfig(rootCAs *x509.CertPool, nodeCert tls.Certificate) *tls.Config {
	return &tls.Config{
		RootCAs:            rootCAs,
		Certificates:       []tls.Certificate{nodeCert},
		InsecureSkipVerify: true,
		VerifyPeerCertificate: func(certificates [][]byte, verifiedChains [][]*x509.Certificate) error {
			certs := make([]*x509.Certificate, len(certificates))
			for i, asn1Data := range certificates {
				// errors should be already checked in
				// https://github.com/golang/gofrontend/blob/master/libgo/go/crypto/tls/handshake_client.go
				// at verifyServerCertificate()
				cert, _ := x509.ParseCertificate(asn1Data)
				certs[i] = cert
			}

			opts := x509.VerifyOptions{
				Roots:       rootCAs,
				CurrentTime: time.Now(),
				// DNSName:       c.config.ServerName,
				Intermediates: x509.NewCertPool(),
			}
			for _, cert := range certs[1:] {
				opts.Intermediates.AddCert(cert)
			}
			_, err := certs[0].Verify(opts)
			return err
		},
	}
}

type Result struct {
	Data utils.FromBytes
	Err  error
}

// https://github.com/Chia-Network/chia-blockchain/blob/latest/chia/server/ws_connection.py
type WSChiaConnection struct {
	peerID           [32]byte
	ws               *websocket.Conn
	isOutbound       bool
	lastRequestNonce uint16
	pendingRequests  map[uint16]chan Result
	closeErr         error
	mutex            *sync.Mutex
}

func ConnectTo(tlsConfig *tls.Config, address string) (*WSChiaConnection, error) {
	dialer := websocket.Dialer{
		HandshakeTimeout: 5 * time.Second,
		TLSClientConfig:  tlsConfig,
	}

	c, _, err := dialer.Dial("wss://"+address+"/ws", nil)
	if err != nil {
		return nil, merry.Wrap(err)
	}

	certs := c.UnderlyingConn().(*tls.Conn).ConnectionState().PeerCertificates
	peerID := sha256.Sum256(certs[0].Raw)

	return &WSChiaConnection{
		peerID:          peerID,
		ws:              c,
		isOutbound:      true,
		pendingRequests: make(map[uint16]chan Result),
		mutex:           &sync.Mutex{},
	}, nil
}

func (c WSChiaConnection) PerformHandshake() (*types.Handshake, error) {
	msgOut := types.Message{
		Type: types.MSG_HANDSHAKE,
		Data: utils.ToByteSlice(types.Handshake{
			NetworkID:       NETWORK_ID,
			ProtocolVersion: PROTOCOL_VERSION,
			SoftwareVersion: SOFTWARE_VERSION,
			ServerPort:      8444,
			NodeType:        types.NODE_FULL,
			Capabilities:    []types.TupleUint16Str{{V0: types.CAP_BASE, V1: "1"}},
		}),
	}

	err := c.ws.WriteMessage(websocket.BinaryMessage, utils.ToByteSlice(msgOut))
	if err != nil {
		return nil, merry.Wrap(err)
	}

	msgType, buf, err := c.ws.ReadMessage()
	if err != nil {
		return nil, merry.Wrap(err)
	}

	if msgType != websocket.BinaryMessage {
		return nil, merry.Errorf("unexpected WS mesage type: expected binary(%d), got %d",
			websocket.BinaryMessage, msgType)
	}

	var msgIn types.Message
	if err := utils.FromByteSliceExact(buf, &msgIn); err != nil {
		return nil, merry.Wrap(err)
	}

	if msgIn.Type != types.MSG_HANDSHAKE {
		return nil, merry.Errorf("unexpected message type: expected handshake(%d), got %d",
			types.MSG_HANDSHAKE, msgIn.Type)
	}

	var hs types.Handshake
	if err := utils.FromByteSliceExact(msgIn.Data, &hs); err != nil {
		return nil, merry.Wrap(err)
	}

	return &hs, nil
}

func (c *WSChiaConnection) StartRoutines() {
	go c.readRoutine()
}

func (c *WSChiaConnection) CloseWithErr(err error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.closeErr == nil {
		c.closeErr = err
		if err := c.ws.Close(); err != nil {
			log.Printf("WARN: closing connection: %s", err)
		}
	}
	log.Printf("DEBUG: closing with: %s", err)

	for msgID, resChan := range c.pendingRequests {
		resChan <- Result{Err: c.closeErr}
		close(resChan)
		delete(c.pendingRequests, msgID)
	}
}

func (c *WSChiaConnection) readRoutine() {
	for {
		if c.closeErr != nil {
			break
		}
		msgType, buf, err := c.ws.ReadMessage()
		if err != nil {
			c.CloseWithErr(err)
			break
		}
		if msgType != websocket.BinaryMessage {
			err := merry.Errorf("WARN: unexpected WS mesage type: expected binary(%d), got %d",
				websocket.BinaryMessage, msgType)
			c.CloseWithErr(err)
			break
		}
		if err := c.processMessage(buf); err != nil {
			c.CloseWithErr(err)
			break
		}
	}
}

func (c *WSChiaConnection) processMessage(msgBuf []byte) error {
	var msg types.Message
	if err := utils.FromByteSliceExact(msgBuf, &msg); err != nil {
		return merry.Wrap(err)
	}
	switch msg.Type {
	case types.MSG_RESPOND_PEERS:
		var peers types.RespondPeers
		if err := utils.FromByteSliceExact(msg.Data, &peers); err != nil {
			return merry.Wrap(err)
		}
		if err := c.handleResponse(msg.ID, &peers); err != nil {
			return merry.Wrap(err)
		}
	default:
		log.Printf("WARN: unsupported message type: %d", msg.Type)
	}
	return nil
}

func (c *WSChiaConnection) handleResponse(msgID uint16, data utils.FromBytes) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	resChan, ok := c.pendingRequests[msgID]
	if !ok {
		return merry.Errorf("unexpected message ID: %d", msgID)
	}
	delete(c.pendingRequests, msgID)

	resChan <- Result{Data: data}
	close(resChan)
	return nil
}

func (c *WSChiaConnection) Send(msgType uint8, request utils.ToBytes) chan Result {
	c.mutex.Lock()

	// The request nonce is an integer between 0 and 2**16 - 1, which is used to match requests to responses
	// If is_outbound, 1 <= nonce < 2^15, else  2^15 <= nonce < 2^16
	// (nonce=0 is not used, differs from https://github.com/Chia-Network/chia-blockchain/blob/latest/chia/server/ws_connection.py)
	if c.isOutbound {
		if c.lastRequestNonce > 0 && c.lastRequestNonce < 1<<15-1 {
			c.lastRequestNonce += 1
		} else {
			c.lastRequestNonce = 1
		}
	} else {
		if c.lastRequestNonce > 0 && c.lastRequestNonce < 1<<16-1 {
			c.lastRequestNonce += 1
		} else {
			c.lastRequestNonce = 1 << 15
		}
	}
	msg := types.Message{
		Type: msgType,
		ID:   c.lastRequestNonce,
		Data: utils.ToByteSlice(request),
	}
	respChan := make(chan Result, 1)
	c.pendingRequests[msg.ID] = respChan
	c.mutex.Unlock()

	err := c.ws.WriteMessage(websocket.BinaryMessage, utils.ToByteSlice(msg))
	if err != nil {
		c.CloseWithErr(err)
	}
	return respChan
}

func (c *WSChiaConnection) SendSync(msgType uint8, request utils.ToBytes) (utils.FromBytes, error) {
	respChan := c.Send(msgType, request)
	res := <-respChan
	if res.Err != nil {
		return nil, merry.Wrap(res.Err)
	}
	return res.Data, nil
}

func (c *WSChiaConnection) RequestPeers() (*types.RespondPeers, error) {
	respChan := c.Send(types.MSG_REQUEST_PEERS, types.RequestPeers{})
	res := <-respChan
	if res.Err != nil {
		return nil, merry.Wrap(res.Err)
	}
	peers, ok := res.Data.(*types.RespondPeers)
	if !ok {
		return nil, utils.WrongRespError(res.Data)
	}
	return peers, nil
}
