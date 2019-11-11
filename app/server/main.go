
package main

import (
	"flag"
	"html/template"
	"net/http"
	"log"
	"github.com/luobangkui/golang-IM/config"
	"github.com/luobangkui/golang-IM/server"
	"github.com/luobangkui/golang-IM/database/mysql"
	"github.com/luobangkui/golang-IM/database/redis"
)

var addr = flag.String("addr", "localhost:8080", "http service address")

var nsqdaddr = flag.String("nsqdaddr", "localhost:9876", "nsqd address")






func home(w http.ResponseWriter, r *http.Request) {
	homeTemplate.Execute(w, "ws://"+r.Host+"/send_msg")
}

func main() {
	flag.Parse()
	log.SetFlags(0)
	cfg,err := config.LoadFile("config/config.toml")
	if err != nil {
		panic(err.Error())
	}
	//
	db := mysql.NewDatabase(*cfg)
	producer := server.NewProducer(*cfg)
	redisClient := redis.NewRedisClient(*cfg)
	options := []server.Option{
		server.WithDatabase(*db),
		server.WithProducer(producer),
		server.WithRedis(*redisClient),
	}
	im := server.NewImServer(options...)

	http.HandleFunc("/send_msg", im.ProcessSendMsg)
	http.HandleFunc("/login",im.ValidateUserPass)
	http.HandleFunc("/", home)
	log.Fatal(http.ListenAndServe(*addr, nil))
}

var homeTemplate = template.Must(template.New("").Parse(`
<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<script>  
window.addEventListener("load", function(evt) {

    var output = document.getElementById("output");
    var input = document.getElementById("input");
    var ws;

    var print = function(message) {
        var d = document.createElement("div");
        d.innerHTML = message;
        output.appendChild(d);
    };

    document.getElementById("open").onclick = function(evt) {
        if (ws) {
            return false;
        }
        ws = new WebSocket("{{.}}");
        ws.onopen = function(evt) {
            print("OPEN");
        }
        ws.onclose = function(evt) {
            print("CLOSE");
            ws = null;
        }
        ws.onmessage = function(evt) {
            print("RESPONSE: " + evt.data);
        }
        ws.onerror = function(evt) {
            print("ERROR: " + evt.data);
        }
        return false;
    };

    document.getElementById("send").onclick = function(evt) {
        if (!ws) {
            return false;
        }
        print("SEND: " + input.value);
        ws.send(input.value);
        return false;
    };

    document.getElementById("close").onclick = function(evt) {
        if (!ws) {
            return false;
        }
        ws.close();
        return false;
    };

});
</script>
</head>
<body>
<table>
<tr><td valign="top" width="50%">
<p>Click "Open" to create a connection to the server, 
"Send" to send a message to the server and "Close" to close the connection. 
You can change the message and send multiple times.
<p>
<form>
<button id="open">Open</button>
<button id="close">Close</button>
<p><input id="input" type="text" value="Hello world!">
<button id="send">Send</button>
</form>
</td><td valign="top" width="50%">
<div id="output"></div>
</td></tr></table>
</body>
</html>
`))
