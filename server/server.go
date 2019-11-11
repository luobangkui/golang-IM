package server

import (
	"github.com/nsqio/go-nsq"
	"database/sql"
	"log"
	"github.com/spf13/cast"
	"net/http"
	message2 "github.com/luobangkui/golang-IM/message"
	"github.com/gorilla/websocket"
	"fmt"
	"github.com/luobangkui/golang-IM/config"
	"time"
	"gopkg.in/redis.v5"
	"io/ioutil"
	"github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary


var upgrader = websocket.Upgrader{} // use default options

type IM_server struct {
	options *ImOptions
}

type ImOptions struct {

	Producer nsq.Producer

	Db  *sql.DB

	Redis *redis.Client
}

var DefaultImOptions = &ImOptions{

}


type Option func(*ImOptions)



func NewProducer(cfg config.Config) nsq.Producer {
	addr := cfg.NsqdAddr
	p, err := nsq.NewProducer(addr,nsq.NewConfig())
	if err != nil {
		log.Fatal(err)
	}
	return *p
}


func NewImServer(opts	...Option) *IM_server  {

	options := DefaultImOptions

	for i := range opts{
		opts[i](options)
	}

	return &IM_server{
		options:options,
	}
}

func WithProducer(p nsq.Producer) Option {
	return func(o *ImOptions) {
		o.Producer = p
	}
}

func WithDatabase(d sql.DB) Option  {
	return func(o *ImOptions) {
		o.Db = &d
	}
}

func WithRedis(d redis.Client) Option  {
	return func(o *ImOptions) {
		o.Redis = &d
	}
}

func (im *IM_server)ValidateUserPass(w http.ResponseWriter, r *http.Request) {
	log.Println("validate user pass!!")
	body, _ := ioutil.ReadAll(r.Body)
	up := &UserPass{}
	err := json.Unmarshal(body,up)
	if err != nil {
		log.Println(err)
		// failed validate
		w.Write([]byte("-1"))
	}
	uid ,err := ValidateUserPass(im.options.Db,up.Username,up.Password)
	w.Write([]byte(uid))
}




func (im *IM_server) ProcessSendMsg(w http.ResponseWriter, r *http.Request) {

	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()
	producer := im.options.Producer
	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		//收到用户A的消息，向用户B去推送
		msg_carrier := &message2.Msg{}
		fmt.Println(string(message))
		err = json.Unmarshal(message,msg_carrier)
		if err != nil {
			log.Printf("%v",err)
		}
		err = msgHandle(im.options,msg_carrier)
		if err != nil {
			fmt.Println(err)
		}

		rid := msg_carrier.RecipientId
		//消息推送给接受人
		producer.Publish(cast.ToString(rid),message)
	}
}

func msgHandle(option *ImOptions, msg *message2.Msg) error {
	db := option.Db
	/* 1.存消息内容 */
	fmt.Println("msgHandle---")
	stmtIn, err := db.Prepare("INSERT INTO IM_MSG_CONTENT " +
		"(content,sender_id,recipient_id,msg_type,create_time) VALUES( ?, ? ,?,?,?)")

	if err != nil {
		return err
	}

	timestamp := time.Now()
	res, err := stmtIn.Exec(msg.Content,msg.SenderId,msg.RecipientId,msg.MsgType,timestamp)
	if err != nil {
		return err
	}
	mid, err := res.LastInsertId()
	fmt.Println(mid)
	if err != nil {
		return err
	}
	stmtIn.Close()
	log.Println("insert msg :",mid)

	//2. 存发件人的发件箱
	stmtIn, err = db.Prepare("INSERT INTO  IM_MSG_RELATION" +
		"(owner_uid,other_uid,mid,type,create_time) VALUES (?,?,?,?,?)")
	if err != nil {
		return err
	}
	res, err = stmtIn.Exec(msg.SenderId,msg.RecipientId,mid,0,timestamp)
	if err != nil {
		return err
	}
	log.Println("sender msg stored")
	stmtIn.Close()


	//3. 收件人收件箱
	stmtIn, err = db.Prepare("INSERT INTO  IM_MSG_RELATION" +
		"(owner_uid,other_uid,mid,type,create_time) VALUES (?,?,?,?,?)")
	if err != nil {
		return err
	}
	res, err = stmtIn.Exec(msg.RecipientId,msg.SenderId,mid,1,timestamp)
	if err != nil {
		return err
	}
	log.Println("reciever msg stored")
	stmtIn.Close()

	//4. 更新发件人最近联系人
	stmtIn, err = db.Prepare("INSERT INTO  IM_MSG_CONTACT" +
		"(owner_uid,other_uid,mid,type,create_time) VALUES (?,?,?,?,?)")
	if err != nil {
		return err
	}
	res, err = stmtIn.Exec(msg.SenderId,msg.RecipientId,mid,0,timestamp)
	if err != nil {
		return err
	}
	log.Println("sender contact stored")
	stmtIn.Close()

	//5. 更新收件人的最近联系人
	stmtIn, err = db.Prepare("INSERT INTO  IM_MSG_CONTACT" +
		"(owner_uid,other_uid,mid,type,create_time) VALUES (?,?,?,?,?)")
	if err != nil {
		return err
	}
	res, err = stmtIn.Exec(msg.RecipientId,msg.SenderId,mid,1,timestamp)
	if err != nil {
		return err
	}
	log.Println("reciever contact stored")
	stmtIn.Close()

	//6. 更新未读
	redisClient := option.Redis
	redisClient.Incr(cast.ToString(msg.RecipientId)+"_T") //加总未读
	redisClient.Incr(cast.ToString(msg.RecipientId)+"_C") //加会话未读

	return nil
}





