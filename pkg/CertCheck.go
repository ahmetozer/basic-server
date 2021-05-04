package pkg

import (
	"log"
	"os"
)

func (config *CertConfig) CertCheck() error {

	var err error
	certCheckOk := false

	if _, err := os.Stat(config.KeyLocation); err == nil {
		log.Printf(config.KeyLocation + " exist\n")
		certCheckOk = true
	} else {
		log.Printf(config.KeyLocation + " not exist\n")
		certCheckOk = false
	}
	if certCheckOk {
		if _, err := os.Stat(config.CertLocation); err == nil {
			log.Printf(config.CertLocation + " exist\n")
			certCheckOk = true
		} else {
			log.Printf(config.CertLocation + " not exist\n")
			certCheckOk = false
		}
	}

	if !certCheckOk {
		log.Printf("Self created certs will be used\n")
		config.Defaults()
		err = config.Generate()
		if err != nil {
			return err
		}
	}
	return nil

}
