package mosquitto

import (
	// Frameworks
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	gopi "github.com/djthorpe/gopi/v2"
	iface "github.com/djthorpe/mosquitto"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type event struct {
	Source_     gopi.Unit
	Type_       iface.Flags
	ReturnCode_ int
	Id_         int
	Topic_      string
	Data_       []byte
}

////////////////////////////////////////////////////////////////////////////////
// NEW MESSAGES

func NewConnect(source gopi.Unit, returnCode int) iface.Event {
	this := new(event)
	this.Source_ = source
	this.Type_ = iface.MOSQ_FLAG_EVENT_CONNECT
	this.ReturnCode_ = returnCode
	return this
}

func NewDisconnect(source gopi.Unit, returnCode int) iface.Event {
	this := new(event)
	this.Source_ = source
	this.Type_ = iface.MOSQ_FLAG_EVENT_DISCONNECT
	this.ReturnCode_ = returnCode
	return this
}

func NewSubscribe(source gopi.Unit, id int) iface.Event {
	this := new(event)
	this.Source_ = source
	this.Type_ = iface.MOSQ_FLAG_EVENT_SUBSCRIBE
	this.Id_ = id
	return this
}

func NewUnsubscribe(source gopi.Unit, id int) iface.Event {
	this := new(event)
	this.Source_ = source
	this.Type_ = iface.MOSQ_FLAG_EVENT_UNSUBSCRIBE
	this.Id_ = id
	return this
}

func NewPublish(source gopi.Unit, id int) iface.Event {
	this := new(event)
	this.Source_ = source
	this.Type_ = iface.MOSQ_FLAG_EVENT_PUBLISH
	this.Id_ = id
	return this
}

func NewMessage(source gopi.Unit, id int, topic string, data []byte) iface.Event {
	this := new(event)
	this.Source_ = source
	this.Type_ = iface.MOSQ_FLAG_EVENT_MESSAGE
	this.Id_ = id
	this.Topic_ = topic
	this.Data_ = data
	return this
}

////////////////////////////////////////////////////////////////////////////////
// IMPLEMENTATION gopi.Event

func (*event) Name() string {
	return "mosquitto.Event"
}

func (*event) NS() gopi.EventNS {
	return gopi.EVENT_NS_DEFAULT
}

func (this *event) Source() gopi.Unit {
	return this.Source_
}

func (this *event) Value() interface{} {
	return this.Data_
}

////////////////////////////////////////////////////////////////////////////////
// IMPLEMENTATION mosquitto.Event

func (this *event) Type() iface.Flags {
	return this.Type_
}

func (this *event) Topic() string {
	return this.Topic_
}

func (this *event) ReturnCode() int {
	return this.ReturnCode_
}

func (this *event) Id() int {
	return this.Id_
}

func (this *event) Data() []byte {
	return this.Data_
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *event) String() string {
	str := "<" + this.Name() +
		" type=" + fmt.Sprint(this.Type_)
	if this.Type_ == iface.MOSQ_FLAG_EVENT_CONNECT || this.Type_ == iface.MOSQ_FLAG_EVENT_DISCONNECT {
		str += " returnCode=" + fmt.Sprint(this.ReturnCode_)
	}
	if this.Id_ != 0 {
		str += " id=" + fmt.Sprint(this.Id_)
	}
	if this.Topic_ != "" {
		str += " topic=" + strconv.Quote(this.Topic_)
	}
	if this.Data_ != nil {
		str += " data=" + strings.ToUpper(hex.EncodeToString(this.Data_))
	}
	return str + ">"
}
