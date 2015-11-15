package cli

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	//"github.com/socialpoint/sprocket/pkg/dumper"
)

type CliSession struct {
	Conn net.Conn
	finish bool
}

type Definition struct{
	Handler Handler
	Description string
}

type Handler func([]interface{}) (interface{}, error)

type CliHandler interface {
	CliHandlers() map[string]Definition
}

type Command struct {
	Name string
	Args []interface{}
	Fn   Handler
}

func (cmd *Command) Execute() (interface{}, error) {
	return cmd.Fn(cmd.Args)
}

type Server struct {
	port int
	handlers map[string]Definition
}

func New(port int) *Server {
	srv :=  &Server{
		port: port,
		handlers: make(map[string]Definition),
	}

	srv.Register("help", Definition{srv.Help, "Display registered commands"})

	return srv
}

func (s *Server) Run() {
	addy, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("0.0.0.0:%d", s.port))
	if err != nil {
		log.Fatal(err)
	}

	inbound, err := net.ListenTCP("tcp", addy)
	if err != nil {
		log.Fatal(err)
	}

	for {
		conn, err := inbound.Accept()
		if err != nil {
			log.Println("rpc.Serve: accept:", err.Error())

			continue
		}
		go s.handleConn(conn)
	}
}

func (s *Server) RegisterCommands(i CliHandler) {
	cmds, ok := i.(CliHandler)
	if ok {
		for cmd, h := range cmds.CliHandlers() {
			s.Register(cmd, h)
		}
	}
}

func (s *Server) Register(name string, h Definition) {
	s.handlers[name] = h
}

func (s *Server) handleConn(conn net.Conn) {
	buf := bufio.NewReader(conn)
	for {
		buffer, err := buf.ReadBytes('\n')
		if err != nil {
			if err != io.ErrClosedPipe && err != io.EOF {
				log.Print("Socket Reader err: ", err)
			}
		}

		//@TODO: Quit command ends session
		if string(buffer) == "quit\r\n" {
			conn.Close()
			return
		}

		result, err := s.HandleReq(buffer)
		if err != nil {
			log.Println("Error ", err)
		}

		if result != nil {
			output := s.prepareResponse(result)
			conn.Write(output)
		}
	}
}

func (s *Server) HandleReq(b []byte) (interface{}, error) {
	cmdParts := strings.Split(string(b), "\r\n")
	cmdParts = strings.Split(cmdParts[0], " ")
	f, ok := s.handlers[cmdParts[0]]
	if !ok {
		return nil, errors.New("Command Not found \n")
	}

	args := []interface{}{}
	for _, v := range cmdParts[1:] {
		args = append(args, v)
	}

	cmd := &Command{
		Name: cmdParts[0],
		Args: args,
		Fn:   f.Handler,
	}

	res, err := cmd.Execute()
	if err != nil {
		log.Println("Error executing", cmd.Name)
		return nil, err
	}

	return res, nil
}

func (s *Server) prepareResponse(rsp interface{}) []byte {
	switch rsp.(type) {
	default:
		r, err := json.Marshal(rsp)
		if err != nil {
			log.Println("Error encoding response ", err)
			return nil
		}
		return []byte(string(r) + "\n")
	case int:
		return []byte(fmt.Sprintf("%d\n", rsp))
	case string:
		return []byte(rsp.(string) + "\n")
	}
}

func (s *Server) Help(args []interface{}) (interface{}, error) {
	content := ""
	for k, v := range s.handlers {
		content = content + fmt.Sprintf("Command %s Description %s \n ", k, v.Description)
	}

	return content, nil
}