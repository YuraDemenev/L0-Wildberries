package mynats

import (
	"L0/pkg/database"
	"L0/pkg/models"
	"encoding/json"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"
)

var Js *nats.JetStreamContext
var Cache map[int]models.Order

func ConnectToNATS() (js nats.JetStreamContext, nc *nats.Conn, err error) {
	nc, err = nats.Connect(nats.DefaultURL)
	if err != nil {
		return nil, nil, err
	}

	js, err = nc.JetStream()
	if err != nil {
		return nil, nil, err
	}
	Js = &js

	return js, nc, nil
}

func CreateStream(js nats.JetStreamContext, name string) error {
	fmt.Printf("Creating stream: %q\n", name)
	_, err := js.AddStream(&nats.StreamConfig{
		Name:     name,
		Subjects: []string{"test"},
	})

	return err
}

func Publish(js nats.JetStreamContext, subj string, f func() []byte, //f is a function that returns message to be published
) error {
	ack, err := js.Publish(subj, f())
	if err != nil {
		return err
	}
	fmt.Printf("%#v\n", ack)
	return nil
}

func SubscribeToJetStream(js nats.JetStreamContext, chanelName string, db *sqlx.DB) (*nats.Subscription, error) {
	sub, err := js.Subscribe(chanelName, func(msg *nats.Msg) {
		var order models.Order
		fmt.Printf("message : %s\n", string(msg.Data))

		//Open json
		err := json.Unmarshal(msg.Data, &order)
		if err != nil {
			logrus.Infof("failed unmarshal json: %s", err.Error())
			return
		}

		//Check Data
		err = order.CheckData()
		if err != nil {
			logrus.Infof("not correct data in: %s", err.Error())
			return
		}

		//Add to DB
		orderId, err := database.SaveOrder(db, order)
		if err != nil {
			logrus.Infof("not correct data in: %s", err.Error())
			return
		}

		order.Id = orderId
		Cache[orderId] = order

	}, nats.Durable("my-dirable"))

	return sub, err
}

//publish(js,"test",testMSG)
// strm.addConsumer(js,strName,"test","test")
