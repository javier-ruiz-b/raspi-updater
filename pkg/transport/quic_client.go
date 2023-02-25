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

	"github.com/javier-ruiz-b/raspi-image-updater/pkg/nlog"
	"github.com/javier-ruiz-b/raspi-image-updater/pkg/testdata"
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

func NewQuicClient(address string, qlogs bool) Client {
	return newClient(newQuicClient(address, qlogs))
}

func newQuicClient(address string, qlogs bool) transportClient {
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
	testdata.AddRootCA(pool)

	roundTripper := &http3.RoundTripper{
		TLSClientConfig: &tls.Config{
			RootCAs:            pool,
			InsecureSkipVerify: false,
		},
		QuicConfig: &qconf,
	}
	hclient := &http.Client{
		Transport: roundTripper,
	}

	if !strings.HasPrefix(address, "https://") {
		address = "https://" + address
	}

	return &QuicClient{
		address:      address,
		client:       hclient,
		roundTripper: roundTripper,
	}
}

func (c *QuicClient) Close() {
	c.roundTripper.Close()
}

func (c *QuicClient) Get(url string) (*http.Response, error) {
	url = c.address + url
	nlog.Debug("Get ", url)
	return c.client.Get(url)
}
