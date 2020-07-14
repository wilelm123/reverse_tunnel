package main

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"time"

	flags "github.com/jessevdk/go-flags"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

func init() {
	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)
}

var opts struct {
	Port    int    `short:"p" long:"port" description:"The local port to forward to" default:"3141"`
	RIP     string `long:"remote_ip" description:"The remote server ip address" required:"true"`
	RPort   int    `short:"r" long:"remote_port" description:"The remote port to listen on the remote server" required:"true"`
	RUser   string `short:"u" long:"remote_user" description:"The user to ssh to the remote server" required:"true"`
	RPasswd string `long:"remote_password" description:"The password to ssh to the remote server" required:"true"`
}

// Endpoint the server endpoint struct
type Endpoint struct {
	Host string
	Port int
}

func (endpoint *Endpoint) String() string {
	return fmt.Sprintf("%s:%d", endpoint.Host, endpoint.Port)
}

// SSHtunnel is the ssh tunnel struct
type SSHtunnel struct {
	Local  *Endpoint
	Server *Endpoint
	Remote *Endpoint

	Config *ssh.ClientConfig
}

// Start method
func (tunnel *SSHtunnel) Start() error {
	log.Info("Start the ssh tunnel service...")

	serverConn, err := ssh.Dial("tcp", tunnel.Server.String(), tunnel.Config)
	if err != nil {
		log.Errorf("Server dial error: %s\n", err)
		return err
	}

	listener, err := serverConn.Listen("tcp", tunnel.Remote.String())
	if err != nil {
		log.Errorf("Listen on the remote server %s failed, %s", tunnel.Remote.String(), err.Error())
		return err
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil && err.Error() != "EOF" {
			log.Errorf("Receive error from connection, %s\n", err.Error())
			return err
		}
		log.Infof("Receive a connection from %s, forward req to %s\n", conn.RemoteAddr(), tunnel.Local.String())
		go tunnel.forward(conn)
	}
}

func (tunnel *SSHtunnel) forward(remoteConn net.Conn) {
	var d net.Dialer
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	localConn, err := d.DialContext(ctx, "tcp", tunnel.Local.String())
	if err != nil {
		log.Errorf("Error to connect to local endpoint: %s\n", err.Error())
		return
	}
	copyConn := func(writer, reader net.Conn) {
		defer writer.Close()
		defer reader.Close()

		_, err := io.Copy(writer, reader)
		if err != nil {
			log.Debugf("io.Copy error: %s", err)
		}
	}

	go copyConn(localConn, remoteConn)
	go copyConn(remoteConn, localConn)
}

func main() {
	_, err := flags.Parse(&opts)
	if err != nil {
		log.Info("User -h or --help to get more help information\n")
		return
	}

	localEndpoint := &Endpoint{
		Host: "127.0.0.1",
		Port: opts.Port,
	}

	serverEndpoint := &Endpoint{
		Host: opts.RIP,
		Port: 22,
	}

	remoteEndpoint := &Endpoint{
		Host: opts.RIP,
		Port: opts.RPort,
	}

	sshConfig := &ssh.ClientConfig{
		User: opts.RUser,
		Auth: []ssh.AuthMethod{
			ssh.Password(opts.RPasswd),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	tunnel := &SSHtunnel{
		Config: sshConfig,
		Local:  localEndpoint,
		Server: serverEndpoint,
		Remote: remoteEndpoint,
	}

	err = tunnel.Start()
	if err != nil {
		log.Fatalf("Service stopped, %s\n", err.Error())
	}
}
