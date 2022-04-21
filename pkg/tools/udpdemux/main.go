package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
	"sync"
	"time"
)

const (
	maxBufferSize = 16 * 1024
)

var (
	listenAddr = flag.String("listen-addr", "0.0.0.0:2004", "UDP address to listen on")

	usageFunc = func() {
		fmt.Fprintf(flag.CommandLine.Output(),
			"Listen for UDP traffic and forward to any number of localhost ports\n\n",
		)
		fmt.Fprintf(flag.CommandLine.Output(),
			"%s [flags] [UDP localhost ports for forwarding...]\n\n", os.Args[0],
		)
		fmt.Fprintf(flag.CommandLine.Output(), "Flags:\n")
		flag.PrintDefaults()
		fmt.Fprintf(flag.CommandLine.Output(),
			"\nExample:\n\t%s --listen-addr 0.0.0.0:8080 8081 8082\n", os.Args[0],
		)
	}
)

func main() {
	flag.Usage = usageFunc
	flag.Parse()

	if flag.NArg() == 0 {
		flag.Usage()
		os.Exit(0)
	}

	ports := parsePorts(flag.Args())
	fmt.Printf("Listening on UDP address %s...\n", *listenAddr)
	err := server(context.Background(), *listenAddr, ports)
	if err != nil {
		panic(err)
	}
}

func parsePorts(osArgs []string) []int {
	var ports []int
	for _, p := range osArgs {
		i, err := strconv.ParseInt(p, 10, 32)
		if err != nil {
			panic("failed parsing: " + p + ", error: " + err.Error())
		}
		ports = append(ports, int(i))
	}
	return ports
}

type Buff struct {
	orig []byte
	data []byte
}

var bufPool = sync.Pool{
	New: func() interface{} {
		b := make([]byte, maxBufferSize)
		return &Buff{
			orig: b,
			data: b,
		}
	},
}

func server(ctx context.Context, address string, ports []int) error {
	// ListenPacket provides us a wrapper around ListenUDP so that
	// we don't need to call `net.ResolveUDPAddr` and then subsequentially
	// perform a `ListenUDP` with the UDP address.
	//
	// The returned value (PacketConn) is pretty much the same as the one
	// from ListenUDP (UDPConn) - the only difference is that `Packet*`
	// methods and interfaces are more broad, also covering `ip`.
	pc, err := net.ListenPacket("udp", address)
	if err != nil {
		return err
	}

	// `Close`ing the packet "connection" means cleaning the data structures
	// allocated for holding information about the listening socket.
	defer pc.Close()

	errChan := make(chan error, 1)
	msgChan := make(chan *Buff, 1024)

	go func() {
		for {
			buff := bufPool.Get().(*Buff)

			n, _, err := pc.ReadFrom(buff.orig)
			buff.data = buff.orig[:n]
			msgChan <- buff
			if err != nil {
				errChan <- err
				return
			}
		}
	}()

	go func() {
		conns, err := getConns(ports)
		if err != nil {
			errChan <- err
			return
		}

		defer func() {
			for _, c := range conns {
				if err := c.Close(); err != nil {
					errChan <- err
				}
			}
		}()

		for {
			select {
			case msg := <-msgChan:
				for _, c := range conns {
					err = c.SetWriteDeadline(time.Now().Add(5 * time.Second))
					if err != nil {
						panic(err)
					}

					n, err := c.Write(msg.data)
					if errors.Is(err, net.ErrClosed) {
						fmt.Printf("Connection to %v already closed: %v\n",
							c.RemoteAddr(), err,
						)
						continue
					}
					if err != nil {
						fmt.Printf("Error printing to %v, err: %v\n",
							c.RemoteAddr(), err,
						)
						continue
					}
					fmt.Printf("%v: forwarded %v bytes\n", c.RemoteAddr(), n)
				}

				bufPool.Put(msg)
			case <-ctx.Done():
				return
			}
		}
	}()

	select {
	case <-ctx.Done():
		fmt.Println("context cancelled")
		err = ctx.Err()
	case err = <-errChan:
	}

	return err
}

func getConns(ports []int) ([]*net.UDPConn, error) {
	var conns []*net.UDPConn
	for _, p := range ports {
		target := fmt.Sprintf("localhost:%d", p)

		addr, err := net.ResolveUDPAddr("udp", target)
		if err != nil {
			return nil, fmt.Errorf("ResolveUDPAddr error: %w", err)
		}

		conn, err := net.DialUDP("udp", nil, addr)
		if err != nil {
			return nil, fmt.Errorf("DialUDP error: %w", err)
		}

		conns = append(conns, conn)
		fmt.Printf("Connected to target: %s...\n", target)
	}

	return conns, nil
}
