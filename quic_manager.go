package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"time"

	"github.com/quic-go/quic-go"
)

type QUICManager struct {
	Listener quic.Listener
	Conn     quic.Connection
}

func InitServer(addr string) (*QUICManager, error) {
	tlsConf := &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         []string{"echo-file-transfer-v1"},
	}

	cert, err := genereateCert()
	if err != nil {
		return nil, err
	}

	tlsConf.Certificates = []tls.Certificate{cert}
	listener, err := quic.ListenAddr(addr, tlsConf, nil)
	if err != nil {
		return nil, err
	}

	return &QUICManager{Listener: *listener}, nil
}

func InitClient(addr string) (*QUICManager, error) {
	tlsConf := &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         []string{"echo-file-transfer-v1"},
	}

	conn, err := quic.DialAddr(context.Background(), addr, tlsConf, nil)
	if err != nil {
		return nil, err
	}

	return &QUICManager{Conn: conn}, nil
}

func (qm *QUICManager) AcceptConnection(ctx context.Context) (quic.Connection, error) {
	return qm.Listener.Accept(ctx)
}

func (qm *QUICManager) GetConnection() quic.Connection {
	return qm.Conn
}

func (qm *QUICManager) Close() error {
	return qm.Listener.Close()
}

func genereateCert() (tls.Certificate, error) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return tls.Certificate{}, err
	}

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"Echo File Transfer"},
		},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().Add(365 * 25 * time.Hour),
		KeyUsage:    x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}

	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		return tls.Certificate{}, err
	}

	return tls.Certificate{
		Certificate: [][]byte{certDER},
		PrivateKey:  key,
	}, nil
}
