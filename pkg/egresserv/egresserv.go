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

func StartTrackEgress(roomName string, company string) (string, error) {
	cfg := config.GetConfig()
	ctx := context.Background()
	egressClient := lksdk.NewEgressClient(
		cfg.LVHost,
		cfg.LVApiKey,
		cfg.LVApiSecret,
	)
	nowTimestamp := utils.GetNowTimestamp()
	fileName := "record/ogg/" + roomName + "/audio_" + utils.RemoveSpaces(roomName) + "_" + strings.ToLower(utils.RemoveSpaces(company)) + "_" + nowTimestamp + ".ogg"

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
						AccessKey: cfg.AWSAccessKey,
						Secret:    cfg.AWSSecret,
						Region:    cfg.AWSRegion,
						Bucket:    cfg.AWSBucket,
					},
				},
			},
		},
	}

	info, err := egressClient.StartRoomCompositeEgress(ctx, fileRequest)
	if err != nil {
		fmt.Println("Error in StartTrackEgress:", err)
		return "", err
	}
	fmt.Println("StartTrackEgress:", info.EgressId)

	return info.EgressId, nil
}

func StopTrackEgress(egressID string) error {
	cfg := config.GetConfig()
	client := lksdk.NewEgressClient(cfg.LVHost, cfg.LVApiKey, cfg.LVApiSecret)
	ctx := context.Background()
	_, err := client.StopEgress(ctx, &livekit.StopEgressRequest{
		EgressId: egressID,
	})
	if err != nil {
		return err
	}
	return nil
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
