package assets

import (
	"time"

	"github.com/jessevdk/go-assets"
)

var _Assets2b7ef9c09e8e120a968348f6a0414b24b022fd53 = "\nfunc init() {\n    {{range $_, $item := .Functions}}\n    {{ if $item.Subscription }}registry.RegisterSubscription(\"{{ $item.Name }}\", subscribeTo{{ $item.Name }}, subscriptionTypes{{ $item.Name }} ){{ else }}registry.RegisterFunction(\"{{ $item.Name }}\", callAdapterFor{{ $item.Name }}){{ end }}{{end}}\n}\n"
var _Assets1ad8fcf1a012f1e976cdf59512da19e040972051 = "/** {{range $_, $item := .Comments}}\n * {{ $item }}{{end}}\n */\nexport type {{ .Name }} = { {{range $_, $item := .Field}} {{range $_, $comment := $item.Comment }}\n    // {{ $comment }} {{end}}\n    {{ $item.Name }}: {{ $item.Type }}, {{end}}\n}\n"
var _Assets693a5f743ee831585a375e293545870ad9775df1 = "{{ if .Subscription }}\n/**{{range $_, $item := .Comments}}\n * {{ $item }}{{end}}\n */\nexport async function {{ .Name }}({{range $index, $item := .Params}}{{if $index}}, {{end}}{{ $item.Name }}: {{ $item.Type }}{{end}}, callback: (e: {{ .Subscription }}) => void ) : GoSubscription {\n   subscription = DeviceEventEmitter.addListener('StageSubscription:', callback);\n\n   try {\n      await subribeApiCall('{{ .Name }}', [{{range $index, $item := .Params}}{{if $index}}, {{end}}{{ $item.Name }}{{end}}], (json: string) => { callback(JSON.parse(json)) })\n   } catch(error) {\n      console.error(\"Call of {{ .Name }} failed\", error)\n   }\n\n   return subscription\n}\n\n{{ else }}\n/**{{range $_, $item := .Comments}}\n * {{ $item }}{{end}}\n */\nexport async function {{ .Name }}({{range $index, $item := .Params}}{{if $index}}, {{end}}{{ $item.Name }}: {{ $item.Type }}{{end}}) : Promise<{{ .ReturnType }}> {\n   try {\n        const jsonString = await runApiCall('{{ .Name }}', [{{range $index, $item := .Params}}{{if $index}}, {{end}}{{ $item.Name }}{{end}}])\n        return JSON.parse(jsonString)\n   } catch(error) {\n        console.error(\"Call of {{ .Name }} failed\", error)\n        throw error\n   }\n}\n{{ end }}\n"
var _Assets3b4dd3b006d667b8c4ecbff30ccd3c1161228224 = "\n{{define \"typeCast\"}}\n   {{if eq .Type \"string\"}} return arg.(string), nil {{end}}\n   {{if eq .Type \"int\"}} \n      str, ok := arg.(string) \n      if ok {\n         return strconv.Atoi(str)\n      }\n      fl, ok := arg.(float64)\n      if ok {\n         return int(fl), nil\n      }\n\n      return arg.(int), nil\n   {{end}}\n   {{if eq .Type \"float32\"}} \n      str, ok := arg.(string) \n      if ok {\n         fl, err := strconv.ParseFloat(float32) \n         if err == nil {\n            return fl\n         }\n         return 0, errors.New(\"invalid data\")\n      }\n      return arg.(float32), nil\n   {{end}}\n{{end}}\n{{define \"atomTypeCast\"}}\n            {{if eq .SimpleType \"string\"}} return arg.(string), nil {{end}}\n            {{if eq .SimpleType \"int\"}} \n               str, ok := arg.(string) \n               if ok {\n                  return strconv.Atoi(str)\n               }\n\n               fl, ok := arg.(float64)\n               if ok {\n                  return int(fl), nil\n               }\n\n               return arg.(int), nil\n            {{end}}\n            {{if eq .SimpleType \"float32\"}} \n               str, ok := arg.(string) \n               if ok {\n                  fl, err := strconv.ParseFloat(float32) \n                  if err == nil {\n                     return fl\n                  }\n                  return 0, errors.New(\"invalid data\")\n               }\n               return arg.(float32), nil\n            {{end}}\n{{end}}\n{{define \"argCast\"}}{{if .RichType.Array}} \n      argsSlice := arg.([]interface{}) \n      outArray := make([]{{ .RichType.SimpleType }}, 0, len(argsSlice))\n\n      for _, element := range argsSlice {\n         item, err := func(arg interface{}) ({{ .RichType.SimpleType }}, error) {\n            {{template \"atomTypeCast\" .RichType}}\n         }(element)\n\n         if err != nil {\n            return outArray, err\n         }\n\n         outArray = append(outArray, item)         \n      }\n\n      return outArray, nil\n{{else}}{{template \"typeCast\" .}}{{end}}\n{{end}}\n\n{{ if .Subscription }}\n/**{{range $_, $item := .Comments}}\n * {{ $item }}{{end}}\n */\nfunc subscribeTo{{ .Name }}(callData map[string]interface{}, event goapi.EventCallback) (goapi.Subscription, error) {\n   args, ok := callData[\"args\"].([]interface{})\n\tif !ok {\n\t\treturn nil,errors.New(\"not able to cast args, wrong type\")\n\t}\n\n   {{range $index, $item := .Params}}\n\n   {{ $item.Name }}, err := func(arg interface{}) ({{ $item.Type }}, error) { {{ template \"argCast\" $item }}\n   }(args[{{$index}}])\n   if err != nil {\n      return nil, err\n   }{{ end }}\n\n   return {{ .Package}}.{{ .Name }}({{range $index, $item := .Params}}{{if $index}}, {{end}}{{ $item.Name }}{{end}},event)\n}\n\n\n/**{{range $_, $item := .Comments}}\n * {{ $item }}{{end}}\n */\nfunc subscriptionTypes{{ .Name }}(args []interface{}) ([]interface{}, error) {\n   {{range $index, $item := .Params}}\n\n   result := make([]interface{}, 0, len(args))\n\n   {{ $item.Name }}, err := func(arg interface{}) ({{ $item.Type }}, error) { {{ template \"argCast\" $item }}\n   }(args[{{$index}}])\n   if err != nil {\n      return nil, err\n   }\n   result = append(result, {{ $item.Name }})\n   {{ end }}\n\n   return result, nil\n}\n\n{{ else }}\nfunc callAdapterFor{{ .Name }}(callData map[string]interface{}, callback goapi.JsCallback) error {\n   args, ok := callData[\"args\"].([]interface{})\n\tif !ok {\n\t\treturn errors.New(\"not able to cast args, wrong type\")\n\t}\n\n   {{range $index, $item := .Params}}\n\n   {{ $item.Name }}, err := func(arg interface{}) ({{ $item.Type }}, error) { {{ template \"argCast\" $item }}\n   }(args[{{$index}}])\n   if err != nil {\n      return err\n   }{{ end }}\n\n   {{ .Package }}.{{ .Name }}({{range $index, $item := .Params}}{{if $index}}, {{end}}{{ $item.Name }}{{end}}, callback)\n   return nil\n}\n\n{{ end }}\n"
var _Assets382e2ce7c28616d04646a0aad561f67e2a6e1c39 = "/**\n * GoCall library binding\n * flow\n */\n\nimport {\n  NativeModules,\n  DeviceEventEmitter,\n  EmitterSubscription\n} from 'react-native';\n\ntype GoSubscription = {\n   subsription: EmitterSubscription,\n   name: string,\n   args: any[],\n   devId: number,\n};\n\n{{if .Dev }}\nclass RemoveDev {\n  server = \"\"\n  requestId = 1\n  call = {}\n  event = {}\n  ws:WebSocket\n  pendingList = []\n\n  constructor(server : string) {\n    this.server = server\n    this.connect()\n  }\n\n  connect = () => {\n    const ws = new WebSocket(this.server)\n    ws.onmessage = this.onMessage\n\n    this.ws = ws\n    this.ws.onopen = this.sendPending\n    ws.onclose = () => {\n       // Try to reconnect in 5 seconds\n       setTimeout(() => { this.connect()  }, 1000);\n   };\n  }\n\n  onMessage = (message: any) => {\n    const messages = message.data.split(\"\\n\")\n    messages.forEach((content: string) => {\n      const response = JSON.parse(content)\n\n      if (this.call[response.ID]) {\n        this.call[response.ID](response)\n      } else  if (response.EventName) {\n        this.event[response.EventName](response)\n      }\n    })\n  }\n\n  sendPending = () => {\n    this.pendingList.forEach(it => this.ws.send(it))\n  }\n\n  callMethod = (name: string, args :any[]) : Promise<any> => {\n    return new Promise((resolve, reject) => {\n      const callData = {\n        args,\n        method: name,\n      }\n\n      const requestID = this.requestId++\n\n      this.call[requestID] = (response: any) => {\n        if (response.Success) {\n          resolve(JSON.stringify(response.Success))\n        } else {\n          reject(JSON.stringify(response.Error))\n        }\n\n        this.call[requestID] = null\n      }\n\n      const body = JSON.stringify({id: requestID, call: callData })\n      try  {\n        this.ws.send(body)\n      } catch (err) {\n        this.pendingList = [...this.pendingList, body]\n      }\n    })\n  }\n\n  cancel = (name: string, args :any[]) : GoSubscriptions => {\n    const callData = {\n      args,\n      event: name,\n    }\n\n    this.requestId++\n\n    this.event[eventName] = (response: any) => {\n      if (response.Body) {\n        callback(JSON.stringify(response.Body))\n      }\n    }\n\n    const body = JSON.stringify({id: this.requestId, cancel: callData })\n    try  {\n      this.ws.send(body)\n    } catch (err) {\n      this.pendingList = [...this.pendingList, body]\n    }\n\n    return { name, eventName, devId: this.requestId }\n  }\n\n  subscribe = (name: string, args :any[], eventName: string, callback: (json : String) => void) : GoSubscriptions => {\n    const callData = {\n      args,\n      event: name,\n    }\n\n    this.requestId++\n\n    this.event[eventName] = (response: any) => {\n      if (response.Body) {\n        callback(JSON.stringify(response.Body))\n      }\n    }\n\n    const body = JSON.stringify({id: this.requestId, subscribe: callData })\n    try  {\n      this.ws.send(body)\n    } catch (err) {\n      this.pendingList = [...this.pendingList, body]\n    }\n\n    return { name, eventName, devId: this.requestId }\n  }\n}\n\nconst devCall = new RemoveDev(\"ws://localhost:{{.Port}}/ws\");\n\n{{end}}\n\nfunction getName(name: string, args :any[]) : string {\n   body = args.reduce((acc:string, value: any) => {\n      if (acc == \"\") {\n         return value\n      }\n\n      return acc + \" : \" + value\n   }, \"\")\n\n   return `${name}:${body}`\n}\n\nfunction subribeApiCall(name: string, args :any[], callback: (json : String) => void) : GoSubscription {\n   {{if .Dev}}\n   const subscriptionName = getName(name, args)\n   return devCall.subscribe(name, args, subscriptionName, callback)\n   {{else}}\n   const subscriptionName = getName(name, args)\n   subsription = DeviceEventEmitter.addListener(subscriptionName, callback);\n\n   const callData = JSON.stringify({\n      args,\n      event: name,\n   })\n\n   NativeModules.GoCall.subscribe(callData)\n   \n   return {\n      args,\n      name,\n      subsription,\n   }\n   {{end}}\n}\n\nasync function cancelSubsriptionApiCall(subscription : GoSubscription) : Promise<any> {\n   {{if .Dev}} \n   const {name, args, subsription} = subscription\n   {{else}}\n    const callData = JSON.stringify({\n      args,\n      event: name,\n    })\n\n    subsription.remove()\n\n    return NativeModules.GoCall.cancel(callData)\n   {{end}}\n}\n\nasync function runApiCall(name: string, args :any[]) : Promise<any> {\n   {{if .Dev}} \n    return devCall.callMethod(name, args)\n   {{else}}\n    const callData = JSON.stringify({\n      args,\n      method: name,\n    })\n\n    return NativeModules.GoCall.callMethod(callData)\n   {{end}}\n}"
var _Assets0533dda96e7fff99871c332d671b47636555c60c = "package {{.Package}}\n\nimport (\n\t\"encoding/json\"\n\t\"log\"\n\t\"errors\"\n\t\"strconv\"\n\t\"gitlab.vmassive.ru/gocallgen/goapi\"\n\t{{if .Dev}}\n\t\"gitlab.vmassive.ru/gocallgen/goapi/remgo\"\n\t{{end}}\n)\n\nvar registry goapi.JsRegistry = goapi.NewJsRegistry()\n\n{{if .Dev }}\nfunc serveHome(w http.ResponseWriter, r *http.Request) {\n\tlog.Println(r.URL)\n\tif r.URL.Path != \"/\" {\n\t\thttp.Error(w, \"Not found\", http.StatusNotFound)\n\t\treturn\n\t}\n\tif r.Method != \"GET\" {\n\t\thttp.Error(w, \"Method not allowed\", http.StatusMethodNotAllowed)\n\t\treturn\n\t}\n\thttp.ServeFile(w, r, \"home.html\")\n}\n\nfunc main() {\n\thub := remgo.NewHub()\n\tgo hub.Run(&registry)\n\n\thttp.HandleFunc(\"/\", serveHome)\n\thttp.HandleFunc(\"/ws\", func(w http.ResponseWriter, r *http.Request) {\n\t\tremgo.ServeWs(&registry, hub, w, r)\n\t})\n\terr := http.ListenAndServe(\"0.0.0.0:9009\", nil)\n\tif err != nil {\n\t\tlog.Fatal(\"ListenAndServe: \", err)\n\t}\n}\n{{end}}\n// JsCallback the interface for any callbacks\ntype JsCallback interface {\n\tOnSuccess(json string)\n\tOnError(json string)\n}\n\n// JsEvent the interface for any events\ntype JsEvent interface {\n\tOnEvent(eventName string, json string)\n}\n\ntype eventerSender struct {\n\tevent\t\tJsEvent\n}\n\nfunc newEventSender(event JsEvent) goapi.JsEvent {\n\treturn &eventerSender{\n\t\tevent: event,\n\t}\n}\n\nfunc (eventer eventerSender) OnEvent(eventName string, data interface{}) {\n\tlog.Printf(\" >> + << %#v\", data)\n\tbytes, _ := json.Marshal(data)\n\teventer.event.OnEvent(eventName, string(bytes))\n}\n\ntype callbackCaller struct {\n\tcallback JsCallback\n}\n\nfunc (caller callbackCaller) OnSuccess(data interface{}) {\n\tlog.Printf(\" >> + << %#v\", data)\n\tbytes, _ := json.Marshal(data)\n\tcaller.callback.OnSuccess(string(bytes))\n}\n\nfunc (caller callbackCaller) OnError(data interface{}) {\n\tbytes, _ := json.Marshal(data)\n\tcaller.callback.OnError(string(bytes))\n}\n\nfunc newCaller(callback JsCallback) goapi.JsCallback {\n\treturn &callbackCaller{\n\t\tcallback: callback,\n\t}\n}\n\n// CallMethod - call from JS\nfunc CallMethod(callData string, callback JsCallback) {\n\tmethodCallData := make(map[string]interface{})\n\tjson.Unmarshal([]byte(callData), &methodCallData)\n\n\tcaller := newCaller(callback)\n\tregistry.Call(methodCallData, caller)\n}\n\n\nfunc RegisterEventCallback(callback JsEvent) {\n\tregistry.RegisterEventCallback(newEventSender(callback))\n}\n\nfunc RemoveEventCallback() {\n\tregistry.RegisterEventCallback(nil)\n}\n\n// Subscribe - subsribe from JS\nfunc Subscribe(callData string) {\n\tmethodCallData := make(map[string]interface{})\n\tjson.Unmarshal([]byte(callData), &methodCallData)\n\n\tregistry.Subscribe(methodCallData)\n}\n\n// Subscribe - subsribe from JS\nfunc Cancel(callData string) {\n\tmethodCallData := make(map[string]interface{})\n\tregistry.CancelSubscription(methodCallData)\n}\n"

