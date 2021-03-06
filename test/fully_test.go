package test

import (
	"context"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"rulex/core"
	"rulex/engine"
	"rulex/plugin/demo_plugin"
	httpserver "rulex/plugin/http_server"
	"rulex/rulexrpc"
	"rulex/typex"
	"syscall"
	"testing"
	"time"

	"github.com/ngaut/log"
	"google.golang.org/grpc"
)

func TestFullyRun(t *testing.T) {
	runTest()
}

//
func runTest() {
	core.InitGlobalConfig()
	Run()
}

//
func Run() {

	core.InitGlobalConfig()
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGABRT)
	engine := engine.NewRuleEngine()
	engine.Start()
	core.InitXQueue(10*(2<<10), engine)

	hh := httpserver.NewHttpApiServer(2580, "plugin/http_server/templates", "./rulex.db", engine)

	// HttpApiServer loaded default
	if err := engine.LoadPlugin(hh); err != nil {
		log.Fatal("Rule load failed:", err)
	}
	// Load a demo plugin
	if err := engine.LoadPlugin(demo_plugin.NewDemoPlugin()); err != nil {
		log.Error("Rule load failed:", err)
	}
	// Grpc Inend
	grpcInend := typex.NewInEnd("GRPC", "Rulex Grpc InEnd", "Rulex Grpc InEnd", &map[string]interface{}{
		"port": "2581",
	})
	if err := engine.LoadInEnd(grpcInend); err != nil {
		log.Error("Rule load failed:", err)
	}
	// CoAP Inend
	coapInend := typex.NewInEnd("COAP", "Rulex COAP InEnd", "Rulex COAP InEnd", &map[string]interface{}{
		"port": "2582",
	})
	if err := engine.LoadInEnd(coapInend); err != nil {
		log.Error("Rule load failed:", err)
	}
	// Http Inend
	httpInend := typex.NewInEnd("HTTP", "Rulex HTTP InEnd", "Rulex HTTP InEnd", &map[string]interface{}{
		"port": "2583",
	})
	if err := engine.LoadInEnd(httpInend); err != nil {
		log.Error("Rule load failed:", err)
	}
	// Udp Inend
	udpInend := typex.NewInEnd("UDP", "Rulex UDP InEnd", "Rulex UDP InEnd", &map[string]interface{}{
		"port": "2584",
	})
	if err := engine.LoadInEnd(udpInend); err != nil {
		log.Error("Rule load failed:", err)
	}
	//
	// Load Rule
	//
	rule := typex.NewRule(engine,
		"Just a test",
		"Just a test",
		[]string{grpcInend.Id},
		`function Success() print("[LUA Success]==========================> OK") end`,
		`
		Actions = {
			function(data)
			    print("[LUA Actions Callback]==========================> ", data)
				return true, data
			end
		}`,
		`function Failed(error) print("[LUA Failed]==========================> OK", error) end`)
	if err := engine.LoadRule(rule); err != nil {
		log.Error(err)
	}
	conn, err := grpc.Dial("127.0.0.1:2581", grpc.WithInsecure())
	if err != nil {
		log.Error("grpc.Dial err: %v", err)
	}
	defer conn.Close()
	client := rulexrpc.NewRulexRpcClient(conn)
	resp, err := client.Work(context.Background(), &rulexrpc.Data{
		Value: `{"temp":100,"hum":30, "co2":123.4, "lex":22.56}`,
	})
	if err != nil {
		log.Error("grpc.Dial err: %v", err)
	}
	log.Debugf("Rulex Rpc Call Result ====>>: %v", resp.GetMessage())
	time.Sleep(2 * time.Second)
	log.Info("Test Http Api===> " + HttpGet("http://127.0.0.1:2580/api/v1/system"))
	engine.Stop()
}

func HttpGet(api string) string {
	var err error
	request, err := http.NewRequest("GET", api, nil)
	if err != nil {
		log.Error(err)
		return ""
	}

	response, err := (&http.Client{}).Do(request)
	if err != nil {
		log.Error(err)
		return ""
	}
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Error(err)
		return ""
	}
	return string(body)
}
