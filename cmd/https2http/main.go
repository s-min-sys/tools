package main

import (
	"context"
	"io"
	"log"
	"net"
	"sync"
	"time"

	"github.com/sgostarter/libconfig"
	"github.com/sgostarter/libeasygo/netutils"
)

type Item struct {
	Listen             string `yaml:"listen"`
	RemoteAddress      string `yaml:"remote_address"`
	RemoteUseTLS       bool   `yaml:"remote_use_tls"`
	InsecureSkipVerify bool   `yaml:"insecure_skip_verify"`
}

type Config struct {
	Items              []Item `yaml:"items"`
	RemoteAddress      string `yaml:"remote_address"`
	RemoteUseTLS       bool   `yaml:"remote_use_tls"`
	InsecureSkipVerify bool   `yaml:"insecure_skip_verify"`
}

func main() {
	var cfg Config

	configFileUsed, err := libconfig.Load("https2http.yaml", &cfg)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("used config file is " + configFileUsed)

	wg := sync.WaitGroup{}

	for idx := 0; idx < len(cfg.Items); idx++ {
		if cfg.Items[idx].Listen == "" {
			continue
		}

		if cfg.Items[idx].RemoteAddress == "" {
			cfg.Items[idx].RemoteAddress = cfg.RemoteAddress
			cfg.Items[idx].RemoteUseTLS = cfg.RemoteUseTLS
			cfg.Items[idx].InsecureSkipVerify = cfg.InsecureSkipVerify
		}

		wg.Add(1)

		go serverRoutine(context.Background(), &wg, cfg.Items[idx])
	}

	wg.Wait()
}

func serverRoutine(ctx context.Context, wg *sync.WaitGroup, item Item) {
	defer wg.Done()

	l, e := net.Listen("tcp", item.Listen)
	if e != nil {
		log.Fatal(e)
	}

	useTLS := ""
	if item.RemoteUseTLS {
		useTLS = "tls"
	}

	log.Printf("start on %s to %s:%s\n", item.Listen, item.RemoteAddress, useTLS)

	defer func() {
		_ = l.Close()

		log.Printf("stop on %s to %s:%s\n", item.Listen, item.RemoteAddress, useTLS)
	}()

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Fatal(err)
		}

		go processConn(ctx, conn, &item)
	}
}

func processConn(ctx context.Context, conn net.Conn, item *Item) {
	defer func() {
		_ = conn.Close()
	}()

	var modifiers []netutils.TLSConfigModifier

	if item.RemoteUseTLS && item.InsecureSkipVerify {
		modifiers = append(modifiers, netutils.TLSConfigModifier4InsecureSkipVerify)
	}

	remoteConn, err := netutils.DialTCPWithTimeout(ctx, item.RemoteUseTLS, item.RemoteAddress,
		time.Second*10, modifiers...)
	if err != nil {
		log.Println("dial failed:", item.RemoteAddress, err)

		return
	}

	go func() {
		defer func() {
			_ = remoteConn.Close()
			_ = conn.Close()
		}()

		_, _ = io.Copy(remoteConn, conn)
	}()

	defer func() {
		_ = remoteConn.Close()
		_ = conn.Close()
	}()

	_, _ = io.Copy(conn, remoteConn)
}
