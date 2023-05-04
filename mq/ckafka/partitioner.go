//@File     partitioner.go
//@Time     2023/04/23
//@Author   #Suyghur,

package ckafka

import (
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"hash"
	"hash/fnv"
	"math/rand"
	"time"
)

type (

	// IPartitioner is anything that, given a Kafka message and a number of partitions indexed [0...numPartitions-1],
	// decides to which partition to send the message. RandomPartitioner, RoundRobinPartitioner and HashPartitioner are provided
	// as simple default implementations.
	IPartitioner interface {
		// Partition takes a message and partition count and chooses a partition
		Partition(message *kafka.Message, numPartitions int32) (int32, error)
	}
	randomPartitioner struct {
		generator *rand.Rand
	}

	roundRobinPartitioner struct {
		partition int32
	}

	hashPartitioner struct {
		random       IPartitioner
		hasher       hash.Hash32
		referenceAbs bool
	}
)

// NewRandomPartitioner returns a Partitioner which chooses a random partition each time.
func NewRandomPartitioner() IPartitioner {
	p := new(randomPartitioner)
	p.generator = rand.New(rand.NewSource(time.Now().UTC().UnixNano()))
	return p
}

func (p *randomPartitioner) Partition(message *kafka.Message, numPartitions int32) (int32, error) {
	return int32(p.generator.Intn(int(numPartitions))), nil
}

// NewRoundRobinPartitioner returns a Partitioner which walks through the available partitions one at a time.
func NewRoundRobinPartitioner() IPartitioner {
	return &roundRobinPartitioner{}
}

func (p *roundRobinPartitioner) Partition(message *kafka.Message, numPartitions int32) (int32, error) {
	if p.partition >= numPartitions {
		p.partition = 0
	}
	ret := p.partition
	p.partition++
	return ret, nil
}

// NewHashPartitioner returns a Partitioner which behaves as follows. If the message's key is nil then a
// random partition is chosen. Otherwise the FNV-1a hash of the encoded bytes of the message key is used,
// modulus the number of partitions. This ensures that messages with the same key always end up on the
// same partition.
func NewHashPartitioner() IPartitioner {
	p := new(hashPartitioner)
	p.random = NewRandomPartitioner()
	p.hasher = fnv.New32a()
	p.referenceAbs = false
	return p
}

func (p *hashPartitioner) Partition(message *kafka.Message, numPartitions int32) (int32, error) {
	if len(message.Key) == 0 {
		return p.random.Partition(message, numPartitions)
	}
	p.hasher.Reset()
	_, err := p.hasher.Write(message.Key)
	if err != nil {
		return -1, err
	}
	var partition int32
	// Turns out we were doing our absolute value in a subtly different way from the upstream
	// implementation, but now we need to maintain backwards compat for people who started using
	// the old version; if referenceAbs is set we are compatible with the reference java client
	// but not past Sarama versions
	if p.referenceAbs {
		partition = (int32(p.hasher.Sum32()) & 0x7fffffff) % numPartitions
	} else {
		partition = int32(p.hasher.Sum32()) % numPartitions
		if partition < 0 {
			partition = -partition
		}
	}
	return partition, nil
}
