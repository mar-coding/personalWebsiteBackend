package configHandler

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"github.com/mar-coding/personalWebsiteBackend/pkg/unmarshaller"
	"google.golang.org/grpc/credentials"
	"io/ioutil"
)

func New[Extra any](configPath string) (*BaseConfig[Extra], error) {
	return newUnmarshal[Extra](configPath)
}

func newUnmarshal[Extra any](configPath string) (*BaseConfig[Extra], error) {
	cfg := new(BaseConfig[Extra])

	unmarshal, err := unmarshaller.NewUnmarshaller(configPath)
	if err != nil {
		return nil, err
	}

	return cfg, unmarshal.Unmarshal(cfg)
}

// LoadGrpcServerCredentials create transport credential for grpc server for TLS handshake
func (c *BaseConfig[T]) LoadGrpcServerCredentials() (credentials.TransportCredentials, error) {
	// Load server's certificate and private key
	serverCert, err := tls.LoadX509KeyPair(c.Grpc.CertFilePath, c.Grpc.CertKeyFilePath)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("service config: load x509 key pair got error %s", err.Error()))
	}

	// Create the credentials and return it
	config := &tls.Config{
		Certificates: []tls.Certificate{serverCert},
		ClientAuth:   tls.NoClientCert,
	}

	return credentials.NewTLS(config), nil
}

// LoadGrpcClientCredentials create transport credential for grpc client to access server side for TLS handshake
func (c *BaseConfig[T]) LoadGrpcClientCredentials(client *GrpcClient) (credentials.TransportCredentials, error) {
	if client == nil {
		return nil, errors.New("service config: client is nil")
	}

	// Load certificate of the CA who signed server's certificate
	pemServerCA, err := ioutil.ReadFile(client.CertCAFilePath)
	if err != nil {
		return nil, err
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(pemServerCA) {
		return nil, errors.New("service config: failed to add server CA's certificate")
	}

	// Create the credentials and return it
	config := &tls.Config{
		RootCAs: certPool,
	}

	return credentials.NewTLS(config), nil
}