// Assets returns go-assets FileSystem
var Assets = assets.NewFileSystem(map[string][]string{"/": []string{"templates"}, "/templates": []string{"callmap.go.tmpl", "struct.js.tmpl", "func.js.tmpl", "func.go.tmpl", "head.js.tmpl", "head.go.tmpl"}}, map[string]*assets.File{
	"/": &assets.File{
		Path:     "/",
		FileMode: 0x800001ed,
		Mtime:    time.Unix(1542987597, 1542987597805803795),
		Data:     nil,
	}, "/templates": &assets.File{
		Path:     "/templates",
		FileMode: 0x800001ed,
		Mtime:    time.Unix(1542982329, 1542982329233381397),
		Data:     nil,
	}, "/templates/callmap.go.tmpl": &assets.File{
		Path:     "/templates/callmap.go.tmpl",
		FileMode: 0x1a4,
		Mtime:    time.Unix(1542982322, 1542982322722000000),
		Data:     []byte(_Assets2b7ef9c09e8e120a968348f6a0414b24b022fd53),
	}, "/templates/struct.js.tmpl": &assets.File{
		Path:     "/templates/struct.js.tmpl",
		FileMode: 0x1a4,
		Mtime:    time.Unix(1542982329, 1542982329233000000),
		Data:     []byte(_Assets1ad8fcf1a012f1e976cdf59512da19e040972051),
	}, "/templates/func.js.tmpl": &assets.File{
		Path:     "/templates/func.js.tmpl",
		FileMode: 0x1a4,
		Mtime:    time.Unix(1542982322, 1542982322721000000),
		Data:     []byte(_Assets693a5f743ee831585a375e293545870ad9775df1),
	}, "/templates/func.go.tmpl": &assets.File{
		Path:     "/templates/func.go.tmpl",
		FileMode: 0x1a4,
		Mtime:    time.Unix(1542982322, 1542982322721000000),
		Data:     []byte(_Assets3b4dd3b006d667b8c4ecbff30ccd3c1161228224),
	}, "/templates/head.js.tmpl": &assets.File{
		Path:     "/templates/head.js.tmpl",
		FileMode: 0x1a4,
		Mtime:    time.Unix(1542982327, 1542982327096000000),
		Data:     []byte(_Assets382e2ce7c28616d04646a0aad561f67e2a6e1c39),
	}, "/templates/head.go.tmpl": &assets.File{
		Path:     "/templates/head.go.tmpl",
		FileMode: 0x1a4,
		Mtime:    time.Unix(1542989531, 1542989531843781995),
		Data:     []byte(_Assets0533dda96e7fff99871c332d671b47636555c60c),
	}}, "")
