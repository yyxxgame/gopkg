//@File     config.go
//@Time     2023/04/04
//@Author   #Suyghur,

package ckafka

type ProducerConfig struct {
	Brokers   []string `json:""`
	Topic     string   `json:""`
	Username  string   `json:",optional"`
	Password  string   `json:",optional"`
	Partition int      `json:",default=1"`
}

type ConsumerConfig struct {
	Brokers []string
	Topics  []string
	GroupId string

	Consumers int    `json:",default=8"`
	Username  string `json:",optional"`
	Password  string `json:",optional"`
}
