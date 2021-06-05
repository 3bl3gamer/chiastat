package network

import (
	"chiastat/chia/types"
	"chiastat/chia/utils"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"os"
	"time"

	"github.com/ansel1/merry"
	"github.com/gorilla/websocket"
)

const NETWORK_ID = "mainnet"
const PROTOCOL_VERSION = "0.0.32"
const SOFTWARE_VERSION = "1.1.6"

type WSChiaConnection struct {
	PeerID [32]byte
	WS     *websocket.Conn
}

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

	return &WSChiaConnection{PeerID: peerID, WS: c}, nil
}

func (c *WSChiaConnection) PerformHandshake() (*types.Handshake, error) {
	msg := types.Message{
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

	err := c.WS.WriteMessage(websocket.BinaryMessage, utils.ToByteSlice(msg))
	if err != nil {
		return nil, merry.Wrap(err)
	}

	msgType, buf, err := c.WS.ReadMessage()
	if err != nil {
		return nil, merry.Wrap(err)
	}

	if msgType != websocket.BinaryMessage {
		return nil, merry.Errorf("unexpected WS mesage type: expected binary(%d), got %d",
			websocket.BinaryMessage, msgType)
	}

	pBuf := utils.NewParseBuf(buf)
	msg = types.MessageFromBytes(pBuf)
	pBuf.EnsureEmpty()
	if pBuf.Err() != nil {
		return nil, merry.Wrap(pBuf.Err())
	}

	if msg.Type != types.MSG_HANDSHAKE {
		return nil, merry.Errorf("unexpected message type: expected handshake(%d), got %d",
			types.MSG_HANDSHAKE, msg.Type)
	}

	pBuf = utils.NewParseBuf(msg.Data)
	hs := types.HandshakeFromBytes(pBuf)
	pBuf.EnsureEmpty()
	if pBuf.Err() != nil {
		return nil, merry.Wrap(pBuf.Err())
	}

	return &hs, nil
}
