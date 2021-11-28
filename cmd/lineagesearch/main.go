package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	gUuid "github.com/google/uuid"
	"github.com/kmgreen2/agglo/pkg/client"
	"google.golang.org/grpc"
	"os"
	"github.com/kmgreen2/agglo/pkg/search"
	"sort"
)

type CommandArgs struct {
	searchConnectionString string
	searchText string
	tickerClient client.TickerClient
}

func usage(msg string, exitCode int) {
	fmt.Println(msg)
	flag.PrintDefaults()
	os.Exit(exitCode)
}

func parseArgs() *CommandArgs {
	var err error
	args := &CommandArgs{}

	searchConnectionStringPtr := flag.String("searchConnectionString", "", "search connection string")
	searchTextPtr := flag.String("searchText", "", "search text")
	tickerEndpointPtr := flag.String("tickerEndpoint", "localhost:8001", "ticker grpc endpoint (host:port)")

	flag.Parse()

	if len(*searchConnectionStringPtr) == 0 {
		panic("Must specify a search string")
	}

	if len(*searchTextPtr) == 0 {
		panic("Must specify a search string")
	}

	if len(*tickerEndpointPtr) == 0 {
		panic("Must specify a ticker endpoint")
	}

	args.searchConnectionString = *searchConnectionStringPtr
	args.searchText = *searchTextPtr
	args.searchText = *searchTextPtr

	args.tickerClient, err = client.NewTickerClient(*tickerEndpointPtr, grpc.WithInsecure())

	if err != nil {
		panic(err)
	}

	return args
}

type SearchResult struct {
	tickerUuid gUuid.UUID
	entwineUuid gUuid.UUID
	subStreamID string
	subStreamIndex int64
	text string
	tickerClient client.TickerClient
}

type SearchResults []SearchResult

func (r SearchResults) Len() int {
	return len(r)
}
func (r SearchResults) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}
func (r SearchResults) Less(i, j int) bool {
	tickerClient := r[i].tickerClient
	result, err := tickerClient.HappenedBefore(context.Background(), r[i].subStreamID, r[i].entwineUuid, r[i].subStreamIndex,
			r[j].subStreamID, r[j].entwineUuid, r[j].subStreamIndex)
	if err != nil {
		panic(err)
	}
	return result
}

func main() {
	var searchResults SearchResults
	args := parseArgs()

	searchIndex, err := search.NewIndexFromConnectionString(args.searchConnectionString)
	if err != nil {
		panic(err)
	}

	queryBuilder := search.NewSimpleElasticQueryBuilder()

	queryBuilder.AddText(args.searchText)

	queryResult, err := searchIndex.Query(context.Background(), queryBuilder.Build())

	if outerHits, outerHitsOk := queryResult["hits"]; outerHitsOk {
		if hits, hitsOk := outerHits.(map[string]interface{})["hits"].([]interface{}); hitsOk {
			for _, hit := range hits {
				if source, sourceOk := hit.(map[string]interface{})["_source"]; sourceOk {
					if payload, payloadOk := source.(map[string]interface{}); payloadOk {
						var payloadMap map[string]interface{}
						blob := payload["blob"].([]interface{})[0]
						blobBytes, err := base64.StdEncoding.DecodeString(blob.(string))
						if err != nil {
							panic(err)
						}
						jsonPayload := bytes.NewBuffer(blobBytes)
						decoder := json.NewDecoder(jsonPayload)
						err = decoder.Decode(&payloadMap)
						if err != nil {
							panic(err)
						}

						searchResults = append(searchResults, SearchResult{
							tickerUuid: gUuid.MustParse(payloadMap["entwinedTickerUuid"].(string)),
							entwineUuid: gUuid.MustParse(payloadMap["entwineUuid"].(string)),
							subStreamID: payloadMap["entwinedSubStreamID"].(string),
							subStreamIndex: int64(payloadMap["entwinedSubstreamIndex"].(float64)),
							text: payloadMap["text"].(string),
							tickerClient: args.tickerClient,
						})
					}
				}
			}
		}
	}

	sort.Sort(searchResults)

	for _, v := range searchResults {
		fmt.Printf("%v: %s\n", v.entwineUuid, v.text)
	}
}


