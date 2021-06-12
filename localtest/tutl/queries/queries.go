package queries

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/olivere/elastic/v7"
	"reflect"
	"sort"
	"sudachen.xyz/pkg/errstr"
	"sudachen.xyz/pkg/localnet/fu"
)

const defaulResultSize = 1000

var ElasticURL = "http://localhost:9200"

type Row struct {
	Text string
	Name string
	M struct {
		LayerID int64 `json:"layer_id"`
		EpochID int64 `json:"epoch_id"`
		NodeID string `json:"node_id"`
		IsRsvpReady bool `json:"is_svp_ready"`
		ProposedSet string `json:"proposed_set"`
		EligibilityCount int64 `json:"eligibility_count"`
		CurrentSet string `json:"current_set"`
		LayerHash string `json:"layer_hash"`
		L string
		N string
	}
}

func GetAppStartedMsgs(genesis string) []Row{
	msg := "App started"
	return GetAllMsgs(genesis, msg)
}

func GetTickMsgs(genesis string, layer int64) []Row {
	msg := "release tick"
	r := GetAllMsgs(genesis, msg,
		map[string]string{
			"m.layer_id": fmt.Sprint(layer),
		})
	return r
}

func GetAllMsgs(genesis, msg string, opt ...map[string]string) []Row{
	q := fmt.Sprintf(`text:"%v"`, msg)
	n := 1
	for _, m := range opt {
		for k,v := range m {
			q = q + fmt.Sprintf(`%v:"%v"`,k,v)
			n++
		}
	}
	text, err := json.Marshal(map[string]interface{}{
		"query_string":map[string]interface{}{
			"query": q,
			"minimum_should_match" : n,
		},
	})
	if err != nil { panic(errstr.Wrapf(0, err, "failed to marshal query: %v", err.Error())) }
	query := elastic.NewRawStringQuery(string(text))
	return Query(genesis,query)
}

func FindErrors(genesis string) []Row {
	text, err := json.Marshal(map[string]interface{}{
		"query_string":map[string]interface{}{
			"query": `m.L:"ERROR"`,
			"minimum_should_match" : 1,
		},
	})
	if err != nil { panic(errstr.Wrapf(0, err, "failed to marshal query: %v", err.Error())) }
	query := elastic.NewRawStringQuery(string(text))
	return Query(genesis,query)
}

func GetLastFinishedLayer(genesis string, count ...int) int64 {
	msg := "new layer hash"
	rs := GetAllMsgs(genesis, msg)
	ls := map[int64]int{}
	for _, r := range rs {
		ls[r.M.LayerID] = ls[r.M.LayerID] + 1
	}
	ks := make([]int64,0,len(ls))
	for i := range ls {
		ks = append(ks,i)
	}
	sort.Slice(ks, func(i,j int)bool{ return ks[i] > ks[j]} )
	c := fu.Fnzi(count...)
	if c == 0 {
		return ks[0]
	}
	for _, x := range ks {
		if ls[x] == c {
			return x
		}
	}
	panic(errstr.Errorf("couldn't find layer for %v nodes: %v",c,ls))
}

func Query(genesis string, query elastic.Query, size ...int) []Row{
	client, err := elastic.NewClient(elastic.SetURL(ElasticURL))
	if err != nil { panic(errstr.Wrapf(0, err, "failed to connect elasticsearch cluster: %v", err.Error())) }
	result, err := client.Search().
		Index("x-spacemesh-"+genesis).
		Query(query).
		Size(fu.Fnzi(fu.Fnzi(size...),defaulResultSize)).
		Do(context.Background())
	if err != nil  { panic(errstr.Wrapf(0, err, "failed to search cluster: %v", err.Error())) }
	hits := make([]Row,0,len(result.Hits.Hits))
	for _, tx := range result.Each(reflect.TypeOf(Row{})) {
		if r, ok := tx.(Row); ok {
			hits = append(hits,r)
		}
	}
	return hits
}
