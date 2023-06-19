package main

import (
	"bufio"
	"bytes"
	"container/heap"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"
)

const batchSize = 288

var (
	filename string
	begin    string
	end      string
	datatype string
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	flag.StringVar(&filename, "f", "", "filename in json format")
	flag.StringVar(&begin, "begin", "2021-06-01", "begin date")
	flag.StringVar(&end, "end", "2021-06-02", "end date")
	flag.StringVar(&datatype, "type", "bandwidth", "specify data type. one of (`bandwidth`, `flux`, `dynbandwidth`, `dynflux`)")

	http.DefaultTransport.(*http.Transport).MaxIdleConnsPerHost = runtime.NumCPU()
}

func main() {
	flag.Parse()

	start, err := time.Parse("2006-01-02", begin)
	if err != nil {
		log.Fatal("time.Parse: ", err)
	}
	end, err := time.Parse("2006-01-02", end)
	if err != nil {
		log.Fatal("time.Parse: ", err)
	}

	// for day := start; day.Before(end); day = day.Add(24 * time.Hour) {
	// 	fetchUserTraffic(uid, day, day.Add(24*time.Hour))
	// }

	// 	var traffics []*UserTraffic

	// 	for _, uid := range getUidsFromFile(filename) {
	// 		traffic := fetchUserTraffic(uid, start, end, datatype)

	// 		traffics = append(traffics, &UserTraffic{uid: uid, traffic: sum(traffic)})
	// 	}
}

func getUidsFromFile(filename string) []uint {
	fp, err := os.Open(filename)
	if err != nil {
		log.Fatal("os.Open: ", err)
	}
	defer fp.Close()

	var uids []uint

	reader := bufio.NewReader(fp)
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if !errors.Is(err, io.EOF) {
				log.Fatal("bufio.ReadBytes: ", err)
			}
			break
		}

		uidstr := strings.TrimRight(string(line), "\n")

		uid, err := strconv.ParseUint(uidstr, 10, 64)
		if err != nil {
			log.Fatal("strconv.ParseUint: ", err)
		}

		uids = append(uids, uint(uid))
	}

	return uids
}

type UserTraffic struct {
	uid     uint
	traffic int64
}

// sort.Interface
// Push(x interface{}) // add x as element Len()
// Pop() interface{}   // remove and return element Len() - 1.

type UserTraffics []*UserTraffic

func (traffics *UserTraffics) Swap(i, j int) {
	(*traffics)[i], (*traffics)[j] = (*traffics)[j], (*traffics)[i]
}

func (traffics *UserTraffics) Len() int {
	return len(*traffics)
}

func (traffics *UserTraffics) Less(i, j int) bool {
	return (*traffics)[i].traffic > (*traffics)[j].traffic
}

func (traffics *UserTraffics) Push(x interface{}) {
	slots := *traffics

	slots = append(slots, x.(*UserTraffic))

	(*traffics) = slots
}

func (traffics *UserTraffics) Pop() interface{} {
	if traffics.Len() > 0 {
		x := (*traffics)[len(*traffics)-1]
		*traffics = (*traffics)[:len(*traffics)-1]
		return x
	}
	return nil
}

func statTopUserTraffics(traffics []*UserTraffic) {
	t := &UserTraffics{}
	*t = UserTraffics(traffics)

	heap.Init(t)

	for t.Len() > 0 {
		top := heap.Pop(t)
		traffic := top.(*UserTraffic)
		fmt.Printf("%d,%d\n", traffic.uid, traffic.traffic)
	}
}

