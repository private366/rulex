package test

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"

	"rulex/engine"
	httpserver "rulex/plugin/http_server"
	"rulex/rulexrpc"

	"testing"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/go-playground/assert/v2"
	"github.com/ngaut/log"
	"google.golang.org/grpc"
)

//
// TestHttpAPi
//
func TestHttpAPi(t *testing.T) {

	engine := engine.NewRuleEngine()
	engine.Start()
	////////
	hh := httpserver.NewHttpApiServer(2580, "../plugin/http_server/templates/*", "./rulx.db", engine)
	if e := engine.LoadPlugin(hh); e != nil {
		log.Fatal("rule load failed:", e)
	}
	hh.Truncate()
	//---------------------------------------------------------------------------------------
	//
	//---------------------------------------------------------------------------------------
	//
	log.Debug("Create MQTT",
		post(map[string]interface{}{
			"type":        "MQTT",
			"name":        "MQTT test",
			"description": "MQTT Test Resource",
			"config": map[string]interface{}{
				"server":   "127.0.0.1",
				"port":     1883,
				"username": "test",
				"password": "test",
				"clientId": "test",
			},
		}, "inends"),
	)

	///////
	mIn_id_1, errs2 := hh.GetMInEnd(1)
	if errs2 != nil {
		log.Fatal(errs2)
	}
	assert.Equal(t, len(hh.AllMInEnd()), int(1))
	assert.Equal(t, mIn_id_1.ID, uint(1))
	assert.Equal(t, mIn_id_1.Type, "MQTT")
	assert.Equal(t, mIn_id_1.Name, "MQTT test")
	assert.Equal(t, mIn_id_1.Description, "MQTT Test Resource")
	//
	log.Debug(
		post(map[string]interface{}{
			"type":        "HTTP",
			"name":        "HTTP API Server",
			"description": "HTTP Resource",
			"config": map[string]interface{}{
				"port": "2581",
			},
		}, "inends"),
	)
	//---------------------------------------------------------------------------------------
	// Create outend
	//---------------------------------------------------------------------------------------
	log.Debug("Create Mongo",
		post(map[string]interface{}{
			"type":        "mongo",
			"name":        "data to mongo",
			"description": "data to mongo",
			"config": map[string]interface{}{
				"mongourl": "mongodb://root:root@localhost:27017/rulex_test_db?authSource=admin&retryWrites=true&w=majority",
			},
		}, "outends"),
	)
	//
	m_Out_id_1, errs2 := hh.GetMOutEnd(1)
	if errs2 != nil {
		log.Fatal(errs2)
	}
	assert.Equal(t, len(hh.AllMInEnd()), int(1))
	assert.Equal(t, m_Out_id_1.ID, uint(1))
	assert.Equal(t, m_Out_id_1.Type, "mongo")
	assert.Equal(t, m_Out_id_1.Name, "data to mongo")
	assert.Equal(t, m_Out_id_1.Description, "data to mongo")

	//
	//
	// Create rule
	//
	log.Debug(
		post(map[string]interface{}{
			"name":        "just_a_test",
			"description": "just_a_test",
			"actions": strings.Replace(`
			            local json = require("json")
						Actions = {
							function(data)
							    local s = '{"temp":100,"hum":30, "co2":123.4, "lex":22.56}'
								print(s == data)
								DataToMongo("$${OUT}", s)
								return true, data
							end
						}`, "$${OUT}", m_Out_id_1.UUID, -1),
			"from": mIn_id_1.UUID,
			"failed": `
		           function Failed(error)
				   print("call error:",error)
		           end`,
			"success": `
		           function Success()
				   print("call success")
				   end`,
		}, "rules"),
	)
	//
	time.Sleep(3 * time.Second)
	log.Debug("Create HTTP",
		post(map[string]interface{}{
			"type":        "HTTP",
			"name":        "HTTP API Server",
			"description": "HTTP Resource",
			"config": map[string]interface{}{
				"port": "2581",
			},
		}, "inends"),
	)
	//
	publish()
	//
	assert.Equal(t, len((get("inends"))) > 100, true)
	//
	time.Sleep(1 * time.Second)
	log.Debug("Create Http Target",
		post(map[string]interface{}{
			"type":        "HTTP",
			"name":        "data to http server",
			"description": "data to http server",
			"config": map[string]interface{}{
				"url": "http://127.0.0.1:3356/data",
			},
		}, "outends"),
	)
}

