package main

import (
	"fmt"
	"log"
	"net"

	"crypto/tls"

	"crypto/x509"
	"io/ioutil"

	"math"
        "time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// Person ...
type Person struct {
	Hero  string
	Skill int
}

func preferSslConnect() (*mgo.Session, error) {
	fmt.Println("Load CA file")

	// --sslCAFile
	rootCerts := x509.NewCertPool()
	if ca, err := ioutil.ReadFile("mongoCA.crt"); err == nil {
		rootCerts.AppendCertsFromPEM(ca)
	}

	fmt.Println("Load Cert file")

	// --sslPEMKeyFile
	clientCerts := []tls.Certificate{}
	if cert, err := tls.LoadX509KeyPair("client.crt", "client.key"); err == nil {
		clientCerts = append(clientCerts, cert)
	}

	fmt.Println("Dial conn")

	// Dial with TLS
	session, err := mgo.DialWithInfo(&mgo.DialInfo{
		Addrs:    []string{"mongos-nww.btpns.com:27017"},
		Timeout:  60 * time.Second,
		Database: "mydb",
		Username: "admin",
		Password: "Karimun!@34",
		DialServer: func(addr *mgo.ServerAddr) (net.Conn, error) {
			return tls.Dial("tcp", addr.String(), &tls.Config{
				RootCAs:      rootCerts,
				Certificates: clientCerts,
				InsecureSkipVerify: true,
			})
		},
	})

	return session, err
}

func noSslConnect() (*mgo.Session, error) {
	session, err := mgo.Dial("mongos-nww.btpns.com:27017")
	if err != nil {
		panic(err)
	}
	return session, err
}

func main() {
	// Connect
	session, err := preferSslConnect()
	if err != nil {
		log.Fatal(err)
	}
	defer session.Close()

	// Optional. Switch the session to a monotonic behavior.
	session.SetMode(mgo.Monotonic, true)

	// Drop collection
	c := session.DB("mydb").C("mgo_test")
	c.DropCollection()

	// Insert a struct
	err = c.Insert(Person{"John Rambo", 90})
	if err != nil {
		log.Fatal(err)
	}

	// Insert a series of docs
	docs := make([]interface{}, 5)
	for i := 0; i < 5; i++ {
		docs[i] = bson.M{"a": i, "b": i * i, "c": math.Pow(2, float64(i))}
	}
	err = c.Insert(docs...)
	if err != nil {
		log.Fatal(err)
	}

	// Get results
	result := Person{}
	err = c.Find(bson.M{"hero": "John Rambo"}).One(&result)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(result)
}