func fetchUserTraffic(uid uint, start, end time.Time, dataType string) []int64 {

	params := map[string]interface{}{
		"uid":         uid,
		"startDate":   start.Format("2006-01-02"),
		"endDate":     end.Format("2006-01-02"),
		"granularity": "5min",
		"type":        dataType,
	}

	var buf bytes.Buffer

	err := json.NewEncoder(&buf).Encode(params)
	if err != nil {
		log.Fatal("json.NewEncoder: ", err)
	}

	endpoint := "/v2/domain/"
	switch datatype {
	case "bandwidth", "dynbandwidth", "":
		endpoint += "bandwidth"
	case "dynflux", "flux":
		endpoint += "flux"
	}

	req, err := http.NewRequest(http.MethodPost, "http://deftonestraffic.fusion.internal.qiniu.io"+endpoint, &buf)
	if err != nil {
		log.Fatal("http.NewRequest: ", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal("http.Client.Do: ", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("io.ReadAll: ", err)
	}

	var res struct {
		Code  int
		Error string
		Data  map[string]struct {
			Oversea []int64
			China   []int64
		}
	}
	err = json.Unmarshal(data, &res)
	if err != nil {
		log.Fatal("json.Unmarshal: ", err)
	}

	days := int(end.Sub(start).Hours()/24 + 1)

	var trafficsum = make([]int64, days)

	for _, regionTraffic := range res.Data {
		if len(regionTraffic.China) > 0 {
			trafficsum = AddInt64Array(regionTraffic.China, trafficsum)
		}
		if len(regionTraffic.Oversea) > 0 {
			trafficsum = AddInt64Array(trafficsum, regionTraffic.Oversea)
		}
	}
	return trafficsum
}

func sum(a []int64) int64 {
	sum := int64(0)
	for _, v := range a {
		sum += v
	}
	return sum
}

func bandwidthStat() {
	var res map[string]interface{}

	fp, err := os.Open(filename)
	if err != nil {
		log.Fatal("os.Open:", err)
	}

	if err = json.NewDecoder(fp).Decode(&res); err != nil {
		log.Fatal("json.Decode:", err)
	}

	var flatBandwidths []int64

	domainBws := res["data"].(map[string]interface{})
	for _, bws := range domainBws {
		var bandwidths []int64
		points := bws.(map[string]interface{})["china"].([]interface{})
		for _, point := range points {
			bandwidths = append(bandwidths, int64(point.(float64)))
		}

		if len(flatBandwidths) == 0 {
			flatBandwidths = make([]int64, len(bandwidths))
		}

		flatBandwidths = AddInt64Array(flatBandwidths, bandwidths)
	}

	batchedBadnwidths := make([][]int64, 30)

	groups := len(flatBandwidths) / batchSize

	log.Println("groups:", groups)

	for i := 0; i < groups; i++ {
		begin := i * batchSize
		end := begin + batchSize
		if end > len(flatBandwidths) {
			end = len(flatBandwidths)
		}
		batchedBadnwidths[i] = flatBandwidths[begin:end]
	}

	var peak95sum int64

	day := 20210401

	for _, bws := range batchedBadnwidths {
		v := peak95(bws)
		fmt.Printf("%d,%v\n", day, v)
		day++
		peak95sum += v

	}

	peak95Avr := float64(peak95sum) / float64(groups)
	fmt.Println("peak95Avr: ", peak95Avr)
}

type sortableInt64s []int64

func peak95(bws []int64) int64 {
	sort.Slice(sortableInt64s(bws), func(i, j int) bool {
		return bws[i] > bws[j]
	})

	peak95Index := len(bws) / 20
	if peak95Index > 0 {
		peak95Index--
	}
	return bws[peak95Index]
}

func AddInt64Array(a, b []int64) []int64 {
	if len(a) != len(b) {
		panic("lengtht mismatch")
	}

	var c = make([]int64, len(a))
	for i, v := range a {
		c[i] = v + b[i]
	}
	return c
}

func fetchUserBandwidth() {
	var buf bytes.Buffer

	param := map[string]interface{}{
		"startDate":   "2021-04-01",
		"endDate":     "2021-04-30",
		"uid":         1381979318,
		"granularity": "5min",
	}
	if err := json.NewEncoder(&buf).Encode(param); err != nil {
		log.Fatal("json.Encode:", err)
	}

	req, err := http.NewRequest("POST", "http://fusion.qiniuapi.com/v2/tune/flux", &buf)
	if err != nil {
		log.Fatal("http.NewRequest:", err)
	}
	req.Header.Set("Authorization", "QBox 557TpseUM8ovpfUhaw8gfa2DQ0104ZScM-BTIcBx:4N8vnXaKH4MYfPX3t1Js3SbT6aA=")
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()

	var res map[string]interface{}

	if err = json.NewDecoder(resp.Body).Decode(&res); err != nil {
		log.Fatal("json.Decode:", err)
	}
}
