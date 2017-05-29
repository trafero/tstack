package tls

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"log"
)

func TLSConfig(cacrt string, clientcrt string, clientkey string) (*tls.Config, error) {
	// Import trusted certificates from CAfile.pem.
	// Alternatively, manually add CA certificates to
	// default openssl CA bundle.
	certpool := x509.NewCertPool()
	log.Printf("Using certificate %s\n", cacrt)
	certbytes, err := ioutil.ReadFile(cacrt)
	if err != nil {
		return nil, err
	}
	certpool.AppendCertsFromPEM(certbytes)

	// Import client certificate/key pair
	cert, err := tls.LoadX509KeyPair(clientcrt, clientkey)
	if err != nil {
		return nil, err
	}

	// Just to print out the client certificate..
	//certInfo, err := x509.ParseCertificate(cert.Certificate[0])
	//checkErr(err)
	//fmt.Printf("certinfo: ",certInfo)

	// Create tls.Config with desired tls properties
	return &tls.Config{
		// RootCAs = certs used to verify server cert.
		RootCAs: certpool,
		// ClientAuth = whether to request cert from server.
		// Since the server is set up for SSL, this happens
		// anyways.
		ClientAuth: tls.NoClientCert,
		// ClientCAs = certs used to validate client cert.
		ClientCAs: nil,
		// InsecureSkipVerify = verify that cert contents
		// match server. IP matches what is in cert etc.
		InsecureSkipVerify: false,
		// Certificates = list of certs client sends to server.
		Certificates: []tls.Certificate{cert},
	}, nil
}
