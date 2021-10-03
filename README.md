# Mosquitto

This repository contains a Golang [mosquitto](https://mosquitto.org/) client library, which conforms to the MQTT standard. This documentation includes the following information:

  * What dependencies are needed in order to use this package;
  * Information about the two command-line tools, `mqttpub` and `mqttsub`;
  * Using the `libmosquitto` bindings;
  * Alternatively, using the higher-level package;
  * Building a REST API frontend to mqtt.

This repository is published under the Apache license. 
Please use the [issues](https://github.com/mutablelogic/mosquitto/issues) tab on Github to file bugs, ask for features or
for general discussion.

## Copyright Notice

> Copyright 2020, 2021 David Thorpe
>
>   Licensed under the Apache License, Version 2.0 (the "License");
>   you may not use this file except in compliance with the License.
>   You may obtain a copy of the License at
>
>   http://www.apache.org/licenses/LICENSE-2.0
>
>   Unless required by applicable law or agreed to in writing, software
>   distributed under the License is distributed on an "AS IS" BASIS,
>   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
>   See the License for the specific language governing permissions and
>   limitations under the License.

This repository requires you to install the mosquitto library, which is
licensed under a separate license.

## Dependencies

In order to test on Mac using homebrew:

```sh
bash# brew install mosquitto
bash# git clone git@github.com:djthorpe/mosquitto.git
bash# cd mosquitto
bash# make test
```

Similarly for Debian:

```sh
bash# sudo apt install libmosquitto-dev
bash# git clone git@github.com:djthorpe/mosquitto.git
bash# cd mosquitto
bash# make test
```

## Command-Line Tools

There are two command-line tools, one for publishing messages and one for
subscribing to topics. In order to make these, use the `make` command
within the repository which results in two binaries, `mqttpub` and
`mqttsub`. For example, in order to subscribe to messages:

```sh
bash# mqttsub -mqtt.broker test.mosquitto.org \$SYS/broker/+

TYPE       TOPIC                                    DATA                                    
---------- ---------------------------------------- ----------------------------------------
CONNECT                                             <nil>                                   
SUBSCRIBE                                           <nil>                                   
MESSAGE    $SYS/broker/version                      "mosquitto version 1.6.8"               
MESSAGE    $SYS/broker/uptime                       "231546 seconds"                        
```

(Make sure you use the backslash character where necessary).

In order to publish use the `-topic` flag and one or more arguments. This will publish UTF-8 data on the broker. You can use the `-qos` parameter to set the quality of service to 0, 1 or 2.

```sh
bash# mqttpub -mqtt.broker test.mosquitto.org -topic test "Hello, World"
[INFO] PUBLISH: 1
[INFO] PUBACK: 1
```

## Using the bindings

You can use the following `libmosquitto` bindings in your code. For informaton
about the C API, please see [here](https://mosquitto.org/api/files/mosquitto-h.html):

```go
package mosquitto

const MOSQ_DEFAULT_PORT = 1883

// Initialize the library
func Init() error

// Cleanup the library
func Cleanup() error

// Return library version as major,minor and revision integers
func Version() (int, int, int)

// New is called to create a new empty client object
func New(clientId string, clean bool, userInfo uintptr) (*Client, error)

// Connect to a broker using host and port, setting the keepalive time in
// seconds and use 'true' for the async parameter to connect asyncronously
func (this *Client) Connect(host string, port int, keepalive int, async bool) error

//Connect to a broker using host and port, setting the keepalive time in
// seconds and use 'true' for the async parameter to connect asyncronously.
// Connects to a specific interface.
func (this *Client) ConnectBind(host, bindAddress string, port int, keepalive int, async bool) error

// Destroy is called when you have finished using a client
func (this *Client) Destroy() error

// Disconnect from a broker
func (this *Client) Disconnect() error

// Start the event loop thread, to be called before Connect
func (this *Client) LoopStart() error

// Stop the event loop thread, to be called after Disconnect has completed
func (this *Client) LoopStop(force bool) error

// Publish a message to the broker in a topic and return the id of the request
func (this *Client) Publish(topic string, data []byte, qos int, retain bool) (int, error)

// Subscribe to one set of topics and return the id of the request
func (this *Client) Subscribe(topics string, qos int) (int, error)

// Unsubscribe from one set of topics and return the id of the request
func (this *Client) Unsubscribe(topics string) (int, error)

```

In order to understand when requests to connect, disconnect, subscribe, unsubscribe and publish have completed, you need to set callback functions. In addition `SetMessageCallback` should be used to receive messages from the broker:

```go
func (this *Client) SetConnectCallback(cb DisconnectCallback) error
func (this *Client) SetDisconnectCallback(cb DisconnectCallback) error

func (this *Client) SetSubscribeCallback(cb SubscribeCallback) error
func (this *Client) SetUnsubscribeCallback(cb UnsubscribeCallback) error

func (this *Client) SetPublishCallback(cb PublishCallback) error
func (this *Client) SetMessageCallback(cb MessageCallback) error
func (this *Client) SetLogCallback(cb LogCallback) error
```

The signatures for these callbacks are as follows:

```go
type ConnectCallback func(userInfo uintptr,rc int)
type DisconnectCallback func(userInfo uintptr,rc int)

type SubscribeCallback func(userInfo uintptr,mid int,qos []int)
type UnsubscribeCallback func(userInfo uintptr,mid int)

type MessageCallback func(userInfo uintptr,message *Message)
type PublishCallback func(userInfo uintptr,mid int)
type LogCallback func(userInfo uintptr,level Level,str string)
```

A `Message` has the following methods in order to receive information:

```go
func (this *Message) Id() int
func (this *Message) Topic() string
func (this *Message) Data() []byte
func (this *Message) Len() uint
func (this *Message) Qos() int
func (this *Message) Retain() bool
```
