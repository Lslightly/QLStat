package db

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"go/token"
	"net"
	"os"
	"sync"
	"time"

	"github.com/barweiss/go-tuple"
	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/lib/pq"
	"golang.org/x/crypto/ssh"
)

type DatabaseConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	Database string
}

type SshConfig struct {
	Host       string
	Port       int
	Username   string
	PrivateKey string
}

type viaSSHDialer struct {
	client *ssh.Client
}

func (self *viaSSHDialer) Open(s string) (_ driver.Conn, err error) {
	return pq.DialOpen(self, s)
}
func (self *viaSSHDialer) Dial(network, address string) (net.Conn, error) {
	return self.client.Dial(network, address)
}
func (self *viaSSHDialer) DialTimeout(network, address string, timeout time.Duration) (net.Conn, error) {
	return self.client.Dial(network, address)
}

type Connection struct {
	_db *sql.DB
	type_mutex sync.RWMutex
	type_cache *lru.Cache[string,int64]
	file_cache *lru.Cache[string,int64]
	expr_cache *lru.Cache[tuple.T2[token.Position,token.Position],int64]
}

func(conn *Connection) Query(query string, args ...any) (*sql.Rows, error) {
	rows, err := conn._db.Query(query, args...)
	if err, ok := err.(*net.OpError); ok {
		panic(err)
	}
	return rows, err
}

func (conn *Connection) Exec(query string, args ...any) (sql.Result, error) {
	result, err := conn._db.Exec(query, args...)
	if err, ok := err.(*net.OpError); ok {
		panic(err)
	}
	return result, err
}

func (conn *Connection) Begin() (*sql.Tx, error) {
	tx, err := conn._db.Begin()
	if err, ok := err.(*net.OpError); ok {
		panic(err)
	}
	return tx, err
}

func WithConnection(dbConfig DatabaseConfig, sshConfig SshConfig, useSSH bool, f func(connection *Connection) error) error {
	if useSSH {
		ssh_cfg := &ssh.ClientConfig{
			User: sshConfig.Username,
			Auth: []ssh.AuthMethod{getSignerByPrivateKey(sshConfig.PrivateKey)},
			HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
				return nil
			},
		}
		ssh_conn, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", sshConfig.Host, sshConfig.Port), ssh_cfg)
		if err != nil {
			return err
		}
		defer ssh_conn.Close()

		sql.Register("postgres+ssh", &viaSSHDialer{ssh_conn})
	}

	dbUser := dbConfig.Username
	dbPass := dbConfig.Password
	dbHost := dbConfig.Host
	dbName := dbConfig.Database

	var err error
	var db *sql.DB

	if useSSH {
		db, err = sql.Open("postgres+ssh", fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", dbUser, dbPass, dbHost, dbName))
	} else {
		db, err = sql.Open("postgres", fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", dbUser, dbPass, dbHost, dbName))
	}

	if err != nil {
		return err
	}
	defer db.Close()

	connection := Connection {
		db,
		sync.RWMutex{},
		new_cache[string,int64](),
		new_cache[string,int64](),
		new_cache[tuple.T2[token.Position,token.Position],int64](),
	}

	return f(&connection)
}

func getSignerByPrivateKey(file string) ssh.AuthMethod {
	buffer, err := os.ReadFile(file)
	if err != nil {
		return nil
	}

	key, err := ssh.ParsePrivateKey(buffer)
	if err != nil {
		return nil
	}
	return ssh.PublicKeys(key)
}
