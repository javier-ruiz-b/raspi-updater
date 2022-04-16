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
	Server(address, "images")
}

func Server(address string, imagesDir string) {
	listen(address)
}

func listen(address string) error {
	// listener, err := net.Listen("tcp", address)
	// if err != nil {
	// 	log.Fatalln(err)
	// }
	// defer listener.Close()

	// nlog.Debug("Listening on ", address)
	// for {
	// 	con, err := listener.Accept()
	// 	if err != nil {
	// 		log.Println(err)
	// 		continue
	// 	}

	// 	go handleClientRequest(con)
	// }
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

	handler := newHandler() //www

	server := http3.Server{
		Server:     &http.Server{Handler: handler, Addr: address},
		QuicConfig: quicConf,
	}
	err := server.ListenAndServeTLS(testdata.GetCertificatePaths())

	return err
}

// func handleClientRequest(con net.Conn) {
// 	defer con.Close()

// 	enc := gob.NewEncoder(bufio.NewWriter(con))
// 	dec := gob.NewDecoder(bufio.NewReader(con))

// 	parser := NewClientParser(enc, dec)

// 	for {

// 		var clientHello protocol.Hello
// 		err := dec.Decode(&clientHello)
// 		if err != nil {
// 			nlog.Error("Failed reading hello packet ", clientHello)
// 			return
// 		}
// 		nlog.Debug("Client hello:", clientHello.Id, " version ", clientHello.Version())

// 		// serverHello := protocol.NewHello("server")

// 		// switch err {
// 		// case nil:
// 		// 	clientRequest := strings.TrimSpace(clientRequest)
// 		// 	if clientRequest == ":QUIT" {
// 		// 		log.Println("client requested server to close the connection so closing")
// 		// 		return
// 		// 	} else {
// 		// 		log.Println(clientRequest)
// 		// 	}
// 		// case io.EOF:
// 		// 	log.Println("client closed the connection by terminating the process")
// 		// 	return
// 		// default:
// 		// 	log.Printf("error: %v\n", err)
// 		// 	return
// 		// }

// 		// // Responding to the client request
// 		// if _, err = con.Write([]byte("GOT IT!\n")); err != nil {
// 		// 	log.Printf("failed to respond to client: %v\n", err)
// 		// }
// 	}
//}
