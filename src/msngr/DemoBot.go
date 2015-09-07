package msngr

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

func genId() string {
	//не привязывайся ко времени, может бть в 1 микросекуну много сообщений и ид долэны ыть разными
	return fmt.Sprintf("%d", time.Now().UnixNano()/int64(time.Millisecond))
}

type Notifier struct {
	address string
}

func NewNotifier(addr string) *Notifier {
	return &Notifier{address: addr}
}

func (n Notifier) Notify(order_id int64, state int) {
	log.Println("Notifier notifying at:", n.address, "that order id:", order_id, "have state: ", state)
	http.Post(n.address, "text", bytes.NewBufferString(fmt.Sprintf("to order %v i have state %v", order_id, state)))
}

func getInPackage(r *http.Request) (InPkg, error) {

	var in InPkg

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("error at reading: %q \n", err)
	}

	err = json.Unmarshal(body, &in)
	if err != nil {
		log.Printf("error at unmarshal: %q \n", err)
	}

	log.Printf("request data is:\n%+v\n", in)
	return in, err
}

func setOutPackage(w http.ResponseWriter, out OutPkg) {

	jsoned_out, err := json.Marshal(&out)
	if err != nil {
		log.Println(jsoned_out, err)
	}
	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)

	fmt.Fprintf(w, "%s", string(jsoned_out))
}

type controllerHandler func(w http.ResponseWriter, r *http.Request)

func FormBotControllerHandler(request_cmds map[string]RequestCommandProcessor, message_cmds map[string]MessageCommandProcessor) controllerHandler {

	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "I can not work with non POST methods", 405)
			return
		}

		in, err := getInPackage(r)
		out := new(OutPkg)

		log.Println("forming response...")

		out.To = in.From

		if in.Request != nil {
			log.Printf("processing request %+v", in)
			action := in.Request.Query.Action
			out.Request = &OutRequest{ID: genId(), Type: "result"}
			out.Request.Query.Action = action
			if commandProcessor, ok := request_cmds[action]; ok {
				out.Request.Query.Result, err = commandProcessor.ProcessRequest(in)
			} else {
				out.Request.Query.Text = "Команда не поддерживается."
			}

		} else if in.Message != nil {
			log.Printf("processing message %+v", in)
			out.Message = &OutMessage{Type: in.Message.Type, Thread: in.Message.Thread, ID: genId()}
			action := in.Message.Command.Action
			if commandProcessor, ok := message_cmds[action]; ok {
				out.Message.Body, out.Message.Commands, err = commandProcessor.ProcessMessage(in)
			} else {
				out.Message.Body = "Команда не поддерживается."
			}

		}
		log.Printf("%+v\n", out)

		if err != nil {
			out.Message = &OutMessage{Type: "error", Thread: "0", ID: genId(), Body: fmt.Sprintf("%+v", err)}
		}

		setOutPackage(w, *out)
	}

}
