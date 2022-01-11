package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type FullMsg struct {
	Ver    string     `json:"jsonrpc"`
	Method string     `json:"method"`
	Params JSONParams `json:"params"`
	Id     int        `json:"id"`
}

type JSONParams struct {
	Key string `json:"apiKey"`
	Num int    `json:"n"`
	Min int    `json:"min"`
	Max int    `json:"max"`
}
type ArrMean struct {
	Stddev float64
	Data   []int
}
type MeanOfMeans struct {
	Slices []ArrMean
	Stddev float64
	Data   []int
}
type ReceivedJson struct {
	Jsonrpc string `json:"jsonrpc"`
	Result  struct {
		Random struct {
			Data           []int  `json:"data"`
			CompletionTime string `json:"completionTime"`
		} `json:"random"`
		BitsUsed      int `json:"bitsUsed"`
		BitsLeft      int `json:"bitsLeft"`
		RequestsLeft  int `json:"requestsLeft"`
		AdvisoryDelay int `json:"advisoryDelay"`
	} `json:"result"`
	ID int `json:"id"`
}

func Get_StdDev(vals []int) float64 {
	//mean
	var number int
	for _, i := range vals {
		number += i
	}
	number = number / (len(vals))
	//variance
	var variance float64
	for _, i := range vals {
		variance += math.Pow(float64(i-number), 2)

	}
	//standard deviation
	return math.Sqrt(variance)
}
func getMeans(c *gin.Context) {
	//mean?requests={r}&length{l}
	rq := c.Query("requests")
	length := c.Query("length")
	j, err := strconv.Atoi(length)

	if err != nil {
		// handle error
		fmt.Println(err)
		c.String(http.StatusBadRequest, string("bad input parameters"))

	}
	req, err := strconv.Atoi(rq)
	if err != nil {
		fmt.Println(err)
		c.String(http.StatusBadRequest, string("bad input parameters"))
	}

	if j < 1 || req < 1 || req > 500 {
		c.String(http.StatusBadRequest, string("bad input parameters"))
	}

	var MeMeans = []ArrMean{}
	SubArr := []int{}
	for i := 0; i < req; i++ {
		MeMeans = append(MeMeans, ArrMean{
			Get_StdDev(_request(j)),
			_request(j),
		})
		SubArr = append(SubArr, MeMeans[i].Data...)
		time.Sleep(100 * time.Millisecond)
	}
	//c.String(http.StatusOK, string("ugioguga"))

	result := &MeanOfMeans{
		Stddev: Get_StdDev(SubArr),
		Slices: MeMeans,
		Data:   SubArr,
	}
	resultBody, err := json.MarshalIndent(result, "", "\t")
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
	}

	c.String(http.StatusOK, string(resultBody))
}

//Send {
func _request(length int) []int {
	url := "https://api.random.org/json-rpc/2/invoke"

	param := &JSONParams{
		Key: "f658e5dc-7ad4-4815-890a-79442cf4a4ae",
		Num: length,
		Min: 1,
		Max: 100,
	}

	user := &FullMsg{
		Ver:    "2.0",
		Method: "generateIntegers",
		Params: *param,
		Id:     123,
	}

	postBody, _ := json.Marshal(user)

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(postBody))

	if err != nil {
		log.Fatalf("An Error Occured %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		log.Fatalln(err)
	}
	recJson := ReceivedJson{}
	sb := string(body)

	json.Unmarshal([]byte(sb), &recJson)
	log.Printf(string(postBody))
	log.Printf(sb)
	return recJson.Result.Random.Data
}

//Main stuff does main stuff
func main() {
	router := gin.Default()
	router.GET("/random/mean", getMeans)
	router.Run(":8080")
}
