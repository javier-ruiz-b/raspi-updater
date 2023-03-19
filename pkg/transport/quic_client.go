package transport

import (
	"bufio"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/javier-ruiz-b/raspi-image-updater/pkg/config"
	"github.com/javier-ruiz-b/raspi-image-updater/pkg/utils"
	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/http3"
	"github.com/quic-go/quic-go/logging"
	"github.com/quic-go/quic-go/qlog"
)

type QuicClient struct {
	address      string
	client       *http.Client
	roundTripper *http3.RoundTripper
}

func NewQuicClient(config *config.ClientConfig) Client {
	return newClient(newQuicClient(*config.Address, *config.Log, *config.CertificatePath))
}

func newQuicClient(address string, qlogs bool, certPath string) *QuicClient {
	var qconf quic.Config
	if qlogs {
		qconf.Tracer = qlog.NewTracer(func(_ logging.Perspective, connID []byte) io.WriteCloser {
			filename := fmt.Sprintf("client_%x.qlog", connID)
			f, err := os.Create(filename)
			if err != nil {
				log.Fatal(err)
			}
			log.Printf("Creating qlog file %s.\n", filename)
			return utils.NewBufferedWriteCloser(bufio.NewWriter(f), f)
		})
	}

	pool, err := x509.SystemCertPool()
	if err != nil {
		log.Fatal(err)
	}
	AddRootCA(pool, certPath)

	roundTripper := &http3.RoundTripper{
		TLSClientConfig: &tls.Config{
			RootCAs:            pool,
			InsecureSkipVerify: false,
		},
		QuicConfig: &qconf,
	}
	client := &http.Client{
		Transport: roundTripper,
	}

	if !strings.HasPrefix(address, "https://") {
		address = "https://" + address
	}

	return &QuicClient{
		address:      address,
		client:       client,
		roundTripper: roundTripper,
	}
}

func (c *QuicClient) Close() {
	c.roundTripper.Close()
}

func (c *QuicClient) NewRequest(method, url string, body io.Reader) (*http.Request, error) {
	return http.NewRequest(method, c.address+url, body)
}

func (c *QuicClient) Get(url string) (*http.Response, error) {
	return c.client.Get(c.address + url)
}

func (c *QuicClient) Do(req *http.Request) (*http.Response, error) {
	return c.client.Do(req)
}

// AddRootCA adds the root CA certificate to a cert pool
func AddRootCA(certPool *x509.CertPool, certPath string) error {
	caCertRaw, err := os.ReadFile(certPath)
	if err != nil {
		return err
	}
	if ok := certPool.AppendCertsFromPEM(caCertRaw); !ok {
		return fmt.Errorf("could not add root ceritificate %s to pool", certPath)
	}
	return nil
}
