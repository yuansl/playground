package rabbitmq

import (
	"fmt"
	"net"
	"net/url"

	amqp "github.com/rabbitmq/amqp091-go"
)

func connect(options *rabbitmqOptions) (*amqp.Connection, error) {
	return amqp.Dial(options.addr)
}

func connect2(options *rabbitmqOptions) (*amqp.Connection, error) {
	u, err := url.Parse(options.addr)
	if err != nil {
		return nil, fmt.Errorf("url.Parse(addr=%q): %v", options.addr, err)
	}
	netconn, err := net.Dial("tcp", u.Host)
	if err != nil {
		return nil, fmt.Errorf("net.Dial: %w", err)
	}
	conf := amqp.Config{
		Dial: func(network, addr string) (net.Conn, error) {
			fmt.Printf("Dialing address: '%s/%s' ...\n", network, addr)
			return net.Dial(network, addr)
		},
		Vhost: "/",
	}
	if options.username != "" {
		conf.SASL = []amqp.Authentication{
			&amqp.AMQPlainAuth{
				Username: options.username, Password: options.password,
			}}
	}
	return amqp.Open(netconn, conf)
}
