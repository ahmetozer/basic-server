package main

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"os/user"
	"runtime"

	certmngr "github.com/ahmetozer/basic-server/pkg"
)

var (
	RunningEnv string = ""
)

func main() {
	log.Printf("github.com/ahmetozer/basic-server")
	var currcertDir, currentPath string
	var err error
	if RunningEnv != "container" {
		currentPath, err = os.Getwd()
		if err != nil {
			log.Println(err)
		}
		currcertDir = "."
	} else {
		currcertDir = "/tmp/cert"
		currentPath = "/web/"
	}

	path := flag.String("path", currentPath, "Share Path")
	listen := flag.String("listen", ":443", "Listen addr")
	serverName := flag.String("server-name", "", "Server name check from TLS")
	clientCert := flag.String("client-cert", "", "Allow only given client certificate")
	serverCert := flag.String("server-cert", "", "Server cert")
	serverKey := flag.String("server-key", "", "Server key")
	flag.Parse()

	if runtime.GOOS == "linux" {
		log.Printf("Current user id %v\n", os.Getuid())
		log.Printf("Current user gid %v\n", os.Getgid())
	} else if runtime.GOOS == "windows" {
		user, err := user.Current()
		if err != nil {
			panic(err)
		}
		log.Printf("Current user %s\n", user.Username)
	}

	certConfig := certmngr.CertConfig{}

	if *serverCert == "" {
		certConfig.CertLocation = currcertDir + "/cert.pem"
	}
	if *serverKey == "" {
		certConfig.KeyLocation = currcertDir + "/key.pem"
	}

	if *serverName != "" {
		certConfig.Hosts = append(certConfig.Hosts, *serverName)
	}
	err = certConfig.CertCheck()
	if err != nil {
		log.Fatalf("Err while creating Cert %s", err)
	}

	httpServer := &http.Server{
		Addr: *listen,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if *serverName != "" {
				if *serverName != r.TLS.ServerName {
					w.WriteHeader(http.StatusBadRequest)
					fmt.Fprintf(w, "The request server name is invalid.")
					return
				}
			}

			if r.RequestURI == "/dump" {
				requestDump, err := httputil.DumpRequest(r, true)
				if err != nil {
					fmt.Println(err)
				}
				fmt.Fprintf(w, fmt.Sprintf("%s", requestDump))
				return
			}
			http.FileServer(http.Dir(*path)).ServeHTTP(w, r)
		}),
	}

	if *clientCert != "" {
		caCert, err := ioutil.ReadFile(*clientCert)
		if err != nil {
			log.Fatal(err)
		}

		roots := x509.NewCertPool()
		ok := roots.AppendCertsFromPEM(caCert)
		if !ok {
			log.Fatal("failed to parse root certificate")
		}
		httpServer.TLSConfig = &tls.Config{
			ClientAuth: tls.RequireAndVerifyClientCert,
			ClientCAs:  roots,
		}

	}

	if *serverName == "" {
		log.Printf("WARN: Flag server-name is not set. The system allows all server names.")
	}
	log.Printf("Starting HTTPS server on %s at %s\n", *listen, *path)
	log.Fatal(httpServer.ListenAndServeTLS(certConfig.CertLocation, certConfig.KeyLocation))

}
