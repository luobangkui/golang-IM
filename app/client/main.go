package main

import (
	"flag"
	"github.com/luobangkui/golang-IM/internal/client"

)

var addr = flag.String("addr", "localhost:8080", "http service address")

var nsqdaddr = flag.String("nsqdaddr", "localhost:9876", "nsqd service address")
var rid = flag.String("rid", "1", "reciever id")

var sender = flag.String("sender", "2", "sender id")

func main() {
	flag.Parse()
	cli := new(client.MsgClient)
	cli.Conn =client.InitMsgSenderConn(*addr)
	cli.Option = client.Option{
		Nsqaddr:*nsqdaddr,
		ServerAddr:*addr,
	}
	cli.Init()

	go cli.ProcessEvent()

	pause := make(chan int)
	//go cli.MessageHandle(*nsqdaddr,*rid)

	<- pause
}