func post(data map[string]interface{}, api string) string {
	p, errs1 := json.Marshal(data)
	if errs1 != nil {
		log.Fatal(errs1)
	}
	r, errs2 := http.Post("http://127.0.0.1:2580/api/v1/"+api,
		"application/json",
		bytes.NewBuffer(p))
	if errs2 != nil {
		log.Fatal(errs2)
	}
	defer r.Body.Close()

	body, errs5 := ioutil.ReadAll(r.Body)
	if errs5 != nil {
		log.Fatal(errs5)
	}
	return string(body)
}
func get(api string) string {
	// Get list
	r, errs := http.Get("http://127.0.0.1:2580/api/v1/" + api)
	if errs != nil {
		log.Fatal(errs)
	}
	defer r.Body.Close()
	body, errs2 := ioutil.ReadAll(r.Body)
	if errs2 != nil {
		log.Fatal(errs2)
	}
	return string(body)
}

func publish() {

	var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
		client.Publish("$X_IN_END", 2, false, `{"temp":100,"hum":30, "co2":123.4, "lex":22.56}`)
	}

	opts := mqtt.NewClientOptions()
	opts.AddBroker("tcp://127.0.0.1:1883")
	//
	opts.SetClientID("x_IN_END_TEST1")
	opts.SetUsername("x_IN_END_TEST1")
	opts.SetPassword("x_IN_END_TEST1")
	//
	opts.OnConnect = connectHandler
	client := mqtt.NewClient(opts)
	client.Connect().Wait()
}

func TestHttpInEnd(t *testing.T) {
	engine := engine.NewRuleEngine()
	engine.Start()
	////////
	hh := httpserver.NewHttpApiServer(2580, "../plugin/http_server/templates/*", "./rulx.db", engine)
	if e := engine.LoadPlugin(hh); e != nil {
		log.Fatal("rule load failed:", e)
	}
	hh.Truncate()
	log.Debug(
		post(map[string]interface{}{
			"type":        "HTTP",
			"name":        "HTTP API Server",
			"description": "HTTP Resource",
			"config": map[string]interface{}{
				"port": "2581",
			},
		}, "inends"),
	)
	log.Debug("Create Http Target",
		post(map[string]interface{}{
			"type":        "HTTP",
			"name":        "data to http server",
			"description": "data to http server",
			"config": map[string]interface{}{
				"url": "http://127.0.0.1:3356/data",
			},
		}, "outends"),
	)
}

func TestGrpcInEnd(t *testing.T) {
	engine := engine.NewRuleEngine()
	engine.Start()
	////////
	hh := httpserver.NewHttpApiServer(2580, "../plugin/http_server/templates/*", "./rulx.db", engine)
	if e := engine.LoadPlugin(hh); e != nil {
		log.Fatal("rule load failed:", e)
	}
	hh.Truncate()
	log.Debug(
		post(map[string]interface{}{
			"type":        "GRPC",
			"name":        "GRPC API Server",
			"description": "GRPC Resource",
			"config": map[string]interface{}{
				"port":      "2583",
				"transport": "tcp",
			},
		}, "inends"),
	)
	//
	conn, err := grpc.Dial("127.0.0.1:2583", grpc.WithInsecure())
	if err != nil {
		t.Error(err)
	}
	defer conn.Close()
	c := rulexrpc.NewRulexRpcClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	resp, err := c.Work(ctx, &rulexrpc.Data{
		Value: `{"key":"value"}`,
	})
	if err != nil {
		t.Error(err)
	}
	t.Log(resp)
}
