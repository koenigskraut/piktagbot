package database

import (
	"fmt"
	"github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/ssh"
	msq "gorm.io/driver/mysql"
	"gorm.io/gorm"
	"net"
	"os"
)

var DB *gorm.DB
var SSHCon *ssh.Client

type ViaSSHDialer struct {
	client *ssh.Client
}

func (dialer *ViaSSHDialer) Dial(addr string) (net.Conn, error) {
	return dialer.client.Dial("tcp", addr)
}

func InitializeDB() {
	viaSSH := os.Getenv("PTB_VIA_SSH") != ""
	if viaSSH {
		sshHost := os.Getenv("sshHost") // SSH Server Hostname/IP
		sshUser := os.Getenv("sshUser") // SSH Username
		sshPass := os.Getenv("sshPass") // Empty string for no password
		dbUser := os.Getenv("dbUser")   // DB username
		dbPass := os.Getenv("dbPass")   // DB Password
		dbHost := "localhost:3306"      // DB Hostname/IP
		dbName := "stickers"            // Database name

		sshConfig := &ssh.ClientConfig{
			User: sshUser,
			Auth: []ssh.AuthMethod{
				ssh.Password(sshPass),
			},
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		}
		SSHCon, _ = ssh.Dial("tcp", sshHost+":22", sshConfig)
		// TODO it's deprecated, change it or maybe remove this code since it is of no use rn
		mysql.RegisterDial("mysql+tcp", (&ViaSSHDialer{SSHCon}).Dial)
		var err error
		if DB, err = gorm.Open(msq.Open(fmt.Sprintf("%s:%s@mysql+tcp(%s)/%s?parseTime=true",
			dbUser, dbPass, dbHost, dbName)),
			&gorm.Config{}); err == nil {
			fmt.Printf("Successfully connected to the db\n")
		} else {
			fmt.Printf("Failed to connect to the db: %s\n", err.Error())
		}
	} else {
		dbUser := os.Getenv("dbUser") // DB username
		dbPass := os.Getenv("dbPass") // DB Password
		dbHost := "localhost:3306"    // DB Hostname/IP
		dbName := "stickers"          // Database name

		var err error
		if DB, err = gorm.Open(msq.Open(fmt.Sprintf("%s:%s@tcp(%s)/%s?parseTime=true",
			dbUser, dbPass, dbHost, dbName)),
			&gorm.Config{}); err == nil {
			fmt.Printf("Successfully connected to the db\n")
		} else {
			fmt.Printf("Failed to connect to the db: %s\n", err.Error())
		}
	}
}
