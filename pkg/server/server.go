package server

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/javier-ruiz-b/raspi-image-updater/pkg/testdata"
	"github.com/javier-ruiz-b/raspi-image-updater/pkg/utils"
	"github.com/javier-ruiz-b/raspi-image-updater/pkg/version"
	"github.com/lucas-clemente/quic-go"
	"github.com/lucas-clemente/quic-go/http3"
	"github.com/lucas-clemente/quic-go/logging"
	"github.com/lucas-clemente/quic-go/qlog"
)

func Main(port int) {
	fmt.Println("Server", version.VERSION)
	var ()
	flag.Parse()

	log.Print("Port: ", port)

	address := "0.0.0.0:" + strconv.Itoa(int(port))
	server := NewServer(address, "images")
	err := server.Listen()
	if err != io.EOF && err != nil {
		log.Print("Server error: ", err)
		os.Exit(1)
	}
}

type Server struct {
	address   string
	imagesDir string
	server    *http3.Server
}

func NewServer(address string, imagesDir string) *Server {
	enableQlog := true

	quicConf := &quic.Config{}
	if enableQlog {
		quicConf.Tracer = qlog.NewTracer(func(_ logging.Perspective, connID []byte) io.WriteCloser {
			filename := fmt.Sprintf("server_%x.qlog", connID)
			f, err := os.Create(filename)
			if err != nil {
				log.Fatal(err)
			}
			log.Printf("Creating qlog file %s.\n", filename)
			return utils.NewBufferedWriteCloser(bufio.NewWriter(f), f)
		})
	}

	handler := newHandler("bin")

	server := &http3.Server{
		Server:     &http.Server{Handler: handler, Addr: address},
		QuicConfig: quicConf,
	}

	return &Server{
		address:   address,
		imagesDir: imagesDir,
		server:    server,
	}
}

func (s *Server) Close() {
	s.server.Close()
}

func (s *Server) Listen() error {
	return s.server.ListenAndServeTLS(testdata.GetCertificatePaths())
}
