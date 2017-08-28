package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/juju/errors"
	"github.com/pingcap/tipb/go-mysqlx"
	"github.com/pingcap/tipb/go-mysqlx/Connection"
	"github.com/pingcap/tipb/go-mysqlx/Crud"
	"github.com/pingcap/tipb/go-mysqlx/Expect"
	"github.com/pingcap/tipb/go-mysqlx/Notice"
	"github.com/pingcap/tipb/go-mysqlx/Resultset"
	"github.com/pingcap/tipb/go-mysqlx/Session"
	"github.com/pingcap/tipb/go-mysqlx/Sql"
)

var (
	clientFile = flag.String("client", "", "client tcpdump file")
	serverFile = flag.String("server", "", "server tcpdump file")
)

func rename(name string) string {
	return fmt.Sprintf("%s.txt", name[:len(name)-8])
}

func inAndOut(name string) (*os.File, *os.File, error) {
	fin, err := os.Open(name)
	if err != nil {
		return nil, nil, errors.Trace(err)
	}
	outName := rename(name)
	fout, err := os.OpenFile(outName, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, nil, errors.Trace(err)
	}
	return fin, fout, nil
}

func dealClientMsg(mtype Mysqlx.ClientMessages_Type, payload []byte, f *os.File) {
	var msg = "client message type: %s, content: %s\n"
	var typeS = mtype.String()
	var contentS = "NO CONTENT"
	switch mtype {
	case Mysqlx.ClientMessages_CON_CAPABILITIES_GET:
		var data Mysqlx_Connection.CapabilitiesGet
		data.Unmarshal(payload)
		contentS = data.String()
	case Mysqlx.ClientMessages_CON_CAPABILITIES_SET:
		var data Mysqlx_Connection.CapabilitiesSet
		data.Unmarshal(payload)
		contentS = data.String()
	case Mysqlx.ClientMessages_CON_CLOSE:
		var data Mysqlx_Connection.Close
		data.Unmarshal(payload)
		contentS = data.String()
	case Mysqlx.ClientMessages_SESS_AUTHENTICATE_START:
		var data Mysqlx_Session.AuthenticateStart
		data.Unmarshal(payload)
		contentS = data.String()
	case Mysqlx.ClientMessages_SESS_AUTHENTICATE_CONTINUE:
		var data Mysqlx_Session.AuthenticateContinue
		data.Unmarshal(payload)
		contentS = data.String()
	case Mysqlx.ClientMessages_SESS_RESET:
		var data Mysqlx_Session.Reset
		data.Unmarshal(payload)
		contentS = data.String()
	case Mysqlx.ClientMessages_SESS_CLOSE:
		var data Mysqlx_Session.Close
		data.Unmarshal(payload)
		contentS = data.String()
	case Mysqlx.ClientMessages_SQL_STMT_EXECUTE:
		var data Mysqlx_Sql.StmtExecute
		data.Unmarshal(payload)
		contentS = data.String()
	case Mysqlx.ClientMessages_CRUD_FIND:
		var data Mysqlx_Crud.Find
		data.Unmarshal(payload)
		contentS = data.String()
	case Mysqlx.ClientMessages_CRUD_INSERT:
		var data Mysqlx_Crud.Insert
		data.Unmarshal(payload)
		contentS = data.String()
	case Mysqlx.ClientMessages_CRUD_UPDATE:
		var data Mysqlx_Crud.Update
		data.Unmarshal(payload)
		contentS = data.String()
	case Mysqlx.ClientMessages_CRUD_DELETE:
		var data Mysqlx_Crud.Delete
		data.Unmarshal(payload)
		contentS = data.String()
	case Mysqlx.ClientMessages_EXPECT_OPEN:
		var data Mysqlx_Expect.Open
		data.Unmarshal(payload)
		contentS = data.String()
	case Mysqlx.ClientMessages_EXPECT_CLOSE:
		var data Mysqlx_Expect.Close
		data.Unmarshal(payload)
		contentS = data.String()
	case Mysqlx.ClientMessages_CRUD_CREATE_VIEW:
		var data Mysqlx_Crud.CreateView
		data.Unmarshal(payload)
		contentS = data.String()
	case Mysqlx.ClientMessages_CRUD_MODIFY_VIEW:
		var data Mysqlx_Crud.ModifyView
		data.Unmarshal(payload)
		contentS = data.String()
	case Mysqlx.ClientMessages_CRUD_DROP_VIEW:
		var data Mysqlx_Crud.DropView
		data.Unmarshal(payload)
		contentS = data.String()
	}
	msg = fmt.Sprintf(msg, typeS, "\"" + contentS + "\"")
	f.WriteString(msg)
}

