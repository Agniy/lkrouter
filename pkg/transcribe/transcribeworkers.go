package transcribe

import (
	"fmt"
	"log"
	"os"
	"sync"

	"lkrouter/pkg/mongodb/mrequests"
)

var transcribeWorkersOnce sync.Once
var workChan chan map[string]interface{}
var logger = log.New(os.Stdout, "Beanstalk: ", 0)

const numOfMessageWorkers = 300

func InitFileTranscribeWorkers() chan map[string]interface{} {
	//Perform connection creation operation only once.
	transcribeWorkersOnce.Do(func() {
		workChan = make(chan map[string]interface{}, numOfMessageWorkers)
		for w := 1; w <= numOfMessageWorkers; w++ {
			go transcriberPutWorker(w, workChan)
		}
	})

	return workChan
}

func transcriberPutWorker(id int, jobs <-chan map[string]interface{}) {
	for jMap := range jobs {
		logger.Println("transcriberPutWorker message: ", id, " - started  job with data:", jMap)

		room, rfound := jMap["room"]
		lang, lFound := jMap["lang"]
		if !lFound {
			lang = ""
		}

		transcriberType, tfound := jMap["type"]
		prompt, pFound := jMap["prompt"]
		if !rfound || !tfound {
			logger.Println("transcriberPutWorker message: ", id, " - job with data:", jMap, " - missing room or lang")
			continue
		}

		roomUrl := room.(string)
		roomLang := lang.(string)
		roomtranscriberType := transcriberType.(string)

		promptStr := ""
		if pFound {
			promptStr = prompt.(string)
		}

		//get audio file url
		call, err := mrequests.GetCallByRoom(roomUrl)
		if err != nil {
			fmt.Printf("transcriberPutWorker message: Error when try to get room %v : %v in \n", roomUrl, err)
			continue
		}

		if call["audioUrl"] == nil {
			fmt.Printf("transcriberPutWorker message: Error when try to get audioUrl from call: %v in \n", call)
			continue
		}

		audioUrl := call["audioUrl"].(string)
		if roomtranscriberType == "google" {
			transcriber := NewGoogleTranscriber(roomUrl, audioUrl, roomLang)
			transcriber.GoogleFileTranscribe()
		} else {
			transcriber := NewWhisperTranscriber(roomUrl, audioUrl, roomLang, promptStr)
			transcriber.WhisperFileTranscribe()
		}
		logger.Println("transcriberPutWorker message: ", id, "finished job", jMap)
	}
}

func SendWorkTask(transcribeData map[string]interface{}) int {
	go func() {
		workChan <- transcribeData
	}()
	return 0
}
