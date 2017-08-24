package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/golang/protobuf/proto"
	"github.com/juju/errors"
	"github.com/pingcap/tipb/go-mysqlx"
	"github.com/pingcap/tipb/go-mysqlx/Connection"
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
	return fmt.Sprintf("%s.txt", name[0:len(name)-8])
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

func deal_with_client_message(mtype Mysqlx.ClientMessages_Type, payload []byte, f *os.File) {
	var msg = "client message type: %s, content: %s\n"
	var typeS = mtype.String()
	var contentS = "NO CONTENT"
	switch mtype {
	case Mysqlx.ClientMessages_CON_CAPABILITIES_SET:
		var caps Mysqlx_Connection.CapabilitiesSet
		proto.Unmarshal(payload, &caps)
		contentS = caps.String()
	case Mysqlx.ClientMessages_CON_CAPABILITIES_GET:
	case Mysqlx.ClientMessages_SESS_AUTHENTICATE_START:
		var authStart Mysqlx_Session.AuthenticateStart
		proto.Unmarshal(payload, &authStart)
		contentS = authStart.String()
	case Mysqlx.ClientMessages_SESS_AUTHENTICATE_CONTINUE:
		var authCont Mysqlx_Session.AuthenticateContinue
		proto.Unmarshal(payload, &authCont)
		contentS = authCont.String()
	case Mysqlx.ClientMessages_SQL_STMT_EXECUTE:
		var sql Mysqlx_Sql.StmtExecute
		proto.Unmarshal(payload, &sql)
		contentS = sql.String()
	}
	msg = fmt.Sprintf(msg, typeS, contentS)
	f.WriteString(msg)
}

func deal_with_server_message(mtype Mysqlx.ServerMessages_Type, payload []byte, f *os.File) {
	var msg = "server message type: %s, content: %s\n"
	var typeS = mtype.String()
	var contentS = "NO CONTENT"
	switch mtype {
	case Mysqlx.ServerMessages_OK:
		var ok Mysqlx.Ok
		proto.Unmarshal(payload, &ok)
		contentS = ok.String()
	case Mysqlx.ServerMessages_CONN_CAPABILITIES:
		var caps Mysqlx_Connection.Capabilities
		proto.Unmarshal(payload, &caps)
		contentS = caps.String()
	case Mysqlx.ServerMessages_SESS_AUTHENTICATE_CONTINUE:
		var authCont Mysqlx_Session.AuthenticateContinue
		proto.Unmarshal(payload, &authCont)
		contentS = authCont.String()
	case Mysqlx.ServerMessages_NOTICE:
		var notice Mysqlx_Notice.Frame
		proto.Unmarshal(payload, &notice)
		contentS = notice.String()
		switch notice.GetType() {
		case 1: // warning
		case 2: // session veriable changed
		case 3: // session state changed
			var ssc Mysqlx_Notice.SessionStateChanged
			proto.Unmarshal(payload, &ssc)
			contentS = fmt.Sprintf("(scope: %s, payload: %s)", notice.GetScope().String(), ssc.String())
		}
	case Mysqlx.ServerMessages_SESS_AUTHENTICATE_OK:
		var authOk Mysqlx_Session.AuthenticateOk
		proto.Unmarshal(payload, &authOk)
		contentS = authOk.String()
	case Mysqlx.ServerMessages_RESULTSET_COLUMN_META_DATA:
		var rcmd Mysqlx_Resultset.ColumnMetaData
		proto.Unmarshal(payload, &rcmd)
		contentS = rcmd.String()
	case Mysqlx.ServerMessages_RESULTSET_ROW:
		var row Mysqlx_Resultset.Row
		proto.Unmarshal(payload, &row)
		contentS = row.String()
	case Mysqlx.ServerMessages_RESULTSET_FETCH_DONE:
		var done Mysqlx_Resultset.FetchDone
		proto.Unmarshal(payload, &done)
		contentS = done.String()
	case Mysqlx.ServerMessages_SQL_STMT_EXECUTE_OK:
		var sqlOk Mysqlx_Sql.StmtExecuteOk
		proto.Unmarshal(payload, &sqlOk)
		contentS = sqlOk.String()
	}
	msg = fmt.Sprintf(msg, typeS, contentS)
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

	var length uint32 = 0
	var message_type uint8 = 0
	var payloadLen int
	var payload = make([]byte, 1e8)
	for {
		if err = binary.Read(fin, binary.LittleEndian, &length); err != nil {
			break
		}
		if err = binary.Read(fin, binary.LittleEndian, &message_type); err != nil {
			break
		}
		if payloadLen, err = fin.Read(payload[0 : length-1]); err != nil {
			break
		}
		if isClient {
			deal_with_client_message(Mysqlx.ClientMessages_Type(message_type), payload[0:payloadLen], fout)
		} else {
			deal_with_server_message(Mysqlx.ServerMessages_Type(message_type), payload[0:payloadLen], fout)
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
