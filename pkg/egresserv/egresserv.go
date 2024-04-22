package egresserv

import (
	"context"
	"fmt"
	"github.com/livekit/protocol/livekit"
	lksdk "github.com/livekit/server-sdk-go/v2"
	"lkrouter/config"
	"lkrouter/utils"
	"strings"
)

func StartTrackEgress(roomName string, company string) string {
	config := config.GetConfig()
	ctx := context.Background()
	egressClient := lksdk.NewEgressClient(
		config.LVHost,
		config.LVApiKey,
		config.LVApiSecret,
	)

	fileName := "audio_" + utils.RemoveSpaces(roomName) + "_" + strings.ToLower(utils.RemoveSpaces(company)) + ".ogg"

	fileRequest := &livekit.RoomCompositeEgressRequest{
		RoomName:  roomName,
		Layout:    "1x1",
		AudioOnly: true,
		Output: &livekit.RoomCompositeEgressRequest_File{
			File: &livekit.EncodedFileOutput{
				FileType: livekit.EncodedFileType_OGG,
				Filepath: fileName,
				Output: &livekit.EncodedFileOutput_S3{
					S3: &livekit.S3Upload{
						AccessKey: config.AWSAccessKey,
						Secret:    config.AWSSecret,
						Region:    config.AWSRegion,
						Bucket:    config.AWSBucket,
					},
				},
			},
		},
	}

	info, err := egressClient.StartRoomCompositeEgress(ctx, fileRequest)
	if err != nil {
		fmt.Println("Error in StartTrackEgress:", err)
	}
	fmt.Println("StartTrackEgress:", info.EgressId)

	return info.EgressId
}

func TrackEgressRequest(roomID string, trackID string, wsURL string) (*livekit.EgressInfo, error) {
	cfg := config.GetConfig()
	client := lksdk.NewEgressClient(cfg.LVHost, cfg.LVApiKey, cfg.LVApiSecret)
	ctx := context.Background()
	req := &livekit.TrackEgressRequest{
		RoomName: roomID,
		TrackId:  trackID,
		Output: &livekit.TrackEgressRequest_WebsocketUrl{
			WebsocketUrl: wsURL,
		},
	}
	info, err := client.StartTrackEgress(ctx, req)
	if err != nil {
		return nil, err
	}
	return info, nil
}
