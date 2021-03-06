package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"text/template"

	"github.com/garyburd/redigo/redis" //Redis
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

type Article struct {
	Title   string `json:"Title"`
	Desc    string `json:"desc"`
	Content string `json:"content"`
}

// let's declare a global Articles array
// that we can then populate in our main function
// to simulate a database
var Articles []Article

func homePage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome to the HomePage!")
	fmt.Println("Endpoint Hit: homePage")
}

func returnAllArticles(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Endpoint Hit: returnAllArticles")
	json.NewEncoder(w).Encode(Articles)
}

func handleRequests() {
	myRouter := mux.NewRouter().StrictSlash(true)
	myRouter.HandleFunc("/apihome", homePage)
	myRouter.HandleFunc("/articles", returnAllArticles)
	log.Fatal(http.ListenAndServe(":10000", myRouter))
}

func RedisTest() {
	fmt.Println("Redis Test GO")
	c, err := redis.Dial("tcp", "localhost:6379")
	if err != nil {
		fmt.Println("conn redis failed, err:", err)
		return
	}
	defer c.Close()
	fmt.Println("Redis Connection Success")

	fmt.Println("Redis Set Get Test")
	_, err = c.Do("Set", "name", "nick")
	if err != nil {
		fmt.Println(err)
		return
	}

	r, err := redis.String(c.Do("Get", "name"))
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(r)
	fmt.Println("Redis Set Get Test Success")
	fmt.Println("Redis mset mget Test")
	_, err = c.Do("MSet", "name", "nick", "age", "18")
	if err != nil {
		fmt.Println("MSet error: ", err)
		return
	}

	r2, err := redis.Strings(c.Do("MGet", "name", "age"))
	if err != nil {
		fmt.Println("MGet error: ", err)
		return
	}
	fmt.Println(r2)
	fmt.Println("Redis mset mget Test Success")

	//hset hget test
	_, err = c.Do("HSet", "names", "nick", "suoning")
	if err != nil {
		fmt.Println("hset error: ", err)
		return
	}

	r, err = redis.String(c.Do("HGet", "names", "nick"))
	if err != nil {
		fmt.Println("hget error: ", err)
		return
	}
	fmt.Println(r)

	// expire test
	_, err = c.Do("expire", "names", 5)
	if err != nil {
		fmt.Println("expire error: ", err)
		return
	}
	//loop
	// 隊列
	_, err = c.Do("lpush", "Queue", "nick", "dawn", 9)
	if err != nil {
		fmt.Println("lpush error: ", err)
		return
	}
	for {
		r, err = redis.String(c.Do("lpop", "Queue"))
		if err != nil {
			fmt.Println("lpop error: ", err)
			break
		}
		fmt.Println(r)
	}
	r3, err := redis.Int(c.Do("llen", "Queue"))
	if err != nil {
		fmt.Println("llen error: ", err)
		return
	}
	fmt.Println(r3)
}
func main() {
	flag.Parse()
	log.SetFlags(0)
	http.HandleFunc("/echo", echo)
	http.HandleFunc("/", home)
	log.Fatal(http.ListenAndServe(*addr, nil))
	fmt.Println("Rest API v2.0 - Mux Routers")
	Articles = []Article{
		Article{Title: "Hello", Desc: "Article Description", Content: "Article Content"},
		Article{Title: "Hello 2", Desc: "Article Description", Content: "Article Content"},
	}

	RedisTest()
	handleRequests()
}

var addr = flag.String("addr", "localhost:8080", "http service address")

var upgrader = websocket.Upgrader{} // use default options

func echo(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()
	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		log.Printf("recv: %s", message)
		err = c.WriteMessage(mt, message)
		if err != nil {
			log.Println("write:", err)
			break
		}
	}
}

func home(w http.ResponseWriter, r *http.Request) {
	homeTemplate.Execute(w, "ws://"+r.Host+"/echo")
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
        d.textContent = message;
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
