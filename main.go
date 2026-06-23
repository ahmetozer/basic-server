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
	"path/filepath"
	"runtime"

	certmngr "github.com/ahmetozer/basic-server/pkg"
)

var (
	RunningEnv string = ""
)

// certDir returns the directory used for self-generated temporary
// certificates. It prefers ~/.basic-server, creating it if needed.
// If the home directory cannot be resolved, the directory cannot be
// created, or it is not writable, it falls back to the working directory.
func certDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Printf("Cannot resolve home directory, using working directory for certs: %v", err)
		return "."
	}

	dir := filepath.Join(home, ".basic-server")

	// If certs are already in place, use the directory as-is without
	// recreating it or probing for writability.
	_, certErr := os.Stat(filepath.Join(dir, "cert.pem"))
	_, keyErr := os.Stat(filepath.Join(dir, "key.pem"))
	if certErr == nil && keyErr == nil {
		return dir
	}

	if err := os.MkdirAll(dir, 0700); err != nil {
		log.Printf("Cannot create %s, using working directory for certs: %v", dir, err)
		return "."
	}

	// Verify the directory is writable before committing to it.
	probe := filepath.Join(dir, ".write-test")
	f, err := os.OpenFile(probe, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Printf("%s is not writable, using working directory for certs: %v", dir, err)
		return "."
	}
	f.Close()
	os.Remove(probe)

	return dir
}

func main() {
	log.Printf("github.com/ahmetozer/basic-server")
	var currcertDir, currentPath string
	var err error
	if RunningEnv != "container" {
		currentPath, err = os.Getwd()
		if err != nil {
			log.Println(err)
		}
		currcertDir = certDir()
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
