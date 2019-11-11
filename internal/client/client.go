package client

import (
	m2 "github.com/luobangkui/im-learn-about/message"
	"github.com/gorilla/websocket"
	"net/url"
	"fmt"
	"encoding/json"
	"github.com/luobangkui/im-learn-about/utils/log"
	"github.com/nsqio/go-nsq"
	"os"
	"math/rand"
	"github.com/luobangkui/im-learn-about/internal/event"
	"github.com/spf13/cast"
	"github.com/luobangkui/im-learn-about/message"
	"time"
	"bufio"
	"os/signal"
	"syscall"
)

var (
	USER_LOGIN = "login"
	USER_LOGOUT = "logout"
	USER_LOGOUT_TEMP = "logout_temp"
	CHAT = "chat"
	CHAT_EXIT = "chat exit"
)


type MsgClient struct {
	Newest_msg_version int64

	MsgCache map[int64]m2.Msg

	Conn *websocket.Conn

	Timeout int

	MessageDump chan m2.Msg

	LastMsg m2.Msg

	Sid 	int64

	AckSucceed	chan int

	Signout chan int

	Option Option

	eventBus *event.EventBus

}

func (client *MsgClient)Init()  {
	client.MessageDump = make(chan m2.Msg)
	client.MsgCache = make(map[int64]m2.Msg)
}

func (client *MsgClient)AddSubscriber()  {
	client.eventBus.Subscribe(client.onHandle)
}

func (client *MsgClient) onHandle(evt *event.Event)  {
	log.Debugf("event fired >> %v", evt)
	evt.Processer(evt)

}

func (client *MsgClient)processEvent()  {




}

func (client *MsgClient)initState(evt *event.Event)  {
	log.Debugf("%v event fired",evt.Type)
	fmt.Println("请输入用户名")
	var username string
	fmt.Scanln(&username)

	fmt.Println("请输入密码")
	var password string
	fmt.Scanln(&password)

	//TODO 验证并登录,应该需要加密传输的
	line := username+":"+password

	msg := message.Msg{
		MsgId:   int64(0),
		Content: line,
		MsgType:2,
	}
	client.MessageDump <- msg
}


//登入事件处理
func (client *MsgClient) processLogin(evt *event.Event)  {

	fmt.Println("联系人")

	//TODO 列出联系人

	fmt.Println("请输入联系人编号")

	var reciever int64
	fmt.Scanln(&reciever)

	//TODO 列出最近聊天记录

	//TODO 得到用户id number，可以开始向他发送消息了
	//登录之后发布chat事件
	chatEvt := event.NewEvent(CHAT,"",event.WithProcessor(client.processChat))
	chatEvt.Sender = evt.Sender
	chatEvt.Reciever = reciever
	client.eventBus.Emit(chatEvt)
}

func (client *MsgClient) processChat(evt *event.Event)  {

	var senderid int64
	var recieverid int64
	//TODO chat process
	go client.MessageHandle(client.Option.Nsqaddr,cast.ToString(recieverid))

	sid := 0
	version := 100000
	r := bufio.NewReader(os.Stdin)
	stopChan := make(chan bool)
	delim := string("\n")[0]

	go func(reciever,sender int64) {
		for  {
			sid +=2
			version += 1
			line, err := r.ReadBytes(delim)
			if err != nil {
				log.Info(err)
			}
			msg := message.Msg{
				MsgId:   int64(sid),
				Content: string(line),
				RecipientId: reciever,
				SenderId:sender,
			}
			client.MessageDump <- msg
			time.Sleep(2*time.Second)
		}
	}(recieverid,senderid)

	select {
	case <-stopChan:

	}
}



func (client *MsgClient) processSignoutTemp(evt *event.Event) {
	//TODO 未读消息提醒
}














func InitMsgSenderConn(addr string) *websocket.Conn  {
	u := url.URL{Scheme: "ws", Host: addr, Path: "/send_msg"}
	log.Debug("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)

	if err != nil {
		fmt.Println(err)
	}

	return c
}

func GenerateId() int64  {
	//TODO
	return int64(0)
}


type SimpleHandler struct {
}

func (sh *SimpleHandler) HandleMessage(m *nsq.Message) error {
	msg := &m2.Msg{}

	err := json.Unmarshal(m.Body,msg)

	if err != nil {
		log.Info(err)
	}

	_, err = os.Stdout.Write([]byte("recieve: "+msg.Content))

	if err != nil {
		fmt.Println(err)
	}
	return nil
}



func (client *MsgClient) MessageHandle(nsqdaddr string, reciver_id string) error  {

	done := make(chan struct{})

	cfg := nsq.NewConfig()

	channel := fmt.Sprintf("tail%06d#ephemeral", rand.Int()%999999)

	c, _ := nsq.NewConsumer(reciver_id,channel,cfg)

	c.AddHandler(&SimpleHandler{})

	c.ConnectToNSQD(nsqdaddr)
	stop := make(chan int)

	sig := make(chan os.Signal, 1)

	//ctrl C 退出当前会话
	signal.Notify(sig, syscall.SIGINT)

	go func() {
		for  {
			select {
			case <- done:
				return
			case a :=<- client.MessageDump:
				client.LastMsg = a
				sid := a.MsgId
				client.Sid = sid
				client.MsgCache[sid] = a
				conn := client.Conn
				reqMsg,_ := json.Marshal(a)
				err := conn.WriteMessage(websocket.TextMessage,reqMsg)
				if err != nil {
					log.Info(err)
				}
				//log.Println("write message :",string(reqMsg))
			case <-client.Signout:
				//退出通知
				return
			case <- sig:
				chatEvt := event.NewEvent(CHAT_EXIT,"",event.WithProcessor(client.processSignoutTemp))

				client.eventBus.Emit(chatEvt)
			}
		}
	}()

	<-stop

	return nil
}

