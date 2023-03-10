// pulp package pruner
package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/pulp"
)

var (
	hostname    = flag.String("hostname", "", "API url to make requests to.")
	username    = flag.String("username", "", "Required for Basic Auth.")
	password    = flag.String("password", "", "Required for Basic Auth.")
	certificate = flag.String("certificate", "", "CA certificate in DER format.")

	dry         = flag.Bool("dry-run", false, "Execute without deletion.")
	keep        = flag.Int("keep", 3, "How many latest minor versions not to delete.")
	repository  = flag.String("repository", "", "Specify repository to interact with.")
	packageType = flag.String("package", "", "Specify package type.")
)

func main() {
	flag.Parse()

	cert, err := os.ReadFile(*certificate)
	if err != nil {
		log.Fatalln("read cert", err)
	}

	client, err := pulp.Login(*hostname, *username, *password, cert)
	if err != nil {
		log.Fatalln("login", err)
	}

	var (
		listFunc func(*http.Client, string, string, uint) ([]string, error)
		rmFunc   func(*http.Client, string, string, string) error
	)
	switch *packageType {
	case "deb":
		listFunc = pulp.Debs
		rmFunc = pulp.RemoveDeb
	case "rpm":
		listFunc = pulp.Rpms
		rmFunc = pulp.RemoveRpm
	default:
		log.Fatalln("unknown package type:", *packageType)
	}

	if *dry {
		rmFunc = func(*http.Client, string, string, string) error { return nil }
	}

	versions, err := listFunc(client, *hostname, *repository, uint(*keep))
	if err != nil {
		log.Fatalln(*packageType, err)
	}
	log.Println(internal.InfoPrefix, "versions to delete:", strings.Join(versions, " "))

	for _, v := range versions {
		log.Println(internal.DebugPrefix, "deleting", v, "from", *repository)
		err := rmFunc(client, *hostname, *repository, v)
		if err != nil {
			log.Println(internal.ErrorPrefix, err)
		}
	}
}
