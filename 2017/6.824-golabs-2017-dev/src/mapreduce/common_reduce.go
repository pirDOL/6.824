package mapreduce

import (
	"encoding/json"
	"io"
	"log"
	"os"
)

// doReduce manages one reduce task: it reads the intermediate
// key/value pairs (produced by the map phase) for this task, sorts the
// intermediate key/value pairs by key, calls the user-defined reduce function
// (reduceF) for each key, and writes the output to disk.
func doReduce(
	jobName string, // the name of the whole MapReduce job
	reduceTaskNumber int, // which reduce task this is
	outFile string, // write the output here
	nMap int, // the number of map tasks that were run ("M" in the paper)
	reduceF func(key string, values []string) string,
) {
	switch solutionVersion {
	case "v1":
		doReduceV1(jobName, reduceTaskNumber, outFile, nMap, reduceF)
	case "v2":
		doReduceV2(jobName, reduceTaskNumber, outFile, nMap, reduceF)
	}
	//
	// You will need to write this function.
	//
	// You'll need to read one intermediate file from each map task;
	// reduceName(jobName, m, reduceTaskNumber) yields the file
	// name from map task m.
	//
	// Your doMap() encoded the key/value pairs in the intermediate
	// files, so you will need to decode them. If you used JSON, you can
	// read and decode by creating a decoder and repeatedly calling
	// .Decode(&kv) on it until it returns an error.
	//
	// You may find the first example in the golang sort package
	// documentation useful.
	//
	// reduceF() is the application's reduce function. You should
	// call it once per distinct key, with a slice of all the values
	// for that key. reduceF() returns the reduced value for that key.
	//
	// You should write the reduce output as JSON encoded KeyValue
	// objects to the file named outFile. We require you to use JSON
	// because that is what the merger than combines the output
	// from all the reduce tasks expects. There is nothing special about
	// JSON -- it is just the marshalling format we chose to use. Your
	// output code will look something like this:
	//
	// enc := json.NewEncoder(file)
	// for key := ... {
	// 	enc.Encode(KeyValue{key, reduceF(...)})
	// }
	// file.Close()
	//
}

func doReduceV1(
	jobName string, // the name of the whole MapReduce job
	reduceTaskNumber int, // which reduce task this is
	outFile string, // write the output here
	nMap int, // the number of map tasks that were run ("M" in the paper)
	reduceF func(key string, values []string) string,
) {
	// 1 read mrtmp.xxx-{0..nMap}-iReduce
	var mrtmpkvs []KeyValue
	key2values := make(map[string][]string)
	for i := 0; i < nMap; i++ {
		filename := reduceName(jobName, i, reduceTaskNumber)
		file, err := os.Open(filename)
		if err != nil {
			log.Print("open mrtmp err", err, filename)
			continue
		}
		dec := json.NewDecoder(file)
		for {
			if err := dec.Decode(&mrtmpkvs); err == io.EOF {
				break
			} else if err != nil {
				log.Fatal("unmarshal mrtmp err", err, filename)
			}
		}
		file.Close()

		for _, kv := range mrtmpkvs {
			key2values[kv.Key] = append(key2values[kv.Key], kv.Value)
		}
	}

	// 2 write mrtmp.xxx-res-iReduce
	file, err := os.OpenFile(outFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		log.Fatal("open reduce output err", err, outFile)
		return
	}
	defer file.Close()

	enc := json.NewEncoder(file)
	for key, values := range key2values {
		res := reduceF(key, values)
		enc.Encode(KeyValue{key, res})
		if err != nil {
			log.Fatal("write reduce output err", err, outFile)
			return
		}
	}
}

func doReduceV2(
	jobName string, // the name of the whole MapReduce job
	reduceTaskNumber int, // which reduce task this is
	outFile string, // write the output here
	nMap int, // the number of map tasks that were run ("M" in the paper)
	reduceF func(key string, values []string) string,
) {
	// 1 read mrtmp.xxx-{0..nMap}-iReduce
	var mrtmpkv KeyValue
	key2values := make(map[string][]string)
	for i := 0; i < nMap; i++ {
		filename := reduceName(jobName, i, reduceTaskNumber)
		file, err := os.Open(filename)
		if err != nil {
			log.Print("open mrtmp err", err, filename)
			continue
		}
		dec := json.NewDecoder(file)
		for {
			err := dec.Decode(&mrtmpkv)
			if err == io.EOF {
				break
			} else if err != nil {
				log.Fatal("unmarshal mrtmp err", err, filename)
			} else {
				key2values[mrtmpkv.Key] = append(key2values[mrtmpkv.Key], mrtmpkv.Value)
			}
		}
		file.Close()
	}

	// 2 write mrtmp.xxx-res-iReduce
	file, err := os.OpenFile(outFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		log.Fatal("open reduce output err", err, outFile)
		return
	}
	defer file.Close()

	enc := json.NewEncoder(file)
	for key, values := range key2values {
		res := reduceF(key, values)
		enc.Encode(KeyValue{key, res})
		if err != nil {
			log.Fatal("write reduce output err", err, outFile)
			return
		}
	}
}
