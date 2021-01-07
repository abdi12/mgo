package main

import (
    "crypto/tls"
    "crypto/x509"
    "fmt"
    "gopkg.in/mgo.v2"
    "io/ioutil"
    "net"
    "time"
)

type Game struct {
    Winner       string    `bson:"winner"`
    OfficialGame bool      `bson:"official_game"`
    Location     string    `bson:"location"`
    StartTime    time.Time `bson:"start"`
    EndTime      time.Time `bson:"end"`
    Players      []Player  `bson:"players"`
}

type Player struct {
    Name   string    `bson:"name"`
    Decks  [2]string `bson:"decks"`
    Points uint8     `bson:"points"`
    Place  uint8     `bson:"place"`
}

func NewPlayer(name, firstDeck, secondDeck string, points, place uint8) Player {
    return Player{
        Name:   name,
        Decks:  [2]string{firstDeck, secondDeck},
        Points: points,
        Place:  place,
    }
}

func main() {
    const (
        Host       = "mongos-nww.btpns.com:27017"
        Username   = "admin"
        Password   = "Karimun!@34"
        Database   = "mydb"
        Collection = "mgo_test"
    )

    game := Game{
        Winner:       "Dave",
        OfficialGame: true,
        Location:     "Austin",
        StartTime:    time.Date(2015, time.February, 12, 04, 11, 0, 0, time.UTC),
        EndTime:      time.Date(2015, time.February, 12, 05, 54, 0, 0, time.UTC),
        Players: []Player{
            NewPlayer("Dave", "Wizards", "Steampunk", 21, 1),
            NewPlayer("Javier", "Zombies", "Ghosts", 18, 2),
            NewPlayer("George", "Aliens", "Dinosaurs", 17, 3),
            NewPlayer("Seth", "Spies", "Leprechauns", 10, 4),
        },
    }

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

    session, err := mgo.DialWithInfo(&mgo.DialInfo{
        Addrs:    []string{Host},
        Timeout:  60 * time.Second,
        Username: Username,
        Password: Password,
        Database: Database,
        DialServer: func(addr *mgo.ServerAddr) (net.Conn, error) {
            return tls.Dial("tcp", addr.String(), &tls.Config{
		RootCAs: rootCerts,
                Certificates: clientCerts,
		InsecureSkipVerify: true,
	    })
        },
    })
    if err != nil {
        panic(err)
    }
    defer session.Close()

    fmt.Printf("Connected to %v\n", session.LiveServers())

    coll := session.DB(Database).C(Collection)
    if err := coll.Insert(game); err != nil {
        panic(err)
    }
    fmt.Println("Document inserted successfully!")
}

