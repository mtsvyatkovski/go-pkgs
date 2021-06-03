// Copyright 2019 SumUp Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package rabbitmq

import (
	"context"

	"github.com/streadway/amqp"
)

type Handler interface {
	GetQueueName() string
	GetConsumerTag() string
	QueueAutoAck() bool
	ExclusiveConsumer() bool
	MustStopOnAckError() bool
	MustStopOnNAckError() bool
	MustStopOnRejectError() bool
	WaitToConsumeInflight() bool
	ReceiveMessage(ctx context.Context, payload []byte) (acknowledgement HandlerAcknowledgement, err error)
	GetConsumeContext(ctx context.Context, d *amqp.Delivery) context.Context
}

type AcknowledgementType int

const (
	Ack AcknowledgementType = iota
	Nack
	Reject
)

type HandlerAcknowledgement struct {
	Acknowledgement AcknowledgementType
	Requeue         bool
}
