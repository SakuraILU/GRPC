package server

import (
	"fmt"
	codec "grpc/Codec"
	"log"
	"net"
	"reflect"
	"sync"
)

type RequestHandler struct {
	codecer  codec.ICodec
	sendlock sync.RWMutex
}

func NewRequestHandler(conn net.Conn) (rh *RequestHandler, err error) {
	opt, err := codec.ParseOption(conn)

	if err != nil {
		log.Println("Init request handler failed")
		return nil, err
	}

	codec_newfun, ok := codec.CodecNewFuncs[opt.Codec_type]
	if !ok {
		log.Fatal("wrong opt type")
	}
	rh = &RequestHandler{
		codecer: codec_newfun(conn),
	}

	return rh, nil
}

func (rh *RequestHandler) Read() (r *codec.Request, err error) {
	var head codec.Head
	if err := rh.codecer.ReadHead(&head); err != nil {
		return nil, err
	}

	argv := reflect.New(reflect.TypeOf(""))
	if err = rh.codecer.ReadBody(argv.Interface()); err != nil {
		return
	}
	log.Printf("argv: %v\n", argv.Elem().Interface())

	r = &codec.Request{
		Head: head,
		Argv: argv,
	}
	return
}

func (rh *RequestHandler) Write(head *codec.Head, body interface{}) (err error) {
	rh.sendlock.Lock()
	defer rh.sendlock.Unlock()

	err = rh.codecer.Write(head, body)
	return
}

func (rh *RequestHandler) Handle() (err error) {
	req, err := rh.Read()
	if err != nil {
		return
	}

	// TODO: handle request
	// now just reply a message to client, assume the argv is string
	reply := fmt.Sprintf("grpc: pong %d", req.Head.Service_id)
	if err = rh.Write(&req.Head, reply); err != nil {
		return
	}

	return
}
