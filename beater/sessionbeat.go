package beater

import (
	"fmt"
	"time"
        "encoding/json"
        "os/exec"
        "os"
        "io/ioutil"
        "log"
        "strings"
	"github.com/elastic/beats/libbeat/beat"
	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/logp"
	"github.com/hyuntori/sessionbeat/config"
)

//beat Struct
type Sessionbeat struct {
	done   chan struct{}
	config config.Config
	client beat.Client
}

//jsonFile
type SessionDoc struct {
    Name    string `json:"name"`
    Command string `json:"command"`
}

//toString
func (s SessionDoc) toString() string {
    return toJson(s)
}

//toJson
func toJson(s interface{}) string {
    bytes, err := json.Marshal(s)
    if err != nil {
        fmt.Println(err.Error())
        os.Exit(1)
    }

    return string(bytes)
}

// Creates beater
func New(b *beat.Beat, cfg *common.Config) (beat.Beater, error) {
	config := config.DefaultConfig
	if err := cfg.Unpack(&config); err != nil {
		return nil, fmt.Errorf("Error reading config file: %v", err)
	}

	bt := &Sessionbeat{
		done:   make(chan struct{}),
		config: config,
	}
	return bt, nil
}

func (bt *Sessionbeat) Run(b *beat.Beat) error {
	logp.Info("sessionbeat is running! Hit CTRL-C to stop it.")

	var err error
	bt.client, err = b.Publisher.Connect()
	if err != nil {
		return err
	}
        
         var commandMap map[string]string
         commandMap = make(map[string]string)
         docs := getDocs()
         for _, s := range docs {
           cmd := s.Command
           out, err := exec.Command("bash","-c",cmd).Output()
           if err != nil {
             log.Fatal(err)
         }
           commandMap[s.Name] = strings.TrimSuffix(string(out),"\n")
         }

        fmt.Println(commandMap)


	ticker := time.NewTicker(bt.config.Period)
	for {
		select {
		case <-bt.done:
			return nil
		case <-ticker.C:
		}

		event := beat.Event{
			Timestamp: time.Now(),
			Fields: common.MapStr{
				"type":"session",
				"sum":common.MapStr{
 					"total":commandMap["sum_total"],
					"tcp_total":commandMap["sum_tcp_total"],
					"tcp_extab":commandMap["sum_tcp_extab"],
					"tcp_closed":commandMap["sum_tcp_closed"],
					"tcp_orphaned":commandMap["sum_tcp_orphaned"],
					"tcp_synrecv":commandMap["sum_tcp_synrecv"],
					"tcp_timewait":commandMap["sum_tcp_timewait"],
				},
				"detail":common.MapStr{
					"raw":common.MapStr{
						"total":commandMap["detail_raw_total"],
 						"ip":commandMap["detail_raw_ip"],
						"ipv6":commandMap["detail_raw_ipv6"],
					},
					"udp":common.MapStr{
                                                "total":commandMap["detail_udp_total"],
                                                "ip":commandMap["detail_udp_ip"],
                                                "ipv6":commandMap["detail_udp_ipv6"],

                                        },
					"tcp":common.MapStr{
                                                 "total":commandMap["detail_tcp_total"],
                                                 "ip":commandMap["detail_tcp_ip"],
                                                 "ipv6":commandMap["detail_tcp_ipv6"],
                                         },
					"inet":common.MapStr{
   						"total":commandMap["detail_inet_total"],
         					"ip":commandMap["detail_inet_ip"],
         					"ipv6":commandMap["detail_inet_ipv6"],
 					},
				},
			},
		}
		bt.client.Publish(event)
		logp.Info("Event sent")
	}
}

//getDocs
func getDocs() []SessionDoc {
    raw, err := ioutil.ReadFile("/home/hduser/projects/src/src/github.com/hyuntori/sessionbeat/beater/commandDoc.json")
    if err != nil {
        fmt.Println(err.Error())
        os.Exit(1)
    }

    var c []SessionDoc
    json.Unmarshal(raw, &c)
    return c
}


func (bt *Sessionbeat) Stop() {
	bt.client.Close()
	close(bt.done)
}
