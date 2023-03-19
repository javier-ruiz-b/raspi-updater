package testdata

import (
	"crypto/tls"
	"crypto/x509"
	"io"

	"github.com/javier-ruiz-b/raspi-image-updater/pkg/transport"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("certificates", func() {
	It("returns certificates", func() {
		ln, err := tls.Listen("tcp", "localhost:4433", GetTLSConfig())
		Expect(err).ToNot(HaveOccurred())

		go func() {
			defer GinkgoRecover()
			conn, err := ln.Accept()
			Expect(err).ToNot(HaveOccurred())
			defer conn.Close()
			_, err = conn.Write([]byte("foobar"))
			Expect(err).ToNot(HaveOccurred())
		}()

		pool := x509.NewCertPool()
		cert, _ := GetCertificatePaths()
		transport.AddRootCA(pool, cert)

		conn, err := tls.Dial("tcp", "localhost:4433", &tls.Config{RootCAs: pool})
		Expect(err).ToNot(HaveOccurred())
		data, err := io.ReadAll(conn)
		Expect(err).ToNot(HaveOccurred())
		Expect(string(data)).To(Equal("foobar"))
	})
})
