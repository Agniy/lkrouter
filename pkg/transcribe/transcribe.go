package transcribe

import (
	speech "cloud.google.com/go/speech/apiv1p1beta1"
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"google.golang.org/api/option"
	speechpb "google.golang.org/genproto/googleapis/cloud/speech/v1p1beta1"
	"lkrouter/config"
	"lkrouter/pkg/mongodb/mrequests"
	"net/url"
)

type Transcribe struct {
	logger       *logrus.Logger
	speechClient *speech.Client
	lang         string
	room         string
	audioUrl     string
}

func NewGoogleTranscriber(room string, audioUrl string, lang string) *Transcribe {
	ctx := context.Background()
	logger := logrus.New()
	cfg := config.GetConfig()

	// Creates a client.
	client, err := speech.NewClient(ctx, option.WithCredentialsFile(cfg.App.GoogleAppCredPath))
	if err != nil {
		logger.Fatalf("Failed to create client: %v", err)
	}
	return &Transcribe{
		room:         room,
		audioUrl:     audioUrl,
		logger:       logger,
		speechClient: client,
		lang:         lang,
	}
}
func (t *Transcribe) GoogleFileTranscribe() {

	mrequests.UpdateTranscribeTextStatus(t.room, "progress")

	ctx := context.Background()
	// [START speech_transcribe_sync]
	// Sample rate: the sample rate in Hertz of the audio data sent
	// to the API. Valid values are: 8000-48000. 16000 is optimal.
	// 16000 Hz is currently the only valid value for sampleRateHz.
	var sampleRate int32 = 8000

	// The language of the supplied audio
	languageCode := t.lang

	// Encoding of audio data sent. This sample uses linear16
	// (raw 16-bit samples)	.
	encoding := speechpb.RecognitionConfig_MP3

	// split url string into parts
	u, err := url.Parse(t.audioUrl)
	if err != nil {
		t.logger.Errorf("Error parsing url: %v", err)
		return
	}
	// get file name from url
	gsPath := "gs://" + u.Path[1:]

	// Detects speech in the audio file.
	req := &speechpb.LongRunningRecognizeRequest{
		Config: &speechpb.RecognitionConfig{
			Encoding:              encoding,
			SampleRateHertz:       sampleRate,
			LanguageCode:          languageCode,
			EnableWordTimeOffsets: true,
		},
		Audio: &speechpb.RecognitionAudio{
			AudioSource: &speechpb.RecognitionAudio_Uri{Uri: gsPath},
		},
	}

	op, err := t.speechClient.LongRunningRecognize(ctx, req)
	if err != nil {
		mrequests.UpdateTranscribeTextStatus(t.room, "error")
		t.logger.Errorf("Failed to start recognize: %v", err)
		return
	}
	resp, err := op.Wait(ctx)
	if err != nil {
		mrequests.UpdateTranscribeTextStatus(t.room, "error")
		t.logger.Errorf("Failed to start recognize: %v", err)
		return
	}

	fmt.Println("Transcription response:", resp)
	mrequests.UpdateCompanySttStatsByRoom(t.room, int32(resp.TotalBilledTime.Seconds)*1000)

	defer t.speechClient.Close()

	transcribeResult := make([]map[string]interface{}, 0)
	for _, result := range resp.Results {
		// There can be several alternative transcripts for a given chunk of speech.
		// Just use the first (most likely) one here.

		wordItems := make([]map[string]interface{}, 0)
		wordsWithTiming := result.Alternatives[0].Words
		transcript := result.Alternatives[0].Transcript
		for i := range wordsWithTiming {
			wordInfo := wordsWithTiming[i] // word with timing
			word := wordInfo.Word
			wordTimestamp := wordInfo.StartTime.Seconds*1000 + int64(wordInfo.StartTime.Nanos/1000000)

			if word == "" {
				continue
			}

			wordItems = append(wordItems, map[string]interface{}{
				"word":         word,
				"utcTimestamp": wordTimestamp,
			})
		}

		itemResult := map[string]interface{}{
			"msgID": uuid.New().String(),
			"lang":  t.lang,
			"words": wordItems,
			"msg":   transcript,
		}
		transcribeResult = append(transcribeResult, itemResult)
	}

	mrequests.UpdateTranscribeText(t.room, transcribeResult)
	fmt.Println("Transcription Results:", transcribeResult)
	// [END speech_transcribe_sync]
}