func dealServerMsg(mtype Mysqlx.ServerMessages_Type, payload []byte, f *os.File) {
	var msg = "server message type: %s, content: %s\n"
	var typeS = mtype.String()
	var contentS = "NO CONTENT"
	switch mtype {
	case Mysqlx.ServerMessages_OK:
		var data Mysqlx.Ok
		data.Unmarshal(payload)
		contentS = data.String()
	case Mysqlx.ServerMessages_ERROR:
		var data Mysqlx.Error
		data.Unmarshal(payload)
		contentS = data.String()
	case Mysqlx.ServerMessages_CONN_CAPABILITIES:
		var data Mysqlx_Connection.Capabilities
		data.Unmarshal(payload)
		contentS = data.String()
	case Mysqlx.ServerMessages_SESS_AUTHENTICATE_CONTINUE:
		var data Mysqlx_Session.AuthenticateContinue
		data.Unmarshal(payload)
		contentS = data.String()
	case Mysqlx.ServerMessages_SESS_AUTHENTICATE_OK:
		var data Mysqlx_Session.AuthenticateOk
		data.Unmarshal(payload)
		contentS = data.String()
	case Mysqlx.ServerMessages_NOTICE:
		var data Mysqlx_Notice.Frame
		data.Unmarshal(payload)
		contentS = data.String()
		switch data.GetType() {
		case 1: // warning
			var data Mysqlx_Notice.Warning
			data.Unmarshal(payload)
			contentS += fmt.Sprintf("\"\nNotice Type: Warning, Notice Payload: \"%s", data.String())
		case 2: // session variable changed
			var data Mysqlx_Notice.SessionVariableChanged
			data.Unmarshal(payload)
			contentS += fmt.Sprintf("\"\nNotice Type: SessionVariableChanged, Notice Payload: \"%s", data.String())
		case 3: // session state changed
			var data Mysqlx_Notice.SessionStateChanged
			data.Unmarshal(payload)
			contentS += fmt.Sprintf("\"\nNotice Type: SessionStateChanged, Notice Payload: \"%s", data.String())
		}
	case Mysqlx.ServerMessages_RESULTSET_COLUMN_META_DATA:
		var data Mysqlx_Resultset.ColumnMetaData
		data.Unmarshal(payload)
		contentS = data.String()
	case Mysqlx.ServerMessages_RESULTSET_ROW:
		var data Mysqlx_Resultset.Row
		data.Unmarshal(payload)
		contentS = data.String()
	case Mysqlx.ServerMessages_RESULTSET_FETCH_DONE:
		var data Mysqlx_Resultset.FetchDone
		data.Unmarshal(payload)
		contentS = data.String()
	case Mysqlx.ServerMessages_RESULTSET_FETCH_SUSPENDED:
	case Mysqlx.ServerMessages_RESULTSET_FETCH_DONE_MORE_RESULTSETS:
		var data Mysqlx_Resultset.FetchDoneMoreResultsets
		data.Unmarshal(payload)
		contentS = data.String()
	case Mysqlx.ServerMessages_SQL_STMT_EXECUTE_OK:
		var data Mysqlx_Sql.StmtExecuteOk
		data.Unmarshal(payload)
		contentS = data.String()
	case Mysqlx.ServerMessages_RESULTSET_FETCH_DONE_MORE_OUT_PARAMS:
		var data Mysqlx_Resultset.FetchDoneMoreOutParams
		data.Unmarshal(payload)
		contentS = data.String()
	}
	msg = fmt.Sprintf(msg, typeS, "\"" + contentS + "\"")
	f.WriteString(msg)
}

func extractMessages(isClient bool) (err error) {
	var file string
	if isClient {
		file = *clientFile
	} else {
		file = *serverFile
	}

	fin, fout, err := inAndOut(file)
	if err != nil {
		return errors.Trace(err)
	}
	defer fin.Close()
	defer fout.Close()

	var length uint32
	var messageType uint8
	var payloadLen int
	var payload = make([]byte, 6553600)
	for {
		if err = binary.Read(fin, binary.LittleEndian, &length); err != nil {
			break
		}
		if err = binary.Read(fin, binary.LittleEndian, &messageType); err != nil {
			break
		}
		if payloadLen, err = fin.Read(payload[:length-1]); err != nil {
			break
		}
		if payloadLen != int(length-1) {
			fmt.Println("yusp", length, payloadLen)
		}
		if isClient {
			dealClientMsg(Mysqlx.ClientMessages_Type(messageType), payload[:payloadLen], fout)
		} else {
			dealServerMsg(Mysqlx.ServerMessages_Type(messageType), payload[:payloadLen], fout)
		}
	}
	if err == io.EOF {
		err = nil
	}
	return errors.Trace(err)
}

func main() {
	flag.Parse()
	if len(*clientFile) == 0 || len(*serverFile) == 0 {
		fmt.Printf("wrong arguments")
		os.Exit(-1)
	}

	if err := extractMessages(true); err != nil {
		fmt.Printf("error: %s\n", err)
	}
	if err := extractMessages(false); err != nil {
		fmt.Printf("error: %s\n", err)
	}
}
