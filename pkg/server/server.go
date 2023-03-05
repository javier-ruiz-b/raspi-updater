package server

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/javier-ruiz-b/raspi-image-updater/pkg/config"
	"github.com/javier-ruiz-b/raspi-image-updater/pkg/testdata"
	"github.com/javier-ruiz-b/raspi-image-updater/pkg/utils"
	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/http3"
	"github.com/quic-go/quic-go/logging"
	"github.com/quic-go/quic-go/qlog"
)

type Server struct {
	options *config.ServerConfig
	server  *http3.Server
}

func NewServer(options *config.ServerConfig) *Server {
	quicConf := &quic.Config{}
	if *options.Log {
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

	server := &http3.Server{
		Handler:    newMainHandler(options),
		Addr:       *options.Address,
		QuicConfig: quicConf,
	}

	return &Server{
		options: options,
		server:  server,
	}
}

func (s *Server) Close() {
	s.server.Close()
}

func (s *Server) Listen() error {
	log.Print("Listening on ", *s.options.Address)
	return s.server.ListenAndServeTLS(testdata.GetCertificatePaths())
}
